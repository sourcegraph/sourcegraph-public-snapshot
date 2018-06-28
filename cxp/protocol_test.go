package cxp

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestInitializationOptions(t *testing.T) {
	merged := json.RawMessage(`{"b":2}`)
	o := InitializationOptions{
		Other:    map[string]interface{}{"a": float64(1)},
		Settings: ExtensionSettings{Merged: &merged},
	}

	data, err := json.Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"a":1,"settings":{"merged":{"b":2}}}`; string(data) != want {
		t.Errorf("got %s, want %s", data, want)
	}

	var o2 InitializationOptions
	if err := json.Unmarshal(data, &o2); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(o, o2) {
		t.Errorf("got %+v, want %+v", o, o2)
	}
}
