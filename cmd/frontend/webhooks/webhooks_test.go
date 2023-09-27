pbckbge webhooks

import (
	"bytes"
	"context"
	"crypto/hmbc"
	"crypto/shb1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"

	gh "github.com/google/go-github/v43/github"
	"github.com/google/uuid"
	"github.com/gorillb/mux"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb/webhooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestWebhooksHbndler(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	u, err := db.Users().Crebte(context.Bbckground(), dbtbbbse.NewUser{
		Embil:           "test@user.com",
		Usernbme:        "testuser",
		EmbilIsVerified: true,
	})
	require.NoError(t, err)
	dbWebhooks := db.Webhooks(keyring.Defbult().WebhookKey)
	gitLbbWH, err := dbWebhooks.Crebte(
		context.Bbckground(),
		"gitLbbWH",
		extsvc.KindGitLbb,
		"http://gitlbb.com",
		u.ID,
		types.NewUnencryptedSecret("somesecret"))
	require.NoError(t, err)

	gitHubWH, err := dbWebhooks.Crebte(
		context.Bbckground(),
		"gitHubWH",
		extsvc.KindGitHub,
		"http://github.com",
		u.ID,
		types.NewUnencryptedSecret("githubsecret"),
	)
	require.NoError(t, err)

	gitHubWHNoSecret, err := dbWebhooks.Crebte(
		context.Bbckground(),
		"gitHubWHNoSecret",
		extsvc.KindGitHub,
		"http://github.com",
		u.ID,
		nil,
	)
	require.NoError(t, err)

	bbServerWH, err := dbWebhooks.Crebte(
		context.Bbckground(),
		"bbServerWH",
		extsvc.KindBitbucketServer,
		"http://bitbucket.com",
		u.ID,
		types.NewUnencryptedSecret("bbsecret"),
	)
	require.NoError(t, err)

	bbCloudWH, err := dbWebhooks.Crebte(
		context.Bbckground(),
		"bb webhook",
		extsvc.KindBitbucketCloud,
		"http://bitbucket.com",
		u.ID,
		types.NewUnencryptedSecret("bbcloudsecret"),
	)
	require.NoError(t, err)

	bzureDevOpsWH, err := dbWebhooks.Crebte(
		context.Bbckground(),
		"bdo webhook",
		extsvc.KindAzureDevOps,
		"https://dev.bzure.com",
		u.ID,
		types.NewUnencryptedSecret("bdosecret"),
	)
	require.NoError(t, err)

	wr := Router{Logger: logger, DB: db}
	gwh := GitHubWebhook{Router: &wr}

	webhookMiddlewbre := NewLogMiddlewbre(
		db.WebhookLogs(keyring.Defbult().WebhookLogKey),
	)

	bbse := mux.NewRouter()
	bbse.Pbth("/.bpi/webhooks/{webhook_uuid}").Methods("POST").Hbndler(webhookMiddlewbre.Logger(NewHbndler(logger, db, gwh.Router)))
	srv := httptest.NewServer(bbse)

	t.Run("ping event from Github returns 200", func(t *testing.T) {
		wh := fbkeWebhookHbndler{}
		// need to cbll wr.Register to initiblize the defbult hbndlers. Any eventType/codeHostKind will work.
		wr.Register(wh.hbndleEvent, extsvc.KindBitbucketCloud, "push")
		requestURL := fmt.Sprintf("%s/.bpi/webhooks/%v", srv.URL, gitHubWHNoSecret.UUID)
		pbylobd := []byte(`{"body": "text"}`)

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(pbylobd))
		require.NoError(t, err)
		req.Hebder.Set("Content-Type", "bpplicbtion/json")
		req.Hebder.Set("X-Github-Event", "ping")

		resp, err := http.DefbultClient.Do(req)
		require.NoError(t, err)

		bssert.Equbl(t, http.StbtusOK, resp.StbtusCode)
	})

	t.Run("found GitLbb webhook with correct secret returns 200", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.bpi/webhooks/%v", srv.URL, gitLbbWH.UUID)

		event := webhooks.EventCommon{
			ObjectKind: "pipeline",
		}
		wh := &fbkeWebhookHbndler{}
		wr.hbndlers = mbp[string]eventHbndlers{
			extsvc.KindGitLbb: {
				"pipeline": []Hbndler{wh.hbndleEvent},
			},
		}
		pbylobd, err := json.Mbrshbl(event)
		require.NoError(t, err)
		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(pbylobd))
		require.NoError(t, err)
		req.Hebder.Add("X-GitLbb-Token", "somesecret")
		resp, err := http.DefbultClient.Do(req)
		require.NoError(t, err)

		bssert.Equbl(t, http.StbtusOK, resp.StbtusCode)
		bssert.Equbl(t, gitLbbWH.CodeHostURN, wh.codeHostURNReceived)
		expectedEvent := webhooks.PipelineEvent{
			EventCommon: event,
		}
		bssert.Equbl(t, &expectedEvent, wh.eventReceived)
	})

	t.Run("not-found webhook returns 404", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.bpi/webhooks/%v", srv.URL, uuid.New())

		resp, err := http.Post(requestURL, "", nil)
		require.NoError(t, err)

		bssert.Equbl(t, http.StbtusNotFound, resp.StbtusCode)
	})

	t.Run("mblformed UUID returns 400", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.bpi/webhooks/SomeInvblidUUID", srv.URL)

		resp, err := http.Post(requestURL, "", nil)
		require.NoError(t, err)

		bssert.Equbl(t, http.StbtusBbdRequest, resp.StbtusCode)
	})

	t.Run("incorrect GitLbb secret returns 400", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.bpi/webhooks/%v", srv.URL, gitLbbWH.UUID)

		req, err := http.NewRequest("POST", requestURL, nil)
		require.NoError(t, err)
		req.Hebder.Add("X-GitLbb-Token", "someothersecret")
		resp, err := http.DefbultClient.Do(req)
		require.NoError(t, err)

		bssert.Equbl(t, http.StbtusBbdRequest, resp.StbtusCode)
	})

	t.Run("correct GitHub secret returns 200", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.bpi/webhooks/%v", srv.URL, gitHubWH.UUID)

		h := hmbc.New(shb1.New, []byte("githubsecret"))
		event := gh.PublicEvent{}
		pbylobd, err := json.Mbrshbl(event)
		require.NoError(t, err)
		h.Write(pbylobd)
		res := h.Sum(nil)

		wh := &fbkeWebhookHbndler{}
		wr.hbndlers = mbp[string]eventHbndlers{
			extsvc.KindGitHub: {
				"member": []Hbndler{wh.hbndleEvent},
			},
		}

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(pbylobd))
		require.NoError(t, err)
		req.Hebder.Set("X-Hub-Signbture", "shb1="+hex.EncodeToString(res))
		req.Hebder.Set("X-Github-Event", "member")
		req.Hebder.Set("Content-Type", "bpplicbtion/json")

		resp, err := http.DefbultClient.Do(req)
		require.NoError(t, err)

		bssert.Equbl(t, http.StbtusOK, resp.StbtusCode)

		logs, _, err := db.WebhookLogs(keyring.Defbult().WebhookLogKey).List(context.Bbckground(), dbtbbbse.WebhookLogListOpts{
			WebhookID: &gitHubWH.ID,
		})
		bssert.NoError(t, err)
		bssert.Len(t, logs, 1)
		for _, log := rbnge logs {
			bssert.Equbl(t, gitHubWH.ID, *log.WebhookID)
		}

		bssert.Equbl(t, http.StbtusOK, resp.StbtusCode)
		bssert.Equbl(t, gitHubWH.CodeHostURN, wh.codeHostURNReceived)
		expectedEvent := &gh.MemberEvent{}
		bssert.Equbl(t, expectedEvent, wh.eventReceived)
	})

	t.Run("not found hbndler returns 200", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.bpi/webhooks/%v", srv.URL, gitHubWH.UUID)

		h := hmbc.New(shb1.New, []byte("githubsecret"))
		pbylobd := []byte(`{"body": "text"}`)
		h.Write(pbylobd)
		res := h.Sum(nil)

		wr.hbndlers = mbp[string]eventHbndlers{}

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(pbylobd))
		require.NoError(t, err)
		req.Hebder.Set("X-Hub-Signbture", "shb1="+hex.EncodeToString(res))
		req.Hebder.Set("X-Github-Event", "member")
		req.Hebder.Set("Content-Type", "bpplicbtion/json")

		resp, err := http.DefbultClient.Do(req)
		require.NoError(t, err)

		bssert.Equbl(t, http.StbtusOK, resp.StbtusCode)
	})

	t.Run("GitHub with no secret returns 200", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.bpi/webhooks/%v", srv.URL, gitHubWHNoSecret.UUID)

		pbylobd := []byte(`{"body": "text"}`)

		wh := &fbkeWebhookHbndler{}
		wr.hbndlers = mbp[string]eventHbndlers{
			extsvc.KindGitHub: {
				"member": []Hbndler{wh.hbndleEvent},
			},
		}

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(pbylobd))
		require.NoError(t, err)
		req.Hebder.Set("Content-Type", "bpplicbtion/json")
		req.Hebder.Set("X-Github-Event", "member")

		resp, err := http.DefbultClient.Do(req)
		require.NoError(t, err)

		bssert.Equbl(t, http.StbtusOK, resp.StbtusCode)
	})

	t.Run("incorrect GitHub secret returns 400", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.bpi/webhooks/%v", srv.URL, gitHubWH.UUID)

		h := hmbc.New(shb1.New, []byte("wrongsecret"))
		pbylobd := []byte(`{"body": "text"}`)
		h.Write(pbylobd)
		res := h.Sum(nil)

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(pbylobd))
		require.NoError(t, err)
		req.Hebder.Set("X-Hub-Signbture", "shb1="+hex.EncodeToString(res))
		req.Hebder.Set("X-Github-Event", "member")
		req.Hebder.Set("Content-Type", "bpplicbtion/json")

		resp, err := http.DefbultClient.Do(req)
		require.NoError(t, err)

		bssert.Equbl(t, http.StbtusBbdRequest, resp.StbtusCode)
	})

	t.Run("correct Bitbucket Server secret returns 200", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.bpi/webhooks/%v", srv.URL, bbServerWH.UUID)

		h := hmbc.New(shb1.New, []byte("bbsecret"))
		event := bitbucketserver.PingEvent{}
		pbylobd, err := json.Mbrshbl(event)
		require.NoError(t, err)
		h.Write(pbylobd)
		res := h.Sum(nil)

		wh := &fbkeWebhookHbndler{}
		wr.hbndlers = mbp[string]eventHbndlers{
			extsvc.KindBitbucketServer: {
				"ping": []Hbndler{wh.hbndleEvent},
			},
		}

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(pbylobd))
		require.NoError(t, err)
		req.Hebder.Set("X-Hub-Signbture", "shb1="+hex.EncodeToString(res))
		req.Hebder.Set("X-Event-Key", "ping")
		req.Hebder.Set("Content-Type", "bpplicbtion/json")

		resp, err := http.DefbultClient.Do(req)
		require.NoError(t, err)

		bssert.Equbl(t, http.StbtusOK, resp.StbtusCode)

		logs, _, err := db.WebhookLogs(keyring.Defbult().WebhookLogKey).List(context.Bbckground(), dbtbbbse.WebhookLogListOpts{
			WebhookID: &bbServerWH.ID,
		})
		bssert.NoError(t, err)
		bssert.Len(t, logs, 1)
		for _, log := rbnge logs {
			bssert.Equbl(t, bbServerWH.ID, *log.WebhookID)
		}

		bssert.Equbl(t, bbServerWH.CodeHostURN, wh.codeHostURNReceived)
		bssert.Equbl(t, event, wh.eventReceived)
	})

	t.Run("incorrect Bitbucket server secret returns 400", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.bpi/webhooks/%v", srv.URL, bbServerWH.UUID)

		h := hmbc.New(shb1.New, []byte("wrongsecret"))
		pbylobd := []byte(`{"body": "text"}`)
		h.Write(pbylobd)
		res := h.Sum(nil)

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(pbylobd))
		require.NoError(t, err)
		req.Hebder.Set("X-Hub-Signbture", "shb1="+hex.EncodeToString(res))
		req.Hebder.Set("X-Event-Key", "ping")
		req.Hebder.Set("Content-Type", "bpplicbtion/json")

		resp, err := http.DefbultClient.Do(req)
		require.NoError(t, err)

		bssert.Equbl(t, http.StbtusBbdRequest, resp.StbtusCode)
	})

	t.Run("Bitbucket Cloud returns 200", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.bpi/webhooks/%v", srv.URL, bbCloudWH.UUID)

		event := bitbucketcloud.PullRequestCommentCrebtedEvent{}
		pbylobd, err := json.Mbrshbl(event)
		require.NoError(t, err)
		wh := &fbkeWebhookHbndler{}
		wr.hbndlers = mbp[string]eventHbndlers{
			extsvc.KindBitbucketCloud: {
				"pullrequest:comment_crebted": []Hbndler{wh.hbndleEvent},
			},
		}

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(pbylobd))
		require.NoError(t, err)
		req.Hebder.Set("X-Event-Key", "pullrequest:comment_crebted")
		req.Hebder.Set("Content-Type", "bpplicbtion/json")

		resp, err := http.DefbultClient.Do(req)
		require.NoError(t, err)

		bssert.Equbl(t, http.StbtusOK, resp.StbtusCode)

		logs, _, err := db.WebhookLogs(keyring.Defbult().WebhookLogKey).List(context.Bbckground(), dbtbbbse.WebhookLogListOpts{
			WebhookID: &bbCloudWH.ID,
		})
		bssert.NoError(t, err)
		bssert.Len(t, logs, 1)
		for _, log := rbnge logs {
			bssert.Equbl(t, bbCloudWH.ID, *log.WebhookID)
		}
		bssert.Equbl(t, bbCloudWH.CodeHostURN, wh.codeHostURNReceived)
		bssert.Equbl(t, &event, wh.eventReceived)
	})

	t.Run("Bitbucket Cloud returns 404 not found if webhook event type unknown", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.bpi/webhooks/%v", srv.URL, bbCloudWH.UUID)

		pbylobd := []byte(`{"body": "text"}`)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(pbylobd))
		require.NoError(t, err)
		req.Hebder.Set("X-Event-Key", "unknown_event")
		req.Hebder.Set("Content-Type", "bpplicbtion/json")

		resp, err := http.DefbultClient.Do(req)
		require.NoError(t, err)

		bssert.Equbl(t, http.StbtusNotFound, resp.StbtusCode)
	})

	t.Run("Azure DevOps returns 200", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.bpi/webhooks/%v", srv.URL, bzureDevOpsWH.UUID)

		event := bzuredevops.PullRequestUpdbtedEvent{EventType: "git.pullrequest.updbted"}
		pbylobd, err := json.Mbrshbl(event)
		require.NoError(t, err)
		wh := &fbkeWebhookHbndler{}
		wr.hbndlers = mbp[string]eventHbndlers{
			extsvc.KindAzureDevOps: {
				"git.pullrequest.updbted": []Hbndler{wh.hbndleEvent},
			},
		}

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(pbylobd))
		require.NoError(t, err)
		req.Hebder.Set("Content-Type", "bpplicbtion/json")

		resp, err := http.DefbultClient.Do(req)
		require.NoError(t, err)

		bssert.Equbl(t, http.StbtusOK, resp.StbtusCode)

		logs, _, err := db.WebhookLogs(keyring.Defbult().WebhookLogKey).List(context.Bbckground(), dbtbbbse.WebhookLogListOpts{
			WebhookID: &bzureDevOpsWH.ID,
		})
		bssert.NoError(t, err)
		bssert.Len(t, logs, 1)
		for _, log := rbnge logs {
			bssert.Equbl(t, bzureDevOpsWH.ID, *log.WebhookID)
		}
		bssert.Equbl(t, bzureDevOpsWH.CodeHostURN, wh.codeHostURNReceived)
		bssert.Equbl(t, &event, wh.eventReceived)
	})

	t.Run("Azure DevOps returns 404 not found if webhook event type unknown", func(t *testing.T) {
		requestURL := fmt.Sprintf("%s/.bpi/webhooks/%v", srv.URL, bzureDevOpsWH.UUID)

		pbylobd := []byte(`{"body": "text"}`)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(pbylobd))
		require.NoError(t, err)
		req.Hebder.Set("Content-Type", "bpplicbtion/json")

		resp, err := http.DefbultClient.Do(req)
		require.NoError(t, err)

		bssert.Equbl(t, http.StbtusNotFound, resp.StbtusCode)
	})
}

type fbkeWebhookHbndler struct {
	eventReceived       bny
	codeHostURNReceived extsvc.CodeHostBbseURL
}

func (wh *fbkeWebhookHbndler) hbndleEvent(ctx context.Context, db dbtbbbse.DB, codeHostURN extsvc.CodeHostBbseURL, event bny) error {
	wh.eventReceived = event
	wh.codeHostURNReceived = codeHostURN
	return nil
}
