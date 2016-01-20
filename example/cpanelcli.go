package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"bufio"

	"io/ioutil"

	"path/filepath"

	"bytes"

	"github.com/letsencrypt-cpanel/cpanelgo"
	"github.com/letsencrypt-cpanel/cpanelgo/cpanel"
	"github.com/letsencrypt-cpanel/cpanelgo/whm"
)

var mode, hostname, username, password, accesshash, impersonate string
var debug, insecure, pretty bool

var version, module, function string

func init() {

	flag.Usage = func() {
		cmd := filepath.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", cmd)
		fmt.Fprintf(os.Stderr, "cPanel example:\n")
		fmt.Fprintf(os.Stderr, "# %s -mode cpanel -hostname 127.0.0.1 -username test -password test -version uapi -module Themes -function get_theme_base\n", cmd)
		fmt.Fprintf(os.Stderr, "\tIf password isn't specified, will be prompted\n\n")
		fmt.Fprintf(os.Stderr, "WHM impersonation example:\n")
		fmt.Fprintf(os.Stderr, "# %s -mode whmimp -hostname 127.0.0.1 -username root -impersonate test -accesshash .accesshash -version uapi -module Themes -function get_theme_base\n\n", cmd)
		fmt.Fprintf(os.Stderr, "WHM example:\n")
		fmt.Fprintf(os.Stderr, "# %s -mode whm -hostname 127.0.0.1 -username root -accesshash .accesshash -function listaccts\n\n", cmd)
		fmt.Fprintf(os.Stderr, "To show extra debug use -debug and to pretty print json result use -pretty\n")
	}

	// required flags
	flag.StringVar(&mode, "mode", "", "cpanel | whm | whmimp")

	// optional flags
	flag.BoolVar(&insecure, "insecure", true, "insecure ssl connection to cpanel/whm")
	flag.BoolVar(&debug, "debug", false, "debug cpanel responses")
	flag.BoolVar(&pretty, "pretty", false, "pretty cpanel json response")

	// flags for cpanel
	flag.StringVar(&password, "password", "", "password for cpanel")

	// flags for cpanel/whm/whmimp
	flag.StringVar(&hostname, "hostname", "", "hostname to connect to")
	flag.StringVar(&username, "username", "", "username to authenticate")

	// flags for whm/whmimp
	flag.StringVar(&accesshash, "accesshash", "", "access hash file path for whm/whmimp")
	flag.StringVar(&impersonate, "impersonate", "", "user to impersonate for whmimp")

	// flags for all
	flag.StringVar(&version, "version", "", "uapi | 2 | 1")
	flag.StringVar(&module, "module", "", "module to run (eg Branding)")
	flag.StringVar(&function, "function", "", "function to run (eg include)")
}

func required(v interface{}, msg string) {
	if v == reflect.Zero(reflect.TypeOf(v)).Interface() {
		log.Fatal(msg)
	}
}

var modes = map[string]func(){
	"cpanel": modeCpanel,
	"whm":    modeWhm,
	"whmimp": modeWhmImp,
}

func ifpanic(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	required(mode, "Please specify a mode")
	required(hostname, "Please specify a hostname")

	if debug {
		os.Setenv("DEBUG_CPANEL_RESPONSES", "1")
	}

	f, ok := modes[mode]
	if !ok {
		log.Fatal("Unknown mode:", mode)
	}

	f()
}

func getArgs() cpanelgo.Args {
	var args = cpanelgo.Args{}

	if flag.NArg() > 0 {
		for _, a := range flag.Args() {
			kv := strings.SplitN(a, "=", 2)
			if len(kv) == 1 {
				args[kv[0]] = ""
			} else if len(kv) == 2 {
				args[kv[0]] = kv[1]
			}
		}
	}

	return args
}

func modeCpanel() {
	required(username, "Please specify a username")
	if password == "" {
		fmt.Print("Please enter password: ")
		password, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		password = strings.Trim(password, "\n")
	}
	required(password, "Please specify a password")
	required(version, "Please specify an api version")

	cl, err := cpanel.NewJsonApi(hostname, username, password, insecure)
	ifpanic(err)

	/*
		theme, err := cl.GetTheme()
		log.Println(theme, err)
		parked, err := cl.ListParkedDomains()
		log.Println(parked, err)
		cl.SetVar("dprefix", "../")
		cl.SetVar("hidehelp", "1")
		branding, err := cl.BrandingInclude("stdheader.html")
		log.Println(branding, err)
	*/

	api(cl)
}

func modeWhmImp() {
	required(username, "Please specify a username")
	required(impersonate, "Please specify a user to impersonate")
	required(accesshash, "Please specify an access hash file")
	required(version, "Please specify an api version")

	ahBytes, err := ioutil.ReadFile(accesshash)
	ifpanic(err)
	if len(ahBytes) == 0 {
		log.Fatal("accesshash file was empty")
	}

	cl := whm.NewWhmImpersonationApi(hostname, username, string(ahBytes), impersonate, insecure)

	/*
		theme, err := cl.GetTheme()
		log.Println(theme, err)
		parked, err := cl.ListParkedDomains()
		log.Println(parked, err)
		branding, err := cl.BrandingInclude("stdheader.html")
		log.Println(branding, err)
	*/

	api(cl)
}

func api(cl cpanel.CpanelApi) {
	var out json.RawMessage
	switch version {
	case "uapi":
		err := cl.Gateway.UAPI(module, function, getArgs(), &out)
		ifpanic(err)
		printResult(out)
		var response cpanelgo.BaseUAPIResponse
		err = json.Unmarshal(out, &response)
		ifpanic(err)
		ifpanic(response.Error())
		fmt.Printf("%+v", response)
	case "2":
		err := cl.Gateway.API2(module, function, getArgs(), &out)
		ifpanic(err)
		printResult(out)
		var response cpanelgo.BaseAPI2Response
		err = json.Unmarshal(out, &response)
		ifpanic(err)
		ifpanic(response.Error())
		fmt.Printf("%+v", response)
	case "1":
		err := cl.Gateway.API1(module, function, flag.Args(), &out)
		ifpanic(err)
		printResult(out)
		var response cpanelgo.BaseAPI1Response
		err = json.Unmarshal(out, &response)
		ifpanic(err)
		ifpanic(response.Error())
		fmt.Printf("%+v", response)
	default:
		log.Fatal("Unknown version: %q, expected uapi, 2 or 1", version)
	}
}

func modeWhm() {
	required(username, "Please specify a username")
	required(accesshash, "Please specify an access hash file")

	ahBytes, err := ioutil.ReadFile(accesshash)
	ifpanic(err)
	if len(ahBytes) == 0 {
		log.Fatal("accesshash file was empty")
	}

	whmcl := whm.NewWhmApi(hostname, username, string(ahBytes), insecure)
	ifpanic(err)

	var out json.RawMessage
	err = whmcl.WHMAPI1(function, getArgs(), &out)
	ifpanic(err)

	printResult(out)
}

func printResult(out json.RawMessage) {
	if pretty {
		var pretty bytes.Buffer
		json.Indent(&pretty, out, "", "\t")

		fmt.Println(string(pretty.Bytes()))
	} else {
		fmt.Println(string(out))
	}
}
