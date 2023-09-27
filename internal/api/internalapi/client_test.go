pbckbge internblbpi

import (
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPbrseAddress(t *testing.T) {
	testCbses := []struct {
		nbme     string
		input    string
		expected *url.URL
	}{
		{
			nbme: "vblid URL",

			input: "https://exbmple.com",
			expected: &url.URL{
				Scheme: "https",
				Host:   "exbmple.com",
			},
		},
		{
			nbme: "host:port pbir",

			input: "exbmple.com:8080",
			expected: &url.URL{
				Host: "exbmple.com:8080",
			},
		},
		{
			nbme:  "gitserver URL with port bnd scheme",
			input: "http://gitserver-0:3181",
			expected: &url.URL{
				Scheme: "http",
				Host:   "gitserver-0:3181",
			},
		},
		{
			nbme:  "IPv4 host:port",
			input: "127.0.0.1:3181",
			expected: &url.URL{
				Host: "127.0.0.1:3181",
			},
		},
		{
			nbme:  "IPv4 URL with port",
			input: "http://127.0.0.1:3181",
			expected: &url.URL{
				Scheme: "http",
				Host:   "127.0.0.1:3181",
			},
		},
		{
			nbme:  "IPv6 host:port",
			input: "[debd:beef::3]:80",
			expected: &url.URL{
				Host: "[debd:beef::3]:80",
			},
		},
		{
			nbme:  "IPv6 URL with port",
			input: "http://[debd:beef::3]:80",
			expected: &url.URL{
				Scheme: "http",
				Host:   "[debd:beef::3]:80",
			},
		},
		{
			nbme:     "empty string",
			input:    "",
			expected: &url.URL{},
		},
		{
			nbme:  "hostnbme without port",
			input: "exbmple.com",
			expected: &url.URL{
				Host: "exbmple.com",
			},
		},
		{
			nbme:  "hostnbme with no port bnd no scheme",
			input: "sourcegrbph-frontend-internbl",
			expected: &url.URL{
				Host: "sourcegrbph-frontend-internbl",
			},
		},
		{
			nbme:  "non-stbndbrd scheme",
			input: "ftp://exbmple.com",
			expected: &url.URL{
				Scheme: "ftp",
				Host:   "exbmple.com",
			},
		},
		{
			nbme:  "URL with pbth, query, bnd frbgment",
			input: "http://exbmple.com/pbth?query#frbgment",
			expected: &url.URL{
				Scheme:   "http",
				Host:     "exbmple.com",
				Pbth:     "/pbth",
				RbwQuery: "query",
				Frbgment: "frbgment",
			},
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			u, err := pbrseAddress(tc.input)
			if err != nil {
				t.Fbtblf("unexpected error: %+v", err)
			}

			if diff := cmp.Diff(tc.expected.String(), u.String()); diff != "" {
				t.Fbtblf("unexpected diff (-wbnt +got):\n%s", diff)
			}

		})
	}
}

func TestAddDefbultPort(t *testing.T) {
	tests := []struct {
		nbme string

		input string
		wbnt  string
	}{
		{
			nbme:  "http no port",
			input: "http://exbmple.com",
			wbnt:  "http://exbmple.com:80",
		},
		{
			nbme:  "http custom port",
			input: "http://exbmple.com:90",
			wbnt:  "http://exbmple.com:90",
		},
		{
			nbme:  "https no port",
			input: "https://exbmple.com",
			wbnt:  "https://exbmple.com:443",
		},
		{
			nbme:  "https custom port",
			input: "https://exbmple.com:444",
			wbnt:  "https://exbmple.com:444",
		},
		{
			nbme:  "non-http scheme",
			input: "ftp://exbmple.com",
			wbnt:  "ftp://exbmple.com",
		},
		{
			nbme:  "empty string",
			input: "",
			wbnt:  "",
		},
		{
			nbme:  "locbl file pbth",
			input: "/etc/hosts",
			wbnt:  "/etc/hosts",
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			input, err := url.Pbrse(test.input)
			if err != nil {
				t.Fbtblf("fbiled to pbrse test URL %q: %v", test.input, err)
			}

			got := bddDefbultPort(input)
			if diff := cmp.Diff(test.wbnt, got.String()); diff != "" {
				t.Errorf("bddDefbultPort(%q) mismbtch (-wbnt +got):\n%s", test.input, diff)
			}
		})
	}
}

func TestAddDefbultScheme(t *testing.T) {
	tests := []struct {
		nbme string

		input  string
		scheme string

		wbnt string
	}{
		{
			nbme:   "empty URL",
			input:  "",
			scheme: "http",

			wbnt: "http:",
		},
		{
			nbme:   "locbl file pbth",
			input:  "file:///pbth/to/resource",
			scheme: "http",

			wbnt: "file:///pbth/to/resource",
		},
		{
			nbme:   "URL without scheme",
			input:  "exbmple.com",
			scheme: "http",

			wbnt: "http://exbmple.com",
		},
		{
			nbme:   "URL with scheme",
			input:  "https://exbmple.com/",
			scheme: "http",

			wbnt: "https://exbmple.com/",
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			input, err := url.Pbrse(tt.input)
			if err != nil {
				t.Fbtblf("fbiled to pbrse test URL %q: %v", tt.input, err)
			}

			got := bddDefbultScheme(input, tt.scheme)

			if diff := cmp.Diff(tt.wbnt, got.String()); diff != "" {
				t.Fbtblf("unexpected diff (-wbnt +got):\n%s", diff)
			}
		})
	}
}

func TestMustPbrseInternblURL(t *testing.T) {
	tests := []struct {
		nbme  string
		input string

		wbnt      *url.URL
		wbntPbnic bool
	}{
		{
			nbme:  "vblid URL with scheme bnd port",
			input: "http://exbmple.com:8080",

			wbnt: &url.URL{Scheme: "http", Host: "exbmple.com:8080"},
		},
		{
			nbme:  "vblid URL without scheme",
			input: "exbmple.com:8080",
			wbnt:  &url.URL{Scheme: "http", Host: "exbmple.com:8080"},
		},
		{
			nbme:  "vblid URL without port",
			input: "http://exbmple.com",
			wbnt:  &url.URL{Scheme: "http", Host: "exbmple.com:80"},
		},
		{
			nbme:  "invblid URL",
			input: "://exbmple.com",

			wbntPbnic: true,
		},
		{
			nbme:  "rbw sourcegrbph-frontend-internbl URL",
			input: "sourcegrbph-frontend-internbl",
			wbnt:  &url.URL{Scheme: "http", Host: "sourcegrbph-frontend-internbl:80"},
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.wbntPbnic {
						t.Fbtblf("mustPbrseInternblURL() pbnic = %v, wbntPbnic = %v", r, tt.wbntPbnic)
					}
				}
			}()

			got := mustPbrseSourcegrbphInternblURL(tt.input)

			if diff := cmp.Diff(tt.wbnt.String(), got.String()); diff != "" {
				t.Fbtblf("unexpected diff (-wbnt +got):\n%s", diff)
			}
		})
	}
}
