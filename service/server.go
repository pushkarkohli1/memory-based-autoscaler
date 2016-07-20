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

package service

import (
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"

	cfClient "github.com/cloudfoundry-community/go-cfclient"
)

var GCFClient *cfClient.Client

// NewServer configures and returns a Server.
func NewServer(cfClient *cfClient.Client) *negroni.Negroni {

	GCFClient = cfClient

	formatter := render.New(render.Options{
		IndentJSON: true,
	})

	n := negroni.Classic()
	mx := mux.NewRouter()

	initRoutes(mx, formatter)

	n.UseHandler(mx)
	return n
}

func initRoutes(mx *mux.Router, formatter *render.Render) {
	//mx.HandleFunc("/api/apps/{org}/{space}/{app}", singleAppHandler(formatter)).Methods("GET")
	mx.HandleFunc("/v2/catalog", catalogHandler(formatter)).Methods("GET")

	mx.HandleFunc("/v2/service_instances/{service_instance_guid}", getServiceInstanceHandler(formatter)).Methods("GET")
	mx.HandleFunc("/v2/service_instances/{service_instance_guid}", createServiceInstanceHandler(formatter)).Methods("PUT")
	mx.HandleFunc("/v2/service_instances/{service_instance_guid}", removeServiceInstanceHandler(formatter)).Methods("DELETE")
	mx.HandleFunc("/v2/service_instances/{service_instance_guid}/service_bindings/{service_binding_guid}", bindServiceInstanceHandler(formatter)).Methods("PUT")
	mx.HandleFunc("/v2/service_instances/{service_instance_guid}/service_bindings/{service_binding_guid}", unbindServiceInstanceHandler(formatter)).Methods("DELETE")
	mx.HandleFunc("/v2/service_instances/{service_instance_guid}/service_bindings/{service_binding_guid}", BoundActionHandler(formatter)).Methods("POST")
	mx.HandleFunc("/v2/service_instances/{service_instance_guid}/service_bindings/{service_binding_guid}", GetAppDataHandler(formatter)).Methods("GET")

}
