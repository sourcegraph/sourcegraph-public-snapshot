package base

import (
	"encoding/json"
	"fmt"
	. "testing"
)

func TestSanitizieForJSON(t *T) {

	tests := map[string]interface{}{
		"null": nil,
		"4":    4,
		"8":    8,

		// Int keys should be mapped to string keys
		`{"1":1,"2":4,"3":9}`: map[int]int{1: 1, 2: 4, 3: 9},

		// float keys should be mapped to string keys
		`{"0.4":true,"21.3":false}`: map[float32]bool{.4: true, 21.3: false},

		// nested maps
		`{"4":{"16":"crayons"},"7":{"1":"A","2":"B","3":"C"},"8":{}}`: map[int]map[int]string{
			4: map[int]string{16: "crayons"},
			8: map[int]string{},
			7: map[int]string{1: "A", 2: "B", 3: "C"},
		},

		// structs (only Public fields are expected)
		`{"alpha":4,"beta":"lemons"}`: struct {
			Alpha int
			Beta  string
			delta bool
		}{4, "lemons", false},

		// structs with pointers
		`{"int_ptr":2}`: struct {
			IntPtr    *int
			FloatPtr  *float64
			StringPtr *string
		}{IntPtr(2), nil, nil},
	}

	runTest := func(obj interface{}) string {
		data := SanitizeForJSON(obj, nil)
		bytes, err := json.Marshal(data)
		if err != nil {
			t.Error("JSON encoding failed")
		}
		return string(bytes)
	}

	for expected, test := range tests {
		actual := runTest(test)
		if actual != expected {
			t.Error(actual + " != " + expected)
		}
	}
}

type testStruct struct {
	IncludeMe int
	ExcludeMe int
}

type santizeWithWhitelistTest struct {
	data      interface{}
	whitelist []string
	expected  string
}

func TestSanitizieForJSONWithWhitelist(t *T) {

	tests := []santizeWithWhitelistTest{
		// Include all fields when the whitelist is nil
		santizeWithWhitelistTest{
			testStruct{4, 5},
			nil,
			`{"exclude_me":5,"include_me":4}`,
		},
		// Exclude specified fields
		santizeWithWhitelistTest{
			testStruct{4, 5},
			[]string{"base.testStruct.IncludeMe"},
			`{"include_me":4}`,
		},
		santizeWithWhitelistTest{
			testStruct{4, 5},
			[]string{"base.testStruct.IncludeMe", "base.testStruct.ExcludeMe"},
			`{"exclude_me":5,"include_me":4}`,
		},
		// The field names are the Go names, not the JSON names
		santizeWithWhitelistTest{
			testStruct{4, 5},
			[]string{"base.testStruct.include_me"},
			`{}`,
		},
		// The package name is required
		santizeWithWhitelistTest{
			testStruct{4, 5},
			[]string{"testStruct.IncludeMe"},
			`{}`,
		},
		// The struct name is required
		santizeWithWhitelistTest{
			testStruct{4, 5},
			[]string{"IncludeMe"},
			`{}`,
		},
		// Non-existent fields are not an error
		santizeWithWhitelistTest{
			testStruct{4, 5},
			[]string{"base.testStruct.IncludeMe", "base.testStruct.DoesntExist"},
			`{"include_me":4}`,
		},
	}

	for i, test := range tests {

		whitelist := make(map[string]bool, len(test.whitelist))
		if test.whitelist == nil {
			whitelist = nil
		} else {
			for _, s := range test.whitelist {
				whitelist[s] = true
			}
		}

		sanitized := SanitizeForJSON(test.data, whitelist)
		bytes, err := json.Marshal(sanitized)
		if err != nil {
			t.Error("JSON encoding failed")
		}
		actual := string(bytes)

		if actual != test.expected {
			t.Error(fmt.Sprintf("Test case %v: %v != %v", i, actual, test.expected))
		}
	}
}
