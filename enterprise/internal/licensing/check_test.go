package licensing

import (
	"bytes"
	"context"
	"encoding/hex"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func Test_checkerShouldSkip(t *testing.T) {
	tests := []struct {
		name string
		info *Info
		want bool
	}{
		{
			name: "skips for older license version",
			info: createTestLicenseInfo(false, nil),
			want: true,
		},
		{
			name: "skips for air-gapped instances",
			info: createTestLicenseInfo(true, []string{AirGappedTag}),
			want: true,
		},
		{
			name: "does not skip for newer license version",
			info: createTestLicenseInfo(true, nil),
			want: false,
		},
	}
	checker := &licenseChecker{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := checker.shouldSkip(test.info)
			require.Equal(t, test.want, got)
		})
	}
}

func Test_checkIsLicenseValid(t *testing.T) {
	tests := []struct {
		name           string
		licenseInfo    *Info
		licenseKey     string
		responseStatus int
		responseBody   []byte
		want           bool
		wantErr        bool
	}{
		{
			name:           "returns error if unable to make a request to license server",
			licenseInfo:    createTestLicenseInfo(true, nil),
			licenseKey:     "test",
			responseStatus: http.StatusInternalServerError,
			responseBody:   []byte(""),
			want:           false,
			wantErr:        true,
		},
		{
			name:           "returns correct result",
			licenseInfo:    createTestLicenseInfo(true, nil),
			licenseKey:     "test",
			responseStatus: http.StatusOK,
			responseBody:   []byte(`{"license":{"expiresAt":"2020-01-01T00:00:00Z"}}`),
			want:           true,
			wantErr:        false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			doer := &mockDoer{
				statusCode: test.responseStatus,
				response:   test.responseBody,
			}
			checker := &licenseChecker{
				doer: doer,
			}
			got, err := checker.check(context.Background(), database.GlobalState{}, test.licenseInfo, test.licenseKey)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.want, got)
			require.True(t, doer.DoCalled)
			require.Equal(t, "https://sourcegraph.com/.api/license/check", doer.Request.URL.String())
			require.Equal(t, "application/json", doer.Request.Header.Get("Content-Type"))

			tokenHexEncoded := hex.EncodeToString(GenerateHashedLicenseKeyAccessToken(test.licenseKey))
			require.Equal(t, "Bearer "+tokenHexEncoded, doer.Request.Header.Get("Authorization"))
		})
	}
}

var strPtr = func(s string) *string { return &s }

func createTestLicenseInfo(newer bool, tags []string) *Info {
	var salesforceSubscriptionID *string
	if newer {
		salesforceSubscriptionID = strPtr("123")
	}

	return &Info{
		license.Info{
			SalesforceSubscriptionID: salesforceSubscriptionID,
			Tags:                     tags,
		},
	}
}

type mockDoer struct {
	DoCalled bool
	Request  *http.Request

	statusCode int
	response   []byte
}

func (d *mockDoer) Do(req *http.Request) (*http.Response, error) {
	d.DoCalled = true
	d.Request = req

	return &http.Response{
		StatusCode: d.statusCode,
		Body:       io.NopCloser(bytes.NewReader(d.response)),
	}, nil
}
