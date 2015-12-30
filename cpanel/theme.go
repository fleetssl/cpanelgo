package cpanel

import "github.com/letsencrypt-cpanel/cpanelgo"

type GetThemeAPIResponse struct {
	cpanelgo.BaseUAPIResponse
	Theme string `json:"data"`
}

func (c LiveApi) GetTheme() (GetThemeAPIResponse, error) {
	var out GetThemeAPIResponse
	err := c.Gateway.UAPI("Themes", "get_theme_base", cpanelgo.Args{}, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}
