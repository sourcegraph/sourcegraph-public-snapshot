package requestclient

import (
	"strings"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
)

func TestClientOriginCountryCode(t *testing.T) {
	for _, tc := range []struct {
		name     string
		client   *Client
		wantCode autogold.Value
	}{
		{
			name: "have trusted geolocation",
			client: &Client{
				wafIPCountryCode: "CA",
			},
			wantCode: autogold.Expect("CA"),
		},
		{
			name: "infer from single ForwardedFor",
			client: &Client{
				ForwardedFor: "93.184.216.34", // ping -c1 example.net
			},
			wantCode: autogold.Expect("GB"),
		},
		{
			name: "infer from multiple ForwardedFor",
			client: &Client{
				ForwardedFor: strings.Join([]string{
					"61.144.235.160", // example from OSS datasets with country code 'CN'
					"93.184.216.34",  // ping -c1 example.net
				}, ","),
			},
			wantCode: autogold.Expect("CN"),
		},
		{
			name: "infer from IP address",
			client: &Client{
				IP: "93.184.216.34", // ping -c1 example.net
			},
			wantCode: autogold.Expect("GB"),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			code, err := tc.client.OriginCountryCode()
			assert.NoError(t, err)
			tc.wantCode.Equal(t, code)

			// Check cached state
			tc.client.countryCodeOnce.Do(func() {
				t.Error("countryCodeOnce should have been called already")
			})
			assert.Equal(t, code, tc.client.countryCode)
			assert.NoError(t, tc.client.countryCodeError)
		})
	}
}
