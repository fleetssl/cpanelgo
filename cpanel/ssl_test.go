package cpanel

import (
	"testing"
	"time"

	"github.com/letsencrypt-cpanel/cpanelgo"
)

const (
	d_Feb_16_2020   = 1581811200
	d_March_16_2020 = 1584316800
	d_May_17_2020   = 1589673600
)

func installedCert(cn string, sans []string, notAfter int64) InstalledCertificate {
	return InstalledCertificate{
		Certificate: CpanelSslCertificate{
			IsSelfSigned: cpanelgo.MaybeInt64(0),
			CommonName:   cpanelgo.MaybeCommonNameString(cn),
			Domains:      sans,
			NotAfter:     cpanelgo.MaybeInt64(notAfter),
		},
	}
}

func TestWildcardDoesntClobberOtherCertificates(t *testing.T) {
	cutoff := time.Unix(d_Feb_16_2020, 0).Add(time.Duration(31) * 24 * time.Hour)
	data := []InstalledCertificate{
		installedCert("home.example.com", []string{"home.example.com"}, d_March_16_2020),
		installedCert("node.example.com", []string{"node.example.com"}, d_March_16_2020),
		installedCert("example.com", []string{"example.com", "*.example.com"}, d_May_17_2020),
		installedCert("fake.test.com", nil, d_May_17_2020),
		installedCert("*.test.com", nil, d_March_16_2020),
	}

	tests := []struct {
		apiResp InstalledHostsApiResponse
		domain  string
		cutoff  time.Time
		result  bool
	}{
		// node.example.com is expiring
		{
			apiResp: InstalledHostsApiResponse{Data: data},
			domain:  "node.example.com",
			cutoff:  cutoff,
			result:  false,
		},
		// node.example.com is expiring
		{
			apiResp: InstalledHostsApiResponse{Data: []InstalledCertificate{data[2], data[1]}},
			domain:  "node.example.com",
			cutoff:  cutoff,
			result:  false,
		},
		// home.example.com is expiring
		{
			apiResp: InstalledHostsApiResponse{Data: data},
			domain:  "home.example.com",
			cutoff:  cutoff,
			result:  false,
		},
		// fake.test.com exists and is not expiring
		{
			apiResp: InstalledHostsApiResponse{Data: data},
			domain:  "fake.test.com",
			cutoff:  cutoff,
			result:  true,
		},
		// example.com exists and is not expiring
		{
			apiResp: InstalledHostsApiResponse{Data: data},
			domain:  "example.com",
			cutoff:  cutoff,
			result:  true,
		},
		// node.example.com exists and is not expiring @ cutoff of d_March_16_2020
		{
			apiResp: InstalledHostsApiResponse{Data: data},
			domain:  "node.example.com",
			cutoff:  time.Unix(d_March_16_2020, 0),
			result:  true,
		},
		// *.test.com exists and is expiring
		{
			apiResp: InstalledHostsApiResponse{Data: data},
			domain:  "*.test.com",
			cutoff:  cutoff,
			result:  false,
		},
		// *.test.com exists and is not expiring
		{
			apiResp: InstalledHostsApiResponse{Data: data},
			domain:  "*.test.com",
			cutoff:  time.Unix(d_March_16_2020, 0),
			result:  true,
		},
	}

	for _, test := range tests {
		valid := test.apiResp.HasValidDomain(test.domain, test.cutoff)
		if valid != test.result {
			t.Fatalf("domain %q HasValidDomain expected: %t, got: %t", test.domain, test.result, valid)
		}
	}
}
