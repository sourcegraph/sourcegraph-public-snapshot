package dbtypes

import (
	"reflect"
	"testing"
)

func TestJSONMapStringString(t *testing.T) {
	tests := []struct {
		data  interface{}
		value JSONMapStringString

		onlyCheckScan bool
	}{
		{data: "null", value: nil},
		{data: nil, value: nil, onlyCheckScan: true},
		{data: []byte(`{"k1":"v1","k2":"v2"}`), value: map[string]string{"k1": "v1", "k2": "v2"}},
	}
	for _, test := range tests {
		var value JSONMapStringString
		if err := value.Scan(test.data); err != nil {
			t.Errorf("%q: Scan failed: %s", test.data, err)
			continue
		}
		if want := test.value; !reflect.DeepEqual(value, want) {
			t.Errorf("%q: Scan: got %+v, want %+v", test.data, value, want)
		}

		if test.onlyCheckScan {
			continue
		}

		data, err := test.value.Value()
		if err != nil {
			t.Errorf("%v: Value failed: %s", test.value, err)
			continue
		}
		if b, ok := test.data.([]byte); ok {
			if want := b; !reflect.DeepEqual(data, want) {
				t.Errorf("%v: Value: got %q, want %q", test.value, data, want)
				continue
			}
		}
	}
}
