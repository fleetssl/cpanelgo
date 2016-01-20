package cpanel

import (
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
}

type DomainsDataApiResponse struct {
	cpanelgo.BaseUAPIResponse
	Data struct {
		MainDomain    DomainsDataDomain   `json:"main_domain"`
		AddonDomains  []DomainsDataDomain `json:"addon_domains"`
		ParkedDomains []string            `json:"parked_domains"`
		Subdomains    []DomainsDataDomain `json:"sub_domains"`
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
