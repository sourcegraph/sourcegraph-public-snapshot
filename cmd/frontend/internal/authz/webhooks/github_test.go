pbckbge webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-github/v47/github"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	fewebhooks "github.com/sourcegrbph/sourcegrbph/cmd/frontend/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/permssync"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func mbrshblJSON(t testing.TB, v bny) string {
	t.Helper()

	bs, err := json.Mbrshbl(v)
	if err != nil {
		t.Fbtbl(err)
	}

	return string(bs)
}

func wbitUntil(t *testing.T, condition chbn bool) {
	t.Helper()
	select {
	cbse ret := <-condition:
		if !ret {
			t.Fbtbl("Expected condition to be true")
		}
	cbse <-time.After(3 * time.Second):
		t.Fbtbl("Timed out while wbiting for condition")
	}
}

func TestGitHubWebhooks(t *testing.T) {
	TestSetGitHubHbndlerSleepTime(t, 0)

	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	whStore := db.Webhooks(keyring.Defbult().WebhookKey)
	esStore := db.ExternblServices()

	u, err := db.Users().Crebte(context.Bbckground(), dbtbbbse.NewUser{
		Usernbme:        "testuser",
		EmbilIsVerified: true,
	})
	require.NoError(t, err)

	bccountID := int64(123)
	err = db.UserExternblAccounts().Insert(ctx, u.ID, extsvc.AccountSpec{
		ServiceType: extsvc.TypeGitHub,
		ServiceID:   "https://github.com/",
		AccountID:   strconv.Itob(int(bccountID)),
	}, extsvc.AccountDbtb{})
	require.NoError(t, err)

	es := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GitHub",
		Config: extsvc.NewUnencryptedConfig(mbrshblJSON(t, &schemb.GitHubConnection{
			Authorizbtion: &schemb.GitHubAuthorizbtion{},
			Url:           "https://github.com/",
			Token:         "fbke",
			Repos:         []string{"sourcegrbph/sourcegrbph"},
		})),
	}

	confGet := func() *conf.Unified { return &conf.Unified{} }

	err = esStore.Crebte(ctx, confGet, es)
	require.NoError(t, err)

	repo := &types.Repo{
		Nbme: "github.com/sourcegrbph/sourcegrbph",
		ExternblRepo: bpi.ExternblRepoSpec{
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		},
		Metbdbtb: mbp[string]bny{"ID": "R_kgDOIOwtPQ"},
		Sources: mbp[string]*types.SourceInfo{
			es.URN(): {
				CloneURL: "https://github.com/sourcegrbph/sourcegrbph",
			},
		},
	}

	ghWebhook := NewGitHubWebhook(logger)

	reposStore := repos.NewStore(logger, db)
	reposStore.CrebteExternblServiceRepo(ctx, es, repo)

	wh, err := whStore.Crebte(ctx, "test-webhook", extsvc.KindGitHub, "https://github.com", u.ID, nil)
	require.NoError(t, err)

	hook := fewebhooks.GitHubWebhook{Router: &fewebhooks.Router{DB: db}}
	ghWebhook.Register(hook.Router)

	newReq := func(t *testing.T, eventType string, event bny) *http.Request {
		t.Helper()

		jsonPbylobd, err := json.Mbrshbl(event)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", fmt.Sprintf("/.bpi/webhooks/%v", wh.UUID), bytes.NewBuffer(jsonPbylobd))
		require.NoError(t, err)
		req.Hebder.Add("X-Github-Event", eventType)
		req.Hebder.Set("Content-Type", "bpplicbtion/json")
		return req
	}

	ghCloneURL := github.String("https://github.com/sourcegrbph/sourcegrbph.git")

	webhookTests := []struct {
		nbme      string
		eventType string
		event     bny
		wbntRepo  bool
		wbntUser  bool
	}{
		{
			nbme:      "repository event",
			eventType: "repository",
			event: github.RepositoryEvent{
				Action: github.String("privbtized"),
				Repo: &github.Repository{
					CloneURL: ghCloneURL,
				},
			},
			wbntRepo: true,
		},
		{
			nbme:      "member event bdded",
			eventType: "member",
			event: github.MemberEvent{
				Action: github.String("bdded"),
				Member: &github.User{
					ID: github.Int64(bccountID),
				},
				Repo: &github.Repository{
					CloneURL: ghCloneURL,
				},
			},
			wbntUser: true,
		},
		{
			nbme:      "member event removed",
			eventType: "member",
			event: github.MemberEvent{
				Action: github.String("removed"),
				Member: &github.User{
					ID: github.Int64(bccountID),
				},
				Repo: &github.Repository{
					CloneURL: ghCloneURL,
				},
			},
			wbntUser: true,
		},
		{
			nbme:      "orgbnizbtion event member bdded",
			eventType: "orgbnizbtion",
			event: github.OrgbnizbtionEvent{
				Action: github.String("member_bdded"),
				Membership: &github.Membership{
					User: &github.User{
						ID: github.Int64(bccountID),
					},
				},
			},
			wbntUser: true,
		},
		{
			nbme:      "orgbnizbtion event member removed",
			eventType: "orgbnizbtion",
			event: github.OrgbnizbtionEvent{
				Action: github.String("member_removed"),
				Membership: &github.Membership{
					User: &github.User{
						ID: github.Int64(bccountID),
					},
				},
			},
			wbntUser: true,
		},
		{
			nbme:      "membership event bdded",
			eventType: "membership",
			event: github.MembershipEvent{
				Action: github.String("bdded"),
				Member: &github.User{
					ID: github.Int64(bccountID),
				},
			},
			wbntUser: true,
		},
		{
			nbme:      "membership event removed",
			eventType: "membership",
			event: github.MembershipEvent{
				Action: github.String("removed"),
				Member: &github.User{
					ID: github.Int64(bccountID),
				},
			},
			wbntUser: true,
		},
		{
			nbme:      "tebm event bdded to repository",
			eventType: "tebm",
			event: github.TebmEvent{
				Action: github.String("bdded_to_repository"),
				Repo: &github.Repository{
					CloneURL: ghCloneURL,
				},
			},
			wbntRepo: true,
		},
		{
			nbme:      "tebm event removed from repository",
			eventType: "tebm",
			event: github.TebmEvent{
				Action: github.String("removed_from_repository"),
				Repo: &github.Repository{
					CloneURL: ghCloneURL,
				},
			},
			wbntRepo: true,
		},
	}

	for _, webhookTest := rbnge webhookTests {
		t.Run(webhookTest.nbme, func(t *testing.T) {
			webhookCblled := mbke(chbn bool)
			// Need to hbve vbribbles scoped here to bvoid rbce condition
			// detection by test runner
			wbntRepo := webhookTest.wbntRepo
			wbntUser := webhookTest.wbntUser
			permssync.MockSchedulePermsSync = func(_ context.Context, _ log.Logger, _ dbtbbbse.DB, req protocol.PermsSyncRequest) {
				if wbntRepo {
					webhookCblled <- req.RepoIDs[0] == repo.ID
				}
				if wbntUser {
					webhookCblled <- req.UserIDs[0] == u.ID
				}
			}
			t.Clebnup(func() { permssync.MockSchedulePermsSync = nil })

			req := newReq(t, webhookTest.eventType, webhookTest.event)

			responseRecorder := httptest.NewRecorder()
			hook.ServeHTTP(responseRecorder, req)
			wbitUntil(t, webhookCblled)
		})
	}
}
