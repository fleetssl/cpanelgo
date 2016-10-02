package whm

import (
	"errors"

	"github.com/letsencrypt-cpanel/cpanelgo"
)

type CreateUserSessionApiResponse struct {
	BaseWhmApiResponse
	Data struct {
		SecurityToken string              `json:"cp_security_token"`
		Expires       cpanelgo.MaybeInt64 `json:"expires"`
		Session       string              `json:"session"`
		Url           string              `json:"url"`
	} `json:"data"`
}

func (a WhmApi) CreateUserSession(username, service string) (CreateUserSessionApiResponse, error) {
	var out CreateUserSessionApiResponse

	err := a.WHMAPI1("create_user_session", cpanelgo.Args{
		"user":    username,
		"service": service,
	}, &out)
	if err == nil && out.Result() != 1 {
		err = errors.New(out.Metadata.Reason)
	}

	return out, err
}
