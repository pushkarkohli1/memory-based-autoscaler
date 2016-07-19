package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	scaler "github.com/ECSTeam/memory-based-autoscaler/scaling"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

type AppData struct {
	AppName   string `json:"appname"`
	AppAction string `json:"action"`
}

type Service struct {
	Name           string   `json:"name"`
	Id             string   `json:"id"`
	Description    string   `json:"description"`
	Bindable       bool     `json:"bindable"`
	PlanUpdateable bool     `json:"plan_updateable, omitempty"`
	Tags           []string `json:"tags, omitempty"`
	Requires       []string `json:"requires, omitempty"`

	Metadata        interface{}   `json:"metadata, omitempty"`
	Plans           []ServicePlan `json:"plans"`
	DashboardClient interface{}   `json:"dashboard_client"`
}

type ServicePlan struct {
	Name        string      `json:"name"`
	Id          string      `json:"id"`
	Description string      `json:"description"`
	Metadata    interface{} `json:"metadata, omitempty"`
	Free        bool        `json:"free, omitempty"`
}

type ServiceInstance struct {
	Id               string `json:"id"`
	DashboardUrl     string `json:"dashboard_url"`
	InternalId       string `json:"internalId, omitempty"`
	ServiceId        string `json:"service_id"`
	PlanId           string `json:"plan_id"`
	OrganizationGuid string `json:"organization_guid"`
	SpaceGuid        string `json:"space_guid"`

	LastOperation *LastOperation `json:"last_operation, omitempty"`

	Parameters interface{} `json:"parameters, omitempty"`
}

type LastOperation struct {
	State                    string `json:"state"`
	Description              string `json:"description"`
	AsyncPollIntervalSeconds int    `json:"async_poll_interval_seconds, omitempty"`
}

type CreateServiceInstanceResponse struct {
	Status string `json:"status"`
}

type DeleteServiceInstanceResponse struct {
	Status string `json:"status"`
}

type ServiceBinding struct {
	Id                string `json:"id"`
	ServiceId         string `json:"service_id"`
	AppId             string `json:"app_id"`
	ServicePlanId     string `json:"service_plan_id"`
	PrivateKey        string `json:"private_key"`
	ServiceInstanceId string `json:"service_instance_id"`
}

type CreateServiceBindingResponse struct {
	// SyslogDrainUrl string      `json:"syslog_drain_url, omitempty"`
	Credentials interface{} `json:"credentials"`
}

type DeleteServiceBindingResponse struct {
	Status string `json:"status"`
}

type Credential struct {
	SubscriptionGUID string `json:"subscription_guid"`
	BindingGUID      string `json:"binding_guid"`
}

type Catalog struct {
	Services []Service `json:"services"`
}

var ScalerMap = make(map[string]scaler.Scaler)

func catalogHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		fmt.Println("Get Service Broker Catalog...")

		var catalog Catalog
		catalogFileName := "catalog.json"

		err := ReadAndUnmarshal(&catalog, "data", catalogFileName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Access-Control-Allow-Origin", req.Header.Get("Origin"))
		w.Header().Add("Access-Control-Allow-Methods", "GET")
		formatter.JSON(w, http.StatusOK, catalog)

	}
}

func getServiceInstanceHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("Get Service Instance")

		response := CreateServiceInstanceResponse{
			Status: "ok",
		}

		w.Header().Add("Access-Control-Allow-Origin", req.Header.Get("Origin"))
		w.Header().Add("Access-Control-Allow-Methods", "GET")
		formatter.JSON(w, http.StatusOK, response)
	}
}

func createServiceInstanceHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("Create Service Instance")

		serviceInstanceGuid := ExtractVarsFromRequest(req, "service_instance_guid")

		scalerInst := scaler.Scaler{GUID: serviceInstanceGuid}

		scalerInst.Hello()
		scalerInst.Initialize(GCFClient)

		ScalerMap[serviceInstanceGuid] = scalerInst

		response := CreateServiceInstanceResponse{
			Status: "ok",
		}

		fmt.Println("Done Creating Service Instance")

		w.Header().Add("Access-Control-Allow-Origin", req.Header.Get("Origin"))
		w.Header().Add("Access-Control-Allow-Methods", "PUT")
		formatter.JSON(w, http.StatusOK, response)
	}
}

func removeServiceInstanceHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("Remove Service Instance")

		response := DeleteServiceInstanceResponse{
			Status: "deleted",
		}

		w.Header().Add("Access-Control-Allow-Origin", req.Header.Get("Origin"))
		w.Header().Add("Access-Control-Allow-Methods", "DELETE")
		formatter.JSON(w, http.StatusOK, response)
	}
}

func bindServiceInstanceHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("Bind Service Instance")

		serviceInstanceGuid := ExtractVarsFromRequest(req, "service_instance_guid")
		bindingId := ExtractVarsFromRequest(req, "service_binding_guid")

		scalerInst, ok := ScalerMap[serviceInstanceGuid]

		if ok {
			scalerInst.SetAppData(bindingId)
		}

		credential := Credential{
			SubscriptionGUID: serviceInstanceGuid,
			BindingGUID:      bindingId,
		}

		response := CreateServiceBindingResponse{
			Credentials: credential,
		}

		w.Header().Add("Access-Control-Allow-Origin", req.Header.Get("Origin"))
		w.Header().Add("Access-Control-Allow-Methods", "PUT")
		formatter.JSON(w, http.StatusOK, response)
	}
}

func unbindServiceInstanceHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("Unbind Service Instance")

		response := DeleteServiceBindingResponse{
			Status: "deleted",
		}

		w.Header().Add("Access-Control-Allow-Origin", req.Header.Get("Origin"))
		w.Header().Add("Access-Control-Allow-Methods", "DELETE")
		formatter.JSON(w, http.StatusOK, response)
	}
}

func BoundActionHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("Starting to listen...")

		serviceInstanceGuid := ExtractVarsFromRequest(req, "service_instance_guid")
		//bindingId := ExtractVarsFromRequest(req, "service_binding_guid")

		var appdata AppData

		b, _ := ioutil.ReadAll(req.Body)
		json.Unmarshal(b, &appdata)

		scalerInst, ok := ScalerMap[serviceInstanceGuid]

		if ok {
			if appdata.AppAction == "START" {
				scalerInst.StartListening(appdata.AppName)
			} else if appdata.AppAction == "STOP" {
				scalerInst.StopListening()
			} else {
				fmt.Println("action not supported")
			}
		}

	}
}

func ReadAndUnmarshal(object interface{}, dir string, fileName string) error {
	path := dir + string(os.PathSeparator) + fileName

	bytes, err := ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, object)
	if err != nil {
		return err
	}

	return nil
}

func ReadFile(path string) (content []byte, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return
	}
	content = bytes

	return
}

func ExtractVarsFromRequest(r *http.Request, varName string) string {
	return mux.Vars(r)[varName]
}
