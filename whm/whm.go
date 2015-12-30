package whm

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/letsencrypt-cpanel/cpanelgo"
)

type BaseWhmApiResponse struct {
	Metadata struct {
		Reason    string      `json:"reason"`
		ResultRaw interface{} `json:"result"`
	} `json:"metadata"`
}

func (r BaseWhmApiResponse) Error() error {
	if r.Result() == 1 {
		return nil
	}
	if len(r.Metadata.Reason) == 0 {
		return errors.New("Unknown")
	}
	return errors.New(r.Metadata.Reason)
}

// WHM randomly returns this as a string, gg
func (r BaseWhmApiResponse) Result() int {
	if v, ok := r.Metadata.ResultRaw.(float64); ok { // default for Number JSON type is f64
		return int(v)
	}

	if s, ok := r.Metadata.ResultRaw.(string); ok {
		if v, err := strconv.Atoi(s); err != nil {
			return -1
		} else {
			return v
		}
	}

	return -1
}

// This implements a standalone WHM client, not for the cPanel API
type WhmApi struct {
	Hostname   string
	Username   string
	AccessHash string
	Insecure   bool
}

func NewWhmApi(hostname, username, accessHash string, insecure bool) WhmApi {
	accessHash = strings.Replace(accessHash, "\n", "", -1)
	accessHash = strings.Replace(accessHash, "\r", "", -1)

	return WhmApi{
		Hostname:   hostname,
		Username:   username,
		AccessHash: accessHash,
		Insecure:   insecure,
	}
}

func (c *WhmApi) WHMAPI1(function string, arguments cpanelgo.Args, out interface{}) error {
	vals := url.Values{
		"api.version": []string{"1"},
	}

	if arguments != nil && len(arguments) > 0 {
		for k, v := range arguments {
			if ver, ok := arguments["cpanel_jsonapi_apiversion"]; ok {
				if ver == "1" {
					kv := strings.SplitN(k, "=", 2)
					if len(kv) == 1 {
						vals.Add(kv[0], "")
					} else if len(kv) == 2 {
						vals.Add(kv[0], kv[1])
					}
					continue
				}
			}
			vals.Add(k, fmt.Sprintf("%v", v))
		}
	}

	reqUrl := fmt.Sprintf("https://%s:2087/json-api/%s?%s", c.Hostname, function, vals.Encode())
	req, err := http.NewRequest("GET", reqUrl, nil)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("WHM %s:%s", c.Username, c.AccessHash))

	cl := &http.Client{}
	cl.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.Insecure,
		},
	}

	resp, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return errors.New(resp.Status)
	}

	// limit maximum response size
	lReader := io.LimitReader(resp.Body, cpanelgo.ResponseSizeLimit)

	bytes, err := ioutil.ReadAll(lReader)
	if err != nil {
		return err
	}

	if os.Getenv("DEBUG_CPANEL_RESPONSES") == "1" {
		log.Println(reqUrl)
		log.Println(resp.Status)
		log.Println(function)
		log.Println(arguments)
		log.Println(string(bytes))
	}

	if len(bytes) == cpanelgo.ResponseSizeLimit {
		return errors.New("API response maximum size exceeded")
	}

	return json.Unmarshal(bytes, out)
}
