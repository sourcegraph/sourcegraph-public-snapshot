package httpapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestCheckClientCodyIgnoreCompatibility(t *testing.T) {
	t.Cleanup(func() { conf.Mock(nil) })

	mockRequest := func(q url.Values) *http.Request {
		target := "/.api/completions/code" + "?" + q.Encode()
		req := httptest.NewRequest("GET", target, nil)
		return req
	}

	ccf := &schema.CodyContextFilters{
		Exclude: []*schema.CodyContextFilterItem{{RepoNamePattern: ".*sensitive.*"}},
	}

	tests := []struct {
		name string
		ccf  *schema.CodyContextFilters
		q    url.Values
		want *clientCodyIgnoreCompatibilityError
	}{
		{
			name: "Cody context filters not defined in the site config",
			q:    url.Values{},
			want: nil,
		},
		{
			name: "missing client name and version",
			ccf:  ccf,
			q:    url.Values{},
			want: &clientCodyIgnoreCompatibilityError{
				reason:     "\"client-name\" query param is required.",
				statusCode: http.StatusNotAcceptable,
			},
		},
		{
			name: "missing client name",
			ccf:  ccf,
			q: url.Values{
				"client-version": []string{"1.1.1"},
			},
			want: &clientCodyIgnoreCompatibilityError{
				reason:     "\"client-name\" query param is required.",
				statusCode: http.StatusNotAcceptable,
			},
		},
		{
			name: "missing client version",
			ccf:  ccf,
			q: url.Values{
				"client-name": []string{string(types.CodyClientVscode)},
			},
			want: &clientCodyIgnoreCompatibilityError{
				reason:     "\"client-version\" query param is required.",
				statusCode: http.StatusNotAcceptable,
			},
		},
		{
			name: "not supported client",
			ccf:  ccf,
			q: url.Values{
				"client-name": []string{"sublime"},
			},
			want: &clientCodyIgnoreCompatibilityError{
				reason:     fmt.Sprintf("please use one of the supported clients: %s, %s, %s.", types.CodyClientVscode, types.CodyClientJetbrains, types.CodyClientWeb),
				statusCode: http.StatusNotAcceptable,
			},
		},
		{
			name: "version doesn't follow semver spec",
			ccf:  ccf,
			q: url.Values{
				"client-name":    []string{string(types.CodyClientVscode)},
				"client-version": []string{"dev"},
			},
			want: &clientCodyIgnoreCompatibilityError{
				reason:     "Cody for vscode version \"dev\" doesn't follow semver spec.",
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "vscode: version doesn't match constraint",
			ccf:  ccf,
			q: url.Values{
				"client-name":    []string{string(types.CodyClientVscode)},
				"client-version": []string{"1.14.0"},
			},
			want: &clientCodyIgnoreCompatibilityError{
				reason:     fmt.Sprintf("Cody for %s version \"1.14.0\" doesn't match version constraint \"1.14.1\"", types.CodyClientVscode),
				statusCode: http.StatusNotAcceptable,
			},
		},
		{
			name: "vscode: version matches constraint",
			ccf:  ccf,
			q: url.Values{
				"client-name":    []string{string(types.CodyClientVscode)},
				"client-version": []string{"1.14.1"},
			},
			want: nil,
		},
		{
			name: "jetbrains: version doesn't match constraint",
			ccf:  ccf,
			q: url.Values{
				"client-name":    []string{string(types.CodyClientJetbrains)},
				"client-version": []string{"1.14.0"},
			},
			want: &clientCodyIgnoreCompatibilityError{
				reason:     fmt.Sprintf("Cody for %s version \"1.14.0\" doesn't match version constraint \"5.5.5\"", types.CodyClientJetbrains),
				statusCode: http.StatusNotAcceptable,
			},
		},
		{
			name: "jetbrains: version matches constraint",
			ccf:  ccf,
			q: url.Values{
				"client-name":    []string{string(types.CodyClientJetbrains)},
				"client-version": []string{"6.0.0"},
			},
			want: nil,
		},
		{
			name: "web: version param not required",
			ccf:  ccf,
			q: url.Values{
				"client-name": []string{string(types.CodyClientWeb)},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ccf != nil {
				conf.Mock(&conf.Unified{
					SiteConfiguration: schema.SiteConfiguration{
						CodyContextFilters: tt.ccf,
					},
				})
			}
			req := mockRequest(tt.q)
			got := checkClientCodyIgnoreCompatibility(req)
			require.Equal(t, tt.want, got)
		})
	}
}
