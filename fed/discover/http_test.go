package discover

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/wellknown"

	"golang.org/x/net/context"
)

func TestDiscoverSiteHTTP(t *testing.T) {
	remoteInfo := func(grpcURLStr, httpURLStr string) *remoteInfo {
		grpcURL, err := url.Parse(grpcURLStr)
		if err != nil {
			t.Fatal(err)
		}
		return &remoteInfo{grpcEndpoint: grpcURL.String()}
	}

	tests := map[string]struct {
		insecureHTTP bool
		httpGet      func(string) (*http.Response, error)
		want         Info
		wantError    string
	}{
		"https success": {
			false,
			fakeHTTPGet(`{"GRPCEndpoint":"http://x","HTTPEndpoint":"http://y"}`, "application/json", true, true, 0),
			remoteInfo("http://x", "http://y"),
			"",
		},
		"https failure": {
			false,
			fakeHTTPGet(``, "application/json", false, false, 0),
			nil,
			(&NotFoundError{Type: "site", Input: "mysite.example.com:1234"}).Error(),
		},
		"http success": {
			true,
			fakeHTTPGet(`{"GRPCEndpoint":"http://x","HTTPEndpoint":"http://y"}`, "application/json", false, true, 0),
			remoteInfo("http://x", "http://y"),
			"",
		},
		"http failure": {
			true,
			fakeHTTPGet(``, "application/json", false, false, http.StatusNotFound),
			nil,
			(&NotFoundError{Type: "site", Input: "mysite.example.com:1234"}).Error(),
		},
		"bad URLs": {
			true,
			fakeHTTPGet(`{"GRPCEndpoint":"x","HTTPEndpoint":"y"}`, "application/json", true, true, 0),
			nil,
			(&NotFoundError{Type: "site", Input: "mysite.example.com:1234"}).Error(),
		},
	}

	defer func() {
		InsecureHTTP = false
	}()

	for label, test := range tests {
		siteCache = nil
		httpGet = test.httpGet
		InsecureHTTP = test.insecureHTTP
		info, err := Site(context.Background(), "mysite.example.com:1234")
		if test.wantError == "" {
			if err != nil {
				t.Errorf("%s: got err == %q, want nil", label, err)
				continue
			}
			if !reflect.DeepEqual(info, test.want) {
				t.Errorf("%s: got Info %#v, want %#v", label, info, test.want)
			}
		} else {
			if err == nil {
				t.Errorf("%s: got err == nil, want %q", label, test.wantError)
				continue
			}
			if !strings.Contains(err.Error(), test.wantError) {
				t.Errorf("%s: got err == %q, want %q", label, err, test.wantError)
			}
		}
	}
}

func TestDiscoverRepoHTTP(t *testing.T) {
	remoteInfo := func(grpcURLStr, httpURLStr string) *remoteInfo {
		grpcURL, err := url.Parse(grpcURLStr)
		if err != nil {
			t.Fatal(err)
		}
		return &remoteInfo{grpcEndpoint: grpcURL.String()}
	}

	tests := map[string]struct {
		insecureHTTP bool
		httpGet      func(string) (*http.Response, error)
		want         Info
		wantError    string
	}{
		"https success": {
			false,
			fakeHTTPGet(`{"GRPCEndpoint":"http://x","HTTPEndpoint":"http://y"}`, "application/json", true, true, 0),
			remoteInfo("http://x", "http://y"),
			"",
		},
		"https failure": {
			false,
			fakeHTTPGet(``, "text/html", false, false, 0),
			nil,
			(&NotFoundError{Type: "repo", Input: "my/repo"}).Error(),
		},
		"http success": {
			true,
			fakeHTTPGet(`{"GRPCEndpoint":"http://x","HTTPEndpoint":"http://y"}`, "application/json", false, true, 0),
			remoteInfo("http://x", "http://y"),
			"",
		},
		"http failure": {
			true,
			fakeHTTPGet(``, "text/html", false, false, http.StatusNotFound),
			nil,
			(&NotFoundError{Type: "repo", Input: "my/repo"}).Error(),
		},
	}

	defer func() {
		InsecureHTTP = false
	}()

	for label, test := range tests {
		siteCache = nil
		httpGet = test.httpGet
		InsecureHTTP = test.insecureHTTP
		info, err := Repo(context.Background(), "my/repo")
		if test.wantError == "" {
			if err != nil {
				t.Errorf("%s: got err == %q, want nil", label, err)
				continue
			}
			if !reflect.DeepEqual(info, test.want) {
				t.Errorf("%s: got Info %#v, want %#v", label, info, test.want)
			}
		} else {
			if err == nil {
				t.Errorf("%s: got err == nil, want %q", label, test.wantError)
				continue
			}
			if !strings.Contains(err.Error(), test.wantError) {
				t.Errorf("%s: got err == %q, want %q", label, err, test.wantError)
			}
		}
	}
}

// fakeHTTPGet is adapted from
// https://github.com/appc/spec/blob/master/discovery/http_test.go.
func fakeHTTPGet(bodyStr, contentType string, httpSuccess bool, httpsSuccess bool, httpErrorCode int) func(uri string) (*http.Response, error) {
	return func(uri string) (*http.Response, error) {
		body := ioutil.NopCloser(strings.NewReader(bodyStr))

		if !strings.HasSuffix(uri, wellknown.ConfigPath) {
			return &http.Response{StatusCode: http.StatusNotFound}, nil
		}

		var resp *http.Response
		switch {
		case strings.HasPrefix(uri, "https://") && httpsSuccess:
			fallthrough
		case strings.HasPrefix(uri, "http://") && httpSuccess:
			resp = &http.Response{
				Status:     "200 OK",
				StatusCode: http.StatusOK,
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
				Header: http.Header{
					"Content-Type": []string{contentType},
				},
				Body: body,
			}
		case httpErrorCode > 0:
			resp = &http.Response{
				Status:     "Error",
				StatusCode: httpErrorCode,
				Proto:      "HTTP/1.1",
				ProtoMajor: 1,
				ProtoMinor: 1,
				Header: http.Header{
					"Content-Type": []string{contentType},
				},
				Body: body,
			}
		default:
			return nil, errors.New("fakeHTTPGet failed as requested")
		}
		return resp, nil
	}
}
