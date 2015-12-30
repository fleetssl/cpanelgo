package cpanel

import (
	"fmt"

	"github.com/letsencrypt-cpanel/cpanelgo"
)

// This is fucking undocumented
func (c LiveApi) BrandingInclude(name string) (cpanelgo.BaseAPI1Response, error) {
	var out cpanelgo.BaseAPI1Response
	err := c.Gateway.API1("Branding", "include", []string{name}, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}

func (c LiveApi) SetVar(key, value string) (cpanelgo.BaseAPI1Response, error) {
	var out cpanelgo.BaseAPI1Response
	err := c.Gateway.API1("setvar", "", []string{fmt.Sprintf("%s=%s", key, value)}, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}
