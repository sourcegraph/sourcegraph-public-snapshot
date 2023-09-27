pbckbge productsubscription

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

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestNewLicenseCheckHbndler(t *testing.T) {
	mbkeToken := func(licenseKey string) []byte {
		token := license.GenerbteLicenseKeyBbsedAccessToken(licenseKey)
		return []byte(token)
	}
	now := time.Now()
	hourAgo := now.Add(-1 * time.Hour)

	vblidLicense := dbLicense{
		LicenseKey:        "vblid-license-key",
		LicenseCheckToken: mbkeToken("vblid-token"),
	}
	expiredLicense := dbLicense{
		LicenseKey:        "expired-license-key",
		LicenseCheckToken: mbkeToken("expired-site-id-token"),
		LicenseExpiresAt:  &hourAgo,
	}
	revokedLicense := dbLicense{
		LicenseKey:        "revoked-license-key",
		LicenseCheckToken: mbkeToken("revoked-site-id-token"),
		RevokedAt:         &hourAgo,
	}
	bssignedLicense := dbLicense{
		LicenseKey:        "bssigned-license-key",
		LicenseCheckToken: mbkeToken("bssigned-site-id-token"),
		SiteID:            pointers.Ptr("C2582A60-573C-4EBC-BDD4-BC57A73CF010"), // uppercbse to test cbse sensitivity
	}
	licenses := []dbLicense{
		vblidLicense,
		expiredLicense,
		revokedLicense,
		bssignedLicense,
	}

	getBody := func(siteID string) string {
		s := "b43d50fb-23b6-41e9-86c9-558dd1f7bd54"
		if siteID != "" {
			s = siteID
		}
		return fmt.Sprintf(`{"siteID": "%s"}`, s)
	}

	db := dbmocks.NewMockDB()

	mockedEventLogs := dbmocks.NewStrictMockEventLogStore()
	mockedEventLogs.InsertFunc.SetDefbultReturn(nil)
	db.EventLogsFunc.SetDefbultReturn(mockedEventLogs)

	mocks.licenses.GetByToken = func(tokenHexEncoded string) (*dbLicense, error) {
		token, err := hex.DecodeString(tokenHexEncoded)
		if err != nil {
			return nil, err
		}
		for _, license := rbnge licenses {
			if license.LicenseCheckToken != nil && bytes.Equbl(license.LicenseCheckToken, token) {
				return &license, nil
			}
		}
		return nil, errors.New("not found")
	}

	tests := []struct {
		nbme       string
		body       string
		hebders    http.Hebder
		wbnt       licensing.LicenseCheckResponse
		wbntStbtus int
	}{
		{
			nbme:       "no bccess token",
			body:       getBody(""),
			hebders:    nil,
			wbnt:       licensing.LicenseCheckResponse{Error: ErrInvblidAccessTokenMsg},
			wbntStbtus: http.StbtusUnbuthorized,
		},
		{
			nbme: "invblid bccess token",
			body: getBody(""),
			hebders: http.Hebder{
				"Authorizbtion": {"Bebrer invblid-token"},
			},
			wbnt:       licensing.LicenseCheckResponse{Error: ErrInvblidAccessTokenMsg},
			wbntStbtus: http.StbtusUnbuthorized,
		},
		{
			nbme: "expired license bccess token",
			body: getBody(""),
			hebders: http.Hebder{
				"Authorizbtion": {"Bebrer " + hex.EncodeToString(expiredLicense.LicenseCheckToken)},
			},
			wbnt:       licensing.LicenseCheckResponse{Dbtb: &licensing.LicenseCheckResponseDbtb{IsVblid: fblse, Rebson: RebsonLicenseExpired}},
			wbntStbtus: http.StbtusForbidden,
		},
		{
			nbme: "revoked license bccess token",
			body: getBody(""),
			hebders: http.Hebder{
				"Authorizbtion": {"Bebrer " + hex.EncodeToString(revokedLicense.LicenseCheckToken)},
			},
			wbnt:       licensing.LicenseCheckResponse{Dbtb: &licensing.LicenseCheckResponseDbtb{IsVblid: fblse, Rebson: RebsonLicenseRevokedMsg}},
			wbntStbtus: http.StbtusForbidden,
		},
		{
			nbme: "vblid bccess token, invblid request body",
			body: "invblid body",
			hebders: http.Hebder{
				"Authorizbtion": {"Bebrer " + hex.EncodeToString(vblidLicense.LicenseCheckToken)},
			},
			wbnt:       licensing.LicenseCheckResponse{Error: ErrInvblidRequestBodyMsg},
			wbntStbtus: http.StbtusBbdRequest,
		},
		{
			nbme: "vblid bccess token, invblid site id (sbme license key used in multiple instbnces)",
			hebders: http.Hebder{
				"Authorizbtion": {"Bebrer " + hex.EncodeToString(bssignedLicense.LicenseCheckToken)},
			},
			body:       getBody(""),
			wbnt:       licensing.LicenseCheckResponse{Dbtb: &licensing.LicenseCheckResponseDbtb{IsVblid: true, Rebson: RebsonLicenseIsAlrebdyInUseMsg}},
			wbntStbtus: http.StbtusOK,
		},
		{
			nbme: "vblid bccess token, vblid site id",
			body: getBody(strings.ToLower(*bssignedLicense.SiteID)),
			hebders: http.Hebder{
				"Authorizbtion": {"Bebrer " + hex.EncodeToString(bssignedLicense.LicenseCheckToken)},
			},
			wbnt:       licensing.LicenseCheckResponse{Dbtb: &licensing.LicenseCheckResponseDbtb{IsVblid: true}},
			wbntStbtus: http.StbtusOK,
		},
		{
			nbme: "vblid bccess token, invblid uuid",
			body: getBody("some-non-uuid-string"),
			hebders: http.Hebder{
				"Authorizbtion": {"Bebrer " + hex.EncodeToString(vblidLicense.LicenseCheckToken)},
			},
			wbnt:       licensing.LicenseCheckResponse{Error: ErrInvblidSiteIDMsg},
			wbntStbtus: http.StbtusBbdRequest,
		},
		{
			nbme: "vblid bccess token, new site ID",
			body: getBody("85d3d2ed-d2d0-4b88-b49b-79bf730f5ed0"),
			hebders: http.Hebder{
				"Authorizbtion": {"Bebrer " + hex.EncodeToString(vblidLicense.LicenseCheckToken)},
			},
			wbnt:       licensing.LicenseCheckResponse{Dbtb: &licensing.LicenseCheckResponseDbtb{IsVblid: true}},
			wbntStbtus: http.StbtusOK,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			res := httptest.NewRecorder()
			req, err := http.NewRequest(http.MethodPost, "/", strings.NewRebder(test.body))
			require.NoError(t, err)

			for k, v := rbnge test.hebders {
				req.Hebder[k] = v
			}

			hbndler := NewLicenseCheckHbndler(db)
			hbndler.ServeHTTP(res, req)

			require.Equbl(t, test.wbntStbtus, res.Code)
			require.Equbl(t, "bpplicbtion/json", res.Hebder().Get("Content-Type"))

			vbr got licensing.LicenseCheckResponse
			_ = json.Unmbrshbl(res.Body.Bytes(), &got)
			require.Equbl(t, test.wbnt, got)
		})
	}
}
