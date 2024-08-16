package redispool

import (
	"flag"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/sourcegraph/log/logtest"
)

func TestSchemeMatcher(t *testing.T) {
	tests := []struct {
		urlMaybe  string
		hasScheme bool
	}{
		{"redis://foo.com", true},
		{"https://foo.com", true},
		{"redis://:password@foo.com/0", true},
		{"redis://foo.com/0?password=foo", true},
		{"foo:1234", false},
	}
	for _, test := range tests {
		hasScheme := schemeMatcher.MatchString(test.urlMaybe)
		if hasScheme != test.hasScheme {
			t.Errorf("for string %q, exp != got: %v != %v", test.urlMaybe, test.hasScheme, hasScheme)
		}
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	logtest.Init(m)
	os.Exit(m.Run())
}

func TestDeleteAllKeysWithPrefix(t *testing.T) {
	t.Helper()

	kv := NewTestKeyValue()

	// If we are not on CI, skip the test if our redis connection fails.
	if os.Getenv("CI") == "" {
		if err := kv.Ping(); err != nil {
			t.Skip("could not connect to redis", err)
		}
	}

	var aKeys, bKeys []string
	var key string
	for i := range 10 {
		if i%2 == 0 {
			key = "a:" + strconv.Itoa(i)
			aKeys = append(aKeys, key)
		} else {
			key = "b:" + strconv.Itoa(i)
			bKeys = append(bKeys, key)
		}

		if err := kv.Set(key, []byte(strconv.Itoa(i))); err != nil {
			t.Fatalf("could not set key %s: %v", key, err)
		}
	}

	if err := DeleteAllKeysWithPrefix(kv, "a"); err != nil {
		t.Fatal(err)
	}

	getMulti := func(keys ...string) []string {
		t.Helper()
		var vals []string
		for _, k := range keys {
			v, _ := kv.Get(k).String()
			vals = append(vals, v)
		}
		return vals
	}

	vals := getMulti(aKeys...)
	if got, exp := vals, []string{"", "", "", "", ""}; !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}

	vals = getMulti(bKeys...)
	if got, exp := vals, []string{"1", "3", "5", "7", "9"}; !reflect.DeepEqual(exp, got) {
		t.Errorf("Expected %v, but got %v", exp, got)
	}
}
