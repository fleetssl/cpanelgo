package cpanelgo

import (
	"net/url"
	"reflect"
	"testing"
)

func TestArgs_Values(t *testing.T) {
	args := Args{
		"key=not": "value",
	}
	expected0 := url.Values{
		"key=not": []string{"value"},
	}
	actual0 := args.Values("0")
	if !reflect.DeepEqual(expected0, actual0) {
		t.Errorf("Unexpected Args.Values(), expected: '%+v', got: '%+v'", expected0, actual0)
	}
	actual1 := args.Values("1")
	expected1 := url.Values{
		"key": []string{"not"},
	}
	if !reflect.DeepEqual(actual1, expected1) {
		t.Errorf("Unexpected Args.Values(), expected: '%+v', got: '%+v'", expected0, actual0)
	}
}
