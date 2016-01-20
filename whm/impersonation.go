package whm

import (
	"strings"

	"encoding/json"

	"github.com/letsencrypt-cpanel/cpanelgo"
	"github.com/letsencrypt-cpanel/cpanelgo/cpanel"
)

type WhmImpersonationApi struct {
	Impersonate string // who we are impersonating
	WhmApi
}

func NewWhmImpersonationApi(hostname, username, accessHash, userToImpersonate string, insecure bool) cpanel.CpanelApi {
	accessHash = strings.Replace(accessHash, "\n", "", -1)
	accessHash = strings.Replace(accessHash, "\r", "", -1)

	return cpanel.CpanelApi{cpanelgo.NewApi(
		&WhmImpersonationApi{
			Impersonate: userToImpersonate,
			WhmApi: WhmApi{
				Hostname:   hostname,
				Username:   username,
				AccessHash: accessHash,
				Insecure:   insecure,
			},
		})}
}

func (c *WhmImpersonationApi) UAPI(module, function string, arguments cpanelgo.Args, out interface{}) error {
	arguments["user"] = c.Impersonate
	arguments["cpanel_jsonapi_apiversion"] = "3"
	arguments["cpanel_jsonapi_module"] = module
	arguments["cpanel_jsonapi_func"] = function

	var result cpanelgo.UAPIResult
	err := c.WHMAPI1("cpanel", arguments, &result)
	if err == nil {
		err = result.Error()
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(result.Result, out)
}

func (c *WhmImpersonationApi) API2(module, function string, arguments cpanelgo.Args, out interface{}) error {
	arguments["user"] = c.Impersonate
	arguments["cpanel_jsonapi_apiversion"] = "2"
	arguments["cpanel_jsonapi_module"] = module
	arguments["cpanel_jsonapi_func"] = function

	var result cpanelgo.API2Result
	err := c.WHMAPI1("cpanel", arguments, &result)
	if err == nil {
		err = result.Error()
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(result.Result, out)
}

func (c *WhmImpersonationApi) API1(module, function string, arguments []string, out interface{}) error {
	args := cpanelgo.Args{}
	args["user"] = c.Impersonate
	args["cpanel_jsonapi_apiversion"] = "1"
	args["cpanel_jsonapi_module"] = module
	args["cpanel_jsonapi_func"] = function

	if arguments != nil && len(arguments) > 0 {
		for _,v := range arguments {
			args[v] = true
		}
	}

	return c.WHMAPI1("cpanel", args, out)
}

func (c *WhmImpersonationApi) Close() error {
	return nil
}
