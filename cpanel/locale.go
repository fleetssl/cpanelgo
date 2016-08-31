package cpanel

import "github.com/letsencrypt-cpanel/cpanelgo"

type LocaleAPIResponse_UAPI struct {
	cpanelgo.BaseUAPIResponse
	Data struct {
		Direction string `json:"direction"`
		Name      string `json:"name"`
		Locale    string `json:"locale"`
		Encoding  string `json:"encoding"`
	} `json:"data"`
}

func (c CpanelApi) GetLocaleAttributes() (LocaleAPIResponse_UAPI, error) {
	var out LocaleAPIResponse_UAPI
	err := c.Gateway.UAPI("Locale", "get_attributes", cpanelgo.Args{}, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}

type LocaleAPIResponse_API2 struct {
	cpanelgo.BaseAPI2Response
	Data []struct {
		Locale string `json:"locale"`
	} `json:"data"`
}

func (c CpanelApi) GetUserLocale() (LocaleAPIResponse_API2, error) {
	var out LocaleAPIResponse_API2
	err := c.Gateway.API2("Locale", "get_user_locale", cpanelgo.Args{}, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}
