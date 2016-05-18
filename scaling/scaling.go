/*
Copyright 2016 Pivotal

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
	"log"
        "os"
	"fmt"
	"time"

	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/Sirupsen/logrus"
        "github.com/cloudfoundry/sonde-go/events"
)

type Event struct {
	Fields logrus.Fields
	Msg    string
	Type   string
}

type MemoryDetails struct {
	Memory uint64
	LastTime int64
}

var MemoryMap = make(map[int32]MemoryDetails)

var LastScaleTime = time.Now().UnixNano()

// ProcessEvents churns through the firehose channel, processing incoming events.
func ProcessEvents(in chan *events.Envelope) {
        for msg := range in {
                processEvent(msg)
        }
}

func processEvent(msg *events.Envelope) {

	logger:= log.New(os.Stdout, "", 0)
        
        eventType := msg.GetEventType()
	if eventType.String() == "ContainerMetric" {

		var event Event

		logger.Println("Recieved message, type == " + eventType.String())

		event = ContainerMetric(msg)

		event.AnnotateWithAppData()
		if event.Fields["cf_app_name"] == "dummy" {
			logger.Println("found a dummy event!")
			var memory = event.Fields["memory_bytes"]
			var instance = event.Fields["instance_index"]
			fmt.Printf("memory == %d and instanceid == %d\n", memory, instance)
			updateMemoryMap(event)
			CheckMemoryAverage()
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

func CheckMemoryAverage() {

	var sum uint64 = 0
	count := 0	

	fmt.Printf("Memory map size == %d\n" , len(MemoryMap))

	for key, value := range MemoryMap {
		fmt.Printf("Memory Map output:  instance, bytes, lastTime = %d, %d, %d\n", key, value.Memory, value.LastTime)
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

	if count > 0 {

		average := float64(sum) / float64(count)

		fmt.Printf("Average Memory consumption for all running instances is %f\n", average)	

		if average > 220000000 {

			scaleElapsed := time.Now().UnixNano() - LastScaleTime
			scaleElapsedSeconds := scaleElapsed / 1000000000

			fmt.Printf("seconds since last scale is %d\n", scaleElapsedSeconds)

			if scaleElapsedSeconds > 120 {
				
				fmt.Printf("Here is where we'd make a call to scale up\n")
			}
		}

	}

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
