/*
Copyright 2016 ECSTeam

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package scaling

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/cloudfoundry-community/firehose-to-syslog/firehose"
	cfClient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/cloudfoundry/sonde-go/events"

	model "github.com/ECSTeam/memory-based-autoscaler/model"
)

//"io/ioutil"

type Scaler struct {
	GUID string
}

type Event struct {
	Fields logrus.Fields
	Msg    string
	Type   string
}

type MemoryDetails struct {
	Memory   uint64
	LastTime int64
}

var MemoryMap = make(map[int32]MemoryDetails)

//var memoryThresholdLimit, _ = strconv.Atoi(os.Getenv("MEMORY_THRESHOLD_LIMIT"))
//var timeBetweenScales, _ = strconv.Atoi(os.Getenv("TIME_BETWEEN_SCALES"))
//var timeOverThreshold, _ = strconv.Atoi(os.Getenv("TIME_OVER_THRESHOLD"))

var LastScaleTime = time.Now().UnixNano()
var TimeFirstOverThreshold int64 = 1

var gcfClientScaler *cfClient.Client
var appGuid string
var bindingGuid string

var appData model.AppData

func (s *Scaler) Hello() string {
	fmt.Println("Hello world ")
	fmt.Println(s.GUID)
	return "hello"
}

func (s *Scaler) GetCfToken() string {
	token := gcfClientScaler.GetToken()
	return token
}

func (s *Scaler) GetCfClient() *cfClient.Client {
	return gcfClientScaler
}


func (s *Scaler) Initialize(cfClient *cfClient.Client) {
	gcfClientScaler = cfClient

	appData.MaxInstanceThreshold = 10
	appData.MinInstanceThreshold = 2
	appData.MaxMemoryThreshold = 100
	appData.MinMemoryThreshold = 30
	appData.TimeBetweenScales = 75
	appData.TimeOverThreshold = 55
}

func (s *Scaler) SetAppIds(bindingguid string, appguid string) {
	appGuid = appguid
	bindingGuid = bindingguid
	fmt.Println("Setting App Data")
}

func (s *Scaler) SetAppData(inAppData model.AppData) {
	appData.MaxInstanceThreshold = inAppData.MaxInstanceThreshold
	appData.MinInstanceThreshold = inAppData.MinInstanceThreshold
	appData.MaxMemoryThreshold = inAppData.MaxMemoryThreshold
	appData.MinMemoryThreshold = inAppData.MinMemoryThreshold
	appData.TimeBetweenScales = inAppData.TimeBetweenScales
	appData.TimeOverThreshold = inAppData.TimeOverThreshold
}

func (s *Scaler) GetAppData() model.AppData {
	return appData
}

func (s *Scaler) StartListening(appName string) {
	fmt.Println("in the starter method")
	firehose := firehose.CreateFirehoseChan(gcfClientScaler.Endpoint.DopplerEndpoint, gcfClientScaler.GetToken(), bindingGuid, true)
	if firehose != nil {
		fmt.Println("Firehose Subscription Succesful! Routing events...")

		go ProcessEvents(firehose)
	} else {
		fmt.Println("Failed connecting to Firehose...Please check settings and try again!")
	}
}

func (s *Scaler) StopListening() {
	fmt.Println("in the stopper method")
}

func ProcessEvents(in chan *events.Envelope) {

	for msg := range in {
		processEvent(msg)
	}
}

func processEvent(msg *events.Envelope) {

	eventType := msg.GetEventType()

	// only container events will contain the memory statistics

	if eventType.String() == "ContainerMetric" {

		var event Event

		event = ContainerMetric(msg)

		event.AnnotateWithAppData()

		// once the event has been annotated with application data, lets see if
		// its for the app we care about

		if event.Fields["cf_app_id"] == appGuid {
			updateMemoryMap(event)
			CheckMemoryAverage(event)
		}
	}
}

func ContainerMetric(msg *events.Envelope) Event {
	containerMetric := msg.GetContainerMetric()

	fields := logrus.Fields{
		"origin":         msg.GetOrigin(),
		"cf_app_id":      containerMetric.GetApplicationId(),
		"cpu_percentage": containerMetric.GetCpuPercentage(),
		"disk_bytes":     containerMetric.GetDiskBytes(),
		"instance_index": containerMetric.GetInstanceIndex(),
		"memory_bytes":   containerMetric.GetMemoryBytes(),
	}

	return Event{
		Fields: fields,
		Msg:    "",
		Type:   msg.GetEventType().String(),
	}
}

// annotates an event with the application data (org, space, app names/ids)

func (e *Event) AnnotateWithAppData() {

	cf_app_id := e.Fields["cf_app_id"]
	appGuid := ""
	if cf_app_id != nil {
		appGuid = fmt.Sprintf("%s", cf_app_id)
	}

	if cf_app_id != nil && appGuid != "<nil>" && cf_app_id != "" {
		appInfo := getAppInfo(appGuid)
		cf_app_name := appInfo.Name
		cf_space_id := appInfo.SpaceGuid
		cf_space_name := appInfo.SpaceName
		cf_org_id := appInfo.OrgGuid
		cf_org_name := appInfo.OrgName

		if cf_app_name != "" {
			e.Fields["cf_app_name"] = cf_app_name
		}

		if cf_space_id != "" {
			e.Fields["cf_space_id"] = cf_space_id
		}

		if cf_space_name != "" {
			e.Fields["cf_space_name"] = cf_space_name
		}

		if cf_org_id != "" {
			e.Fields["cf_org_id"] = cf_org_id
		}

		if cf_org_name != "" {
			e.Fields["cf_org_name"] = cf_org_name
		}
	}
}

func getAppInfo(appGuid string) caching.App {
	if app := caching.GetAppInfo(appGuid); app.Name != "" {
		return app
	} else {
		caching.GetAppByGuid(appGuid)
	}
	return caching.GetAppInfo(appGuid)
}

func (e Event) ShipEvent() {

	logrus.WithFields(e.Fields).Info(e.Msg)
}

func CheckMemoryAverage(ctrEvent Event) {

	var sum uint64 = 0
	count := 0

	// loop over every event in the map

	for key, value := range MemoryMap {
		meminMb := value.Memory / 1000000
		fmt.Printf("Memory Map output:  instance, bytes, lastTime = %d, %d, %d\n", key, meminMb, value.LastTime)
		totalElapsed := time.Now().UnixNano() - value.LastTime
		elapsedSeconds := totalElapsed / 1000000000
		// elapsedSeconds shows the last time the map was updated with a container metric.
		// if that's more than ten minutes old, we assume the app instance has gone away
		// and shouldn't be accounted for in the average calculations
		if elapsedSeconds < 600 {
			sum += value.Memory
			count += 1
		}
	}

	if (count < appData.MinInstanceThreshold) && (count > 0) {
		fmt.Printf("**************** Under minimum instances, scaling to %d **************************\n", appData.MinInstanceThreshold)
		scaleApp(appData.MinInstanceThreshold-1, ctrEvent)
		LastScaleTime = time.Now().UnixNano()
	} else if count > 0 {

		average := float64(sum) / float64(count)
		averageInMb := average / 1000000

		fmt.Printf("Average Memory consumption for all running instances is %f\n", averageInMb)

		// see if that average is more than our threshold

		if int(averageInMb) > appData.MaxMemoryThreshold {

			// check to see if this is the first crossing of the memory threshold by
			// checking that TimeFirstOverThreshold value.  1 is a magic number to
			// note that it hasn't been crossed yet

			if TimeFirstOverThreshold == 1 {
				TimeFirstOverThreshold = time.Now().UnixNano()
				fmt.Printf("****************   First crossing of memory threshold! *********************\n")
			} else {

				// we've been over that threshold for at least a few seconds, lets find
				// out how long we've been over the memory threshold

				thresholdElapsed := time.Now().UnixNano() - TimeFirstOverThreshold
				thresholdElapsedSeconds := thresholdElapsed / 1000000000

				fmt.Printf("seconds since first threshold crossing is %d\n", thresholdElapsedSeconds)

				if thresholdElapsedSeconds > int64(appData.TimeOverThreshold) {

					// we've been over the memory threshold for quite a while.  Let's
					// see how long its been since we've scaled
					fmt.Printf("**************** Been over memory Threshold for too long!  *********************\n")

					scaleElapsed := time.Now().UnixNano() - LastScaleTime
					scaleElapsedSeconds := scaleElapsed / 1000000000

					fmt.Printf("seconds since last scale is %d\n", scaleElapsedSeconds)

					if scaleElapsedSeconds > int64(appData.TimeBetweenScales) {
						fmt.Printf("**************** Need to scale!  *********************\n")

						// we've been over the threshold for a while and haven't scaled
						// for a while.  time to scale it up.
						scaleApp(count, ctrEvent)
						LastScaleTime = time.Now().UnixNano()

					} else {
						fmt.Printf("**************** Need to scale but already did recently, waiting a bit...  *********************\n")
					}
				}
			}

		} else {

			// average memory usage is not over the threshold, so reset the first
			// time over threshold variable back to the magic number 1

			TimeFirstOverThreshold = 1
		}

	}

}

func scaleApp(aiCount int, ctrEvent Event) {
	token := gcfClientScaler.GetToken()
	//fmt.Printf("Token: %s\n", token)

	cf_app_id := ctrEvent.Fields["cf_app_id"]
	appGuid := ""
	if cf_app_id != nil {
		appGuid = fmt.Sprintf("%s", cf_app_id)
	}

	apiEndpoint := os.Getenv("CF_TARGET")

	url := fmt.Sprintf("%s/v2/apps/%s", apiEndpoint, appGuid)
	//fmt.Println("URL:", url)

	scaleCount := aiCount + 1

	fmt.Printf("**************** Scaling from %d instance(s) to %d instances **************************\n", aiCount, scaleCount)

	var jsonStr = []byte(fmt.Sprintf(`{"instances":%d}`, scaleCount))
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", token)
	//req.Header.Set("Host", "bosh-lite.com")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", "")

	// need to figure out a better way to skip the ssl validation
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	//fmt.Println("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)
	//body, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println("response Body:", string(body))

}

func updateMemoryMap(ctrEvent Event) {

	memory := ctrEvent.Fields["memory_bytes"].(uint64)
	instance := ctrEvent.Fields["instance_index"].(int32)
	lastTime := time.Now().UnixNano()

	memDetails := MemoryMap[instance]

	memDetails.Memory = memory
	memDetails.LastTime = lastTime

	MemoryMap[instance] = memDetails
}
