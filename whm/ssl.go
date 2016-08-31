package whm

import (
	"github.com/letsencrypt-cpanel/cpanelgo"
	"github.com/letsencrypt-cpanel/cpanelgo/cpanel"
)

type VhostEntry struct {
	User        string                      `json:"user"`
	Docroot     string                      `json:"docroot"`
	Certificate cpanel.CpanelSslCertificate `json:"crt"`
}

type FetchSslVhostsApiResponse struct {
	BaseWhmApiResponse
	Data struct {
		Vhosts []VhostEntry `json:"vhosts"`
	} `json:"data"`
}

func (a WhmApi) FetchSslVhosts() (FetchSslVhostsApiResponse, error) {
	var out FetchSslVhostsApiResponse

	err := a.WHMAPI1("fetch_ssl_vhosts", cpanelgo.Args{}, &out)
	if err == nil {
		err = out.Error()
	}

	return out, err
}
