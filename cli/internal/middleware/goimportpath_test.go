package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/cli/internal/middleware"
)

func TestGoImportPath(t *testing.T) {
	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			path:       "/sourcegraph/sourcegraph/usercontent",
			wantStatus: http.StatusOK,
			wantBody:   `<meta name="go-import" content="example.com/sourcegraph/sourcegraph git https://github.com/sourcegraph/sourcegraph">`,
		},
		{
			path:       "/sourcegraph/srclib/ann",
			wantStatus: http.StatusOK,
			wantBody:   `<meta name="go-import" content="example.com/sourcegraph/srclib git https://github.com/sourcegraph/srclib">`,
		},
		{
			path:       "/sourcegraph/srclib-go",
			wantStatus: http.StatusOK,
			wantBody:   `<meta name="go-import" content="example.com/sourcegraph/srclib-go git https://github.com/sourcegraph/srclib-go">`,
		},
		{
			path:       "/sourcegraph/doesntexist/foobar",
			wantStatus: http.StatusOK,
			wantBody:   `<meta name="go-import" content="example.com/sourcegraph/doesntexist git https://github.com/sourcegraph/doesntexist">`,
		},
		{
			path:       "/sqs/pbtypes",
			wantStatus: http.StatusOK,
			wantBody:   `<meta name="go-import" content="example.com/sqs/pbtypes git https://github.com/sqs/pbtypes">`,
		},
		{
			path:       "/gorilla/mux",
			wantStatus: http.StatusNotFound,
		},
		{
			path:       "/github.com/gorilla/mux",
			wantStatus: http.StatusNotFound,
		},
	}
	for _, test := range tests {
		rw := httptest.NewRecorder()

		req, err := http.NewRequest("GET", test.path+"?go-get=1", nil)
		if err != nil {
			panic(err)
		}

		middleware.SourcegraphComGoGetHandler(nil).ServeHTTP(rw, req)

		if got, want := rw.Code, test.wantStatus; got != want {
			t.Errorf("%s:\ngot  %#v\nwant %#v", test.path, got, want)
		}

		if test.wantBody != "" && !strings.Contains(rw.Body.String(), test.wantBody) {
			t.Errorf("response body %q doesn't contain expected substring %q", rw.Body.String(), test.wantBody)
		}
	}
}
