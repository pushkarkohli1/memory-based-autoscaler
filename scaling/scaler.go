package scaling

import (
	"fmt"

	cfClient "github.com/cloudfoundry-community/go-cfclient"
)

//"io/ioutil"

type Scaler struct {
	GUID string
}

var gcfClientScaler *cfClient.Client
var appguid string

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
}

func (s *Scaler) StopListening() {
	fmt.Println("in the stopper method")
}
