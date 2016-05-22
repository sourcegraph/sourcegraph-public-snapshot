// +build exectest

// Package integration_test starts a single test server during the duration of all tests,
// allowing there to be many quick integration checks for easily checkable things that can otherwise regress.
//
// Since each test reuses the same server, all tests should be idempotent and order-independent.
package integration_test

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/testserver"
)

// h is the host:port address of the test server.
// There will be only one server available during all tests in this package, so each test
// must be idempotent.
var h *url.URL

func TestMain(m *testing.M) {
	var code int
	defer func() {
		switch err := recover(); err {
		case nil:
			os.Exit(code)
		default:
			log.Println(err)
			os.Exit(1)
		}
	}()

	s, _ := testserver.NewUnstartedServer()
	s.Config.Serve.NoWorker = true
	if err := s.Start(); err != nil {
		panic(err)
	}
	defer s.Close() // This will kill the started src process. It's important that this func runs before os.Exit, otherwise there will be a runaway zombie process.
	var err error
	h, err = url.Parse(s.Config.Endpoint.URL)
	if err != nil {
		panic(err)
	}

	code = m.Run()
}

func TestRobotsTxt(t *testing.T) {
	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("no redirects expected") },
	}
	resp, err := client.Get(u("/robots.txt"))
	if err != nil {
		t.Fatal(err)
	}
	n, err := io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	err = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusOK; resp.StatusCode != want {
		t.Errorf("got HTTP status %v, want %v\n", resp.Status, want)
	}
	if got, want := resp.Header.Get("Content-Type"), "text/plain"; got != want {
		t.Errorf("got Content-Type %v, want %v\n", got, want)
	}
	if n == 0 {
		t.Error("got empty body, want non-empty")
	}
}

func TestFavicon(t *testing.T) {
	resp, err := http.Get(u("/favicon.ico"))
	if err != nil {
		t.Fatal(err)
	}
	n, err := io.Copy(ioutil.Discard, resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	err = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if want := http.StatusOK; resp.StatusCode != want {
		t.Errorf("got HTTP status %v, want %v\n", resp.Status, want)
	}
	if got, want := resp.Header.Get("Content-Type"), "image/png"; got != want {
		t.Errorf("got Content-Type %v, want %v\n", got, want)
	}
	if n == 0 {
		t.Error("got empty body, want non-empty")
	}
}

// u converts an absolute path to a full url, including scheme, host and port of test server.
// For example, "/robots.txt" will become "http://localhost:10001/robots.txt".
func u(path string) string {
	return h.ResolveReference(&url.URL{Path: path}).String()
}
