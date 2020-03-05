package cpanel

import (
	"strconv"
	"strings"
	"time"

	"errors"
	"fmt"

	"crypto/x509"
	"encoding/pem"

	"github.com/letsencrypt-cpanel/cpanelgo"
)

type CpanelSslCertificate struct {
	Domains            []string                       `json:"domains"`
	CommonName         cpanelgo.MaybeCommonNameString `json:"subject.commonName"`
	IsSelfSigned       cpanelgo.MaybeInt64            `json:"is_self_signed"`
	Id                 string                         `json:"id"`
	NotAfter           cpanelgo.MaybeInt64            `json:"not_after"`
	OrgName            string                         `json:"issuer.organizationName"`
	DomainIsConfigured cpanelgo.MaybeInt64            `json:"domain_is_configured"` // Doesn't actually work
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
	FQDNs           []string             `json:"fqdns"`
}

type InstalledHostsApiResponse struct {
	cpanelgo.BaseUAPIResponse
	Data []InstalledCertificate `json:"data"`
}

func (r InstalledHostsApiResponse) HasDomain(d string) bool {
	for _, h := range r.Data {
		if strings.ToLower(d) == strings.ToLower(string(h.Certificate.CommonName)) {
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

func (r InstalledHostsApiResponse) GetCertificateForDomain(d string) (CpanelSslCertificate, bool) {
	for _, h := range r.Data {
		if strings.ToLower(d) == strings.ToLower(string(h.Certificate.CommonName)) {
			return h.Certificate, true
		}
		for _, v := range h.Certificate.Domains {
			if strings.ToLower(d) == strings.ToLower(v) {
				return h.Certificate, true
			}
		}
	}
	return CpanelSslCertificate{}, false
}

func (r InstalledHostsApiResponse) HasValidDomain(wanted string, expiryCutoff time.Time) bool {
	wanted = strings.ToLower(wanted)
	splitWanted := strings.Split(wanted, ".")

	isDomainCoveredByName := func(name string) bool {
		name = strings.ToLower(name)

		if wanted == name {
			return true
		}

		// Only other way this name can cover the wanted name is if the name is a wildcard
		if !strings.HasPrefix(name, "*.") {
			return false
		}

		// Strip off the first identifier
		splitName := strings.Split(name, ".")

		if len(splitWanted) < 2 || len(splitName) < 2 {
			return false
		}

		// Compare the name without the first identifier (because we know one of them is * already)
		return strings.Join(splitWanted[1:], ".") == strings.Join(splitName[1:], ".")
	}

	hasExactDomain := false

	for _, h := range r.Data {
		if wanted == strings.ToLower(string(h.Certificate.CommonName)) {
			hasExactDomain = true
		}
		// Ignore self-signed and 'expiring'/expired certificates
		if h.Certificate.IsSelfSigned == 1 || h.Certificate.Expiry().Before(expiryCutoff) {
			continue
		}
		if isDomainCoveredByName(string(h.Certificate.CommonName)) {
			return !hasExactDomain
		}
		for _, v := range h.Certificate.Domains {
			if wanted == strings.ToLower(v) {
				hasExactDomain = true
			}
			if isDomainCoveredByName(v) {
				return !hasExactDomain
			}
		}
	}
	return false
}

// In AutoSSL we want to avoid issuing certificates into virtual hosts that already have
// a valid certificate installed, whether or not that certificate actually covers `domain`
func (r InstalledHostsApiResponse) DoesAnyValidCertificateOverlapVhostsWith(domain string, expiryCutoff time.Time) bool {
	domain = strings.ToLower(domain)
	split := strings.Split(domain, ".")

	isDomainCoveredByName := func(name string) bool {
		name = strings.ToLower(name)
		if domain == name {
			return true
		}
		if !strings.HasPrefix(name, "*.") {
			return false
		}
		splitName := strings.Split(name, ".")
		if len(split) < 2 || len(splitName) < 2 {
			return false
		}
		return strings.Join(split[1:], ".") == strings.Join(splitName[1:], ".")
	}

	for _, h := range r.Data {
		// Intentionally not paying attention to the validity
		if h.Certificate.IsSelfSigned == 1 || h.Certificate.Expiry().Before(expiryCutoff) {
			continue
		}
		for _, fqdn := range h.FQDNs {
			if isDomainCoveredByName(fqdn) {
				return true
			}
		}
	}

	return false
}

func (c CpanelApi) InstalledHosts() (InstalledHostsApiResponse, error) {
	var out InstalledHostsApiResponse

	if err := c.Gateway.UAPI("SSL", "installed_hosts", nil, &out); err != nil {
		return out, err
	}

	// If there is a non-transport error/warning and we didn't find any SSL
	// virtual hosts, report this as a fatal error.
	if err := out.Error(); err != nil && len(out.Data) == 0 {
		return out, err
	}

	return out, nil
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

	// err = errors.New("FAKE unknown error")
	// out.Data.CertId = ""
	// out.Data.KeyId = ""

	// unknown error ocsp failing condition
	// certificate is installed but no certid/keyid returned
	// attempt to find the certid for installed status
	// TODO: remove this prior to pushing to github
	if err != nil && strings.Contains(err.Error(), "unknown error") {
		// if the api actually returned the cert id proper, we can just ignore the error and continue
		if out.Data.CertId != "" {
			err = nil
			out.Data.Message = fmt.Sprintf("The SSL certificate is now installed onto the domain “%s”", domain)
			goto DORETURN
		}
		// otherwise try to find the installed certid of the given cert
		installedCertId, findCertErr := c.findExistingCertificate(cert)
		if findCertErr != nil {
			err = fmt.Errorf("Error checking installed ssl certificate: %v", findCertErr)
			goto DORETURN
		}
		if installedCertId == "" {
			err = errors.New("Unable to find installed certificate")
			goto DORETURN
		}
		out.Data.CertId = installedCertId
		out.Data.Message = fmt.Sprintf("The SSL certificate is now installed onto the domain “%s”", domain)
		err = nil
	}

DORETURN:
	return out, err
}

// TODO: remove this prior to pushing to github
func decodeToCert(s string) (*x509.Certificate, error) {
	b, _ := pem.Decode([]byte(s))
	if b == nil {
		return nil, errors.New("Unable to decode pem")
	}

	cert, err := x509.ParseCertificate(b.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// TODO: remove this prior to pushing to github
func (c CpanelApi) findExistingCertificate(certPem string) (string, error) {

	hosts, err := c.InstalledHosts()
	if err != nil {
		return "", err
	}

	cert, err := decodeToCert(certPem)
	if err != nil {
		return "", err
	}

	for _, h := range hosts.Data {
		c, err := decodeToCert(h.CertificateText)
		if err == nil {
			if cert.SerialNumber.Cmp(c.SerialNumber) == 0 {
				return h.Certificate.Id, nil
			}
		}
	}

	return "", errors.New("Unable to find installed certificate for domain")
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

func (c CpanelApi) DeleteKey(keyId string) (cpanelgo.BaseUAPIResponse, error) {
	var out cpanelgo.BaseUAPIResponse
	err := c.Gateway.UAPI("SSL", "delete_key", cpanelgo.Args{
		"id": keyId,
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
