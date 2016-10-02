package cpanelgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	megabyte          = 1 * 1024 * 1024
	ResponseSizeLimit = (5 * megabyte) + 1337
	ErrorUnknown      = "Unknown"
)

type BaseResult struct {
	ErrorString string `json:"error"`
}

func (r BaseResult) Error() error {
	if r.ErrorString == "" {
		return nil
	}
	return errors.New(r.ErrorString)
}

type UAPIResult struct {
	BaseResult
	Result json.RawMessage `json:"result"`
}

type API2Result struct {
	BaseResult
	Result json.RawMessage `json:"cpanelresult"`
}

type BaseUAPIResponse struct {
	BaseResult
	StatusCode int      `json:"status"`
	Errors     []string `json:"errors"`
	Messages   []string `json:"messages"`
}

func (r BaseUAPIResponse) Error() error {
	if r.StatusCode == 1 {
		return nil
	}
	err := r.BaseResult.Error()
	if err != nil {
		return err
	}
	if len(r.Errors) == 0 {
		return errors.New(ErrorUnknown)
	}
	return errors.New(strings.Join(r.Errors, "\n"))
}

func (r BaseUAPIResponse) Message() string {
	if r.Messages == nil || len(r.Messages) == 0 {
		return ""
	}
	return strings.Join(r.Messages, "\n")
}

type BaseAPI2Response struct {
	BaseResult
	Event struct {
		Result int    `json:"result"`
		Reason string `json:"reason"`
	} `json:"event"`
}

func (r BaseAPI2Response) Error() error {
	if r.Event.Result == 1 {
		return nil
	}
	err := r.BaseResult.Error()
	if err != nil {
		return err
	}
	if len(r.Event.Reason) == 0 {
		return errors.New("Unknown")
	}
	return errors.New(r.Event.Reason)
}

type BaseAPI1Response struct {
	// other stuff here "apiversion":"1","type":"event","module":"Serverinfo","func":"servicestatus","source":"module"
	Data struct {
		Result string `json:"result"`
	} `json:"data"`
	ErrorString string `json:"error"`
	// "event":{"result":1,"reason":"blah blah"}}
	Event struct {
		Result int    `json:"result"`
		Reason string `json:"reason"`
	} `json:"event"`
}

func (r BaseAPI1Response) Error() error {
	if r.ErrorString != "" {
		return errors.New(r.ErrorString)
	}
	if r.Event.Result != 1 {
		// if the result != 1 the reason usually present in error ^ so kinda redundant to check, but check just in case
		if len(r.Event.Reason) == 0 {
			return errors.New("Unknown")
		}
		return errors.New(r.Event.Reason)
	}
	return nil
}

type Args map[string]interface{}

func (a Args) Values(apiVersion string) url.Values {
	vals := url.Values{}
	for k, v := range a {
		if apiVersion == "1" {
			kv := strings.SplitN(k, "=", 2)
			if len(kv) == 1 {
				vals.Add(kv[0], "")
			} else if len(kv) == 2 {
				vals.Add(kv[0], kv[1])
			}
		} else {
			vals.Add(k, fmt.Sprintf("%v", v))
		}
	}
	return vals
}

type ApiGateway interface {
	UAPI(module, function string, arguments Args, out interface{}) error
	API2(module, function string, arguments Args, out interface{}) error
	API1(module, function string, arguments []string, out interface{}) error
	Close() error
}

type Api struct {
	Gateway ApiGateway
}

func NewApi(gw ApiGateway) Api {
	return Api{
		Gateway: gw,
	}
}

func (a Api) Close() error {
	if a.Gateway != nil {
		return a.Gateway.Close()
	} else {
		return nil
	}
}

type MaybeInt64 int64

func (m *MaybeInt64) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(*m))
}

func (m *MaybeInt64) UnmarshalJSON(buf []byte) error {
	var out interface{}
	if err := json.Unmarshal(buf, &out); err != nil {
		return err
	}

	switch v := out.(type) {
	case string:
		if len(v) == 0 {
			*m = 0
			break
		}

		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		*m = MaybeInt64(f)
	case float64:
		*m = MaybeInt64(v)
	case nil:
		*m = 0
	default:
		return errors.New("Not a string or int64")
	}

	return nil
}
