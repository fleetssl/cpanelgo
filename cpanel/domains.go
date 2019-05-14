package cpanel

import (
	"encoding/json"

	"github.com/letsencrypt-cpanel/cpanelgo"
)

const (
	ParkedStatusNotRedirected = "not redirected"
)

type DomainsDataDomain struct {
	Domain       string `json:"domain"`
	Ip           string `json:"ip"`
	DocumentRoot string `json:"documentroot"`
	User         string `json:"user"`
	ServerAlias  string `json:"serveralias"`
	ServerName   string `json:"servername"`
}

type DomainsDataApiResponse struct {
	cpanelgo.BaseUAPIResponse
	Data struct {
		MainDomain    DomainsDataDomain   `json:"main_domain"`
		AddonDomains  []DomainsDataDomain `json:"addon_domains"`
		ParkedDomains []string            `json:"parked_domains"`
		Sub_Domains   []json.RawMessage   `json:"sub_domains"`
		Subdomains    []DomainsDataDomain `json:"-"`
	} `json:"data"`
}

func (dd DomainsDataApiResponse) DataList() []DomainsDataDomain {
	doms := append(dd.Data.AddonDomains, dd.Data.MainDomain)
	doms = append(doms, dd.Data.Subdomains...)
	return doms
}

func (r DomainsDataApiResponse) DomainList() []string {
	out := []string{}
	out = append(out, r.Data.MainDomain.Domain)
	out = append(out, r.Data.ParkedDomains...)
	for _, v := range r.Data.AddonDomains {
		out = append(out, v.Domain)
	}
	for _, v := range r.Data.Subdomains {
		out = append(out, v.Domain)
	}
	return out
}

func (c CpanelApi) DomainsData() (DomainsDataApiResponse, error) {
	var out DomainsDataApiResponse

	err := c.Gateway.UAPI("DomainInfo", "domains_data", cpanelgo.Args{
		"format": "hash",
	}, &out)
	if err == nil {
		err = out.Error()
	}
	if err == nil {
		out.Data.Subdomains = []DomainsDataDomain{}
		for _, v := range out.Data.Sub_Domains {
			dec := DomainsDataDomain{}
			if err := json.Unmarshal(v, &dec); err == nil {
				out.Data.Subdomains = append(out.Data.Subdomains, dec)
			}
		}

	}
	return out, err
}

type SingleDomainDataApiResponse struct {
	Status int `json:"status"`
	Data   struct {
		Domain       string `json:"domain"`
		DocumentRoot string `json:"documentroot"`
	} `json:"data"`
}

func (c CpanelApi) SingleDomainData(domain string) (SingleDomainDataApiResponse, error) {
	var out SingleDomainDataApiResponse

	err := c.Gateway.UAPI("DomainInfo", "single_domain_data", cpanelgo.Args{
		"domain": domain,
	}, &out)

	return out, err
}

type ParkedDomain struct {
	Domain string `json:"domain"`
	Status string `json:"status"`
	Dir    string `json:"dir"`
}

type ListParkedDomainsApiResponse struct {
	cpanelgo.BaseAPI2Response
	Data []ParkedDomain `json:"data"`
}

func (c CpanelApi) ListParkedDomains() (ListParkedDomainsApiResponse, error) {
	var out ListParkedDomainsApiResponse

	err := c.Gateway.API2("Park", "listparkeddomains", cpanelgo.Args{}, &out)

	if err == nil {
		err = out.Error()
	}

	return out, err
}

type WebVhostsListDomainsApiResponse struct {
	cpanelgo.BaseUAPIResponse
	Data []VhostEntry `json:"data"`
}

type VhostEntry struct {
	Domain          string   `json:"domain"`
	VhostName       string   `json:"vhost_name"`
	VhostIsSsl      int      `json:"vhost_is_ssl"`
	ProxySubdomains []string `json:"proxy_subdomains"`
}

// put them into map for easy access
func (vhapi WebVhostsListDomainsApiResponse) GetProxySubdomainsMap() map[string][]string {
	proxyDomainsMap := map[string][]string{}
	for _, vhd := range vhapi.Data {
		if len(vhd.ProxySubdomains) > 0 {
			proxyDomainsMap[vhd.Domain] = vhd.ProxySubdomains
		}
	}
	return proxyDomainsMap
}

func (r WebVhostsListDomainsApiResponse) GetAllProxySubdomains() []string {
	m := map[string]struct{}{}
	for _, d := range r.Data {
		for _, proxy := range d.ProxySubdomains {
			m[proxy] = struct{}{}
		}
	}
	res := []string{}
	for p := range m {
		res = append(res, p)
	}
	return res
}

func (c CpanelApi) WebVhostsListDomains() (WebVhostsListDomainsApiResponse, error) {
	var out WebVhostsListDomainsApiResponse

	err := c.Gateway.UAPI("WebVhosts", "list_domains", cpanelgo.Args{}, &out)

	if err == nil {
		err = out.Error()
	}

	return out, err
}
