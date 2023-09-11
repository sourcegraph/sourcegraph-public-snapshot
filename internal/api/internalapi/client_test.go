package internalapi

import (
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseAddress(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected *url.URL
	}{
		{
			name: "valid URL",

			input: "https://example.com",
			expected: &url.URL{
				Scheme: "https",
				Host:   "example.com",
			},
		},
		{
			name: "host:port pair",

			input: "example.com:8080",
			expected: &url.URL{
				Host: "example.com:8080",
			},
		},
		{
			name:  "gitserver URL with port and scheme",
			input: "http://gitserver-0:3181",
			expected: &url.URL{
				Scheme: "http",
				Host:   "gitserver-0:3181",
			},
		},
		{
			name:  "IPv4 host:port",
			input: "127.0.0.1:3181",
			expected: &url.URL{
				Host: "127.0.0.1:3181",
			},
		},
		{
			name:  "IPv4 URL with port",
			input: "http://127.0.0.1:3181",
			expected: &url.URL{
				Scheme: "http",
				Host:   "127.0.0.1:3181",
			},
		},
		{
			name:  "IPv6 host:port",
			input: "[dead:beef::3]:80",
			expected: &url.URL{
				Host: "[dead:beef::3]:80",
			},
		},
		{
			name:  "IPv6 URL with port",
			input: "http://[dead:beef::3]:80",
			expected: &url.URL{
				Scheme: "http",
				Host:   "[dead:beef::3]:80",
			},
		},
		{
			name:     "empty string",
			input:    "",
			expected: &url.URL{},
		},
		{
			name:  "hostname without port",
			input: "example.com",
			expected: &url.URL{
				Host: "example.com",
			},
		},
		{
			name:  "hostname with no port and no scheme",
			input: "sourcegraph-frontend-internal",
			expected: &url.URL{
				Host: "sourcegraph-frontend-internal",
			},
		},
		{
			name:  "non-standard scheme",
			input: "ftp://example.com",
			expected: &url.URL{
				Scheme: "ftp",
				Host:   "example.com",
			},
		},
		{
			name:  "URL with path, query, and fragment",
			input: "http://example.com/path?query#fragment",
			expected: &url.URL{
				Scheme:   "http",
				Host:     "example.com",
				Path:     "/path",
				RawQuery: "query",
				Fragment: "fragment",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u, err := parseAddress(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %+v", err)
			}

			if diff := cmp.Diff(tc.expected.String(), u.String()); diff != "" {
				t.Fatalf("unexpected diff (-want +got):\n%s", diff)
			}

		})
	}
}

func TestAddDefaultPort(t *testing.T) {
	tests := []struct {
		name string

		input string
		want  string
	}{
		{
			name:  "http no port",
			input: "http://example.com",
			want:  "http://example.com:80",
		},
		{
			name:  "http custom port",
			input: "http://example.com:90",
			want:  "http://example.com:90",
		},
		{
			name:  "https no port",
			input: "https://example.com",
			want:  "https://example.com:443",
		},
		{
			name:  "https custom port",
			input: "https://example.com:444",
			want:  "https://example.com:444",
		},
		{
			name:  "non-http scheme",
			input: "ftp://example.com",
			want:  "ftp://example.com",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "local file path",
			input: "/etc/hosts",
			want:  "/etc/hosts",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input, err := url.Parse(test.input)
			if err != nil {
				t.Fatalf("failed to parse test URL %q: %v", test.input, err)
			}

			got := addDefaultPort(input)
			if diff := cmp.Diff(test.want, got.String()); diff != "" {
				t.Errorf("addDefaultPort(%q) mismatch (-want +got):\n%s", test.input, diff)
			}
		})
	}
}

func TestAddDefaultScheme(t *testing.T) {
	tests := []struct {
		name string

		input  string
		scheme string

		want string
	}{
		{
			name:   "empty URL",
			input:  "",
			scheme: "http",

			want: "http:",
		},
		{
			name:   "local file path",
			input:  "file:///path/to/resource",
			scheme: "http",

			want: "file:///path/to/resource",
		},
		{
			name:   "URL without scheme",
			input:  "example.com",
			scheme: "http",

			want: "http://example.com",
		},
		{
			name:   "URL with scheme",
			input:  "https://example.com/",
			scheme: "http",

			want: "https://example.com/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := url.Parse(tt.input)
			if err != nil {
				t.Fatalf("failed to parse test URL %q: %v", tt.input, err)
			}

			got := addDefaultScheme(input, tt.scheme)

			if diff := cmp.Diff(tt.want, got.String()); diff != "" {
				t.Fatalf("unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMustParseInternalURL(t *testing.T) {
	tests := []struct {
		name  string
		input string

		want      *url.URL
		wantPanic bool
	}{
		{
			name:  "valid URL with scheme and port",
			input: "http://example.com:8080",

			want: &url.URL{Scheme: "http", Host: "example.com:8080"},
		},
		{
			name:  "valid URL without scheme",
			input: "example.com:8080",
			want:  &url.URL{Scheme: "http", Host: "example.com:8080"},
		},
		{
			name:  "valid URL without port",
			input: "http://example.com",
			want:  &url.URL{Scheme: "http", Host: "example.com:80"},
		},
		{
			name:  "invalid URL",
			input: "://example.com",

			wantPanic: true,
		},
		{
			name:  "raw sourcegraph-frontend-internal URL",
			input: "sourcegraph-frontend-internal",
			want:  &url.URL{Scheme: "http", Host: "sourcegraph-frontend-internal:80"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantPanic {
						t.Fatalf("mustParseInternalURL() panic = %v, wantPanic = %v", r, tt.wantPanic)
					}
				}
			}()

			got := mustParseSourcegraphInternalURL(tt.input)

			if diff := cmp.Diff(tt.want.String(), got.String()); diff != "" {
				t.Fatalf("unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}
