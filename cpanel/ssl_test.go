package cpanel

import (
	"testing"
	"time"

	"github.com/letsencrypt-cpanel/cpanelgo"
)

func TestWildcardDoesntClobberOtherCertificates(t *testing.T) {
	cutoff := time.Unix(1581811200, 0).Add(time.Duration(31) * 24 * time.Hour)
	data := []InstalledCertificate{
		{
			Certificate: CpanelSslCertificate{
				IsSelfSigned: cpanelgo.MaybeInt64(0),
				CommonName:   cpanelgo.MaybeCommonNameString("home.example.com"),
				Domains:      []string{"home.example.com"},
				NotAfter:     cpanelgo.MaybeInt64(1584316800),
			},
		},
		{
			Certificate: CpanelSslCertificate{
				IsSelfSigned: cpanelgo.MaybeInt64(0),
				CommonName:   cpanelgo.MaybeCommonNameString("node.example.com"),
				Domains:      []string{"node.example.com"},
				NotAfter:     cpanelgo.MaybeInt64(1584316800),
			},
		},
		{
			Certificate: CpanelSslCertificate{
				IsSelfSigned: cpanelgo.MaybeInt64(0),
				CommonName:   cpanelgo.MaybeCommonNameString("example.com"),
				Domains:      []string{"example.com", "*.example.com"},
				NotAfter:     cpanelgo.MaybeInt64(1589673600),
			},
		},
		{
			Certificate: CpanelSslCertificate{
				IsSelfSigned: cpanelgo.MaybeInt64(0),
				CommonName:   cpanelgo.MaybeCommonNameString("fake.test.com"),
				NotAfter:     cpanelgo.MaybeInt64(1589673600),
			},
		},
	}

	tests := []struct {
		apiResp InstalledHostsApiResponse
		domain  string
		cutoff  time.Time
		result  bool
	}{
		{
			apiResp: InstalledHostsApiResponse{Data: data},
			domain:  "node.example.com",
			cutoff:  cutoff,
			result:  false,
		},
		{
			apiResp: InstalledHostsApiResponse{Data: []InstalledCertificate{data[2], data[1]}},
			domain:  "node.example.com",
			cutoff:  cutoff,
			result:  false,
		},
		{
			apiResp: InstalledHostsApiResponse{Data: data},
			domain:  "home.example.com",
			cutoff:  cutoff,
			result:  false,
		},
		{
			apiResp: InstalledHostsApiResponse{Data: data},
			domain:  "fake.test.com",
			cutoff:  cutoff,
			result:  true,
		},
		{
			apiResp: InstalledHostsApiResponse{Data: data},
			domain:  "example.com",
			cutoff:  cutoff,
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
