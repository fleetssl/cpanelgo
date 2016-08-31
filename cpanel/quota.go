package cpanel

import (
	"encoding/json"

	"strconv"

	"github.com/letsencrypt-cpanel/cpanelgo"
)

type GetQuotaInfoApiResponse struct {
	cpanelgo.BaseUAPIResponse
	Data struct {
		UnderQuotaOverall *json.RawMessage `json:"under_quota_overall"`
	} `json:"data"`
}

func (q GetQuotaInfoApiResponse) IsUnderQuota() bool {
	// seems to be a string when "1" but an int when 0
	// cover all the bases

	// not present, assume under quota
	if q.Data.UnderQuotaOverall == nil {
		return true
	}

	// string, parse
	s := ""
	if err := json.Unmarshal(*q.Data.UnderQuotaOverall, &s); err == nil {
		n, _ := strconv.ParseInt(s, 10, 0)
		return n == 1
	}

	// int
	i := 0
	if err := json.Unmarshal(*q.Data.UnderQuotaOverall, &s); err == nil {
		return i == 1
	}

	return false
}

func (c CpanelApi) GetQuotaInfo() (GetQuotaInfoApiResponse, error) {
	var out GetQuotaInfoApiResponse
	err := c.Gateway.UAPI("Quota", "get_quota_info", cpanelgo.Args{}, &out)
	if err == nil {
		err = out.Error()
	}
	return out, err
}
