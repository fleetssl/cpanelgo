package cpanel

import (
	"errors"
	"strings"

	"github.com/letsencrypt-cpanel/cpanelgo"
)

type ZoneRecord struct {
	Name   string `json:"name"`
	Record string `json:"record"`
	Type   string `json:"type"`
	Line   int    `json:"line"`
}

type FetchZoneApiResponse struct {
	cpanelgo.BaseAPI2Response
	Data []struct {
		Records       []ZoneRecord `json:"record"`
		Status        int          `json:"status"`
		StatusMessage string       `json:"statusmsg"`
	} `json:"data"`
}

// Returns line number
func (r FetchZoneApiResponse) Find(name, rrType string) (bool, []int) {
	if len(r.Data) == 0 {
		return false, []int{}
	}
	var lines []int
	name = strings.ToLower(name)
	for _, v := range r.Data[0].Records {
		if v.Type == rrType && strings.ToLower(v.Name) == name {
			lines = append(lines, v.Line)
		}
	}
	return len(lines) > 0, lines
}

func (c CpanelApi) FetchZone(domain, types string) (FetchZoneApiResponse, error) {
	var out FetchZoneApiResponse

	err := c.Gateway.API2("ZoneEdit", "fetchzone", cpanelgo.Args{
		"domain": domain,
		"type":   types, // can be multiple CNAME,A,AAAA
	}, &out)

	if err == nil && out.Event.Result != 1 {
		err = errors.New(out.Event.Reason)
	}

	if err == nil && len(out.Data) > 0 && out.Data[0].Status != 1 {
		err = errors.New(out.Data[0].StatusMessage)
	}

	return out, err
}

type AddZoneTextRecordResponse struct {
	cpanelgo.BaseAPI2Response
	Data []struct {
		Result struct {
			Status        int    `json:"status"`
			StatusMessage string `json:"statusmsg"`
		} `json:"result"`
	} `json:"data"`
}

func (c CpanelApi) AddZoneTextRecord(zone, name, txtData string) error {
	var out AddZoneTextRecordResponse

	err := c.Gateway.API2("ZoneEdit", "add_zone_record", cpanelgo.Args{
		"domain":  zone,
		"name":    name,
		"type":    "TXT",
		"txtdata": txtData,
		"ttl":     "360",
	}, &out)

	if err == nil && out.Event.Result != 1 {
		err = errors.New(out.Event.Reason)
	}

	if err == nil && len(out.Data) > 0 && out.Data[0].Result.Status != 1 {
		err = errors.New(out.Data[0].Result.StatusMessage)
	}

	return err
}

type EditZoneTextRecordResponse struct {
	cpanelgo.BaseAPI2Response
	Data []struct {
		Result struct {
			Status        int    `json:"status"`
			StatusMessage string `json:"statusmsg"`
		} `json:"result"`
	} `json:"data"`
}

func (c CpanelApi) EditZoneTextRecord(line int, zone, txtData string) error {
	var out EditZoneTextRecordResponse

	err := c.Gateway.API2("ZoneEdit", "edit_zone_record", cpanelgo.Args{
		"domain":  zone,
		"type":    "TXT",
		"txtdata": txtData,
		"line":    line,
		"ttl":     360,
	}, &out)

	if err == nil && out.Event.Result != 1 {
		err = errors.New(out.Event.Reason)
	}

	if err == nil && len(out.Data) > 0 && out.Data[0].Result.Status != 1 {
		err = errors.New(out.Data[0].Result.StatusMessage)
	}

	return err
}

type FetchZonesApiResponse struct {
	cpanelgo.BaseAPI2Response
	Data []struct {
		Status        int                 `json:"status"`
		StatusMessage string              `json:"statusmsg"`
		Zones         map[string][]string `json:"zones"`
	} `json:"data"`
}

func (r FetchZonesApiResponse) FindRootForName(name string) string {
	if len(r.Data) == 0 {
		return ""
	}
	zones := r.Data[0].Zones
	// Strip labels until we find one that actually has records
	for {
		list, exists := zones[name]
		if exists && len(list) > 0 {
			return name
		}

		idx := strings.Index(name, ".")
		if idx == -1 || idx == (len(name)-1) {
			return ""
		}

		name = name[idx+1:]
	}
}

func (c CpanelApi) FetchZones() (FetchZonesApiResponse, error) {
	var out FetchZonesApiResponse

	err := c.Gateway.API2("ZoneEdit", "fetchzones", cpanelgo.Args{}, &out)

	if err == nil && out.Event.Result != 1 {
		err = errors.New(out.Event.Reason)
	}

	if err == nil && len(out.Data) > 0 && out.Data[0].Status != 1 {
		err = errors.New(out.Data[0].StatusMessage)
	}

	return out, err
}
