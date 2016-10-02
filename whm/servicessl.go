package whm

import "github.com/letsencrypt-cpanel/cpanelgo"

func (a WhmApi) InstallServiceSslCertificate(service, crt, key, cabundle string) (BaseWhmApiResponse, error) {
	var out BaseWhmApiResponse

	err := a.WHMAPI1("install_service_ssl_certificate", cpanelgo.Args{
		"service":  service,
		"crt":      crt,
		"key":      key,
		"cabundle": cabundle,
	}, &out)
	if err == nil {
		err = out.Error()
	}

	return out, err
}

type FetchServiceSslComponentsAPIResponse struct {
	BaseWhmApiResponse
	Data struct {
		Services []struct {
			Service         string `json:"service"`
			Certificate     string `json:"certificate"`
			CABundle        string `json:"cabundle"`
			CertificateInfo struct {
				IsSelfSigned int                 `json:"is_self_signed"`
				NotAfter     cpanelgo.MaybeInt64 `json:"not_after"`
				Domains      []string            `json:"domains"`
			} `json:"certificate_info"`
		} `json:"services"`
	} `json:"data"`
}

func (r FetchServiceSslComponentsAPIResponse) Services() []string {
	out := []string{}
	for _, v := range r.Data.Services {
		out = append(out, v.Service)
	}
	return out
}

func (a WhmApi) FetchServiceSslComponents() (FetchServiceSslComponentsAPIResponse, error) {
	var out FetchServiceSslComponentsAPIResponse

	err := a.WHMAPI1("fetch_service_ssl_components", cpanelgo.Args{}, &out)
	if err == nil {
		err = out.Error()
	}

	return out, err
}

func (a WhmApi) RestartService(name string) (BaseWhmApiResponse, error) {
	var out BaseWhmApiResponse

	err := a.WHMAPI1("restartservice", cpanelgo.Args{
		"service": name,
	}, &out)
	if err == nil {
		err = out.Error()
	}

	return out, err
}
