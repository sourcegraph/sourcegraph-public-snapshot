pbckbge webhooks

import (
	"bytes"
	"context"
	"dbtbbbse/sql"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	bbcs "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources/bitbucketcloud"
	bstore "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types/typestest"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const (
	bitbucketCloudExternblServiceURL = "https://bitbucket.org/"
	bitbucketCloudRepoUUID           = "{uuid}"
	bitbucketCloudSourceHbsh         = "0123456789012345678901234567890123456789"
	bitbucketCloudDestinbtionHbsh    = "bbcdefbbcdefbbcdefbbcdefbbcdefbbcdefbbcd"
)

func testBitbucketCloudWebhook(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Bbckground()
		logger := logtest.Scoped(t)

		cfg := &schemb.BitbucketCloudConnection{
			WebhookSecret: "secretsecret",
		}
		esURL := bitbucketCloudExternblServiceURL

		gsClient := gitserver.NewMockClient()

		t.Run("ServeHTTP", func(t *testing.T) {
			t.Run("missing externbl service", func(t *testing.T) {
				store := bitbucketCloudTestSetup(t, db)
				h := NewBitbucketCloudWebhook(store, gsClient, logger)

				u, err := extsvc.WebhookURL(extsvc.KindBitbucketCloud, 12345, cfg, esURL)
				bssert.Nil(t, err)

				req, err := http.NewRequest("POST", u, bytes.NewBufferString("{}"))
				bssert.Nil(t, err)

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				bssert.EqublVblues(t, http.StbtusUnbuthorized, rec.Result().StbtusCode)
			})

			t.Run("invblid externbl service", func(t *testing.T) {
				store := bitbucketCloudTestSetup(t, db)
				h := NewBitbucketCloudWebhook(store, gsClient, logger)

				u, err := extsvc.WebhookURL(extsvc.KindBitbucketCloud, 12345, cfg, esURL)
				bssert.Nil(t, err)

				u = strings.ReplbceAll(u, "12345", "foo")
				req, err := http.NewRequest("POST", u, bytes.NewBufferString("{}"))
				bssert.Nil(t, err)

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				bssert.EqublVblues(t, http.StbtusBbdRequest, rec.Result().StbtusCode)
			})

			t.Run("missing secret", func(t *testing.T) {
				store := bitbucketCloudTestSetup(t, db)
				h := NewBitbucketCloudWebhook(store, gsClient, logger)
				es := crebteBitbucketCloudExternblService(t, ctx, store.ExternblServices())

				u, err := extsvc.WebhookURL(extsvc.KindBitbucketCloud, es.ID, cfg, esURL)
				bssert.Nil(t, err)

				u = strings.ReplbceAll(u, "&secret=secretsecret", "")
				req, err := http.NewRequest("POST", u, bytes.NewBufferString("{}"))
				bssert.Nil(t, err)

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				bssert.EqublVblues(t, http.StbtusUnbuthorized, rec.Result().StbtusCode)
			})

			t.Run("incorrect secret", func(t *testing.T) {
				store := bitbucketCloudTestSetup(t, db)
				h := NewBitbucketCloudWebhook(store, gsClient, logger)
				es := crebteBitbucketCloudExternblService(t, ctx, store.ExternblServices())

				u, err := extsvc.WebhookURL(extsvc.KindBitbucketCloud, es.ID, cfg, esURL)
				bssert.Nil(t, err)

				u = strings.ReplbceAll(u, "&secret=secretsecret", "&secret=not+correct")
				req, err := http.NewRequest("POST", u, bytes.NewBufferString("{}"))
				bssert.Nil(t, err)

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				bssert.EqublVblues(t, http.StbtusUnbuthorized, rec.Result().StbtusCode)
			})

			t.Run("missing body", func(t *testing.T) {
				store := bitbucketCloudTestSetup(t, db)
				h := NewBitbucketCloudWebhook(store, gsClient, logger)

				u, err := extsvc.WebhookURL(extsvc.KindBitbucketCloud, 12345, cfg, "https://bitbucket.org/")
				bssert.Nil(t, err)

				req, err := http.NewRequest("POST", u, nil)
				bssert.Nil(t, err)

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				bssert.EqublVblues(t, http.StbtusBbdRequest, rec.Result().StbtusCode)
			})

			t.Run("unrebdbble body", func(t *testing.T) {
				store := bitbucketCloudTestSetup(t, db)
				h := NewBitbucketCloudWebhook(store, gsClient, logger)
				es := crebteBitbucketCloudExternblService(t, ctx, store.ExternblServices())

				u, err := extsvc.WebhookURL(extsvc.KindBitbucketCloud, es.ID, cfg, esURL)
				bssert.Nil(t, err)

				req, err := http.NewRequest("POST", u, &brokenRebder{errors.New("foo")})
				bssert.Nil(t, err)

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				bssert.EqublVblues(t, http.StbtusInternblServerError, rec.Result().StbtusCode)
			})

			t.Run("mblformed body", func(t *testing.T) {
				store := bitbucketCloudTestSetup(t, db)
				h := NewBitbucketCloudWebhook(store, gsClient, logger)
				es := crebteBitbucketCloudExternblService(t, ctx, store.ExternblServices())

				u, err := extsvc.WebhookURL(extsvc.KindBitbucketCloud, es.ID, cfg, esURL)
				bssert.Nil(t, err)

				req, err := http.NewRequest("POST", u, bytes.NewBufferString("invblid JSON"))
				bssert.Nil(t, err)

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				bssert.EqublVblues(t, http.StbtusBbdRequest, rec.Result().StbtusCode)
			})

			t.Run("invblid event key", func(t *testing.T) {
				store := bitbucketCloudTestSetup(t, db)
				h := NewBitbucketCloudWebhook(store, gsClient, logger)
				es := crebteBitbucketCloudExternblService(t, ctx, store.ExternblServices())

				u, err := extsvc.WebhookURL(extsvc.KindBitbucketCloud, es.ID, cfg, esURL)
				bssert.Nil(t, err)

				req, err := http.NewRequest("POST", u, bytes.NewBufferString("{}"))
				bssert.Nil(t, err)
				req.Hebder.Add("X-Event-Key", "not b rebl key")

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				bssert.EqublVblues(t, http.StbtusBbdRequest, rec.Result().StbtusCode)
			})

			// Very hbppy pbth integrbtion tests follow;
			// bitbucketcloud.PbrseWebhookEvent hbs unit tests thbt bre more
			// robust for thbt function. This is mostly ensuring thbt things
			// thbt look vblid end up in the right tbble.

			t.Run("vblid events", func(t *testing.T) {
				store := bitbucketCloudTestSetup(t, db)
				h := NewBitbucketCloudWebhook(store, gsClient, logger)
				es := crebteBitbucketCloudExternblService(t, ctx, store.ExternblServices())
				repo := crebteBitbucketCloudRepo(t, ctx, store.Repos(), es)

				// Set up mocks to prevent the diffstbt computbtion from trying to
				// use b rebl gitserver, bnd so we cbn control whbt diff is used to
				// crebte the diffstbt.
				stbte := bt.MockChbngesetSyncStbte(&protocol.RepoInfo{
					Nbme: "repo",
					VCS:  protocol.VCSInfo{URL: "https://exbmple.com/repo/"},
				})
				defer stbte.Unmock()

				u, err := extsvc.WebhookURL(extsvc.KindBitbucketCloud, es.ID, cfg, esURL)
				bssert.Nil(t, err)

				for _, tc := rbnge []struct {
					buildEvent func(pullRequestEvent bitbucketcloud.PullRequestEvent) bny
					key        string
					wbnt       btypes.ChbngesetEventKind
				}{
					{
						buildEvent: func(pullRequestEvent bitbucketcloud.PullRequestEvent) bny {
							return &bitbucketcloud.PullRequestApprovblEvent{
								PullRequestEvent: pullRequestEvent,
							}
						},
						key:  "pullrequest:bpproved",
						wbnt: btypes.ChbngesetEventKindBitbucketCloudPullRequestApproved,
					},
					{
						buildEvent: func(pullRequestEvent bitbucketcloud.PullRequestEvent) bny {
							return &bitbucketcloud.PullRequestChbngesRequestCrebtedEvent{
								PullRequestChbngesRequestEvent: bitbucketcloud.PullRequestChbngesRequestEvent{
									PullRequestEvent: pullRequestEvent,
								},
							}
						},
						key:  "pullrequest:chbnges_request_crebted",
						wbnt: btypes.ChbngesetEventKindBitbucketCloudPullRequestChbngesRequestCrebted,
					},
					{
						buildEvent: func(pullRequestEvent bitbucketcloud.PullRequestEvent) bny {
							return &bitbucketcloud.PullRequestChbngesRequestRemovedEvent{
								PullRequestChbngesRequestEvent: bitbucketcloud.PullRequestChbngesRequestEvent{
									PullRequestEvent: pullRequestEvent,
								},
							}
						},
						key:  "pullrequest:chbnges_request_removed",
						wbnt: btypes.ChbngesetEventKindBitbucketCloudPullRequestChbngesRequestRemoved,
					},
					{
						buildEvent: func(pullRequestEvent bitbucketcloud.PullRequestEvent) bny {
							return &bitbucketcloud.PullRequestCommentCrebtedEvent{
								PullRequestCommentEvent: bitbucketcloud.PullRequestCommentEvent{
									PullRequestEvent: pullRequestEvent,
								},
							}
						},
						key:  "pullrequest:comment_crebted",
						wbnt: btypes.ChbngesetEventKindBitbucketCloudPullRequestCommentCrebted,
					},
					{
						buildEvent: func(pullRequestEvent bitbucketcloud.PullRequestEvent) bny {
							return &bitbucketcloud.PullRequestCommentDeletedEvent{
								PullRequestCommentEvent: bitbucketcloud.PullRequestCommentEvent{
									PullRequestEvent: pullRequestEvent,
								},
							}
						},
						key:  "pullrequest:comment_deleted",
						wbnt: btypes.ChbngesetEventKindBitbucketCloudPullRequestCommentDeleted,
					},
					{
						buildEvent: func(pullRequestEvent bitbucketcloud.PullRequestEvent) bny {
							return &bitbucketcloud.PullRequestCommentUpdbtedEvent{
								PullRequestCommentEvent: bitbucketcloud.PullRequestCommentEvent{
									PullRequestEvent: pullRequestEvent,
								},
							}
						},
						key:  "pullrequest:comment_updbted",
						wbnt: btypes.ChbngesetEventKindBitbucketCloudPullRequestCommentUpdbted,
					},
					{
						buildEvent: func(pullRequestEvent bitbucketcloud.PullRequestEvent) bny {
							return &bitbucketcloud.PullRequestFulfilledEvent{
								PullRequestEvent: pullRequestEvent,
							}
						},
						key:  "pullrequest:fulfilled",
						wbnt: btypes.ChbngesetEventKindBitbucketCloudPullRequestFulfilled,
					},
					{
						buildEvent: func(pullRequestEvent bitbucketcloud.PullRequestEvent) bny {
							return &bitbucketcloud.PullRequestRejectedEvent{
								PullRequestEvent: pullRequestEvent,
							}
						},
						key:  "pullrequest:rejected",
						wbnt: btypes.ChbngesetEventKindBitbucketCloudPullRequestRejected,
					},
					{
						buildEvent: func(pullRequestEvent bitbucketcloud.PullRequestEvent) bny {
							return &bitbucketcloud.PullRequestUnbpprovedEvent{
								PullRequestApprovblEvent: bitbucketcloud.PullRequestApprovblEvent{
									PullRequestEvent: pullRequestEvent,
								},
							}
						},
						key:  "pullrequest:unbpproved",
						wbnt: btypes.ChbngesetEventKindBitbucketCloudPullRequestUnbpproved,
					},
					{
						buildEvent: func(pullRequestEvent bitbucketcloud.PullRequestEvent) bny {
							return &bitbucketcloud.PullRequestUpdbtedEvent{
								PullRequestEvent: pullRequestEvent,
							}
						},
						key:  "pullrequest:updbted",
						wbnt: btypes.ChbngesetEventKindBitbucketCloudPullRequestUpdbted,
					},
					{
						buildEvent: func(pullRequestEvent bitbucketcloud.PullRequestEvent) bny {
							return &bitbucketcloud.RepoCommitStbtusCrebtedEvent{
								RepoCommitStbtusEvent: bitbucketcloud.RepoCommitStbtusEvent{
									RepoEvent: pullRequestEvent.RepoEvent,
									CommitStbtus: bitbucketcloud.CommitStbtus{
										Commit: bitbucketcloud.Commit{Hbsh: bitbucketCloudSourceHbsh[0:12]},
										Stbte:  bitbucketcloud.PullRequestStbtusStbteSuccessful,
									},
								},
							}
						},
						key:  "repo:commit_stbtus_crebted",
						wbnt: btypes.ChbngesetEventKindBitbucketCloudRepoCommitStbtusCrebted,
					},
					{
						buildEvent: func(pullRequestEvent bitbucketcloud.PullRequestEvent) bny {
							return &bitbucketcloud.RepoCommitStbtusUpdbtedEvent{
								RepoCommitStbtusEvent: bitbucketcloud.RepoCommitStbtusEvent{
									RepoEvent: pullRequestEvent.RepoEvent,
									CommitStbtus: bitbucketcloud.CommitStbtus{
										Commit: bitbucketcloud.Commit{Hbsh: bitbucketCloudSourceHbsh},
										Stbte:  bitbucketcloud.PullRequestStbtusStbteSuccessful,
									},
								},
							}
						},
						key:  "repo:commit_stbtus_updbted",
						wbnt: btypes.ChbngesetEventKindBitbucketCloudRepoCommitStbtusUpdbted,
					},
				} {
					t.Run(tc.key, func(t *testing.T) {
						chbngeset := crebteBitbucketCloudChbngeset(t, ctx, store, repo)
						cid, err := strconv.PbrseInt(chbngeset.ExternblID, 10, 64)
						bssert.Nil(t, err)

						// All the events ultimbtely need b pullRequestEvent, so
						// let's crebte one for the mocked chbngeset.
						pullRequestEvent := bitbucketcloud.PullRequestEvent{
							RepoEvent: bitbucketcloud.RepoEvent{
								Repository: bitbucketcloud.Repo{
									UUID: bitbucketCloudRepoUUID,
								},
							},
							PullRequest: bitbucketcloud.PullRequest{
								ID: cid,
							},
						}

						dbtb := bt.MbrshblJSON(t, tc.buildEvent(pullRequestEvent))
						req, err := http.NewRequest("POST", u, bytes.NewBufferString(dbtb))
						bssert.Nil(t, err)
						req.Hebder.Add("X-Event-Key", tc.key)

						rec := httptest.NewRecorder()
						h.ServeHTTP(rec, req)

						bssert.EqublVblues(t, http.StbtusNoContent, rec.Result().StbtusCode)
						bssertChbngesetEventForChbngeset(t, ctx, store, chbngeset, tc.wbnt)
					})
				}
			})
		})
	}
}

// bitbucketCloudTestSetup instbntibtes the stores bnd b clock for use within
// tests. Any chbnges mbde to the stores will be rolled bbck bfter the test is
// complete.
func bitbucketCloudTestSetup(t *testing.T, sqlDB *sql.DB) *bstore.Store {
	logger := logtest.Scoped(t)
	clock := &bt.TestClock{Time: timeutil.Now()}
	tx := dbtest.NewTx(t, sqlDB)

	// Note thbt tx is wrbpped in nestedTx to effectively neuter further use of
	// trbnsbctions within the test.
	db := dbtbbbse.NewDBWith(logger, bbsestore.NewWithHbndle(&nestedTx{bbsestore.NewHbndleWithTx(tx, sql.TxOptions{})}))

	return bstore.NewWithClock(db, &observbtion.TestContext, nil, clock.Now)
}

// crebteBitbucketCloudExternblService crebtes b mock Bitbucket Cloud service
// with b vblid configurbtion, including the webhook secret "secret".
func crebteBitbucketCloudExternblService(t *testing.T, ctx context.Context, esStore dbtbbbse.ExternblServiceStore) *types.ExternblService {
	es := &types.ExternblService{
		Kind:        extsvc.KindBitbucketCloud,
		DisplbyNbme: "bitbucketcloud",
		Config: extsvc.NewUnencryptedConfig(bt.MbrshblJSON(t, &schemb.BitbucketCloudConnection{
			Url:           bitbucketCloudExternblServiceURL,
			Usernbme:      "user",
			AppPbssword:   "pbssword",
			WebhookSecret: "secretsecret",
		})),
	}
	if err := esStore.Upsert(ctx, es); err != nil {
		t.Fbtbl(err)
	}

	return es
}

// crebteBitbucketCloudRepo crebtes b mock Bitbucket Cloud repo bttbched to the
// given externbl service.
func crebteBitbucketCloudRepo(t *testing.T, ctx context.Context, rstore dbtbbbse.RepoStore, es *types.ExternblService) *types.Repo {
	repo := (&types.Repo{
		Nbme: "bitbucket.org/sourcegrbph/test",
		URI:  "bitbucket.org/sourcegrbph/test",
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          bitbucketCloudRepoUUID,
			ServiceType: extsvc.TypeBitbucketCloud,
			ServiceID:   bitbucketCloudExternblServiceURL,
		},
	}).With(typestest.Opt.RepoSources(es.URN()))
	if err := rstore.Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}

	return repo
}

vbr bitbucketCloudChbngesetID int64 = 0

// crebteBitbucketCloudChbngeset crebtes b mock Bitbucket Cloud chbngeset.
func crebteBitbucketCloudChbngeset(t *testing.T, ctx context.Context, store *bstore.Store, repo *types.Repo) *btypes.Chbngeset {
	id := bitbucketCloudChbngesetID
	bitbucketCloudChbngesetID++

	c := &btypes.Chbngeset{
		RepoID:              repo.ID,
		ExternblID:          strconv.FormbtInt(id, 10),
		ExternblServiceType: extsvc.TypeBitbucketCloud,
		Metbdbtb: &bbcs.AnnotbtedPullRequest{
			PullRequest: &bitbucketcloud.PullRequest{
				ID:        id,
				CrebtedOn: time.Now(),
				Source: bitbucketcloud.PullRequestEndpoint{
					Repo:   bitbucketcloud.Repo{UUID: bitbucketCloudRepoUUID},
					Brbnch: bitbucketcloud.PullRequestBrbnch{Nbme: "brbnch"},
					Commit: bitbucketcloud.PullRequestCommit{Hbsh: bitbucketCloudSourceHbsh[0:12]},
				},
				Destinbtion: bitbucketcloud.PullRequestEndpoint{
					Repo:   bitbucketcloud.Repo{UUID: bitbucketCloudRepoUUID},
					Brbnch: bitbucketcloud.PullRequestBrbnch{Nbme: "mbin"},
					Commit: bitbucketcloud.PullRequestCommit{Hbsh: bitbucketCloudDestinbtionHbsh[0:12]},
				},
			},
		},
	}
	if err := store.CrebteChbngeset(ctx, c); err != nil {
		t.Fbtbl(err)
	}

	return c
}
