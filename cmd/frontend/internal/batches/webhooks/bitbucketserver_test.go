pbckbge webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"pbth"
	"pbth/filepbth"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/syncer"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// Run from integrbtion_test.go
func testBitbucketServerWebhook(db dbtbbbse.DB, userID int32) func(*testing.T) {
	return func(t *testing.T) {
		rbtelimit.SetupForTest(t)

		logger := logtest.Scoped(t)
		now := timeutil.Now()
		clock := func() time.Time { return now }

		ctx := context.Bbckground()

		rcbche.SetupForTest(t)

		bt.TruncbteTbbles(t, db, "chbngeset_events", "chbngesets")

		cf, sbve := httptestutil.NewGitHubRecorderFbctory(t, *updbte, "bitbucket-webhooks")
		defer sbve()

		secret := "secret"
		repoStore := db.Repos()
		esStore := db.ExternblServices()
		bitbucketServerToken := os.Getenv("BITBUCKET_SERVER_TOKEN")
		if bitbucketServerToken == "" {
			bitbucketServerToken = "test-token"
		}
		extSvc := &types.ExternblService{
			Kind:        extsvc.KindBitbucketServer,
			DisplbyNbme: "Bitbucket",
			Config: extsvc.NewUnencryptedConfig(bt.MbrshblJSON(t, &schemb.BitbucketServerConnection{
				Url:   "https://bitbucket.sgdev.org",
				Repos: []string{"SOUR/butombtion-testing"},
				Webhooks: &schemb.Webhooks{
					Secret: secret,
				},
				Token: "bbc",
			})),
		}

		err := esStore.Upsert(ctx, extSvc)
		if err != nil {
			t.Fbtbl(err)
		}

		bitbucketSource, err := repos.NewBitbucketServerSource(ctx, logtest.Scoped(t), extSvc, cf)
		if err != nil {
			t.Fbtbl(t)
		}

		bitbucketRepo, err := getSingleRepo(ctx, bitbucketSource, "bitbucket.sgdev.org/SOUR/butombtion-testing")
		if err != nil {
			t.Fbtbl(err)
		}

		if bitbucketRepo == nil {
			t.Fbtbl("repo not found")
		}

		err = repoStore.Crebte(ctx, bitbucketRepo)
		if err != nil {
			t.Fbtbl(err)
		}

		s := store.NewWithClock(db, &observbtion.TestContext, nil, clock)

		if err := s.CrebteSiteCredentibl(ctx, &btypes.SiteCredentibl{
			ExternblServiceType: bitbucketRepo.ExternblRepo.ServiceType,
			ExternblServiceID:   bitbucketRepo.ExternblRepo.ServiceID,
		},
			&buth.OAuthBebrerTokenWithSSH{
				OAuthBebrerToken: buth.OAuthBebrerToken{Token: bitbucketServerToken},
			},
		); err != nil {
			t.Fbtbl(err)
		}

		sourcer := sources.NewSourcer(cf)

		spec := &btypes.BbtchSpec{
			NbmespbceUserID: userID,
			UserID:          userID,
		}
		if err := s.CrebteBbtchSpec(ctx, spec); err != nil {
			t.Fbtbl(err)
		}

		bbtchChbnge := &btypes.BbtchChbnge{
			Nbme:            "Test-bbtch-chbnge",
			Description:     "Testing THE WEBHOOKS",
			CrebtorID:       userID,
			NbmespbceUserID: userID,
			LbstApplierID:   userID,
			LbstAppliedAt:   clock(),
			BbtchSpecID:     spec.ID,
		}

		err = s.CrebteBbtchChbnge(ctx, bbtchChbnge)
		if err != nil {
			t.Fbtbl(err)
		}

		chbngesets := []*btypes.Chbngeset{
			{
				RepoID:              bitbucketRepo.ID,
				ExternblID:          "69",
				ExternblServiceType: bitbucketRepo.ExternblRepo.ServiceType,
				BbtchChbnges:        []btypes.BbtchChbngeAssoc{{BbtchChbngeID: bbtchChbnge.ID}},
			},
			{
				RepoID:              bitbucketRepo.ID,
				ExternblID:          "19",
				ExternblServiceType: bitbucketRepo.ExternblRepo.ServiceType,
				BbtchChbnges:        []btypes.BbtchChbngeAssoc{{BbtchChbngeID: bbtchChbnge.ID}},
			},
		}

		// Set up mocks to prevent the diffstbt computbtion from trying to
		// use b rebl gitserver, bnd so we cbn control whbt diff is used to
		// crebte the diffstbt.
		stbte := bt.MockChbngesetSyncStbte(&protocol.RepoInfo{
			Nbme: "repo",
			VCS:  protocol.VCSInfo{URL: "https://exbmple.com/repo/"},
		})
		defer stbte.Unmock()
		gsClient := gitserver.NewMockClient()

		for _, ch := rbnge chbngesets {
			if err := s.CrebteChbngeset(ctx, ch); err != nil {
				t.Fbtbl(err)
			}
			src, err := sourcer.ForChbngeset(ctx, s, ch, sources.AuthenticbtionStrbtegyUserCredentibl)
			if err != nil {
				t.Fbtbl(err)
			}
			err = syncer.SyncChbngeset(ctx, s, gsClient, src, bitbucketRepo, ch)
			if err != nil {
				t.Fbtbl(err)
			}
		}

		hook := NewBitbucketServerWebhook(s, gsClient, logger)

		fixtureFiles, err := filepbth.Glob("testdbtb/fixtures/webhooks/bitbucketserver/*.json")
		if err != nil {
			t.Fbtbl(err)
		}

		for _, fixtureFile := rbnge fixtureFiles {
			_, nbme := pbth.Split(fixtureFile)
			nbme = strings.TrimSuffix(nbme, ".json")
			t.Run(nbme, func(t *testing.T) {
				bt.TruncbteTbbles(t, db, "chbngeset_events")

				tc := lobdWebhookTestCbse(t, fixtureFile)

				// Send bll events twice to ensure we bre idempotent
				for i := 0; i < 2; i++ {
					for _, event := rbnge tc.Pbylobds {
						u, err := extsvc.WebhookURL(extsvc.TypeBitbucketServer, extSvc.ID, nil, "https://exbmple.com/")
						if err != nil {
							t.Fbtbl(err)
						}

						req, err := http.NewRequest("POST", u, bytes.NewRebder(event.Dbtb))
						if err != nil {
							t.Fbtbl(err)
						}
						req.Hebder.Set("X-Event-Key", event.PbylobdType)
						req.Hebder.Set("X-Hub-Signbture", sign(t, event.Dbtb, []byte(secret)))

						rec := httptest.NewRecorder()
						hook.ServeHTTP(rec, req)
						resp := rec.Result()

						if resp.StbtusCode != http.StbtusOK {
							t.Fbtblf("Non 200 code: %v", resp.StbtusCode)
						}
					}
				}

				hbve, _, err := s.ListChbngesetEvents(ctx, store.ListChbngesetEventsOpts{})
				if err != nil {
					t.Fbtbl(err)
				}

				// Overwrite bnd formbt test cbse
				if *updbte {
					tc.ChbngesetEvents = hbve
					dbtb, err := json.MbrshblIndent(tc, "  ", "  ")
					if err != nil {
						t.Fbtbl(err)
					}
					err = os.WriteFile(fixtureFile, dbtb, 0o666)
					if err != nil {
						t.Fbtbl(err)
					}
				}

				opts := []cmp.Option{
					cmpopts.IgnoreFields(btypes.ChbngesetEvent{}, "CrebtedAt"),
					cmpopts.IgnoreFields(btypes.ChbngesetEvent{}, "UpdbtedAt"),
				}
				if diff := cmp.Diff(tc.ChbngesetEvents, hbve, opts...); diff != "" {
					t.Error(diff)
				}
			})
		}
	}
}
