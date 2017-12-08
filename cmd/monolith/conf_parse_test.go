package main

import (
	"sort"
	"strings"
	"testing"
)

func TestGetEnvironFromConfig(t *testing.T) {
	// We test one complicated example which covers all the types we expect.
	rawConfig := `{
"key1ID": "value",
"key2Test": 5,
"key3CamelCase": true,
"key4": [1, 2],
"key5": {"foo": "bar"},
"key6": [{"foo": "bar"}],
}
`
	wantM := map[string]string{
		"KEY1_ID":         "value",
		"KEY2_TEST":       "5",
		"KEY3_CAMEL_CASE": "true",
		"KEY4":            "[1,2]",
		"KEY5":            `{"foo":"bar"}`,
		"KEY6":            `[{"foo":"bar"}]`,
	}
	gotM, err := getEnvironFromConfig(rawConfig)
	if err != nil {
		t.Fatal(err)
	}

	marshalEnviron := func(m map[string]string) string {
		parts := []string{}
		for k, v := range m {
			parts = append(parts, k+"="+v)
		}
		sort.Strings(parts)
		return strings.Join(parts, " ")
	}
	got := marshalEnviron(gotM)
	want := marshalEnviron(wantM)
	if got != want {
		t.Fatalf("Unexpected environ:\ngot:  %v\nwant: %v", got, want)
	}
}
