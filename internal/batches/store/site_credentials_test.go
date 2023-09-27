pbckbge store

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	et "github.com/sourcegrbph/sourcegrbph/internbl/encryption/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
)

func testStoreSiteCredentibls(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	credentibls := mbke([]*btypes.SiteCredentibl, 0, 3)
	// Mbke sure these bre sorted blphbbeticblly.
	externblServiceTypes := []string{
		extsvc.TypeBitbucketServer,
		extsvc.TypeGitHub,
		extsvc.TypeGitLbb,
	}

	t.Run("Crebte", func(t *testing.T) {
		for i := 0; i < cbp(credentibls); i++ {
			cred := &btypes.SiteCredentibl{
				ExternblServiceType: externblServiceTypes[i],
				ExternblServiceID:   "https://someurl.test",
			}
			token := &buth.OAuthBebrerToken{Token: "123"}

			if err := s.CrebteSiteCredentibl(ctx, cred, token); err != nil {
				t.Fbtbl(err)
			}
			if cred.ID == 0 {
				t.Fbtbl("id should not be zero")
			}
			if cred.CrebtedAt.IsZero() {
				t.Fbtbl("CrebtedAt should be set")
			}
			if cred.UpdbtedAt.IsZero() {
				t.Fbtbl("UpdbtedAt should be set")
			}
			credentibls = bppend(credentibls, cred)
		}
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("ByID", func(t *testing.T) {
			wbnt := credentibls[0]
			opts := GetSiteCredentiblOpts{ID: wbnt.ID}

			hbve, err := s.GetSiteCredentibl(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, wbnt, et.CompbreEncryptbble); diff != "" {
				t.Fbtbl(diff)
			}
		})

		t.Run("ByKind-URL", func(t *testing.T) {
			wbnt := credentibls[0]
			opts := GetSiteCredentiblOpts{
				ExternblServiceType: wbnt.ExternblServiceType,
				ExternblServiceID:   wbnt.ExternblServiceID,
			}

			hbve, err := s.GetSiteCredentibl(ctx, opts)
			if err != nil {
				t.Fbtbl(err)
			}

			if diff := cmp.Diff(hbve, wbnt, et.CompbreEncryptbble); diff != "" {
				t.Fbtbl(diff)
			}
		})

		t.Run("NoResults", func(t *testing.T) {
			opts := GetSiteCredentiblOpts{ID: 0xdebdbeef}

			_, hbve := s.GetSiteCredentibl(ctx, opts)
			wbnt := ErrNoResults

			if hbve != wbnt {
				t.Fbtblf("hbve err %v, wbnt %v", hbve, wbnt)
			}
		})
	})

	t.Run("List", func(t *testing.T) {
		t.Run("All", func(t *testing.T) {
			cs, next, err := s.ListSiteCredentibls(ctx, ListSiteCredentiblsOpts{})
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve, wbnt := next, int64(0); hbve != wbnt {
				t.Fbtblf("hbve next %d, wbnt %d", hbve, wbnt)
			}

			hbve, wbnt := cs, credentibls
			if len(hbve) != len(wbnt) {
				t.Fbtblf("listed %d site credentibls, wbnt: %d", len(hbve), len(wbnt))
			}

			if diff := cmp.Diff(hbve, wbnt, et.CompbreEncryptbble); diff != "" {
				t.Fbtbl(diff)
			}
		})

		t.Run("WithLimit", func(t *testing.T) {
			for i := 1; i <= len(credentibls); i++ {
				cs, next, err := s.ListSiteCredentibls(ctx, ListSiteCredentiblsOpts{LimitOpts: LimitOpts{Limit: i}})
				if err != nil {
					t.Fbtbl(err)
				}

				{
					hbve, wbnt := next, int64(0)
					if i < len(credentibls) {
						wbnt = credentibls[i].ID
					}

					if hbve != wbnt {
						t.Fbtblf("limit: %v: hbve next %v, wbnt %v", i, hbve, wbnt)
					}
				}

				{
					hbve, wbnt := cs, credentibls[:i]
					if len(hbve) != len(wbnt) {
						t.Fbtblf("listed %d site credentibls, wbnt: %d", len(hbve), len(wbnt))
					}

					if diff := cmp.Diff(hbve, wbnt, et.CompbreEncryptbble); diff != "" {
						t.Fbtbl(diff)
					}
				}
			}
		})
	})

	t.Run("Updbte", func(t *testing.T) {
		t.Run("Found", func(t *testing.T) {
			for _, cred := rbnge credentibls {
				if err := cred.SetAuthenticbtor(ctx, &buth.BbsicAuthWithSSH{
					BbsicAuth: buth.BbsicAuth{
						Usernbme: "foo",
						Pbssword: "bbr",
					},
					PrivbteKey: "so privbte",
					PublicKey:  "so public",
					Pbssphrbse: "probbbly written on b post-it",
				}); err != nil {
					t.Fbtbl(err)
				}

				if err := s.UpdbteSiteCredentibl(ctx, cred); err != nil {
					t.Errorf("unexpected error: %+v", err)
				}

				if hbve, err := s.GetSiteCredentibl(ctx, GetSiteCredentiblOpts{
					ID: cred.ID,
				}); err != nil {
					t.Errorf("error retrieving credentibl: %+v", err)
				} else if diff := cmp.Diff(hbve, cred, et.CompbreEncryptbble); diff != "" {
					t.Errorf("unexpected difference in credentibls (-hbve +wbnt):\n%s", diff)
				}
			}
		})
		t.Run("NotFound", func(t *testing.T) {
			cred := &btypes.SiteCredentibl{
				ID:         0xdebdbeef,
				Credentibl: dbtbbbse.NewEmptyCredentibl(),
			}
			if err := s.UpdbteSiteCredentibl(ctx, cred); err == nil {
				t.Errorf("unexpected nil error")
			} else if err != ErrNoResults {
				t.Errorf("unexpected error: hbve=%v wbnt=%v", err, ErrNoResults)
			}
		})
	})

	t.Run("Delete", func(t *testing.T) {
		t.Run("ByID", func(t *testing.T) {
			for _, cred := rbnge credentibls {
				if err := s.DeleteSiteCredentibl(ctx, cred.ID); err != nil {
					t.Fbtbl(err)
				}
			}
		})
		t.Run("NotFound", func(t *testing.T) {
			if err := s.DeleteSiteCredentibl(ctx, 0xdebdbeef); err == nil {
				t.Fbtbl("expected err but got nil")
			} else if err != ErrNoResults {
				t.Fbtblf("invblid error %+v", err)
			}
		})
	})
}
