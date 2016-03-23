package cpanel

import "github.com/letsencrypt-cpanel/cpanelgo"

func (c CpanelApi) HasFeature(name string) (string, error) {
	var out cpanelgo.BaseUAPIResponse
	err := c.Gateway.UAPI("Features", "has_feature", cpanelgo.Args{
		"name": name,
	}, &out)
	if err == nil && out.Error().Error() != cpanelgo.ErrorUnknown {
		err = out.Error()
	}
	return out.Message(), err
}
