package server

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

type Test struct {
	Name            string
	Request         *http.Request
	ExpectedCode    int
	ExpectedBody    string
	ExpectedHeaders http.Header
}

func TestRequest(t *testing.T) {
	tests := []Test{
		{
			Name:         "Command",
			Request:      httptest.NewRequest("POST", "/exec", strings.NewReader(`{"repo": "github.com/gorilla/mux", "args": ["testcommand"]}`)),
			ExpectedCode: http.StatusOK,
			ExpectedBody: "teststdout",
			ExpectedHeaders: http.Header{
				"Trailer":            {"X-Exec-Error, X-Exec-Exit-Status, X-Exec-Stderr"},
				"X-Exec-Error":       {""},
				"X-Exec-Exit-Status": {"42"},
				"X-Exec-Stderr":      {"teststderr"},
			},
		},
		{
			Name:         "NonexistingRepo",
			Request:      httptest.NewRequest("POST", "/exec", strings.NewReader(`{"repo": "github.com/gorilla/doesnotexist", "args": ["testcommand"]}`)),
			ExpectedCode: http.StatusNotFound,
			ExpectedBody: `{"cloneInProgress":false}`,
		},
		{
			Name:         "UnclonedRepo",
			Request:      httptest.NewRequest("POST", "/exec", strings.NewReader(`{"repo": "github.com/nicksnyder/go-i18n", "args": ["testcommand"]}`)),
			ExpectedCode: http.StatusNotFound,
			ExpectedBody: `{"cloneInProgress":true}`,
		},
		{
			Name:         "Error",
			Request:      httptest.NewRequest("POST", "/exec", strings.NewReader(`{"repo": "github.com/gorilla/mux", "args": ["testerror"]}`)),
			ExpectedCode: http.StatusOK,
			ExpectedHeaders: http.Header{
				"Trailer":            {"X-Exec-Error, X-Exec-Exit-Status, X-Exec-Stderr"},
				"X-Exec-Error":       {"testerror"},
				"X-Exec-Exit-Status": {"0"},
				"X-Exec-Stderr":      {""},
			},
		},
		{
			Name:         "EmptyBody",
			Request:      httptest.NewRequest("POST", "/exec", nil),
			ExpectedCode: http.StatusBadRequest,
			ExpectedBody: `EOF`,
		},
		{
			Name:         "EmptyInput",
			Request:      httptest.NewRequest("POST", "/exec", strings.NewReader("{}")),
			ExpectedCode: http.StatusNotFound,
			ExpectedBody: `{"cloneInProgress":false}`,
		},
	}

	s := &Server{ReposDir: "/testroot"}
	h := s.Handler()

	repoCloned = func(dir string) bool {
		return dir == "/testroot/github.com/gorilla/mux"
	}

	testRepoExists = func(ctx context.Context, origin string, opt *vcs.RemoteOpts) bool {
		return origin == "https://github.com/nicksnyder/go-i18n.git"
	}
	defer func() {
		testRepoExists = nil
	}()

	runCommand = func(cmd *exec.Cmd) (error, int) {
		switch cmd.Args[1] {
		case "testcommand":
			cmd.Stdout.Write([]byte("teststdout"))
			cmd.Stderr.Write([]byte("teststderr"))
			return nil, 42
		case "testerror":
			return errors.New("testerror"), 0
		}
		return nil, 0
	}
	skipCloneForTests = true
	defer func() {
		skipCloneForTests = false
	}()

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			w := httptest.ResponseRecorder{Body: new(bytes.Buffer)}
			h.ServeHTTP(&w, test.Request)

			if w.Code != test.ExpectedCode {
				t.Errorf("wrong status: expected %d, got %d", test.ExpectedCode, w.Code)
			}

			body := strings.TrimSpace(w.Body.String())
			if body != test.ExpectedBody {
				t.Errorf("wrong body: expected %q, got %q", test.ExpectedBody, body)
			}

			for k, v := range test.ExpectedHeaders {
				if got := w.HeaderMap.Get(k); got != v[0] {
					t.Errorf("wrong header %q: expected %q, got %q", k, v[0], got)
				}
			}
		})
	}
}
