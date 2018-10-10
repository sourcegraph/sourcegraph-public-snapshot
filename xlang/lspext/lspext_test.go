package lspext

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegraph/go-lsp"
)

func TestWalkURIFields(t *testing.T) {
	tests := map[string][]lsp.DocumentURI{
		`{"textDocument":{"uri":"u1"}}`: []lsp.DocumentURI{"u1"},
		`{"uri":"u1"}`:                  []lsp.DocumentURI{"u1"},
	}
	for objStr, wantURIs := range tests {
		var obj interface{}
		if err := json.Unmarshal([]byte(objStr), &obj); err != nil {
			t.Error(err)
			continue
		}

		var uris []lsp.DocumentURI
		collect := func(uri lsp.DocumentURI) { uris = append(uris, uri) }
		update := func(uri lsp.DocumentURI) lsp.DocumentURI { return "XXX" }
		WalkURIFields(obj, collect, update)

		if !reflect.DeepEqual(uris, wantURIs) {
			t.Errorf("%s: got URIs %q, want %q", objStr, uris, wantURIs)
		}

		wantObj := objStr
		for _, uri := range uris {
			wantObj = strings.Replace(wantObj, string(uri), "XXX", -1)
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

func TestWalkURIFields_struct(t *testing.T) {
	v := lsp.PublishDiagnosticsParams{URI: "u1"}

	var uris []lsp.DocumentURI
	collect := func(uri lsp.DocumentURI) { uris = append(uris, uri) }
	update := func(uri lsp.DocumentURI) lsp.DocumentURI { return "XXX" }
	WalkURIFields(&v, collect, update)

	if want := []lsp.DocumentURI{"u1"}; !reflect.DeepEqual(uris, want) {
		t.Errorf("got %v, want %v", uris, want)
	}

	if want := "XXX"; string(v.URI) != want {
		t.Errorf("got %q, want %q", v.URI, want)
	}
}
