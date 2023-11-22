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

	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestNewLicenseCheckHandler(t *testing.T) {
	makeToken := func(licenseKey string) []byte {
		token := license.GenerateLicenseKeyBasedAccessToken(licenseKey)
		return []byte(token)
	}
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
		SiteID:            pointers.Ptr("C2582A60-573C-4EBC-BDD4-BC57A73CF010"), // uppercase to test case sensitivity
	}
	licenses := []dbLicense{
		validLicense,
		expiredLicense,
		revokedLicense,
		assignedLicense,
	}

	getBody := func(siteID string) string {
		s := "a43d50fa-23b6-41e9-86c9-558dd1f7ad54"
		if siteID != "" {
			s = siteID
		}
		return fmt.Sprintf(`{"siteID": "%s"}`, s)
	}

	db := dbmocks.NewMockDB()

	mockedEventLogs := dbmocks.NewStrictMockEventLogStore()
	mockedEventLogs.InsertFunc.SetDefaultReturn(nil)
	db.EventLogsFunc.SetDefaultReturn(mockedEventLogs)

	mocks.licenses.GetByToken = func(tokenHexEncoded string) (*dbLicense, error) {
		token, err := hex.DecodeString(tokenHexEncoded)
		if err != nil {
			return nil, err
		}
		for _, license := range licenses {
			if license.LicenseCheckToken != nil && bytes.Equal(license.LicenseCheckToken, token) {
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
			body:       getBody(""),
			headers:    nil,
			want:       licensing.LicenseCheckResponse{Error: ErrInvalidAccessTokenMsg},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid access token",
			body: getBody(""),
			headers: http.Header{
				"Authorization": {"Bearer invalid-token"},
			},
			want:       licensing.LicenseCheckResponse{Error: ErrInvalidAccessTokenMsg},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "expired license access token",
			body: getBody(""),
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(expiredLicense.LicenseCheckToken)},
			},
			want:       licensing.LicenseCheckResponse{Data: &licensing.LicenseCheckResponseData{IsValid: false, Reason: ReasonLicenseExpired}},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "revoked license access token",
			body: getBody(""),
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(revokedLicense.LicenseCheckToken)},
			},
			want:       licensing.LicenseCheckResponse{Data: &licensing.LicenseCheckResponseData{IsValid: false, Reason: ReasonLicenseRevokedMsg}},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "valid access token, invalid request body",
			body: "invalid body",
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(validLicense.LicenseCheckToken)},
			},
			want:       licensing.LicenseCheckResponse{Error: ErrInvalidRequestBodyMsg},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "valid access token, invalid site id (same license key used in multiple instances)",
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(assignedLicense.LicenseCheckToken)},
			},
			body:       getBody(""),
			want:       licensing.LicenseCheckResponse{Data: &licensing.LicenseCheckResponseData{IsValid: true, Reason: ReasonLicenseIsAlreadyInUseMsg}},
			wantStatus: http.StatusOK,
		},
		{
			name: "valid access token, valid site id",
			body: getBody(strings.ToLower(*assignedLicense.SiteID)),
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(assignedLicense.LicenseCheckToken)},
			},
			want:       licensing.LicenseCheckResponse{Data: &licensing.LicenseCheckResponseData{IsValid: true}},
			wantStatus: http.StatusOK,
		},
		{
			name: "valid access token, invalid uuid",
			body: getBody("some-non-uuid-string"),
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(validLicense.LicenseCheckToken)},
			},
			want:       licensing.LicenseCheckResponse{Error: ErrInvalidSiteIDMsg},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "valid access token, new site ID",
			body: getBody("85d3d2ed-d2d0-4a88-a49a-79af730f5ed0"),
			headers: http.Header{
				"Authorization": {"Bearer " + hex.EncodeToString(validLicense.LicenseCheckToken)},
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
			_ = json.Unmarshal(res.Body.Bytes(), &got)
			require.Equal(t, test.want, got)
		})
	}
}
