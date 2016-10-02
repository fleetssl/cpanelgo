package cpanel

import (
	"strconv"
	"strings"
	"time"

	"github.com/letsencrypt-cpanel/cpanelgo"
)

type CpanelSslCertificate struct {
	Domains      []string            `json:"domains"`
	CommonName   string              `json:"subject.commonName"`
	IsSelfSigned cpanelgo.MaybeInt64 `json:"is_self_signed"`
	Id           string              `json:"id"`
	NotAfter     cpanelgo.MaybeInt64 `json:"not_after"`
	OrgName      string              `json:"issuer.organizationName"`
}

func (s CpanelSslCertificate) Expiry() time.Time {
	return time.Unix(int64(s.NotAfter), 0)
}

type ListSSLKeysAPIResponse struct {
	cpanelgo.BaseUAPIResponse
	Data struct {
		Created       cpanelgo.MaybeInt64 `json:"created"`
		Modulus       string              `json:"modulus"`
		Id            string              `json:"id"`
		FriendlyName  string              `json:"friendly_name"`
		ModulusLength int                 `json:"modulus_length"`
	} `json:"data"`
}

func (c CpanelApi) ListSSLKeys() (ListSSLKeysAPIResponse, error) {
	var out ListSSLKeysAPIResponse
	err := c.Gateway.UAPI("SSL", "list_keys", nil, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}

type ListSSLCertsAPIResponse struct {
	cpanelgo.BaseUAPIResponse
	Data []CpanelSslCertificate `json:"data"`
}

func (c CpanelApi) ListSSLCerts() (ListSSLCertsAPIResponse, error) {
	var out ListSSLCertsAPIResponse
	err := c.Gateway.UAPI("SSL", "list_certs", nil, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}

type InstalledCertificate struct {
	Certificate     CpanelSslCertificate `json:"certificate"`
	CertificateText string               `json:"certificate_text"`
}

type InstalledHostsApiResponse struct {
	cpanelgo.BaseUAPIResponse
	Data []InstalledCertificate `json:"data"`
}

func (r InstalledHostsApiResponse) HasDomain(d string) bool {
	for _, h := range r.Data {
		if strings.ToLower(d) == strings.ToLower(h.Certificate.CommonName) {
			return true
		}
		for _, v := range h.Certificate.Domains {
			if strings.ToLower(d) == strings.ToLower(v) {
				return true
			}
		}
	}
	return false
}

func (c CpanelApi) InstalledHosts() (InstalledHostsApiResponse, error) {
	var out InstalledHostsApiResponse
	err := c.Gateway.UAPI("SSL", "installed_hosts", nil, &out)
	return out, err
}

type GenerateSSLKeyAPIResponse struct {
	cpanelgo.BaseUAPIResponse
	Data struct {
		Created       cpanelgo.MaybeInt64 `json:"created"`
		Modulus       string              `json:"modulus"`
		Text          string              `json:"text"`
		Id            string              `json:"id"`
		FriendlyName  string              `json:"friendly_name"`
		ModulusLength int                 `json:"modulus_length"`
	} `json:"data"`
}

func (c CpanelApi) GenerateSSLKey(keySize int, friendlyName string) (GenerateSSLKeyAPIResponse, error) {
	var out GenerateSSLKeyAPIResponse
	err := c.Gateway.UAPI("SSL", "generate_key", cpanelgo.Args{
		"key_size":      strconv.Itoa(keySize),
		"friendly_name": friendlyName,
	}, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}

type InstallSSLKeyAPIResponse struct {
	cpanelgo.BaseUAPIResponse
	Data struct {
		Action                  string   `json:"action"`
		CertId                  string   `json:"cert_id"`
		Domain                  string   `json:"domain"`
		Html                    string   `json:"html"`
		Ip                      string   `json:"ip"`
		KeyId                   string   `json:"key_id"`
		Message                 string   `json:"message"`
		StatusMsg               string   `json:"statusmsg"`
		User                    string   `json:"user"`
		WarningDomains          []string `json:"warning_domains"`
		WorkingDomains          []string `json:"working_domains"`
		ExtraCertificateDomains []string `json:"extra_certificate_domains"`
	} `json:"data"`
}

func (c CpanelApi) InstallSSLKey(domain string, cert string, key string, cabundle string) (InstallSSLKeyAPIResponse, error) {
	var out InstallSSLKeyAPIResponse
	err := c.Gateway.UAPI("SSL", "install_ssl", cpanelgo.Args{
		"domain":   domain,
		"cert":     cert,
		"key":      key,
		"cabundle": cabundle,
	}, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}

func (c CpanelApi) DeleteSSL(domain string) (cpanelgo.BaseUAPIResponse, error) {
	var out cpanelgo.BaseUAPIResponse
	err := c.Gateway.UAPI("SSL", "delete_ssl", cpanelgo.Args{
		"domain": domain,
	}, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}

func (c CpanelApi) DeleteCert(certId string) (cpanelgo.BaseUAPIResponse, error) {
	var out cpanelgo.BaseUAPIResponse
	err := c.Gateway.UAPI("SSL", "delete_cert", cpanelgo.Args{
		"id": certId,
	}, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}

func (c CpanelApi) DeleteKey(certId string) (cpanelgo.BaseUAPIResponse, error) {
	var out cpanelgo.BaseUAPIResponse
	err := c.Gateway.UAPI("SSL", "delete_key", cpanelgo.Args{
		"id": certId,
	}, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}

type EnableMailSNIAPIResponse struct {
	cpanelgo.BaseUAPIResponse
	Data struct {
		UpdatedDomains map[string]int         `json:"updated_domains"`
		FailedDomains  map[string]interface{} `json:"failed_domains"`
	} `json:"data"`
}

func (c CpanelApi) EnableMailSNI(domains ...string) (EnableMailSNIAPIResponse, error) {
	var out EnableMailSNIAPIResponse
	err := c.Gateway.UAPI("SSL", "enable_mail_sni", cpanelgo.Args{
		"domains": strings.Join(domains, "|"),
	}, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}

type IsMailSNISupportedAPIResponse struct {
	cpanelgo.BaseUAPIResponse
	Data int `json:"data"`
}

func (c CpanelApi) IsMailSNISupported() (IsMailSNISupportedAPIResponse, error) {
	var out IsMailSNISupportedAPIResponse
	err := c.Gateway.UAPI("SSL", "is_mail_sni_supported", cpanelgo.Args{}, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}

type MailSNIStatusAPIResponse struct {
	cpanelgo.BaseUAPIResponse
	Data struct {
		Enabled int `json:"enabled"`
	} `json:"data"`
}

func (c CpanelApi) MailSNIStatus(domain string) (MailSNIStatusAPIResponse, error) {
	var out MailSNIStatusAPIResponse
	err := c.Gateway.UAPI("SSL", "mail_sni_status", cpanelgo.Args{
		"domain": domain,
	}, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}

type RebuildMailSNIConfigAPIResponse struct {
	cpanelgo.BaseUAPIResponse
	Data struct {
		Success int `json:"success"`
	}
}

func (c CpanelApi) RebuildMailSNIConfig() (RebuildMailSNIConfigAPIResponse, error) {
	var out RebuildMailSNIConfigAPIResponse
	err := c.Gateway.UAPI("SSL", "rebuild_mail_sni_config", cpanelgo.Args{
		"reload_dovecot": 1,
	}, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}
