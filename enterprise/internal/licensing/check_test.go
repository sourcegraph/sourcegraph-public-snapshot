package licensing

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_maxDelayOrZero(t *testing.T) {
	now := time.Now()

	tests := map[string]struct {
		before time.Time
		after  time.Time
		delay  time.Duration
		want   time.Duration
	}{
		"returns 0 if before is zero": {
			before: time.Time{},
			after:  now,
			delay:  12 * time.Hour,
			want:   0,
		},
		"returns 0 if after is zero": {
			before: now.Add(-1 * time.Hour),
			after:  time.Time{},
			delay:  12 * time.Hour,
			want:   0,
		},
		"returns 0 if before is in the future": {
			before: now.Add(1 * time.Hour),
			after:  now,
			delay:  12 * time.Hour,
			want:   0,
		},
		"returns 0 if before is in the past but more than 12 hours ago": {
			before: now.Add(-13 * time.Hour),
			after:  now,
			delay:  12 * time.Hour,
			want:   0,
		},
		"returns 0 hours if before is 12 hours ago": {
			before: now.Add(-12 * time.Hour),
			after:  now,
			delay:  12 * time.Hour,
			want:   0,
		},
		"returns 10h if before is 2 hours ago": {
			before: now.Add(-2 * time.Hour),
			after:  now,
			delay:  12 * time.Hour,
			want:   10 * time.Hour,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := maxDelayOrZero(test.before, now, 12*time.Hour)
			require.Equal(t, test.want, got.Round(time.Hour))
		})
	}
}

func Test_checkLicenseValidity(t *testing.T) {
	tests := map[string]struct {
		response []byte
		status   int
		want     bool
		err      bool
	}{
		"returns error if unable to make a request to license server": {
			response: []byte(`{"error": "some error"}`),
			status:   http.StatusInternalServerError,
			want:     false,
			err:      true,
		},
		"returns error if got error": {
			response: []byte(`{"error": "some error"}`),
			status:   http.StatusOK,
			want:     false,
			err:      true,
		},
		`returns correct result for "true"`: {
			response: []byte(`{"data": {"is_valid": true}}`),
			status:   http.StatusOK,
			want:     true,
		},
		`returns correct result for "false"`: {
			response: []byte(`{"data": {"is_valid": false, "reason": "some reason"}}`),
			status:   http.StatusOK,
			want:     false,
		},
	}

	siteID := "some-site-id"
	licenseKey := "test-license-key"
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			doer := &mockDoer{
				status:   test.status,
				response: test.response,
			}
			got, err := checkLicenseValidity(context.Background(), doer, siteID, licenseKey)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// check doer called
			require.True(t, doer.DoCalled)

			// check called with correct method, url and content type
			require.Equal(t, "https://sourcegraph.com/.api/license/check", doer.Request.URL.String())
			require.Equal(t, "POST", doer.Request.Method)
			require.Equal(t, "application/json", doer.Request.Header.Get("Content-Type"))

			// check called with correct authorization header
			tokenHexEncoded := hex.EncodeToString(GenerateHashedLicenseKeyAccessToken(licenseKey))
			require.Equal(t, "Bearer "+tokenHexEncoded, doer.Request.Header.Get("Authorization"))

			// check called with correct request body
			var body struct {
				SiteID string `json:"siteID"`
			}
			resBody, err := io.ReadAll(doer.Request.Body)
			require.NoError(t, err)
			json.Unmarshal([]byte(resBody), &body)
			require.Equal(t, siteID, body.SiteID)

			// check result
			require.Equal(t, test.want, got)
		})
	}
}

var strPtr = func(s string) *string { return &s }

type mockDoer struct {
	DoCalled bool
	Request  *http.Request

	status   int
	response []byte
}

func (d *mockDoer) Do(req *http.Request) (*http.Response, error) {
	d.DoCalled = true
	d.Request = req

	return &http.Response{
		StatusCode: d.status,
		Body:       io.NopCloser(bytes.NewReader(d.response)),
	}, nil
}
