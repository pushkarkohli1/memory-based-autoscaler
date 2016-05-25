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

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/CrowdSurge/banner"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/boltdb/bolt"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/cloudfoundry-community/firehose-to-syslog/firehose"
	"github.com/cloudfoundry-community/go-cfclient"

        "github.com/ECSTeam/memory-based-autoscaler/scaling"
        "github.com/ECSTeam/memory-based-autoscaler/service"
)

var (
	debug             = kingpin.Flag("debug", "Enable debug mode. This disables forwarding to syslog").Default("false").OverrideDefaultFromEnvar("DEBUG").Bool()
	apiEndpoint       = kingpin.Flag("api-endpoint", "Api endpoint address. For bosh-lite installation of CF: https://api.10.244.0.34.xip.io").OverrideDefaultFromEnvar("API_ENDPOINT").Required().String()
	dopplerEndpoint   = kingpin.Flag("doppler-endpoint", "Overwrite default doppler endpoint return by /v2/info").OverrideDefaultFromEnvar("DOPPLER_ENDPOINT").String()
	subscriptionID    = kingpin.Flag("subscription-id", "Id for the subscription.").Default("firehose").OverrideDefaultFromEnvar("FIREHOSE_SUBSCRIPTION_ID").String()
	user              = kingpin.Flag("user", "Admin user.").Default("admin").OverrideDefaultFromEnvar("FIREHOSE_USER").String()
	password          = kingpin.Flag("password", "Admin password.").Default("admin").OverrideDefaultFromEnvar("FIREHOSE_PASSWORD").String()
	skipSSLValidation = kingpin.Flag("skip-ssl-validation", "Please don't").Default("false").OverrideDefaultFromEnvar("SKIP_SSL_VALIDATION").Bool()
	boltDatabasePath  = kingpin.Flag("boltdb-path", "Bolt Database path ").Default("my.db").OverrideDefaultFromEnvar("BOLTDB_PATH").String()
	tickerTime        = kingpin.Flag("cc-pull-time", "CloudController Polling time in sec").Default("60s").OverrideDefaultFromEnvar("CF_PULL_TIME").Duration()
)

const (
	version = "0.0.1"
)

func main() {

	banner.Print("memory auto scaler")
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "3000"
	}

	// Start web server
	go func() {
		server := service.NewServer()
		server.Run(":" + port)
	}()

	kingpin.Version(version)
	kingpin.Parse()

	logger := log.New(os.Stdout, "", 0)

	logger.Println(fmt.Sprintf("Starting app-usage-nozzle %s ", version))

	c := cfclient.Config{
		ApiAddress:        *apiEndpoint,
		Username:          *user,
		Password:          *password,
		SkipSslValidation: *skipSSLValidation,
	}
	cfClient := cfclient.NewClient(&c)

	if len(*dopplerEndpoint) > 0 {
		cfClient.Endpoint.DopplerEndpoint = *dopplerEndpoint
	}
	logger.Println(fmt.Sprintf("Using %s as doppler endpoint", cfClient.Endpoint.DopplerEndpoint))

	//Use bolt for in-memory  - file caching
	db, err := bolt.Open(*boltDatabasePath, 0600, &bolt.Options{Timeout: 3 * time.Second})
	if err != nil {
		logger.Fatal("Error opening bolt db: ", err)
		os.Exit(1)

	}
	defer db.Close()

	caching.SetCfClient(cfClient)
	caching.SetAppDb(db)
	caching.CreateBucket()

	// Ticker Polling the CC every X sec
	//ccPolling := time.NewTicker(*tickerTime)

	//go func() {
//		for range ccPolling.C {
//			logger.Println("Re-loading application cache.")
//			apps = caching.GetAllApp()
//		}
//	}()

	firehose := firehose.CreateFirehoseChan(cfClient.Endpoint.DopplerEndpoint, cfClient.GetToken(), *subscriptionID, *skipSSLValidation)
	if firehose != nil {
		logger.Println("Firehose Subscription Succesfull! Routing events...")
		//usageevents.ProcessEvents(firehose)
		scaling.SetCfClient(cfClient)
    scaling.ProcessEvents(firehose)
	} else {
		logger.Fatal("Failed connecting to Firehose...Please check settings and try again!")
	}
}
