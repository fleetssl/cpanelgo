package whm

import (
	"bytes"
	"crypto/hmac"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"encoding/base64"

	"time"

	"crypto/sha1"
	"encoding/base32"

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
	Password   string
	Insecure   bool
	TotpSecret string
	cl         *http.Client
}

func NewWhmApiAccessHash(hostname, username, accessHash string, insecure bool) WhmApi {
	accessHash = strings.Replace(accessHash, "\n", "", -1)
	accessHash = strings.Replace(accessHash, "\r", "", -1)

	return WhmApi{
		Hostname:   hostname,
		Username:   username,
		AccessHash: accessHash,
		Insecure:   insecure,
	}
}

func NewWhmApiAccessHashTotp(hostname, username, accessHash string, insecure bool, secret string) WhmApi {
	accessHash = strings.Replace(accessHash, "\n", "", -1)
	accessHash = strings.Replace(accessHash, "\r", "", -1)

	return WhmApi{
		Hostname:   hostname,
		Username:   username,
		AccessHash: accessHash,
		Insecure:   insecure,
		TotpSecret: secret,
	}
}

func NewWhmApiPassword(hostname, username, password string, insecure bool) WhmApi {
	return WhmApi{
		Hostname: hostname,
		Username: username,
		Password: password,
		Insecure: insecure,
	}
}

// Force POST method for these WHM API1 functions
var forcePost = map[string]bool{
	"cpanel": true,
}

func (c *WhmApi) WHMAPI1(function string, arguments cpanelgo.Args, out interface{}) error {
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

	method := "GET"
	if _, ok := forcePost[function]; ok {
		method = "POST"
	}

	version := "0"
	if arguments["cpanel_jsonapi_apiversion"] == "1" {
		version = "1"
	}
	vals := arguments.Values(version)
	vals.Set("api.version", "1")

	var req *http.Request
	var reqUrl string
	var err error

	if method == "GET" {
		reqUrl = fmt.Sprintf("https://%s:2087/json-api/%s?%s", c.Hostname, function, vals.Encode())
		req, err = http.NewRequest("GET", reqUrl, nil)
		if err != nil {
			return err
		}

	} else if method == "POST" {
		reqUrl = fmt.Sprintf("https://%s:2087/json-api/%s", c.Hostname, function)
		req, err = http.NewRequest("POST", reqUrl, strings.NewReader(vals.Encode()))
	}

	if c.AccessHash != "" {
		req.Header.Add("Authorization", fmt.Sprintf("WHM %s:%s", c.Username, c.AccessHash))
	} else if c.Password != "" {
		req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", c.Username, c.Password))))
	}

	if c.TotpSecret != "" {
		decodedSecret, _ := base32.StdEncoding.DecodeString(c.TotpSecret)
		otp, _ := totp(decodedSecret, time.Now().Unix(), sha1.New, 6)

		req.Header.Add("X-CPANEL-OTP", otp)
	}

	resp, err := c.cl.Do(req)
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
		log.Println(function)
		log.Println(arguments)
		log.Println(string(bytes))
	}

	if len(bytes) == cpanelgo.ResponseSizeLimit {
		return errors.New("API response maximum size exceeded")
	}

	return json.Unmarshal(bytes, out)
}

type VersionApiResponse struct {
	BaseWhmApiResponse
	Data struct {
		Version string `json:"version"`
	} `json:"data"`
}

func (a WhmApi) Version() (VersionApiResponse, error) {
	var out VersionApiResponse
	err := a.WHMAPI1("version", cpanelgo.Args{}, &out)
	if err == nil && out.Result() != 1 {
		err = out.Error()
	}
	return out, err
}

func totp(k []byte, t int64, h func() hash.Hash, l int64) (string, error) {
	if l > 9 || l < 1 {
		return "", errors.New("Totp: Length out of range.")
	}

	time := new(bytes.Buffer)

	err := binary.Write(time, binary.BigEndian, (t-int64(0))/int64(30))
	if err != nil {
		return "", err
	}

	hash := hmac.New(h, k)
	hash.Write(time.Bytes())
	v := hash.Sum(nil)

	o := v[len(v)-1] & 0xf
	c := (int32(v[o]&0x7f)<<24 | int32(v[o+1])<<16 | int32(v[o+2])<<8 | int32(v[o+3])) % 1000000000

	return fmt.Sprintf("%010d", c)[10-l : 10], nil
}
