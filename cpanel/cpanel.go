package cpanel

import "github.com/letsencrypt-cpanel/cpanelgo"

type CpanelApi struct {
	cpanelgo.Api
}

type CpanelApiRequest struct {
	Module      string        `json:"module"`
	RequestType string        `json:"reqtype"`
	Function    string        `json:"func"`
	ApiVersion  string        `json:"apiversion"`
	Arguments   cpanelgo.Args `json:"args"`
}
