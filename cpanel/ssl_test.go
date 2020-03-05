package cpanel

import (
	"testing"
	"time"

	"github.com/letsencrypt-cpanel/cpanelgo"
)

func TestWildcardDoesntClobberOtherCertificates(t *testing.T) {
	sslVhosts := InstalledHostsApiResponse{
		Data: []InstalledCertificate{
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
		},
	}

	cutoff := time.Unix(1581811200, 0).Add(time.Duration(31) * 24 * time.Hour)

	if sslVhosts.HasValidDomain("node.example.com", cutoff) {
		t.Error("node.example.com should be up for renewal, but wasn't")
	}
}
