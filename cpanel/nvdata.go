package cpanel

import (
	"encoding/json"

	"github.com/letsencrypt-cpanel/cpanelgo"
)

const NVDatastoreName = "letsencrypt-cpanel"

type NVDataDomainCerts struct {
	Domain     string   `json:"domain"`
	CertUrl    string   `json:"url"`
	DomainKey  string   `json:"key"`
	DomainCert string   `json:"cert"`
	IssuerCert string   `json:"issuer"`
	KeyId      string   `json:"key_id"`  // the cpanel key_id from install_ssl
	CertId     string   `json:"cert_id"` // the cpanel cert_id from install_ssl
	CertExpiry int64    `json:"cert_expiry"`
	AltNames   []string `json:"alt_names"`
}

type NVDataAccount struct {
	AccountKey       string                        `json:"accountkey"`
	LastRenewalCheck int64                         `json:"last_renewal_check"`
	Certs            map[string]*NVDataDomainCerts `json:"certs"`
	DisableMail      bool                          `json:"disable_mail"`
}

type NVDataGetApiResult struct {
	cpanelgo.BaseUAPIResponse
	Data []struct {
		FileName     string `json:"name"`
		FileContents string `json:"value"`
	} `json:"data"`
}

func (c LiveApi) GetNVData() (*NVDataAccount, error) {
	var out NVDataGetApiResult
	err := c.Gateway.UAPI("NVData", "get", cpanelgo.Args{
		"names": NVDatastoreName,
	}, &out)
	if err == nil {
		err = out.Error()
	}

	if err != nil {
		return nil, err
	}

	var acct NVDataAccount
	if len(out.Data) == 0 {
		return &acct, nil
	}

	if len(out.Data[0].FileContents) == 0 {
		return &acct, nil
	}

	err = json.Unmarshal([]byte(out.Data[0].FileContents), &acct)

	return &acct, err
}


func (c LiveApi) SetNVData(data *NVDataAccount) (cpanelgo.BaseAPI2Response, error) {
	var out cpanelgo.BaseAPI2Response

	buf, err := json.Marshal(data)
	if err != nil {
		return out, err
	}

	err = c.Gateway.API2("NVData", "set", cpanelgo.Args{
		"names":         NVDatastoreName,
		NVDatastoreName: string(buf),
	}, &out)

	if err == nil {
		err = out.Error()
	}

	return out, err
}
