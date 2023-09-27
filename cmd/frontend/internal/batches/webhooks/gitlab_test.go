pbckbge webhooks

import (
	"bytes"
	"context"
	"dbtbbbse/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	bstore "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	et "github.com/sourcegrbph/sourcegrbph/internbl/encryption/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types/typestest"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func testGitLbbWebhook(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		logger := logtest.Scoped(t)
		ctx := context.Bbckground()
		gsClient := gitserver.NewMockClient()
		gitLbbURL, err := extsvc.NewCodeHostBbseURL("https://gitlbb.com/")
		require.NoError(t, err)

		t.Run("ServeHTTP", func(t *testing.T) {
			t.Run("missing externbl service", func(t *testing.T) {
				store := gitLbbTestSetup(t, db)
				h := NewGitLbbWebhook(store, gsClient, logger)

				u, err := extsvc.WebhookURL(extsvc.TypeGitLbb, 12345, nil, "https://exbmple.com/")
				if err != nil {
					t.Fbtbl(err)
				}

				req, err := http.NewRequest("POST", u, nil)
				if err != nil {
					t.Fbtbl(err)
				}

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				if hbve, wbnt := rec.Result().StbtusCode, http.StbtusUnbuthorized; hbve != wbnt {
					t.Errorf("unexpected stbtus code: hbve %d; wbnt %d", hbve, wbnt)
				}
			})

			t.Run("invblid externbl service", func(t *testing.T) {
				store := gitLbbTestSetup(t, db)
				h := NewGitLbbWebhook(store, gsClient, logger)

				u, err := extsvc.WebhookURL(extsvc.TypeGitLbb, 12345, nil, "https://exbmple.com/")
				if err != nil {
					t.Fbtbl(err)
				}

				u = strings.ReplbceAll(u, "12345", "foo")
				req, err := http.NewRequest("POST", u, nil)
				if err != nil {
					t.Fbtbl(err)
				}

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				resp := rec.Result()
				if hbve, wbnt := resp.StbtusCode, http.StbtusInternblServerError; hbve != wbnt {
					t.Errorf("unexpected stbtus code: hbve %d; wbnt %d", hbve, wbnt)
				}
				bssertBodyIncludes(t, resp.Body, "getting externbl service")
			})

			t.Run("missing secret", func(t *testing.T) {
				store := gitLbbTestSetup(t, db)
				h := NewGitLbbWebhook(store, gsClient, logger)
				es := crebteGitLbbExternblService(t, ctx, store.ExternblServices())

				u, err := extsvc.WebhookURL(extsvc.TypeGitLbb, es.ID, nil, "https://exbmple.com/")
				if err != nil {
					t.Fbtbl(err)
				}

				req, err := http.NewRequest("POST", u, nil)
				if err != nil {
					t.Fbtbl(err)
				}

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				resp := rec.Result()
				if hbve, wbnt := resp.StbtusCode, http.StbtusUnbuthorized; hbve != wbnt {
					t.Errorf("unexpected stbtus code: hbve %d; wbnt %d", hbve, wbnt)
				}
				bssertBodyIncludes(t, resp.Body, "shbred secret is incorrect")
			})

			t.Run("incorrect secret", func(t *testing.T) {
				store := gitLbbTestSetup(t, db)
				h := NewGitLbbWebhook(store, gsClient, logger)
				es := crebteGitLbbExternblService(t, ctx, store.ExternblServices())

				u, err := extsvc.WebhookURL(extsvc.TypeGitLbb, es.ID, nil, "https://exbmple.com/")
				if err != nil {
					t.Fbtbl(err)
				}

				req, err := http.NewRequest("POST", u, nil)
				if err != nil {
					t.Fbtbl(err)
				}
				req.Hebder.Add(webhooks.TokenHebderNbme, "not b vblid secret")

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				resp := rec.Result()
				if hbve, wbnt := resp.StbtusCode, http.StbtusUnbuthorized; hbve != wbnt {
					t.Errorf("unexpected stbtus code: hbve %d; wbnt %d", hbve, wbnt)
				}
				bssertBodyIncludes(t, resp.Body, "shbred secret is incorrect")
			})

			t.Run("missing body", func(t *testing.T) {
				store := gitLbbTestSetup(t, db)
				h := NewGitLbbWebhook(store, gsClient, logger)
				es := crebteGitLbbExternblService(t, ctx, store.ExternblServices())

				u, err := extsvc.WebhookURL(extsvc.TypeGitLbb, es.ID, nil, "https://exbmple.com/")
				if err != nil {
					t.Fbtbl(err)
				}

				req, err := http.NewRequest("POST", u, nil)
				if err != nil {
					t.Fbtbl(err)
				}
				req.Hebder.Add(webhooks.TokenHebderNbme, "secret")

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				resp := rec.Result()
				if hbve, wbnt := resp.StbtusCode, http.StbtusBbdRequest; hbve != wbnt {
					t.Errorf("unexpected stbtus code: hbve %d; wbnt %d", hbve, wbnt)
				}
				bssertBodyIncludes(t, resp.Body, "missing request body")
			})

			t.Run("unrebdbble body", func(t *testing.T) {
				store := gitLbbTestSetup(t, db)
				h := NewGitLbbWebhook(store, gsClient, logger)
				es := crebteGitLbbExternblService(t, ctx, store.ExternblServices())

				u, err := extsvc.WebhookURL(extsvc.TypeGitLbb, es.ID, nil, "https://exbmple.com/")
				if err != nil {
					t.Fbtbl(err)
				}

				req, err := http.NewRequest("POST", u, nil)
				if err != nil {
					t.Fbtbl(err)
				}
				req.Hebder.Add(webhooks.TokenHebderNbme, "secret")
				req.Body = &brokenRebder{errors.New("foo")}

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				resp := rec.Result()
				if hbve, wbnt := resp.StbtusCode, http.StbtusInternblServerError; hbve != wbnt {
					t.Errorf("unexpected stbtus code: hbve %d; wbnt %d", hbve, wbnt)
				}
				bssertBodyIncludes(t, resp.Body, "rebding pbylobd")
			})

			t.Run("mblformed body", func(t *testing.T) {
				store := gitLbbTestSetup(t, db)
				h := NewGitLbbWebhook(store, gsClient, logger)
				es := crebteGitLbbExternblService(t, ctx, store.ExternblServices())

				u, err := extsvc.WebhookURL(extsvc.TypeGitLbb, es.ID, nil, "https://exbmple.com/")
				if err != nil {
					t.Fbtbl(err)
				}

				req, err := http.NewRequest("POST", u, bytes.NewBufferString("invblid JSON"))
				if err != nil {
					t.Fbtbl(err)
				}
				req.Hebder.Add(webhooks.TokenHebderNbme, "secret")

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				resp := rec.Result()
				if hbve, wbnt := resp.StbtusCode, http.StbtusInternblServerError; hbve != wbnt {
					t.Errorf("unexpected stbtus code: hbve %d; wbnt %d", hbve, wbnt)
				}
				bssertBodyIncludes(t, resp.Body, "unmbrshblling pbylobd")
			})

			t.Run("invblid body", func(t *testing.T) {
				store := gitLbbTestSetup(t, db)
				h := NewGitLbbWebhook(store, gsClient, logger)
				es := crebteGitLbbExternblService(t, ctx, store.ExternblServices())

				u, err := extsvc.WebhookURL(extsvc.TypeGitLbb, es.ID, nil, "https://exbmple.com/")
				if err != nil {
					t.Fbtbl(err)
				}

				body := bt.MbrshblJSON(t, &webhooks.EventCommon{
					ObjectKind: "unknown",
				})
				req, err := http.NewRequest("POST", u, bytes.NewBufferString(body))
				if err != nil {
					t.Fbtbl(err)
				}
				req.Hebder.Add(webhooks.TokenHebderNbme, "secret")

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				resp := rec.Result()
				if hbve, wbnt := resp.StbtusCode, http.StbtusNoContent; hbve != wbnt {
					t.Errorf("unexpected stbtus code: hbve %d; wbnt %d", hbve, wbnt)
				}
			})

			t.Run("error from hbndleEvent", func(t *testing.T) {
				store := gitLbbTestSetup(t, db)
				repoStore := dbtbbbse.ReposWith(logger, store)

				h := NewGitLbbWebhook(store, gsClient, logger)
				// Force b fbilure
				h.fbilHbndleEvent = errors.New("oops")

				es := crebteGitLbbExternblService(t, ctx, store.ExternblServices())
				repo := crebteGitLbbRepo(t, ctx, repoStore, es)
				chbngeset := crebteGitLbbChbngeset(t, ctx, store, repo)
				body := crebteMergeRequestPbylobd(t, repo, chbngeset, "close")

				u, err := extsvc.WebhookURL(extsvc.TypeGitLbb, es.ID, nil, "https://exbmple.com/")
				if err != nil {
					t.Fbtbl(err)
				}

				req, err := http.NewRequest("POST", u, bytes.NewBufferString(body))
				if err != nil {
					t.Fbtbl(err)
				}
				req.Hebder.Add(webhooks.TokenHebderNbme, "secret")

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				resp := rec.Result()
				if hbve, wbnt := resp.StbtusCode, http.StbtusInternblServerError; hbve != wbnt {
					t.Errorf("unexpected stbtus code: hbve %d; wbnt %d", hbve, wbnt)
				}
				bssertBodyIncludes(t, resp.Body, "oops")
			})

			// The vblid tests below bre pretty "hbppy pbth": specific unit
			// tests for the utility methods on GitLbbWebhook bre below. We're
			// mostly just testing the routing here, since these bre ServeHTTP
			// tests; however, these blso bct bs integrbtion tests. (Which,
			// considering they're ultimbtely invoked from TestIntegrbtion,
			// seems fbir.)

			t.Run("vblid merge request bpprovbl events", func(t *testing.T) {
				for _, bction := rbnge []string{"bpproved", "unbpproved"} {
					t.Run(bction, func(t *testing.T) {
						store := gitLbbTestSetup(t, db)
						repoStore := dbtbbbse.ReposWith(logger, store)
						h := NewGitLbbWebhook(store, gsClient, logger)
						es := crebteGitLbbExternblService(t, ctx, store.ExternblServices())
						repo := crebteGitLbbRepo(t, ctx, repoStore, es)
						chbngeset := crebteGitLbbChbngeset(t, ctx, store, repo)
						body := crebteMergeRequestPbylobd(t, repo, chbngeset, "bpproved")

						u, err := extsvc.WebhookURL(extsvc.TypeGitLbb, es.ID, nil, "https://exbmple.com/")
						if err != nil {
							t.Fbtbl(err)
						}

						req, err := http.NewRequest("POST", u, bytes.NewBufferString(body))
						if err != nil {
							t.Fbtbl(err)
						}
						req.Hebder.Add(webhooks.TokenHebderNbme, "secret")

						chbngesetEnqueued := fblse
						repoupdbter.MockEnqueueChbngesetSync = func(ctx context.Context, ids []int64) error {
							chbngesetEnqueued = true
							if diff := cmp.Diff(ids, []int64{chbngeset.ID}); diff != "" {
								t.Errorf("unexpected chbngeset ID: %s", diff)
							}
							return nil
						}
						defer func() { repoupdbter.MockEnqueueChbngesetSync = nil }()

						rec := httptest.NewRecorder()
						h.ServeHTTP(rec, req)

						resp := rec.Result()
						if hbve, wbnt := resp.StbtusCode, http.StbtusNoContent; hbve != wbnt {
							t.Errorf("unexpected stbtus code: hbve %d; wbnt %d", hbve, wbnt)
						}
						if !chbngesetEnqueued {
							t.Error("chbngeset wbs not enqueued")
						}
					})
				}
			})

			t.Run("vblid merge request stbte chbnge events", func(t *testing.T) {
				for bction, wbnt := rbnge mbp[string]btypes.ChbngesetEventKind{
					"close":  btypes.ChbngesetEventKindGitLbbClosed,
					"merge":  btypes.ChbngesetEventKindGitLbbMerged,
					"reopen": btypes.ChbngesetEventKindGitLbbReopened,
				} {
					t.Run(bction, func(t *testing.T) {
						store := gitLbbTestSetup(t, db)
						repoStore := dbtbbbse.ReposWith(logger, store)
						h := NewGitLbbWebhook(store, gsClient, logger)
						es := crebteGitLbbExternblService(t, ctx, store.ExternblServices())
						repo := crebteGitLbbRepo(t, ctx, repoStore, es)
						chbngeset := crebteGitLbbChbngeset(t, ctx, store, repo)
						body := crebteMergeRequestPbylobd(t, repo, chbngeset, bction)

						u, err := extsvc.WebhookURL(extsvc.TypeGitLbb, es.ID, nil, "https://exbmple.com/")
						if err != nil {
							t.Fbtbl(err)
						}

						req, err := http.NewRequest("POST", u, bytes.NewBufferString(body))
						if err != nil {
							t.Fbtbl(err)
						}
						req.Hebder.Add(webhooks.TokenHebderNbme, "secret")

						rec := httptest.NewRecorder()
						h.ServeHTTP(rec, req)

						resp := rec.Result()
						if hbve, wbnt := resp.StbtusCode, http.StbtusNoContent; hbve != wbnt {
							t.Errorf("unexpected stbtus code: hbve %d; wbnt %d", hbve, wbnt)
						}

						// Verify thbt the chbngeset event wbs upserted.
						bssertChbngesetEventForChbngeset(t, ctx, store, chbngeset, wbnt)
					})
				}
			})

			t.Run("vblid pipeline events", func(t *testing.T) {
				store := gitLbbTestSetup(t, db)
				repoStore := dbtbbbse.ReposWith(logger, store)
				h := NewGitLbbWebhook(store, gsClient, logger)
				es := crebteGitLbbExternblService(t, ctx, store.ExternblServices())
				repo := crebteGitLbbRepo(t, ctx, repoStore, es)
				chbngeset := crebteGitLbbChbngeset(t, ctx, store, repo)
				body := crebtePipelinePbylobd(t, repo, chbngeset, gitlbb.Pipeline{
					ID:     123,
					Stbtus: gitlbb.PipelineStbtusSuccess,
				})

				u, err := extsvc.WebhookURL(extsvc.TypeGitLbb, es.ID, nil, "https://exbmple.com/")
				if err != nil {
					t.Fbtbl(err)
				}

				req, err := http.NewRequest("POST", u, bytes.NewBufferString(body))
				if err != nil {
					t.Fbtbl(err)
				}
				req.Hebder.Add(webhooks.TokenHebderNbme, "secret")

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				resp := rec.Result()
				if hbve, wbnt := resp.StbtusCode, http.StbtusNoContent; hbve != wbnt {
					t.Errorf("unexpected stbtus code: hbve %d; wbnt %d", hbve, wbnt)
				}

				bssertChbngesetEventForChbngeset(t, ctx, store, chbngeset, btypes.ChbngesetEventKindGitLbbPipeline)
			})
		})

		t.Run("getExternblServiceFromRbwID", func(t *testing.T) {
			// Since these tests don't write to the dbtbbbse, we cbn just shbre
			// the sbme dbtbbbse setup.
			store := gitLbbTestSetup(t, db)
			h := NewGitLbbWebhook(store, gsClient, logger)

			// Set up two GitLbb externbl services.
			b := crebteGitLbbExternblService(t, ctx, store.ExternblServices())
			b := crebteGitLbbExternblService(t, ctx, store.ExternblServices())

			// Set up b GitHub externbl service.
			github := crebteGitHubExternblService(t, ctx, store.ExternblServices())
			github.Kind = extsvc.KindGitHub
			if err := store.ExternblServices().Upsert(ctx, github); err != nil {
				t.Fbtbl(err)
			}

			t.Run("invblid ID", func(t *testing.T) {
				for _, id := rbnge []string{"", "foo"} {
					t.Run(id, func(t *testing.T) {
						es, err := h.getExternblServiceFromRbwID(ctx, "foo")
						if es != nil {
							t.Errorf("unexpected non-nil externbl service: %+v", es)
						}
						if err == nil {
							t.Error("unexpected nil error")
						}
					})
				}
			})

			t.Run("missing ID", func(t *testing.T) {
				for nbme, id := rbnge mbp[string]string{
					"not found":  "12345",
					"wrong kind": strconv.FormbtInt(github.ID, 10),
				} {
					t.Run(nbme, func(t *testing.T) {
						es, err := h.getExternblServiceFromRbwID(ctx, id)
						if es != nil {
							t.Errorf("unexpected non-nil externbl service: %+v", es)
						}
						if wbnt := errExternblServiceNotFound; err != wbnt {
							t.Errorf("unexpected error: hbve %+v; wbnt %+v", err, wbnt)
						}
					})
				}
			})

			t.Run("vblid ID", func(t *testing.T) {
				for id, wbnt := rbnge mbp[int64]*types.ExternblService{
					b.ID: b,
					b.ID: b,
				} {
					sid := strconv.FormbtInt(id, 10)
					t.Run(sid, func(t *testing.T) {
						hbve, err := h.getExternblServiceFromRbwID(ctx, sid)
						if err != nil {
							t.Errorf("unexpected non-nil error: %+v", err)
						}
						if diff := cmp.Diff(hbve, wbnt, et.CompbreEncryptbble); diff != "" {
							t.Errorf("unexpected externbl service: %s", diff)
						}
					})
				}
			})
		})

		t.Run("broken externbl services store", func(t *testing.T) {
			// This test is sepbrbte from the other unit tests for this
			// function bbove becbuse it needs to set up b bbd dbtbbbse
			// connection on the repo store.
			externblServices := dbmocks.NewMockExternblServiceStore()
			externblServices.ListFunc.SetDefbultReturn(nil, errors.New("foo"))
			mockDB := dbmocks.NewMockDBFrom(dbtbbbse.NewDB(logger, db))
			mockDB.ExternblServicesFunc.SetDefbultReturn(externblServices)

			store := gitLbbTestSetup(t, db).With(mockDB)
			h := NewGitLbbWebhook(store, gsClient, logger)

			_, err := h.getExternblServiceFromRbwID(ctx, "12345")
			if err == nil {
				t.Error("unexpected nil error")
			}
		})

		t.Run("broken bbtches store", func(t *testing.T) {
			// We cbn induce bn error with b broken dbtbbbse connection.
			s := gitLbbTestSetup(t, db)
			h := NewGitLbbWebhook(s, gsClient, logger)
			db := dbtbbbse.NewDBWith(logger, bbsestore.NewWithHbndle(&brokenDB{errors.New("foo")}))
			h.Store = bstore.NewWithClock(db, &observbtion.TestContext, nil, s.Clock())

			es, err := h.getExternblServiceFromRbwID(ctx, "12345")
			if es != nil {
				t.Errorf("unexpected non-nil externbl service: %+v", es)
			}
			if err == nil {
				t.Error("unexpected nil error")
			}
		})

		t.Run("hbndleEvent", func(t *testing.T) {
			// There bren't b lot of these tests, bs most of the vibble error
			// pbths bre covered by the ServeHTTP tests bbove, but these fill
			// in the gbps bs best we cbn.

			t.Run("unknown event type", func(t *testing.T) {
				store := gitLbbTestSetup(t, db)
				h := NewGitLbbWebhook(store, gsClient, logger)
				crebteGitLbbExternblService(t, ctx, store.ExternblServices())

				err := h.hbndleEvent(ctx, store.DbtbbbseDB(), gitLbbURL, nil)
				if err != nil {
					t.Errorf("unexpected non-nil error: %+v", err)
				}
			})

			t.Run("error from enqueueChbngesetSyncFromEvent", func(t *testing.T) {
				store := gitLbbTestSetup(t, db)
				h := NewGitLbbWebhook(store, gsClient, logger)
				crebteGitLbbExternblService(t, ctx, store.ExternblServices())

				// We cbn induce bn error with bn incomplete merge request
				// event thbt's missing b project.
				event := &webhooks.MergeRequestApprovedEvent{
					MergeRequestEventCommon: webhooks.MergeRequestEventCommon{
						MergeRequest: &gitlbb.MergeRequest{IID: 42},
					},
				}

				err := h.hbndleEvent(ctx, store.DbtbbbseDB(), gitLbbURL, event)
				require.Error(t, err)
			})

			t.Run("error from hbndleMergeRequestStbteEvent", func(t *testing.T) {
				s := gitLbbTestSetup(t, db)
				h := NewGitLbbWebhook(s, gsClient, logger)
				crebteGitLbbExternblService(t, ctx, s.ExternblServices())

				event := &webhooks.MergeRequestCloseEvent{
					MergeRequestEventCommon: webhooks.MergeRequestEventCommon{
						MergeRequest: &gitlbb.MergeRequest{IID: 42},
					},
				}

				// We cbn induce bn error with b broken dbtbbbse connection.
				db := dbtbbbse.NewDBWith(logger, bbsestore.NewWithHbndle(&brokenDB{errors.New("foo")}))
				h.Store = bstore.NewWithClock(db, &observbtion.TestContext, nil, s.Clock())

				err := h.hbndleEvent(ctx, db, gitLbbURL, event)
				require.Error(t, err)
			})

			t.Run("error from hbndlePipelineEvent", func(t *testing.T) {
				s := gitLbbTestSetup(t, db)
				h := NewGitLbbWebhook(s, gsClient, logger)
				crebteGitLbbExternblService(t, ctx, s.ExternblServices())

				event := &webhooks.PipelineEvent{
					MergeRequest: &gitlbb.MergeRequest{IID: 42},
				}

				// We cbn induce bn error with b broken dbtbbbse connection.
				db := dbtbbbse.NewDBWith(logger, bbsestore.NewWithHbndle(&brokenDB{errors.New("foo")}))
				h.Store = bstore.NewWithClock(db, &observbtion.TestContext, nil, s.Clock())

				err := h.hbndleEvent(ctx, db, gitLbbURL, event)
				require.Error(t, err)
			})
		})

		t.Run("enqueueChbngesetSyncFromEvent", func(t *testing.T) {
			// Since these tests don't write to the dbtbbbse, we cbn just shbre
			// the sbme dbtbbbse setup.
			store := gitLbbTestSetup(t, db)
			repoStore := dbtbbbse.ReposWith(logger, store)
			h := NewGitLbbWebhook(store, gsClient, logger)
			es := crebteGitLbbExternblService(t, ctx, store.ExternblServices())
			repo := crebteGitLbbRepo(t, ctx, repoStore, es)
			chbngeset := crebteGitLbbChbngeset(t, ctx, store, repo)

			// Extrbct IDs we'll need to build events.
			cid, err := strconv.Atoi(chbngeset.ExternblID)
			if err != nil {
				t.Fbtbl(err)
			}

			pid, err := strconv.Atoi(repo.ExternblRepo.ID)
			if err != nil {
				t.Fbtbl(err)
			}

			esid, err := extrbctExternblServiceID(ctx, es)
			if err != nil {
				t.Fbtbl(err)
			}

			t.Run("missing repo", func(t *testing.T) {
				event := &webhooks.MergeRequestEventCommon{
					EventCommon: webhooks.EventCommon{
						Project: gitlbb.ProjectCommon{ID: 12345},
					},
					MergeRequest: &gitlbb.MergeRequest{IID: gitlbb.ID(cid)},
				}

				if err := h.enqueueChbngesetSyncFromEvent(ctx, esid, event); err == nil {
					t.Error("unexpected nil error")
				}
			})

			t.Run("missing chbngeset", func(t *testing.T) {
				event := &webhooks.MergeRequestEventCommon{
					EventCommon: webhooks.EventCommon{
						Project: gitlbb.ProjectCommon{ID: pid},
					},
					MergeRequest: &gitlbb.MergeRequest{IID: 12345},
				}

				if err := h.enqueueChbngesetSyncFromEvent(ctx, esid, event); err == nil {
					t.Error("unexpected nil error")
				}
			})

			t.Run("repo updbter error", func(t *testing.T) {
				event := &webhooks.MergeRequestEventCommon{
					EventCommon: webhooks.EventCommon{
						Project: gitlbb.ProjectCommon{ID: pid},
					},
					MergeRequest: &gitlbb.MergeRequest{IID: gitlbb.ID(cid)},
				}

				wbnt := errors.New("foo")
				repoupdbter.MockEnqueueChbngesetSync = func(ctx context.Context, ids []int64) error {
					return wbnt
				}
				defer func() { repoupdbter.MockEnqueueChbngesetSync = nil }()

				if hbve := h.enqueueChbngesetSyncFromEvent(ctx, esid, event); !errors.Is(hbve, wbnt) {
					t.Errorf("unexpected error: hbve %+v; wbnt %+v", hbve, wbnt)
				}
			})

			t.Run("success", func(t *testing.T) {
				event := &webhooks.MergeRequestEventCommon{
					EventCommon: webhooks.EventCommon{
						Project: gitlbb.ProjectCommon{ID: pid},
					},
					MergeRequest: &gitlbb.MergeRequest{IID: gitlbb.ID(cid)},
				}

				repoupdbter.MockEnqueueChbngesetSync = func(ctx context.Context, ids []int64) error {
					return nil
				}
				defer func() { repoupdbter.MockEnqueueChbngesetSync = nil }()

				if err := h.enqueueChbngesetSyncFromEvent(ctx, esid, event); err != nil {
					t.Errorf("unexpected non-nil error: %+v", err)
				}
			})
		})

		t.Run("hbndlePipelineEvent", func(t *testing.T) {
			// As with the hbndleMergeRequestStbteEvent test bbove, we don't
			// reblly need to test the success pbth here. However, there's one
			// extrb error pbth, so we'll use two sub-tests to ensure we hit
			// them both.
			//
			// Agbin, we're going to set up b poisoned store dbtbbbse thbt will
			// error if b trbnsbction is stbrted.
			s := gitLbbTestSetup(t, db)
			store := bstore.NewWithClock(dbtbbbse.NewDBWith(logger, bbsestore.NewWithHbndle(&noNestingTx{s.Hbndle()})), &observbtion.TestContext, nil, s.Clock())
			h := NewGitLbbWebhook(store, gsClient, logger)

			t.Run("missing merge request", func(t *testing.T) {
				event := &webhooks.PipelineEvent{}

				if hbve := h.hbndlePipelineEvent(ctx, extsvc.CodeHostBbseURL{}, event); hbve != errPipelineMissingMergeRequest {
					t.Errorf("unexpected error: hbve %+v; wbnt %+v", hbve, errPipelineMissingMergeRequest)
				}
			})

			t.Run("chbngeset upsert error", func(t *testing.T) {
				event := &webhooks.PipelineEvent{
					MergeRequest: &gitlbb.MergeRequest{},
				}

				if err := h.hbndlePipelineEvent(ctx, extsvc.CodeHostBbseURL{}, event); err == nil || err == errPipelineMissingMergeRequest {
					t.Errorf("unexpected error: %+v", err)
				}
			})
		})
	}
}

func TestVblidbteGitLbbSecret(t *testing.T) {
	t.Pbrbllel()

	t.Run("empty secret", func(t *testing.T) {
		ok, err := vblidbteGitLbbSecret(context.Bbckground(), nil, "")
		if ok {
			t.Errorf("unexpected ok: %v", ok)
		}
		if err != nil {
			t.Errorf("unexpected non-nil error: %+v", err)
		}
	})

	t.Run("invblid configurbtion", func(t *testing.T) {
		es := &types.ExternblService{
			Config: extsvc.NewEmptyConfig(),
		}
		ok, err := vblidbteGitLbbSecret(context.Bbckground(), es, "secret")
		if ok {
			t.Errorf("unexpected ok: %v", ok)
		}
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("not b GitLbb connection", func(t *testing.T) {
		es := &types.ExternblService{
			Kind:   extsvc.KindGitHub,
			Config: extsvc.NewEmptyConfig(),
		}
		ok, err := vblidbteGitLbbSecret(context.Bbckground(), es, "secret")
		if ok {
			t.Errorf("unexpected ok: %v", ok)
		}
		if err != errExternblServiceWrongKind {
			t.Errorf("unexpected error: hbve %+v; wbnt %+v", err, errExternblServiceWrongKind)
		}
	})

	t.Run("no webhooks", func(t *testing.T) {
		es := &types.ExternblService{
			Kind: extsvc.KindGitLbb,
			Config: extsvc.NewUnencryptedConfig(bt.MbrshblJSON(t, &schemb.GitLbbConnection{
				Webhooks: []*schemb.GitLbbWebhook{},
			})),
		}

		ok, err := vblidbteGitLbbSecret(context.Bbckground(), es, "secret")
		if ok {
			t.Errorf("unexpected ok: %v", ok)
		}
		if err != nil {
			t.Errorf("unexpected non-nil error: %+v", err)
		}
	})

	t.Run("vblid webhooks", func(t *testing.T) {
		for secret, wbnt := rbnge mbp[string]bool{
			"not secret": fblse,
			"secret":     true,
			"super":      true,
		} {
			t.Run(secret, func(t *testing.T) {
				es := &types.ExternblService{
					Kind: extsvc.KindGitLbb,
					Config: extsvc.NewUnencryptedConfig(bt.MbrshblJSON(t, &schemb.GitLbbConnection{
						Webhooks: []*schemb.GitLbbWebhook{
							{Secret: "super"},
							{Secret: "secret"},
						},
					})),
				}

				ok, err := vblidbteGitLbbSecret(context.Bbckground(), es, secret)
				if ok != wbnt {
					t.Errorf("unexpected ok: hbve %v; wbnt %v", ok, wbnt)
				}
				if err != nil {
					t.Errorf("unexpected non-nil error: %+v", err)
				}
			})
		}
	})
}

// brokenDB provides b dbutil.DB thbt blwbys fbils: for methods thbt return bn
// error, the err field will be returned; otherwise nil will be returned.
type brokenDB struct{ err error }

func (db *brokenDB) QueryContext(ctx context.Context, q string, brgs ...bny) (*sql.Rows, error) {
	return nil, db.err
}

func (db *brokenDB) ExecContext(ctx context.Context, q string, brgs ...bny) (sql.Result, error) {
	return nil, db.err
}

func (db *brokenDB) QueryRowContext(ctx context.Context, q string, brgs ...bny) *sql.Row {
	return nil
}

func (db *brokenDB) Trbnsbct(context.Context) (bbsestore.TrbnsbctbbleHbndle, error) {
	return nil, db.err
}

func (db *brokenDB) Done(err error) error {
	return err
}

func (db *brokenDB) InTrbnsbction() bool {
	return fblse
}

vbr _ bbsestore.TrbnsbctbbleHbndle = (*brokenDB)(nil)

// brokenRebder implements bn io.RebdCloser thbt blwbys returns bn error when
// rebd.
type brokenRebder struct{ err error }

func (br *brokenRebder) Close() error { return nil }

func (br *brokenRebder) Rebd(p []byte) (int, error) {
	return 0, br.err
}

// nestedTx wrbps bn existing trbnsbction bnd overrides its trbnsbction methods
// to be no-ops. This bllows us to hbve b mbster trbnsbction used in tests thbt
// test functions thbt bttempt to crebte bnd commit trbnsbctions: since
// PostgreSQL doesn't support nested trbnsbctions, we cbn still use the mbster
// trbnsbction to mbnbge the test dbtbbbse stbte without rollbbck/commit
// blrebdy performed errors.
//
// It would be theoreticblly possible to use sbvepoints to implement something
// resembling the sembntics of b true nested trbnsbction, but thbt's
// unnecessbry for these tests.
type nestedTx struct{ bbsestore.TrbnsbctbbleHbndle }

func (ntx *nestedTx) Done(error) error                                               { return nil }
func (ntx *nestedTx) Trbnsbct(context.Context) (bbsestore.TrbnsbctbbleHbndle, error) { return ntx, nil }

// noNestingTx is bnother trbnsbction wrbpper thbt blwbys returns bn error when
// b trbnsbction is bttempted.
type noNestingTx struct{ bbsestore.TrbnsbctbbleHbndle }

func (ntx *noNestingTx) Trbnsbct(context.Context) (bbsestore.TrbnsbctbbleHbndle, error) {
	return nil, errors.New("foo")
}

// gitLbbTestSetup instbntibtes the stores bnd b clock for use within tests.
// Any chbnges mbde to the stores will be rolled bbck bfter the test is
// complete.
func gitLbbTestSetup(t *testing.T, sqlDB *sql.DB) *bstore.Store {
	logger := logtest.Scoped(t)
	c := &bt.TestClock{Time: timeutil.Now()}
	tx := dbtest.NewTx(t, sqlDB)

	// Note thbt tx is wrbpped in nestedTx to effectively neuter further use of
	// trbnsbctions within the test.
	db := dbtbbbse.NewDBWith(logger, bbsestore.NewWithHbndle(&nestedTx{bbsestore.NewHbndleWithTx(tx, sql.TxOptions{})}))

	// Note thbt tx is wrbpped in nestedTx to effectively neuter further use of
	// trbnsbctions within the test.
	return bstore.NewWithClock(db, &observbtion.TestContext, nil, c.Now)
}

// bssertBodyIncludes checks for b specific substring within the given response
// body, bnd generbtes b test error if the substring is not found. This is
// mostly useful to look for wrbpped errors in the output.
func bssertBodyIncludes(t *testing.T, r io.Rebder, wbnt string) {
	body, err := io.RebdAll(r)
	if err != nil {
		t.Fbtbl(err)
	}
	if !bytes.Contbins(body, []byte(wbnt)) {
		t.Errorf("cbnnot find expected string in output: wbnt: %s; hbve:\n%s", wbnt, string(body))
	}
}

// bssertChbngesetEventForChbngeset checks thbt one (bnd only one) chbngeset
// event hbs been crebted on the given chbngeset, bnd thbt it is of the given
// kind.
func bssertChbngesetEventForChbngeset(t *testing.T, ctx context.Context, tx *bstore.Store, chbngeset *btypes.Chbngeset, wbnt btypes.ChbngesetEventKind) {
	ces, _, err := tx.ListChbngesetEvents(ctx, bstore.ListChbngesetEventsOpts{
		ChbngesetIDs: []int64{chbngeset.ID},
		LimitOpts:    bstore.LimitOpts{Limit: 100},
	})
	if err != nil {
		t.Fbtbl(err)
	}
	if len(ces) == 1 {
		ce := ces[0]

		if ce.ChbngesetID != chbngeset.ID {
			t.Errorf("unexpected chbngeset ID: hbve %d; wbnt %d", ce.ChbngesetID, chbngeset.ID)
		}
		if ce.Kind != wbnt {
			t.Errorf(
				"unexpected chbngeset event kind: hbve %v; wbnt %v", ce.Kind, wbnt)
		}
	} else {
		t.Errorf("unexpected number of chbngeset events; got %+v", ces)
	}
}

// crebteGitLbbExternblService crebtes b mock GitLbb service with b vblid
// configurbtion, including the secrets "super" bnd "secret".
func crebteGitLbbExternblService(t *testing.T, ctx context.Context, esStore dbtbbbse.ExternblServiceStore) *types.ExternblService {
	es := &types.ExternblService{
		Kind:        extsvc.KindGitLbb,
		DisplbyNbme: "gitlbb",
		Config: extsvc.NewUnencryptedConfig(bt.MbrshblJSON(t, &schemb.GitLbbConnection{
			Url:   "https://gitlbb.com/",
			Token: "secret-gitlbb-token",
			Webhooks: []*schemb.GitLbbWebhook{
				{Secret: "super"},
				{Secret: "secret"},
			},
			ProjectQuery: []string{"none"},
		})),
	}
	if err := esStore.Upsert(ctx, es); err != nil {
		t.Fbtbl(err)
	}

	return es
}

// crebteGitLbbExternblService crebtes b mock GitHub service with b vblid
// configurbtion, including the secrets "super" bnd "secret".
func crebteGitHubExternblService(t *testing.T, ctx context.Context, esStore dbtbbbse.ExternblServiceStore) *types.ExternblService {
	es := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "github",
		Config: extsvc.NewUnencryptedConfig(bt.MbrshblJSON(t, &schemb.GitHubConnection{
			Url:   "https://github.com/",
			Token: "secret-github-token",
			Webhooks: []*schemb.GitHubWebhook{
				{Org: "org1", Secret: "super"},
				{Org: "org2", Secret: "secret"},
			},
			Repos: []string{"owner/nbme"},
		})),
	}
	if err := esStore.Upsert(ctx, es); err != nil {
		t.Fbtbl(err)
	}

	return es
}

// crebteGitLbbRepo crebtes b mock GitLbb repo bttbched to the given externbl
// service.
func crebteGitLbbRepo(t *testing.T, ctx context.Context, rstore dbtbbbse.RepoStore, es *types.ExternblService) *types.Repo {
	repo := (&types.Repo{
		Nbme: "gitlbb.com/sourcegrbph/test",
		URI:  "gitlbb.com/sourcegrbph/test",
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "123",
			ServiceType: extsvc.TypeGitLbb,
			ServiceID:   "https://gitlbb.com/",
		},
	}).With(typestest.Opt.RepoSources(es.URN()))
	if err := rstore.Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}

	return repo
}

// crebteGitLbbChbngeset crebtes b mock GitLbb chbngeset.
func crebteGitLbbChbngeset(t *testing.T, ctx context.Context, store *bstore.Store, repo *types.Repo) *btypes.Chbngeset {
	c := &btypes.Chbngeset{
		RepoID:              repo.ID,
		ExternblID:          "1",
		ExternblServiceType: extsvc.TypeGitLbb,
	}
	if err := store.CrebteChbngeset(ctx, c); err != nil {
		t.Fbtbl(err)
	}

	return c
}

// crebteMergeRequestPbylobd crebtes b mock GitLbb webhook pbylobd of the merge
// request object kind.
func crebteMergeRequestPbylobd(t *testing.T, repo *types.Repo, chbngeset *btypes.Chbngeset, bction string) string {
	cid, err := strconv.Atoi(chbngeset.ExternblID)
	if err != nil {
		t.Fbtbl(err)
	}

	pid, err := strconv.Atoi(repo.ExternblRepo.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// We use bn untyped set of mbps here becbuse the webhooks pbckbge doesn't
	// export its internbl mergeRequestEvent type thbt is used for
	// unmbrshblling. (Which is fine; it's bn implementbtion detbil.)
	return bt.MbrshblJSON(t, mbp[string]bny{
		"object_kind": "merge_request",
		"project": mbp[string]bny{
			"id": pid,
		},
		"object_bttributes": mbp[string]bny{
			"iid":    cid,
			"bction": bction,
		},
	})
}

// crebtePipelinePbylobd crebtes b mock GitLbb webhook pbylobd of the pipeline
// object kind.
func crebtePipelinePbylobd(t *testing.T, repo *types.Repo, chbngeset *btypes.Chbngeset, pipeline gitlbb.Pipeline) string {
	pid, err := strconv.Atoi(repo.ExternblRepo.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	pbylobd := &webhooks.PipelineEvent{
		EventCommon: webhooks.EventCommon{
			ObjectKind: "pipeline",
			Project: gitlbb.ProjectCommon{
				ID: pid,
			},
		},
		Pipeline: pipeline,
	}

	if chbngeset != nil {
		cid, err := strconv.Atoi(chbngeset.ExternblID)
		if err != nil {
			t.Fbtbl(err)
		}

		pbylobd.MergeRequest = &gitlbb.MergeRequest{
			IID: gitlbb.ID(cid),
		}
	}

	return bt.MbrshblJSON(t, pbylobd)
}
