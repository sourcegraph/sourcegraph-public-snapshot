package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/internal/middleware"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httptestutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
)

func TestGoImportPath(t *testing.T) {
	_, mock := httptestutil.NewTest(nil)
	defer httptestutil.ResetGlobals()
	mock.Repos.Get_ = func(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		switch repo.URI {
		case "sourcegraph/sourcegraph": // Hosted repo.
			return &sourcegraph.Repo{}, nil
		case "sourcegraph/srclib-go": // Mirror repo.
			return &sourcegraph.Repo{Mirror: true}, nil
		default:
			return nil, &store.RepoNotFoundError{Repo: repo.URI}
		}
	}
	mock.Ctx = conf.WithURL(mock.Ctx, &url.URL{Scheme: "https", Host: "sourcegraph.com", Path: "/"})

	tests := []struct {
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			path:       "/sourcegraph/sourcegraph/usercontent",
			wantStatus: http.StatusOK,
			wantBody:   `<meta name="go-import" content="sourcegraph.com/sourcegraph/sourcegraph git https://sourcegraph.com/sourcegraph/sourcegraph">`,
		},
		{
			path:       "/sourcegraph/srclib/ann",
			wantStatus: http.StatusOK,
			wantBody:   `<meta name="go-import" content="sourcegraph.com/sourcegraph/srclib git https://github.com/sourcegraph/srclib">`,
		},
		{
			path:       "/sourcegraph/srclib-go",
			wantStatus: http.StatusOK,
			wantBody:   `<meta name="go-import" content="sourcegraph.com/sourcegraph/srclib-go git https://github.com/sourcegraph/srclib-go">`,
		},
		{
			path:       "/sourcegraph/doesntexist/foobar",
			wantStatus: http.StatusOK,
			wantBody:   `<meta name="go-import" content="sourcegraph.com/sourcegraph/doesntexist git https://github.com/sourcegraph/doesntexist">`,
		},
		{
			path:       "/sqs/pbtypes",
			wantStatus: http.StatusOK,
			wantBody:   `<meta name="go-import" content="sourcegraph.com/sqs/pbtypes git https://github.com/sqs/pbtypes">`,
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
		httpctx.SetForRequest(req, mock.Ctx)

		middleware.SourcegraphComGoGetHandler(nil).ServeHTTP(rw, req)

		if got, want := rw.Code, test.wantStatus; got != want {
			t.Errorf("%s:\ngot  %#v\nwant %#v", test.path, got, want)
		}

		if test.wantBody != "" && !strings.Contains(rw.Body.String(), test.wantBody) {
			t.Errorf("response body %q doesn't contain expected substring %q", rw.Body.String(), test.wantBody)
		}
	}
}

// Test the following behavior inside sourcegraphComGoGetHandler:
//
// 	If there are 3 path elements, e.g., "/alpha/beta/gamma", start by checking
// 	repo path "alpha", then "alpha/beta", and finally "alpha/beta/gamma".
func TestGoImportPath_repoCheckSequence(t *testing.T) {
	_, mock := httptestutil.NewTest(nil)
	defer httptestutil.ResetGlobals()
	var attemptedRepoURIs []string
	mock.Repos.Get_ = func(ctx context.Context, repo *sourcegraph.RepoSpec) (*sourcegraph.Repo, error) {
		attemptedRepoURIs = append(attemptedRepoURIs, repo.URI)
		return nil, &store.RepoNotFoundError{Repo: repo.URI}
	}

	rw := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "/alpha/beta/gamma?go-get=1", nil)
	if err != nil {
		panic(err)
	}
	httpctx.SetForRequest(req, mock.Ctx)

	middleware.SourcegraphComGoGetHandler(nil).ServeHTTP(rw, req)

	got := attemptedRepoURIs
	want := []string{"alpha", "alpha/beta", "alpha/beta/gamma"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("\ngot  %#v\nwant %#v", got, want)
	}

	if got, want := rw.Code, http.StatusNotFound; got != want {
		t.Errorf("\ngot  %#v\nwant %#v", got, want)
	}
}
