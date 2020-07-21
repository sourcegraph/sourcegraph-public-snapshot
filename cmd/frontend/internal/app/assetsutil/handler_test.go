package assetsutil

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMount(t *testing.T) {
	m := http.NewServeMux()
	Mount(m)
	ts := httptest.NewServer(m)
	t.Cleanup(ts.Close)

	r, err := http.Get(ts.URL + "/.assets/scripts/app.bundle.js")
	if err != nil {
		t.Fatal(err)
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(body, []byte("<center>cloudflare</center>")) {
		if len(body) > 500 {
			body = append(body[:500], []byte("...")...)
		}
		t.Fatal(string(body))
	}
}
