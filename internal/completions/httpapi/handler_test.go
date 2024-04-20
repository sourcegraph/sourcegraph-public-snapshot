package httpapi

import (
	"fmt"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestCheckClientCodyIgnoreCompatibility(t *testing.T) {
	mockRequest := func(q url.Values) *http.Request {
		target := "/.api/completions/code" + "?" + q.Encode()
		req := httptest.NewRequest("GET", target, nil)
		return req
	}

	tests := []struct {
		name string
		q    url.Values
		want *clientCodyIgnoreCompatibilityError
	}{
		{
			name: "missing client name and version",
			q:    url.Values{},
			want: &clientCodyIgnoreCompatibilityError{
				reason:     "\"client-name\" query param is required.",
				statusCode: http.StatusNotAcceptable,
			},
		},
		{
			name: "missing client name",
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
			q: url.Values{
				"client-name": []string{"sublime"},
			},
			want: &clientCodyIgnoreCompatibilityError{
				reason:     fmt.Sprintf("please use one of the supported clients: %s, %s.", types.CodyClientVscode, types.CodyClientJetbrains),
				statusCode: http.StatusNotAcceptable,
			},
		},
		{
			name: "version doesn't follow semver spec",
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
			q: url.Values{
				"client-name":    []string{string(types.CodyClientVscode)},
				"client-version": []string{"1.14.0"},
			},
			want: &clientCodyIgnoreCompatibilityError{
				reason:     fmt.Sprintf("Cody for %s version \"1.14.0\" doesn't match version constraint \"%s\"", types.CodyClientVscode, vscodeCodyIgnoreVersionConstraint),
				statusCode: http.StatusNotAcceptable,
			},
		},
		{
			name: "vscode: version matches constraint",
			q: url.Values{
				"client-name":    []string{string(types.CodyClientVscode)},
				"client-version": []string{"1.14.1"},
			},
			want: nil,
		},
		{
			name: "jetbrains: version doesn't match constraint",
			q: url.Values{
				"client-name":    []string{string(types.CodyClientJetbrains)},
				"client-version": []string{"1.14.0"},
			},
			want: &clientCodyIgnoreCompatibilityError{
				reason:     fmt.Sprintf("Cody for %s version \"1.14.0\" doesn't match version constraint \"%s\"", types.CodyClientJetbrains, jetbrainsCodyIgnoreVersionConstraint),
				statusCode: http.StatusNotAcceptable,
			},
		},
		{
			name: "jetbrains: version matches constraint",
			q: url.Values{
				"client-name":    []string{string(types.CodyClientJetbrains)},
				"client-version": []string{"6.0.0"},
			},
			want: nil,
		},
		{
			name: "web: version param not required",
			q: url.Values{
				"client-name": []string{string(types.CodyClientWeb)},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mockRequest(tt.q)
			got := checkClientCodyIgnoreCompatibility(req)
			require.Equal(t, tt.want, got)
		})
	}
}
