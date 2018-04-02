package main

import (
	"encoding/json"
	"net/url"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

func TestProbablyFileURI(t *testing.T) {
	tests := map[string]bool{
		"file:///a.py":               true,
		"file:///a.py#line=1,char=2": true,
		"/a.py":    true,
		"file:///": true,

		// We don't want to rewrite uris with an explicit non-file scheme
		"git:///a.py": false,

		// We don't want to rewrite uris that have empty paths
		"file://": false,
		"":        false,
	}

	for rawURIStr, expected := range tests {
		parsedURI, err := url.Parse(rawURIStr)
		if err != nil {
			t.Fatalf("got error %s when parsing test uri %s", err, rawURIStr)
		}
		actual := probablyFileURI(parsedURI)
		if actual != expected {
			t.Errorf("for uri %v, expected %v, actual %v", parsedURI, expected, actual)
		}
	}
}

func TestClientToServerURI(t *testing.T) {
	cacheDir := "/tmp"
	projectFileLoc := "/a.py"
	cacheFileLoc := filepath.Join(cacheDir, projectFileLoc)

	tests := map[string]string{
		"file://" + projectFileLoc: "file://" + cacheFileLoc,
		projectFileLoc:             cacheFileLoc,

		// ensure that only the path is modified when rewriting a uri
		"file://" + projectFileLoc + "#line=1,char=2": "file://" + cacheFileLoc + "#line=1,char=2",
		projectFileLoc + "#line=1,char=2":             cacheFileLoc + "#line=1,char=2",

		"file:///": "file://" + filepath.Join(cacheDir, "/"),
		"/":        filepath.Join(cacheDir, "/"),

		// don't rewrite uris with an explicit non-file scheme
		"git:///sourcegraph.com/sourcegraph?SHA": "git:///sourcegraph.com/sourcegraph?SHA",

		// don't rewruite uris with empty paths
		"file://": "file://",
		"":        "",
	}

	for clientURI, rewrittenURI := range tests {
		actual := clientToServerURI(lsp.DocumentURI(clientURI), cacheDir)
		if actual != lsp.DocumentURI(rewrittenURI) {
			t.Errorf("for uri %s, expected %s, actual %s", clientURI, rewrittenURI, actual)
		}
	}
}

func TestServerToClientURI(t *testing.T) {
	cacheDir := "/tmp"
	projectFileLoc := "/a.py"
	cacheFileLoc := filepath.Join(cacheDir, projectFileLoc)

	tests := map[string]string{
		"file://" + cacheFileLoc: "file://" + projectFileLoc,
		cacheFileLoc:             projectFileLoc,

		// ensure that only the path is modified when rewriting a uri
		"file://" + cacheFileLoc + "#line=1,char=2": "file://" + projectFileLoc + "#line=1,char=2",
		cacheFileLoc + "#line=1,char=2":             projectFileLoc + "#line=1,char=2",

		// don't rewrite uris with an explicit non-file scheme
		"git:///sourcegraph.com/sourcegraph?SHA": "git:///sourcegraph.com/sourcegraph?SHA",

		// don't rewruite uris with empty paths
		"file://": "file://",
		"":        "",

		// don't rewrite uris that have paths that don't contain cacheDir as the first element
		"file://" + projectFileLoc:    "file://" + projectFileLoc,
		projectFileLoc:                projectFileLoc,
		"file:///some/other/location": "file:///some/other/location",
		"/some/other/location":        "/some/other/location",

		// no, REALLY don't rewrite uris with an explicit non-file scheme
		"git://" + cacheFileLoc + "?SHA": "git://" + cacheFileLoc + "?SHA",
	}

	for serverURI, rewrittenURI := range tests {
		actual := serverToClientURI(lsp.DocumentURI(serverURI), cacheDir)
		if actual != lsp.DocumentURI(rewrittenURI) {
			t.Errorf("for uri %s, expected %s, actual %s", serverURI, rewrittenURI, actual)
		}
	}
}

func TestWalkURIFields(t *testing.T) {
	tests := map[string][]lsp.DocumentURI{
		`{"textDocument":{"uri":"u1"}}`: []lsp.DocumentURI{"u1"},
		`{"uri":"u1"}`:                  []lsp.DocumentURI{"u1"},

		// `initialize` specific fields
		`{"method":"initialize","rootPath":"u1"}`:                []lsp.DocumentURI{"u1"},
		`{"method":"initialize","rootUri":"u1"}`:                 []lsp.DocumentURI{"u1"},
		`{"method":"initialize","rootPath":"u1","rootUri":"u2"}`: []lsp.DocumentURI{"u1", "u2"},
	}

	for objStr, wantURIs := range tests {
		var obj interface{}
		if err := json.Unmarshal([]byte(objStr), &obj); err != nil {
			t.Error(err)
			continue
		}

		var collectedURIs []string
		update := func(uri lsp.DocumentURI) lsp.DocumentURI {
			collectedURIs = append(collectedURIs, string(uri))
			return "XXX"
		}

		WalkURIFields(obj, update)

		var wantURIStrs []string
		for _, wantURI := range wantURIs {
			wantURIStrs = append(wantURIStrs, string(wantURI))
		}

		sort.Strings(collectedURIs)
		sort.Strings(wantURIStrs)

		if !reflect.DeepEqual(collectedURIs, wantURIStrs) {
			t.Errorf("%s: got URIs %q, want %q", objStr, collectedURIs, wantURIStrs)
		}

		wantObj := objStr
		for _, uri := range collectedURIs {
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
