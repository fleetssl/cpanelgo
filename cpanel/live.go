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

	readStr := read.String()

	if n := strings.Index(readStr, "<cpanelresult>{"); n != -1 {
		asJson := readStr[strings.Index(readStr, "<cpanelresult>")+14:]
		asJson = asJson[:strings.LastIndex(asJson, "</cpanelresult>")]

		if out != nil {
			return json.Unmarshal([]byte(asJson), out)
		} else {
			return nil
		}
	}

	return fmt.Errorf("Failed to unmarshal LiveAPI response: %v", readStr)
}
