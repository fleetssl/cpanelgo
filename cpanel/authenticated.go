package cpanel

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/letsencrypt-cpanel/cpanelgo"
)

type JsonApiGateway struct {
	Hostname string
	Username string
	Password string
	Insecure bool
	cl       *http.Client
}

func NewJsonApi(hostname, username, password string, insecure bool) (CpanelApi, error) {
	c := &JsonApiGateway{
		Hostname: hostname,
		Username: username,
		Password: password,
		Insecure: insecure,
	}
	// todo: a way to check the username/password here
	return CpanelApi{cpanelgo.NewApi(c)}, nil
}

func (c *JsonApiGateway) UAPI(module, function string, arguments cpanelgo.Args, out interface{}) error {
	req := CpanelApiRequest{
		ApiVersion: "uapi",
		Module:     module,
		Function:   function,
		Arguments:  arguments,
	}

	return c.api(req, out)
}

func (c *JsonApiGateway) API2(module, function string, arguments cpanelgo.Args, out interface{}) error {
	req := CpanelApiRequest{
		ApiVersion: "2",
		Module:     module,
		Function:   function,
		Arguments:  arguments,
	}

	var result cpanelgo.API2Result
	err := c.api(req, &result)
	if err == nil {
		err = result.Error()
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(result.Result, out)
}

func (c *JsonApiGateway) API1(module, function string, arguments []string, out interface{}) error {
	args := cpanelgo.Args{}
	for _, v := range arguments {
		args[v] = true
	}

	req := CpanelApiRequest{
		ApiVersion: "1",
		Module:     module,
		Function:   function,
		Arguments:  args,
	}

	return c.api(req, out)
}

func (c *JsonApiGateway) Close() error {
	return nil
}

func (c *JsonApiGateway) api(req CpanelApiRequest, out interface{}) error {
	vals := req.Arguments.Values(req.ApiVersion)
	reqUrl := fmt.Sprintf("https://%s:2083/", c.Hostname)
	switch req.ApiVersion {
	case "uapi":
		// https://hostname.example.com:2083/cpsess##########/execute/Module/function?parameter=value&parameter=value&parameter=value
		reqUrl += fmt.Sprintf("execute/%s/%s?%s", req.Module, req.Function, vals.Encode())
	case "2":
		fallthrough
	case "1":
		// https://hostname.example.com:2083/cpsess##########/json-api/cpanel?cpanel_jsonapi_user=user&cpanel_jsonapi_apiversion=2&cpanel_jsonapi_module=Module&cpanel_jsonapi_func=function&parameter="value"
		vals.Add("cpanel_jsonapi_user", c.Username)
		vals.Add("cpanel_jsonapi_apiversion", req.ApiVersion)
		vals.Add("cpanel_jsonapi_module", req.Module)
		vals.Add("cpanel_jsonapi_func", req.Function)
		reqUrl += fmt.Sprintf("json-api/cpanel?%s", vals.Encode())
	default:
		return fmt.Errorf("Unknown api version: %s", req.ApiVersion)
	}

	httpReq, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return err
	}

	httpReq.SetBasicAuth(c.Username, c.Password)

	if c.cl == nil {
		c.cl = &http.Client{}
		c.cl.Transport = &http.Transport{
			DisableKeepAlives:   true,
			MaxIdleConns:        1,
			MaxIdleConnsPerHost: 1,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: c.Insecure,
			},
		}
	}

	resp, err := c.cl.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return errors.New(resp.Status)
	}

	// limit maximum response size
	lReader := io.LimitReader(resp.Body, int64(cpanelgo.ResponseSizeLimit))

	bytes, err := ioutil.ReadAll(lReader)
	if err != nil {
		return err
	}

	if os.Getenv("DEBUG_CPANEL_RESPONSES") == "1" {
		log.Println(reqUrl)
		log.Println(resp.Status)
		log.Println(req.Function)
		log.Println(req.Arguments)
		log.Println(vals)
		log.Println(string(bytes))
	}

	if len(bytes) == cpanelgo.ResponseSizeLimit {
		return errors.New("API response maximum size exceeded")
	}

	return json.Unmarshal(bytes, out)
}
