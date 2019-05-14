package cpanel

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"log"
	"os"

	"github.com/letsencrypt-cpanel/cpanelgo"
)

type LiveApiGateway struct {
	net.Conn
}

func NewLiveApi(network, address string) (CpanelApi, error) {
	c := &LiveApiGateway{}

	conn, err := net.Dial(network, address)
	if err != nil {
		return CpanelApi{}, err
	}
	c.Conn = conn

	if err := c.exec(`<cpaneljson enable="1">`, nil); err != nil {
		return CpanelApi{}, fmt.Errorf("Enabling JSON: %v", err)
	}

	return CpanelApi{cpanelgo.NewApi(c)}, nil
}

func (c *LiveApiGateway) UAPI(module, function string, arguments cpanelgo.Args, out interface{}) error {
	req := CpanelApiRequest{
		RequestType: "exec",
		ApiVersion:  "uapi",
		Module:      module,
		Function:    function,
		Arguments:   arguments,
	}

	return c.api(req, out)
}

func (c *LiveApiGateway) API2(module, function string, arguments cpanelgo.Args, out interface{}) error {
	req := CpanelApiRequest{
		RequestType: "exec",
		ApiVersion:  "2",
		Module:      module,
		Function:    function,
		Arguments:   arguments,
	}

	return c.api(req, out)
}

func (c *LiveApiGateway) API1(module, function string, arguments []string, out interface{}) error {
	req := map[string]interface{}{
		"module":     module,
		"reqtype":    "exec",
		"func":       function,
		"apiversion": "1",
		"args":       arguments,
	}
	bytes, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return c.exec("<cpanelaction>"+string(bytes)+"</cpanelaction>", out)
}

func (c *LiveApiGateway) Close() error {
	return c.Conn.Close()
}

func (c *LiveApiGateway) api(req CpanelApiRequest, out interface{}) error {
	buf, err := json.Marshal(req)
	if err != nil {
		return err
	}
	if os.Getenv("DEBUG_CPANEL_RESPONSES") == "1" {
		log.Println("[Lets Encrypt for cPanel] Request: ", string(buf))
	}
	switch req.ApiVersion {
	case "uapi":
		var result cpanelgo.UAPIResult
		err := c.exec("<cpanelaction>"+string(buf)+"</cpanelaction>", &result)
		if err == nil {
			err = result.Error()
		}
		if err != nil {
			return err
		}

		if os.Getenv("DEBUG_CPANEL_RESPONSES") == "1" {
			log.Println("[Lets Encrypt for cPanel] UResult: ", string(result.Result))
		}
		return json.Unmarshal(result.Result, out)
	case "2":
		var result cpanelgo.API2Result
		err := c.exec("<cpanelaction>"+string(buf)+"</cpanelaction>", &result)
		if err == nil {
			err = result.Error()
		}
		if err != nil {
			return err
		}
		if os.Getenv("DEBUG_CPANEL_RESPONSES") == "1" {
			log.Println("[Lets Encrypt for cPanel] 2Result: ", string(result.Result))
		}
		return json.Unmarshal(result.Result, out)
	default:
		return c.exec("<cpanelaction>"+string(buf)+"</cpanelaction>", out)
	}
}

func endsWith(where []byte, what string) bool {
	n := 0
	i := len(where) - len(what)
	if i < 0 {
		return false
	}
	for ; i >= 0 && i < len(where); i++ {
		if where[i] != what[n] {
			return false
		}
		n++
	}
	return true
}

func extractJSONString(s string) (string, error) {
	needles := []string{
		"<cpanelresult>{",
		"</error>{",
		">{",
	}
	var found bool
	for _, needle := range needles {
		pos := strings.Index(s, needle)
		if pos == -1 {
			continue
		}
		s = s[pos+len(needle)-1:]
		found = true
		break
	}
	if !found {
		return "", fmt.Errorf("Could not find start of JSON in: %s", s)
	}
	eof := strings.Index(s, "</cpanelresult>")
	if eof == -1 {
		return "", fmt.Errorf("Does not appear to be well-formed: %s", s)
	}
	return s[:eof], nil
}

func (c *LiveApiGateway) exec(req string, out interface{}) error {
	if _, err := fmt.Fprintf(c, "%d\n%s", len(req), req); err != nil {
		return err
	}

	var read bytes.Buffer
	rd := bufio.NewReader(c)

	line, _, err := rd.ReadLine() // ignore isprefix
	for err == nil {
		read.Write(line)

		if endsWith(read.Bytes(), "</cpanelresult>") {
			break
		}

		// limit memory footprint of any api response
		if read.Len() >= cpanelgo.ResponseSizeLimit {
			return errors.New("Exceeded maximum API response size")
		}
		line, _, err = rd.ReadLine()
	}
	if err != nil && err != io.EOF {
		return err
	}

	if out == nil {
		return nil
	}

	asJSON, err := extractJSONString(read.String())
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(asJSON), out)
}
