package cpanel

import (
	"encoding/json"

	"github.com/letsencrypt-cpanel/cpanelgo"
)

type NVDataGetApiResult struct {
	cpanelgo.BaseUAPIResponse
	Data []struct {
		FileName     string `json:"name"`
		FileContents string `json:"value"`
	} `json:"data"`
}

func (c CpanelApi) GetNVData(name string) (NVDataGetApiResult, error) {
	var out NVDataGetApiResult
	err := c.Gateway.UAPI("NVData", "get", cpanelgo.Args{
		"names": name,
	}, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}

type NVDataSetApiResult struct {
	cpanelgo.BaseUAPIResponse
	Data []struct {
		Set string `json:"set"`
	} `json:"data"`
}

func (c CpanelApi) SetNVData(name string, data interface{}) (NVDataSetApiResult, error) {
	var out NVDataSetApiResult

	buf, err := json.Marshal(data)
	if err != nil {
		return out, err
	}

	err = c.Gateway.UAPI("NVData", "set", cpanelgo.Args{
		"names": name,
		name:    string(buf),
	}, &out)

	if err == nil {
		err = out.Error()
	}

	return out, err
}
