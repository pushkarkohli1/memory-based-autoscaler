package model

import (
	"encoding/json"
	"os"
)

type VcapCredentials struct {
	Host     string `json:"hostname"`
	Port     int    `json:"port"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type VcapServiceEntry struct {
	Credentials    VcapCredentials `json:"credentials"`
	Label          string          `json:"label"`
	Name           string          `json:"name"`
	SyslogDrainUrl string          `json:"syslog_drain_url"`
	Tags           []string        `json:"tags"`
}

type VcapServices struct {
	UserProvided []VcapServiceEntry `json:"p-mysql"`
}

var VcapService *VcapServices

func PopulateVcapServices() (*VcapServices, error) {

	var err error

	str := os.Getenv("VCAP_SERVICES")
	if str != "" {

		//fmt.Println(str)
		VcapService = &VcapServices{}

		err = json.Unmarshal([]byte(str), VcapService)
		if err != nil {
			return nil, err
		}

	} else {
		return nil, err
	}

	return VcapService, nil
}
