package vcsclient

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"
)

// Uses HTTP client testing code adapted from github.com/google/go-github.

var (
	// mux is the HTTP request multiplexer used with the test server.
	mux *http.ServeMux

	// client is the vcsstore client being tested.
	vcsclient *Client

	// server is a test HTTP server used to provide mock API responses.
	server *httptest.Server
)

// setup sets up a test HTTP server along with a Client that is
// configured to talk to that test server. Tests should register handlers on
// mux which provide mock responses for the API method being tested.
func setup() {
	// test server
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	// vcsstore client configured to use test server
	url, _ := url.Parse(server.URL)
	vcsclient = New(url, nil)
}

// teardown closes the test HTTP server.
func teardown() {
	server.Close()
}

func urlPath(t *testing.T, routeName string, repo *repository, routeVars map[string]string) string {
	url, err := repo.url(routeName, routeVars, nil)
	if err != nil {
		t.Fatalf("Error constructing URL path for route %q with vars %+v: %s", routeName, routeVars, err)
	}
	return "/" + url.Path
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		panic("writeJSON: " + err.Error())
	}
}

func testMethod(t *testing.T, r *http.Request, want string) {
	if want != r.Method {
		t.Errorf("Request method = %v, want %v", r.Method, want)
	}
}

type values map[string]string

func testFormValues(t *testing.T, r *http.Request, values values) {
	want := url.Values{}
	for k, v := range values {
		want.Add(k, v)
	}

	r.ParseForm()
	if !reflect.DeepEqual(want, r.Form) {
		t.Errorf("Request parameters = %v, want %v", r.Form, want)
	}
}

func testBody(t *testing.T, r *http.Request, want string) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Errorf("Unable to read body")
	}
	str := string(b)
	if want != str {
		t.Errorf("Body = %s, want: %s", str, want)
	}
}

func normalizeTime(tm *time.Time) {
	*tm = tm.In(time.UTC)
}
