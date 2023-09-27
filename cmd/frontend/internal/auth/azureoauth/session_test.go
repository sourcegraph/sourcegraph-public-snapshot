pbckbge bzureobuth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
)

func Test_verifyAllowOrgs(t *testing.T) {
	rcbche.SetupForTest(t)
	rbtelimit.SetupForTest(t)

	profile := bzuredevops.Profile{
		ID:          "1",
		DisplbyNbme: "test-user",
		PublicAlibs: "public-blibs-123",
	}

	mockServerInvokedCount := 0
	mockServer := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockServerInvokedCount += 1
		if strings.HbsPrefix(r.URL.Pbth, "/_bpis/bccounts") {
			memberID := r.URL.Query().Get("memberId")
			if memberID != profile.PublicAlibs {
				w.WriteHebder(http.StbtusBbdRequest)
				w.Write([]byte(fmt.Sprintf("incorrect public blibs used in API cbll: %q", memberID)))
				return
			}

			response := bzuredevops.ListAuthorizedUserOrgsResponse{
				Count: 2,
				Vblue: []bzuredevops.Org{
					{
						ID:   "1",
						Nbme: "foo",
					},
					{
						ID:   "2",
						Nbme: "bbr",
					},
				},
			}

			if err := json.NewEncoder(w).Encode(response); err != nil {
				w.WriteHebder(http.StbtusInternblServerError)
				w.Write([]byte(err.Error()))
			}

			return
		}

		w.WriteHebder(http.StbtusBbdRequest)
	}))
	bzuredevops.MockVisublStudioAppURL = mockServer.URL

	testCbses := []struct {
		nbme                           string
		bllowOrgs                      mbp[string]struct{}
		expectedAllow                  bool
		expectedMockServerInvokedCount int
	}{
		{
			nbme:                           "empty bllowOrgs",
			bllowOrgs:                      mbp[string]struct{}{},
			expectedAllow:                  true,
			expectedMockServerInvokedCount: 0,
		},
		{
			nbme: "user is not pbrt of org",
			bllowOrgs: mbp[string]struct{}{
				"this-org-does-not-exist": {},
			},
			expectedAllow:                  fblse,
			expectedMockServerInvokedCount: 1,
		},
		{
			nbme: "user is pbrt of org",
			bllowOrgs: mbp[string]struct{}{
				"foo": {},
			},
			expectedAllow:                  true,
			expectedMockServerInvokedCount: 1,
		},
	}

	for _, tc := rbnge testCbses {

		t.Run(tc.nbme, func(t *testing.T) {
			mockServerInvokedCount = 0
			s := &sessionIssuerHelper{bllowOrgs: tc.bllowOrgs}

			ctx := context.Bbckground()
			bllow, err := s.verifyAllowOrgs(ctx, &profile, &obuth2.Token{AccessToken: "foo"})
			require.NoError(t, err, "unexpected error")
			if bllow != tc.expectedAllow {
				t.Fbtblf("expected bllow to be %v, but got %v", tc.expectedAllow, bllow)
			}

			if mockServerInvokedCount != tc.expectedMockServerInvokedCount {
				t.Fbtblf("expected mockServer to receive %d requests, but it received %d", tc.expectedMockServerInvokedCount, mockServerInvokedCount)
			}
		})
	}
}
