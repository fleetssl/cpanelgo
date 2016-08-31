package cpanel

import "github.com/letsencrypt-cpanel/cpanelgo"

func (c CpanelApi) HasFeature(name string) (string, error) {
	var out cpanelgo.BaseUAPIResponse
	err := c.Gateway.UAPI("Features", "has_feature", cpanelgo.Args{
		"name": name,
	}, &out)
	if err == nil {
		err = out.Error()
	}
	// discard the error if its the 'unknown error' as its irrelevant to the result
	if err != nil && err.Error() == cpanelgo.ErrorUnknown {
		err = nil
	}
	return out.Message(), err
}
