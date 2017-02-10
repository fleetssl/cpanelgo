package cpanel

import (
	"errors"

	"github.com/letsencrypt-cpanel/cpanelgo"
)

type ZoneRecord struct {
	Name   string              `json:"name"`
	Record string              `json:"record"`
	Type   string              `json:"type"`
	Raw    string              `json:"raw"`
	TTL    string              `json:"raw"`
	Serial string              `json:"serial"`
	Line   cpanelgo.MaybeInt64 `json:"line"`
}

type FetchZoneApiResponse struct {
	cpanelgo.BaseAPI2Response
	Data []struct {
		StatusMessage string       `json:"statusmsg"`
		Records       []ZoneRecord `json:"record"`
	} `json:"data"`
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

	return out, err
}

type FetchZoneRecordsApiResponse struct {
	cpanelgo.BaseAPI2Response
	Data []ZoneRecord `json:"data"`
}

func (c CpanelApi) FetchZoneRecords(domain string, args cpanelgo.Args) (FetchZoneRecordsApiResponse, error) {
	var out FetchZoneRecordsApiResponse

	if args == nil {
		args = cpanelgo.Args{}
	}
	args["domain"] = domain

	err := c.Gateway.API2("ZoneEdit", "fetchzone_records", args, &out)

	if err == nil && out.Event.Result != 1 {
		err = errors.New(out.Event.Reason)
	}

	return out, err
}

type EditZoneRecordApiResponse struct {
	cpanelgo.BaseAPI2Response
	Data []struct {
		Result struct {
			NewSerial     cpanelgo.MaybeInt64 `json:"newserial"`
			StatusMessage string              `json:"statusmsg"`
			Status        cpanelgo.MaybeInt64 `json:"status"`
		} `json:"result"`
	} `json:"data"`
}

func (c CpanelApi) EditZoneRecord(args cpanelgo.Args) (EditZoneRecordApiResponse, error) {
	var out EditZoneRecordApiResponse
	err := c.Gateway.API2("ZoneEdit", "edit_zone_record", args, &out)

	if err == nil && out.Event.Result != 1 {
		err = errors.New(out.Event.Reason)
	}

	return out, err
}
