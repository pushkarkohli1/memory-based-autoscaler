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

package model

type AppData struct {
	AppName              string `json:"appname"`
	AppAction            string `json:"action"`
	MinMemoryThreshold   int    `json:"min_memory_threshold"`
	MaxMemoryThreshold   int    `json:"max_memory_threshold"`
	MinInstanceThreshold int    `json:"min_instance_threshold"`
	MaxInstanceThreshold int    `json:"max_instance_threshold"`
	TimeBetweenScales    int    `json:"time_between_scales"`
	TimeOverThreshold    int    `json:"time_over_threshold"`
}
