package cpanel

import (
	"encoding/json"
	"testing"
)

func TestLiveAPIUnmarshal(t *testing.T) {
	inputs := []string{
		`<?xml version="1.0" ?><cpanelresult><error>A warning occurred while processing this directive.</error>{"func":"installed_hosts"}</cpanelresult>`,
		`<?xml version="1.0" ?><cpanelresult>{"func":"installed_hosts"}</cpanelresult>`,
		`<?xml version="1.0" ?><cpanelresult>{}</cpanelresult>`,
		`<?xml version="1.0" ?><cpanelresult><yeet>haha</yeet>{}</cpanelresult>`,
	}

	var out interface{}
	for _, input := range inputs {
		s, err := extractJSONString(input)
		if err != nil {
			t.Errorf("%s when extracting JSON from: %s", err, input)
			continue
		}
		if err := json.Unmarshal([]byte(s), &out); err != nil {
			t.Errorf("%s when unmarshaling JSON from: %s", err, s)
		}
	}
}
