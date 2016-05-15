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

package service

import (
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
)

// NewServer configures and returns a Server.
func NewServer() *negroni.Negroni {

	//formatter := render.New(render.Options{
//		IndentJSON: true,
//	})

	n := negroni.Classic()
	mx := mux.NewRouter()

//	initRoutes(mx, formatter)

	n.UseHandler(mx)
	return n
}

func initRoutes(mx *mux.Router, formatter *render.Render) {
//	mx.HandleFunc("/api/apps/{org}/{space}/{app}", singleAppHandler(formatter)).Methods("GET")
//	mx.HandleFunc("/api/apps", appCollectionHandler(formatter)).Methods("GET")
}
