package productsubscription

import (
	"bytes"
	"encoding/hex"
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
	allLicenses := []dbLicense{
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
		for _, license := range allLicenses {
			if license.LicenseCheckToken != nil && bytes.Equal(*license.LicenseCheckToken, token) {
				return &license, nil
			}
		}
		return nil, errors.New("not found")
	}

	tests := []struct {
		name       string
		headers    http.Header
		body       string
		wantStatus int
	}{
		{
			name:       "no access token",
			headers:    nil,
			body:       `{"siteID": "some-site-id"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid access token",
			headers: http.Header{
				"Authorization": {"Bearer invalid-token"},
			},
			body:       `{"siteID": "some-site-id"}`,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "expired license access token",
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(*expiredLicense.LicenseCheckToken)},
			},
			body:       `{"siteID": "some-site-id"}`,
			wantStatus: http.StatusForbidden,
		},
		{
			name: "revoked license access token",
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(*revokedLicense.LicenseCheckToken)},
			},
			body:       `{"siteID": "some-site-id"}`,
			wantStatus: http.StatusForbidden,
		},
		{
			name: "valid access token, invalid request body",
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(*validLicense.LicenseCheckToken)},
			},
			body:       "invalid body",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "valid access token, incorrect site id",
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(*assignedLicense.LicenseCheckToken)},
			},
			body:       `{"siteID": "some-site-id"}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "valid access token, valid site id",
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(*assignedLicense.LicenseCheckToken)},
			},
			body:       fmt.Sprintf(`{"siteID": "%s"}`, *assignedLicense.SiteID),
			wantStatus: http.StatusOK,
		},
		{
			name: "valid access token, new site id",
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(*validLicense.LicenseCheckToken)},
			},
			body:       `{"siteID": "some-site-id"}`,
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
		})
	}

	// todo: test rate limiting
}
