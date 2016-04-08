package app_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/apptest"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/ui"
)

func TestServeUI(t *testing.T) {
	tests := []struct {
		debugMode      bool
		renderResult   ui.RenderResult
		wantStatusCode int
		wantLocation   string
		want           func(t *testing.T, label, body string)
	}{
		{
			renderResult: ui.RenderResult{
				StatusCode:  http.StatusOK,
				Body:        "abc",
				ContentType: "text/html",
			},
			wantStatusCode: http.StatusOK,
			want: func(t *testing.T, label, body string) {
				if want := "abc"; !strings.Contains(body, want) {
					t.Errorf("%s: body does not contain %q\n\n%s", label, want, body)
				}
			},
		},
		{
			renderResult: ui.RenderResult{
				StatusCode:  http.StatusNotFound,
				Body:        "abc",
				ContentType: "text/html",
			},
			wantStatusCode: http.StatusNotFound,
			want: func(t *testing.T, label, body string) {
				if want := "abc"; !strings.Contains(body, want) {
					t.Errorf("%s: body does not contain %q\n\n%s", label, want, body)
				}
			},
		},
		{
			renderResult: ui.RenderResult{
				StatusCode:       http.StatusMovedPermanently,
				RedirectLocation: "/abc",
			},
			wantStatusCode: http.StatusMovedPermanently,
			wantLocation:   "/abc",
		},
		{
			debugMode: false,
			renderResult: ui.RenderResult{
				StatusCode: http.StatusInternalServerError,
				Error:      "abc",
			},
			wantStatusCode: http.StatusInternalServerError,
			want: func(t *testing.T, label, body string) {
				if dontWant := "abc"; strings.Contains(body, dontWant) {
					t.Errorf("%s: body contains error string %q in non-debug mode", label, dontWant)
				}
			},
		},
		{
			debugMode: true,
			renderResult: ui.RenderResult{
				StatusCode: http.StatusInternalServerError,
				Error:      "abc",
			},
			wantStatusCode: http.StatusInternalServerError,
			want: func(t *testing.T, label, body string) {
				if want := "abc"; !strings.Contains(body, want) {
					t.Errorf("%s: body does not contain error string %q in debug mode\n\n%s", label, want, body)
				}
			},
		},
	}

	for _, test := range tests {
		func() {
			c, _ := apptest.New()

			orig := ui.RenderRouter
			defer func() {
				ui.RenderRouter = orig
			}()
			ui.RenderRouter = func(ctx context.Context, req *http.Request, extraProps map[string]interface{}) (*ui.RenderResult, error) {
				return &test.renderResult, nil
			}

			if test.debugMode {
				orig := os.Getenv("DEBUG")
				os.Setenv("DEBUG", "t")
				defer os.Setenv("DEBUG", orig)
			}

			resp, err := c.GetNoFollowRedirects("/")
			if err != nil {
				t.Fatalf("%v: Get: %s", test.renderResult, err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != test.wantStatusCode {
				t.Errorf("%v: got HTTP status %d, want %d", test.renderResult, resp.StatusCode, test.wantStatusCode)
			}
			if resp.Header.Get("location") != test.wantLocation {
				t.Errorf("%v: got location %q, want %q", test.renderResult, resp.Header.Get("location"), test.wantLocation)
			}
			if test.want != nil {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					t.Fatal(err)
				}
				test.want(t, fmt.Sprintf("%v", test.renderResult), string(body))
			}
		}()
	}
}
