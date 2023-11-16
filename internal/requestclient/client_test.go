package requestclient

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientOriginCountryCode(t *testing.T) {
	for _, tc := range []struct {
		name     string
		client   *Client
		wantCode string
	}{
		{
			name: "have trusted geolocation",
			client: &Client{
				wafGeolocationCountryCode: "CA",
			},
			wantCode: "CA",
		},
		{
			name: "infer from single ForwardedFor",
			client: &Client{
				ForwardedFor: "93.184.216.34", // ping -c1 example.net
			},
			wantCode: "US",
		},
		{
			name: "infer from multiple ForwardedFor",
			client: &Client{
				ForwardedFor: strings.Join([]string{
					"61.144.235.160", // example from OSS datasets with country code 'CN'
					"93.184.216.34",  // ping -c1 example.net
				}, ","),
			},
			wantCode: "CN",
		},
		{
			name: "infer from IP address",
			client: &Client{
				IP: "93.184.216.34", // ping -c1 example.net
			},
			wantCode: "US",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code, err := tc.client.OriginCountryCode()
			assert.NoError(t, err)
			assert.Equal(t, tc.wantCode, code)

			// Check cached state
			tc.client.countryCodeOnce.Do(func() {
				t.Error("countryCodeOnce should not be called")
			})
			assert.Equal(t, tc.wantCode, tc.client.countryCode)
			assert.NoError(t, tc.client.countryCodeError)
		})
	}
}
