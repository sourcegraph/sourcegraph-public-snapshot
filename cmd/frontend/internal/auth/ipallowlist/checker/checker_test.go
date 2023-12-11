package checker

import (
	"net/http"
	"testing"
	"text/template"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func Test_isAuthorized(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		enabled          bool
		reqRemoteAddr    string
		reqHeaders       map[string][]string
		userIps          []string
		clientIps        []string
		trustedClientIps []string
		userHeaders      []string
		tmpl             *template.Template
		wantErr          autogold.Value
	}{
		{
			name:          "authorized - it's disabled",
			reqRemoteAddr: "127.0.0.1:1234",
		},
		{
			name:          "authorized - ip allow list is empty",
			enabled:       true,
			reqRemoteAddr: "10.0.0.1:1234",
		},
		{
			name:             "authorized - trusted ip allowlist bypass",
			enabled:          true,
			reqRemoteAddr:    "10.0.0.1:1234",
			trustedClientIps: []string{"10.0.0.1/25"},
			clientIps:        []string{"100.100.100.0/25"},
			userIps:          []string{"100.100.101.0/25"},
		},
		{
			name:          "unauthorized - ip allow list does not contain client ip",
			enabled:       true,
			reqRemoteAddr: "10.0.0.1:1234",
			userIps:       []string{"100.100.100.0"},
			wantErr:       autogold.Expect(`You are not allowed to access this Sourcegraph instance: "10.0.0.1"`),
		},
		{
			name:          "unauthorized - ip allow list does not contain client ip",
			enabled:       true,
			reqRemoteAddr: "10.0.0.1:1234",
			userIps:       []string{"100.100.100.0/25"},
			wantErr:       autogold.Expect(`You are not allowed to access this Sourcegraph instance: "10.0.0.1"`),
		},
		{
			name:          "unauthorized - ip allow list does not contain client ip from x-forwarded-for headers",
			enabled:       true,
			reqRemoteAddr: "10.0.0.1:1234",
			reqHeaders: map[string][]string{
				"X-Forwarded-For": {"1.2.3.4", "4.5.6.7"},
			},
			userHeaders: []string{"X-Forwarded-For"},
			userIps:     []string{"100.100.100.0/25"},
			wantErr:     autogold.Expect(`You are not allowed to access this Sourcegraph instance: "1.2.3.4"`),
		},
		{
			name:          "unauthorized - ip allow list does not contain client ip from x-forwarded-for headers but fallback to remote add",
			enabled:       true,
			reqRemoteAddr: "10.0.0.1:1234",
			reqHeaders: map[string][]string{
				"CF-Connecting-IP": {"1.2.3.4", "4.5.6.7"},
			},
			userHeaders: []string{"X-Forwarded-For"},
			userIps:     []string{"100.100.100.0/25"},
			wantErr:     autogold.Expect(`You are not allowed to access this Sourcegraph instance: "10.0.0.1"`),
		},
		{
			name:          "authorized - ip allow list contains client ip from secondaryheaders",
			enabled:       true,
			reqRemoteAddr: "10.0.0.1:1234",
			reqHeaders: map[string][]string{
				"CF-Connecting-IP": {"100.100.100.0"},
			},
			userHeaders: []string{"X-Forwarded-For", "CF-Connecting-IP"},
			userIps:     []string{"100.100.100.0/25"},
		},
		{
			name:          "authorized - ip allow list contains client ip with lower case user headers",
			enabled:       true,
			reqRemoteAddr: "10.0.0.1:1234",
			reqHeaders: map[string][]string{
				"X-Forwarded-For": {"100.100.100.0"},
			},
			userHeaders: []string{"x-forwarded-for"},
			userIps:     []string{"100.100.100.0/25"},
		},
		{
			name:          "authorized - both user ip and client ip are in the allow list",
			enabled:       true,
			reqRemoteAddr: "10.0.0.1:1234",
			reqHeaders: map[string][]string{
				"X-Forwarded-For": {"100.100.100.0"},
			},
			userHeaders: []string{"x-forwarded-for"},
			userIps:     []string{"100.100.100.0/25"},
			clientIps:   []string{"10.0.0.1/25"},
		},
		{
			name:          "unauthorized - with custom err msg templ",
			enabled:       true,
			reqRemoteAddr: "10.0.0.1:1234",
			userIps:       []string{"100.100.100.0"},
			tmpl:          template.Must(template.New("").Parse("my IP is {{.UserIP}}; my custom error message is {{.Error}}")),
			wantErr:       autogold.Expect(`my IP is 10.0.0.1; my custom error message is You are not allowed to access this Sourcegraph instance: "10.0.0.1"`),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userIps, userRanges, err := parseIPs(tc.userIps)
			require.NoError(t, err)
			clientIps, clientRanges, err := parseIPs(tc.clientIps)
			require.NoError(t, err)
			trustedClientIps, trustedClientRanges, err := parseIPs(append(tc.trustedClientIps, defaultTrustedClientIpAllowlist...))
			require.NoError(t, err)

			r := &http.Request{
				Header: make(http.Header),
			}
			if tc.reqHeaders != nil {
				for k, v := range tc.reqHeaders {
					for _, vv := range v {
						r.Header.Add(k, vv)
					}
				}
			}
			if tc.reqRemoteAddr != "" {
				r.RemoteAddr = tc.reqRemoteAddr
			}

			err = isAuthorized(
				config{
					authorizedUserIps:      userIps,
					authorizedUserRanges:   userRanges,
					authorizedClientIps:    clientIps,
					authorizedClientRanges: clientRanges,
					trustedClientIps:       trustedClientIps,
					trustedClientRanges:    trustedClientRanges,
					userHeaders:            tc.userHeaders,
					errorMessageTmpl:       tc.tmpl,
				},
				r,
			)
			if tc.wantErr != nil {
				require.Error(t, err)
				tc.wantErr.Equal(t, err.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func Test_getConfig(t *testing.T) {
	tests := []struct {
		name    string
		mock    *schema.AuthAllowedIpAddress
		want    bool
		wantErr autogold.Value
	}{
		{
			name: "empty",
		},
		{
			name: "empty - disabled explicitly",
			mock: &schema.AuthAllowedIpAddress{
				Enabled: false,
			},
		},
		{
			name: "valid",
			mock: &schema.AuthAllowedIpAddress{
				Enabled:       true,
				UserIpAddress: []string{"100.100.100.0/25", "1.1.1.1"},
			},
			want: true,
		},
		{
			name: "valid with warning",
			mock: &schema.AuthAllowedIpAddress{
				Enabled:       true,
				UserIpAddress: []string{"invalid-ip"},
			},
			want:    true,
			wantErr: autogold.Expect(`invalid ip addr: "invalid-ip": invalid CIDR address: invalid-ip`),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conf.Mock(&conf.Unified{
				SiteConfiguration: schema.SiteConfiguration{
					AuthAllowedIpAddress: tc.mock,
				},
			})

			cfg := getConfig()
			if tc.wantErr != nil {
				require.Error(t, cfg.err)
				tc.wantErr.Equal(t, cfg.err.Error())
				return
			}
			if tc.want {
				require.NotNil(t, cfg)
				cfg.err = nil // prevent autogold from checking err
				autogold.ExpectFile(t, cfg)
			} else {
				require.Nil(t, cfg)
				return
			}
			require.NoError(t, cfg.err)
		})
	}
}
