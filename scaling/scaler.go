package scaling

import (
	"fmt"

	"github.com/cloudfoundry-community/firehose-to-syslog/firehose"
	cfClient "github.com/cloudfoundry-community/go-cfclient"
)

//"io/ioutil"

type Scaler struct {
	GUID string
}

var gcfClientScaler *cfClient.Client
var appguid string

//var skipSSLValidation = kingpin.Flag("skip-ssl-validation", "Please don't").Default("false").OverrideDefaultFromEnvar("SKIP_SSL_VALIDATION").Bool()

func (s *Scaler) Hello() string {
	fmt.Println("Hello world ")
	fmt.Println(s.GUID)
	return "hello"
}

func (s *Scaler) Initialize(cfClient *cfClient.Client) {
	gcfClientScaler = cfClient
}

func (s *Scaler) SetAppData(guid string) {
	appguid = guid
	fmt.Println("Setting App Data")
	fmt.Println(appguid)
}

func (s *Scaler) StartListening(appName string) {
	fmt.Println("in the starter method")
	firehose := firehose.CreateFirehoseChan(gcfClientScaler.Endpoint.DopplerEndpoint, gcfClientScaler.GetToken(), appguid, true)
	if firehose != nil {
		fmt.Println("Firehose Subscription Succesful! Routing events...")

		//scaling.ProcessEvents(firehose)
	} else {
		fmt.Println("Failed connecting to Firehose...Please check settings and try again!")
	}
}

func (s *Scaler) StopListening() {
	fmt.Println("in the stopper method")
}
