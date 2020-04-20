package bundles

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestExists(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("unexpected method. want=%s have=%s", "GET", r.Method)
		}
		if r.URL.Path != "/dbs/42/exists" {
			t.Errorf("unexpected path. want=%s have=%s", "/dbs/42/exists", r.URL.Path)
		}

		_, _ = w.Write([]byte(`true`))
	}))
	defer ts.Close()

	client := &bundleClientImpl{bundleManagerURL: ts.URL, bundleID: 42}
	exists, err := client.Exists(context.Background(), "main.go")
	if err != nil {
		t.Fatalf("unexpected error querying exists: %s", err)
	} else if !exists {
		t.Errorf("unexpected path to exist")
	}
}

func TestExistsBadResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := &bundleClientImpl{bundleManagerURL: ts.URL, bundleID: 42}
	_, err := client.Exists(context.Background(), "main.go")
	if err == nil {
		t.Fatalf("unexpected nil error querying exists")
	}
}

func TestDefinitions(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedQuery := map[string]string{
			"path":      "main.go",
			"line":      "10",
			"character": "20",
		}

		if r.Method != "GET" {
			t.Errorf("unexpected method. want=%s have=%s", "GET", r.Method)
		}
		if r.URL.Path != "/dbs/42/definitions" {
			t.Errorf("unexpected path. want=%s have=%s", "/dbs/42/exists", r.URL.Path)
		}
		if !compareQuery(r.URL.Query(), expectedQuery) {
			t.Errorf("unexpected query. want=%v have=%s", expectedQuery, r.URL.Query().Encode())
		}

		_, _ = w.Write([]byte(`[
			{"path": "foo.go", "range": {"start": {"line": 1, "character": 2}, "end": {"line": 3, "character": 4}}},
			{"path": "bar.go", "range": {"start": {"line": 5, "character": 6}, "end": {"line": 7, "character": 8}}}
		]`))
	}))
	defer ts.Close()

	expected := []Location{
		{DumpID: 42, Path: "foo.go", Range: Range{Start: Position{1, 2}, End: Position{3, 4}}},
		{DumpID: 42, Path: "bar.go", Range: Range{Start: Position{5, 6}, End: Position{7, 8}}},
	}

	client := &bundleClientImpl{bundleManagerURL: ts.URL, bundleID: 42}
	definitions, err := client.Definitions(context.Background(), "main.go", 10, 20)
	if err != nil {
		t.Fatalf("unexpected error querying definitions: %s", err)
	} else if !reflect.DeepEqual(definitions, expected) {
		t.Errorf("unexpected definitions. want=%v have=%v", expected, definitions)
	}
}

func TestReferences(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedQuery := map[string]string{
			"path":      "main.go",
			"line":      "10",
			"character": "20",
		}

		if r.Method != "GET" {
			t.Errorf("unexpected method. want=%s have=%s", "GET", r.Method)
		}
		if r.URL.Path != "/dbs/42/references" {
			t.Errorf("unexpected path. want=%s have=%s", "/dbs/42/exists", r.URL.Path)
		}
		if !compareQuery(r.URL.Query(), expectedQuery) {
			t.Errorf("unexpected query. want=%v have=%s", expectedQuery, r.URL.Query().Encode())
		}

		_, _ = w.Write([]byte(`[
			{"path": "foo.go", "range": {"start": {"line": 1, "character": 2}, "end": {"line": 3, "character": 4}}},
			{"path": "bar.go", "range": {"start": {"line": 5, "character": 6}, "end": {"line": 7, "character": 8}}}
		]`))
	}))
	defer ts.Close()

	expected := []Location{
		{DumpID: 42, Path: "foo.go", Range: Range{Start: Position{1, 2}, End: Position{3, 4}}},
		{DumpID: 42, Path: "bar.go", Range: Range{Start: Position{5, 6}, End: Position{7, 8}}},
	}

	client := &bundleClientImpl{bundleManagerURL: ts.URL, bundleID: 42}
	references, err := client.References(context.Background(), "main.go", 10, 20)
	if err != nil {
		t.Fatalf("unexpected error querying references: %s", err)
	} else if !reflect.DeepEqual(references, expected) {
		t.Errorf("unexpected references. want=%v have=%v", expected, references)
	}
}

func TestHover(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedQuery := map[string]string{
			"path":      "main.go",
			"line":      "10",
			"character": "20",
		}

		if r.Method != "GET" {
			t.Errorf("unexpected method. want=%s have=%s", "GET", r.Method)
		}
		if r.URL.Path != "/dbs/42/hover" {
			t.Errorf("unexpected path. want=%s have=%s", "/dbs/42/exists", r.URL.Path)
		}
		if !compareQuery(r.URL.Query(), expectedQuery) {
			t.Errorf("unexpected query. want=%v have=%s", expectedQuery, r.URL.Query().Encode())
		}

		_, _ = w.Write([]byte(`{
			"text": "starts the program",
			"range": {"start": {"line": 1, "character": 2}, "end": {"line": 3, "character": 4}}
		}`))
	}))
	defer ts.Close()

	expectedText := "starts the program"
	expectedRange := Range{
		Start: Position{1, 2},
		End:   Position{3, 4},
	}

	client := &bundleClientImpl{bundleManagerURL: ts.URL, bundleID: 42}
	text, r, exists, err := client.Hover(context.Background(), "main.go", 10, 20)
	if err != nil {
		t.Fatalf("unexpected error querying hover: %s", err)
	}

	if !exists {
		t.Errorf("expected hover text to exist")
	} else {
		if text != expectedText {
			t.Errorf("unexpected hover text. want=%v have=%v", expectedText, text)
		} else if !reflect.DeepEqual(r, expectedRange) {
			t.Errorf("unexpected hover range. want=%v have=%v", expectedRange, r)
		}
	}
}

func TestHoverNull(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedQuery := map[string]string{
			"path":      "main.go",
			"line":      "10",
			"character": "20",
		}

		if r.Method != "GET" {
			t.Errorf("unexpected method. want=%s have=%s", "GET", r.Method)
		}
		if r.URL.Path != "/dbs/42/hover" {
			t.Errorf("unexpected path. want=%s have=%s", "/dbs/42/exists", r.URL.Path)
		}
		if !compareQuery(r.URL.Query(), expectedQuery) {
			t.Errorf("unexpected query. want=%v have=%s", expectedQuery, r.URL.Query().Encode())
		}

		_, _ = w.Write([]byte(`null`))
	}))
	defer ts.Close()

	client := &bundleClientImpl{bundleManagerURL: ts.URL, bundleID: 42}
	_, _, exists, err := client.Hover(context.Background(), "main.go", 10, 20)
	if err != nil {
		t.Fatalf("unexpected error querying hover: %s", err)
	} else if exists {
		t.Errorf("unexpected hover text")
	}
}

func TestMonikersByPosition(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedQuery := map[string]string{
			"path":      "main.go",
			"line":      "10",
			"character": "20",
		}

		if r.Method != "GET" {
			t.Errorf("unexpected method. want=%s have=%s", "GET", r.Method)
		}
		if r.URL.Path != "/dbs/42/monikersByPosition" {
			t.Errorf("unexpected path. want=%s have=%s", "/dbs/42/monikersByPosition", r.URL.Path)
		}
		if !compareQuery(r.URL.Query(), expectedQuery) {
			t.Errorf("unexpected query. want=%v have=%s", expectedQuery, r.URL.Query().Encode())
		}

		_, _ = w.Write([]byte(`[
			[{
				"kind": "import",
				"scheme": "gomod",
				"identifier": "pad1"
			}],
			[{
				"kind": "import",
				"scheme": "gomod",
				"identifier": "pad2",
				"packageInformationID": "123"
			}, {
				"kind": "export",
				"scheme": "gomod",
				"identifier": "pad2",
				"packageInformationID": "123"
			}]
		]`))
	}))
	defer ts.Close()

	expected := [][]MonikerData{
		{
			{Kind: "import", Scheme: "gomod", Identifier: "pad1"},
		},
		{
			{Kind: "import", Scheme: "gomod", Identifier: "pad2", PackageInformationID: "123"},
			{Kind: "export", Scheme: "gomod", Identifier: "pad2", PackageInformationID: "123"},
		},
	}

	client := &bundleClientImpl{bundleManagerURL: ts.URL, bundleID: 42}
	monikers, err := client.MonikersByPosition(context.Background(), "main.go", 10, 20)
	if err != nil {
		t.Fatalf("unexpected error querying monikers by position: %s", err)
	} else if !reflect.DeepEqual(monikers, expected) {
		t.Errorf("unexpected moniker data. want=%v have=%v", expected, monikers)
	}
}

func TestMonikerResults(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedQuery := map[string]string{
			"modelType":  "definitions",
			"scheme":     "gomod",
			"identifier": "leftpad",
			"take":       "25",
		}

		if r.Method != "GET" {
			t.Errorf("unexpected method. want=%s have=%s", "GET", r.Method)
		}
		if r.URL.Path != "/dbs/42/monikerResults" {
			t.Errorf("unexpected path. want=%s have=%s", "/dbs/42/monikerResults", r.URL.Path)
		}
		if !compareQuery(r.URL.Query(), expectedQuery) {
			t.Errorf("unexpected query. want=%v have=%s", expectedQuery, r.URL.Query().Encode())
		}

		_, _ = w.Write([]byte(`{
			"locations": [
				{"path": "foo.go", "range": {"start": {"line": 1, "character": 2}, "end": {"line": 3, "character": 4}}},
				{"path": "bar.go", "range": {"start": {"line": 5, "character": 6}, "end": {"line": 7, "character": 8}}}
			],
			"count": 5
		}`))

	}))
	defer ts.Close()

	expected := []Location{
		{DumpID: 42, Path: "foo.go", Range: Range{Start: Position{1, 2}, End: Position{3, 4}}},
		{DumpID: 42, Path: "bar.go", Range: Range{Start: Position{5, 6}, End: Position{7, 8}}},
	}

	client := &bundleClientImpl{bundleManagerURL: ts.URL, bundleID: 42}
	locations, count, err := client.MonikerResults(context.Background(), "definitions", "gomod", "leftpad", 0, 25)
	if err != nil {
		t.Fatalf("unexpected error querying moniker results: %s", err)
	}
	if count != 5 {
		t.Errorf("unexpected count. want=%v have=%v", 2, count)
	}
	if !reflect.DeepEqual(locations, expected) {
		t.Errorf("unexpected locations. want=%v have=%v", expected, locations)
	}
}

func TestPackageInformation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedQuery := map[string]string{
			"path":                 "main.go",
			"packageInformationId": "123",
		}

		if r.Method != "GET" {
			t.Errorf("unexpected method. want=%s have=%s", "GET", r.Method)
		}
		if r.URL.Path != "/dbs/42/packageInformation" {
			t.Errorf("unexpected path. want=%s have=%s", "/dbs/42/packageInformation", r.URL.Path)
		}
		if !compareQuery(r.URL.Query(), expectedQuery) {
			t.Errorf("unexpected query. want=%v have=%s", expectedQuery, r.URL.Query().Encode())
		}

		_, _ = w.Write([]byte(`{"name": "leftpad", "version": "0.1.0"}`))
	}))
	defer ts.Close()

	expected := PackageInformationData{
		Name:    "leftpad",
		Version: "0.1.0",
	}

	client := &bundleClientImpl{bundleManagerURL: ts.URL, bundleID: 42}
	packageInformation, err := client.PackageInformation(context.Background(), "main.go", "123")
	if err != nil {
		t.Fatalf("unexpected error querying package information: %s", err)
	} else if !reflect.DeepEqual(packageInformation, expected) {
		t.Errorf("unexpected package information. want=%v have=%v", expected, packageInformation)
	}
}

func compareQuery(query url.Values, expected map[string]string) bool {
	values := map[string]string{}
	for k, v := range query {
		values[k] = v[0]
	}

	return reflect.DeepEqual(values, expected)
}
