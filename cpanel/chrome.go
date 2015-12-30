package cpanel

import (
	"github.com/letsencrypt-cpanel/cpanelgo"
)

type GetDomAPIResponse struct {
	cpanelgo.BaseUAPIResponse
	Data struct {
		Header string `json:"header"`
		Footer string `json:"footer"`
	} `json:"data"`
}

func (c LiveApi) GetDom(pageTitle string) (GetDomAPIResponse, error) {
	var out GetDomAPIResponse
	err := c.Gateway.UAPI("Chrome", "get_dom", cpanelgo.Args{
		"page_title": pageTitle,
	}, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}
