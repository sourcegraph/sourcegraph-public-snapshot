package productsubscription

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestNewLicenseCheckHandler(t *testing.T) {
	makeToken := func(licenseKey string) *[]byte {
		token := licensing.GenerateHashedLicenseKeyAccessToken(licenseKey)
		return &token
	}
	strPtr := func(s string) *string { return &s }
	now := time.Now()
	hourAgo := now.Add(-1 * time.Hour)

	validLicense := dbLicense{
		LicenseKey:        "valid-license-key",
		LicenseCheckToken: makeToken("valid-token"),
	}
	expiredLicense := dbLicense{
		LicenseKey:        "expired-license-key",
		LicenseCheckToken: makeToken("expired-site-id-token"),
		LicenseExpiresAt:  &hourAgo,
	}
	revokedLicense := dbLicense{
		LicenseKey:        "revoked-license-key",
		LicenseCheckToken: makeToken("revoked-site-id-token"),
		RevokedAt:         &hourAgo,
	}
	assignedLicense := dbLicense{
		LicenseKey:        "assigned-license-key",
		LicenseCheckToken: makeToken("assigned-site-id-token"),
		SiteID:            strPtr("assigned-site-id"),
	}
	licenses := []dbLicense{
		validLicense,
		expiredLicense,
		revokedLicense,
		assignedLicense,
	}

	db := database.NewMockDB()
	mocks.licenses.GetByToken = func(tokenHexEncoded string) (*dbLicense, error) {
		token, err := hex.DecodeString(tokenHexEncoded)
		if err != nil {
			return nil, err
		}
		for _, license := range licenses {
			if license.LicenseCheckToken != nil && bytes.Equal(*license.LicenseCheckToken, token) {
				return &license, nil
			}
		}
		return nil, errors.New("not found")
	}

	tests := []struct {
		name       string
		body       string
		headers    http.Header
		want       licensing.LicenseCheckResponse
		wantStatus int
	}{
		{
			name:       "no access token",
			body:       `{"siteID": "some-site-id"}`,
			headers:    nil,
			want:       licensing.LicenseCheckResponse{Error: "invalid access token"},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid access token",
			body: `{"siteID": "some-site-id"}`,
			headers: http.Header{
				"Authorization": {"Bearer invalid-token"},
			},
			want:       licensing.LicenseCheckResponse{Error: "invalid access token"},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "expired license access token",
			body: `{"siteID": "some-site-id"}`,
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(*expiredLicense.LicenseCheckToken)},
			},
			want:       licensing.LicenseCheckResponse{Error: "license expired"},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "revoked license access token",
			body: `{"siteID": "some-site-id"}`,
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(*revokedLicense.LicenseCheckToken)},
			},
			want:       licensing.LicenseCheckResponse{Data: &licensing.LicenseCheckResponseData{IsValid: false, Reason: "license revoked"}},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "valid access token, invalid request body",
			body: "invalid body",
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(*validLicense.LicenseCheckToken)},
			},
			want:       licensing.LicenseCheckResponse{Error: "invalid request body"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "valid access token, invalid site id (abuse)",
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(*assignedLicense.LicenseCheckToken)},
			},
			body:       `{"siteID": "some-site-id"}`,
			want:       licensing.LicenseCheckResponse{Data: &licensing.LicenseCheckResponseData{IsValid: false, Reason: "license is already in use"}},
			wantStatus: http.StatusOK,
		},
		{
			name: "valid access token, valid site id",
			body: fmt.Sprintf(`{"siteID": "%s"}`, *assignedLicense.SiteID),
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(*assignedLicense.LicenseCheckToken)},
			},
			want:       licensing.LicenseCheckResponse{Data: &licensing.LicenseCheckResponseData{IsValid: true}},
			wantStatus: http.StatusOK,
		},
		{
			name: "valid access token, new site id",
			body: `{"siteID": "some-site-id"}`,
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(*validLicense.LicenseCheckToken)},
			},
			want:       licensing.LicenseCheckResponse{Data: &licensing.LicenseCheckResponseData{IsValid: true}},
			wantStatus: http.StatusOK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			require.NoError(t, err)

			for k, v := range test.headers {
				req.Header[k] = v
			}

			handler := NewLicenseCheckHandler(db)
			handler.ServeHTTP(res, req)

			require.Equal(t, test.wantStatus, res.Code)
			require.Equal(t, "application/json", res.Header().Get("Content-Type"))

			var got licensing.LicenseCheckResponse
			json.Unmarshal([]byte(res.Body.String()), &got)
			require.Equal(t, test.want, got)
		})
	}
}
