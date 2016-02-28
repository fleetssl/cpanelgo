package whm

import "github.com/letsencrypt-cpanel/cpanelgo"

type GetTweakSettingApiResponse struct {
	BaseWhmApiResponse
	Data struct {
		TweakSetting struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"tweaksetting"`
	} `json:"data"`
}

func (a WhmApi) GetTweakSetting(key, module string) (GetTweakSettingApiResponse, error) {
	var out GetTweakSettingApiResponse

	err := a.WHMAPI1("get_tweaksetting", cpanelgo.Args{
		"key":    key,
		"module": module,
	}, &out)
	if err == nil {
		err = out.Error()
	}

	return out, err
}

func (a WhmApi) SetTweakSetting(key, module, value string) (BaseWhmApiResponse, error) {
	var out BaseWhmApiResponse

	err := a.WHMAPI1("set_tweaksetting", cpanelgo.Args{
		"key":    key,
		"module": module,
		"value":  value,
	}, &out)
	if err == nil {
		err = out.Error()
	}

	return out, err
}
