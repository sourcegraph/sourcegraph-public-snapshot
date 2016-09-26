package lspx

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestWalkURIFields(t *testing.T) {
	tests := map[string][]string{
		`{"textDocument":{"uri":"u1"}}`: []string{"u1"},
		`{"uri":"u1"}`:                  []string{"u1"},
	}
	for objStr, wantURIs := range tests {
		var obj interface{}
		if err := json.Unmarshal([]byte(objStr), &obj); err != nil {
			t.Error(err)
			continue
		}

		var uris []string
		collect := func(uri string) { uris = append(uris, uri) }
		update := func(uri string) string { return "XXX" }
		WalkURIFields(obj, collect, update)

		if !reflect.DeepEqual(uris, wantURIs) {
			t.Errorf("%s: got URIs %q, want %q", objStr, uris, wantURIs)
		}

		wantObj := objStr
		for _, uri := range uris {
			wantObj = strings.Replace(wantObj, uri, "XXX", -1)
		}
		gotObj, err := json.Marshal(obj)
		if err != nil {
			t.Error(err)
			continue
		}
		if string(gotObj) != wantObj {
			t.Errorf("%s: got obj %q, want %q after updating URI pointers", objStr, gotObj, wantObj)
		}
	}
}
