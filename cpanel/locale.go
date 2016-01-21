package cpanel

import "github.com/letsencrypt-cpanel/cpanelgo"

type LocaleAPIResponse struct {
	cpanelgo.BaseUAPIResponse
	Data struct {
		Direction string `json:"direction"`
		Name      string `json:"name"`
		Locale    string `json:"locale"`
		Encoding  string `json:"encoding"`
	} `json:"data"`
}

func (c CpanelApi) GetLocaleAttributes() (LocaleAPIResponse, error) {
	var out LocaleAPIResponse
	err := c.Gateway.UAPI("Locale", "get_attributes", cpanelgo.Args{}, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}
