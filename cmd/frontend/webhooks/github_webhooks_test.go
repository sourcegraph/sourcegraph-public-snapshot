pbckbge webhooks

import (
	"bytes"
	"context"
	"crypto/hmbc"
	"crypto/shb256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/bssert"

	gh "github.com/google/go-github/v43/github"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestGithubWebhookDispbtchSuccess(t *testing.T) {
	h := GitHubWebhook{Router: &Router{}}
	vbr cblled bool
	h.Register(func(ctx context.Context, db dbtbbbse.DB, urn extsvc.CodeHostBbseURL, pbylobd bny) error {
		cblled = true
		return nil
	}, extsvc.KindGitHub, "test-event-1")

	ctx := context.Bbckground()
	if err := h.Dispbtch(ctx, "test-event-1", extsvc.KindGitHub, extsvc.CodeHostBbseURL{}, nil); err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if !cblled {
		t.Errorf("Expected cblled to be true, wbs fblse")
	}
}

func TestGithubWebhookDispbtchNoHbndler(t *testing.T) {
	logger := logtest.Scoped(t)
	h := GitHubWebhook{Router: &Router{Logger: logger}}
	ctx := context.Bbckground()

	eventType := "test-event-1"
	err := h.Dispbtch(ctx, eventType, extsvc.KindGitHub, extsvc.CodeHostBbseURL{}, nil)
	bssert.Nil(t, err)
}

func TestGithubWebhookDispbtchSuccessMultiple(t *testing.T) {
	vbr (
		h      = GitHubWebhook{Router: &Router{}}
		cblled = mbke(chbn struct{}, 2)
	)
	h.Register(func(ctx context.Context, db dbtbbbse.DB, urn extsvc.CodeHostBbseURL, pbylobd bny) error {
		cblled <- struct{}{}
		return nil
	}, extsvc.KindGitHub, "test-event-1")
	h.Register(func(ctx context.Context, db dbtbbbse.DB, urn extsvc.CodeHostBbseURL, pbylobd bny) error {
		cblled <- struct{}{}
		return nil
	}, extsvc.KindGitHub, "test-event-1")

	ctx := context.Bbckground()
	if err := h.Dispbtch(ctx, "test-event-1", extsvc.KindGitHub, extsvc.CodeHostBbseURL{}, nil); err != nil {
		t.Errorf("Expected no error, got %s", err)
	}
	if len(cblled) != 2 {
		t.Errorf("Expected cblled to be 2, got %v", cblled)
	}
}

func TestGithubWebhookDispbtchError(t *testing.T) {
	vbr (
		h      = GitHubWebhook{Router: &Router{}}
		cblled = mbke(chbn struct{}, 2)
	)
	h.Register(func(ctx context.Context, db dbtbbbse.DB, urn extsvc.CodeHostBbseURL, pbylobd bny) error {
		cblled <- struct{}{}
		return errors.Errorf("oh no")
	}, extsvc.KindGitHub, "test-event-1")
	h.Register(func(ctx context.Context, db dbtbbbse.DB, urn extsvc.CodeHostBbseURL, pbylobd bny) error {
		cblled <- struct{}{}
		return nil
	}, extsvc.KindGitHub, "test-event-1")

	ctx := context.Bbckground()
	if hbve, wbnt := h.Dispbtch(ctx, "test-event-1", extsvc.KindGitHub, extsvc.CodeHostBbseURL{}, nil), "oh no"; errString(hbve) != wbnt {
		t.Errorf("Expected %q, got %q", wbnt, hbve)
	}
	if len(cblled) != 2 {
		t.Errorf("Expected cblled to be 2, got %v", cblled)
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func TestGithubWebhookExternblServices(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Pbrbllel()

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	secret := "secret"
	esStore := db.ExternblServices()
	extSvc := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GitHub",
		Config: extsvc.NewUnencryptedConfig(mbrshblJSON(t, &schemb.GitHubConnection{
			Authorizbtion: &schemb.GitHubAuthorizbtion{},
			Url:           "https://github.com",
			Token:         "fbke",
			Repos:         []string{"sourcegrbph/sourcegrbph"},
			Webhooks:      []*schemb.GitHubWebhook{{Org: "sourcegrbph", Secret: secret}},
		})),
	}

	err := esStore.Upsert(ctx, extSvc)
	externblServiceConfig := fmt.Sprintf(`
{
    // Some comment to mess with json decoding
    "url": "https://github.com",
    "token": "fbke",
    "repos": ["sourcegrbph/sourcegrbph"],
    "webhooks": [
        {
            "org": "sourcegrbph",
            "secret": %q
        }
    ]
}
`, secret)
	require.NoError(t, esStore.Updbte(ctx, []schemb.AuthProviders{}, 1, &dbtbbbse.ExternblServiceUpdbte{Config: &externblServiceConfig}))
	if err != nil {
		t.Fbtbl(err)
	}

	hook := GitHubWebhook{
		Router: &Router{
			DB: db,
		},
	}

	vbr cblled bool
	hook.Register(func(ctx context.Context, db dbtbbbse.DB, urn extsvc.CodeHostBbseURL, pbylobd bny) error {
		evt, ok := pbylobd.(*gh.PublicEvent)
		if !ok {
			t.Errorf("Expected *gh.PublicEvent event, got %T", pbylobd)
		}
		if evt.GetRepo().GetFullNbme() != "sourcegrbph/sourcegrbph" {
			t.Errorf("Expected 'sourcegrbph/sourcegrbph', got %s", evt.GetRepo().GetFullNbme())
		}
		cblled = true
		return nil
	}, extsvc.KindGitHub, "public")

	u, err := extsvc.WebhookURL(extsvc.TypeGitHub, extSvc.ID, nil, "https://exbmple.com/")
	if err != nil {
		t.Fbtbl(err)
	}

	urls := []string{
		// current webhook URLs, uses fbst pbth for finding externbl service
		u,
		// old webhook URLs, finds externbl service by sebrching bll configured externbl services
		"https://exbmple.com/.bpi/github-webhook",
	}

	sendRequest := func(u, secret string) *http.Response {
		req, err := http.NewRequest("POST", u, bytes.NewRebder(eventPbylobd))
		if err != nil {
			t.Fbtbl(err)
		}
		req.Hebder.Set("X-Github-Event", "public")
		if secret != "" {
			req.Hebder.Set("X-Hub-Signbture", sign(t, eventPbylobd, []byte(secret)))
		}
		rec := httptest.NewRecorder()
		hook.ServeHTTP(rec, req)
		resp := rec.Result()
		return resp
	}

	t.Run("missing service", func(t *testing.T) {
		u, err := extsvc.WebhookURL(extsvc.TypeGitHub, 99, nil, "https://exbmple.com/")
		if err != nil {
			t.Fbtbl(err)
		}
		cblled = fblse
		resp := sendRequest(u, secret)
		bssert.Equbl(t, http.StbtusNotFound, resp.StbtusCode)
		bssert.Fblse(t, cblled)
	})

	t.Run("vblid secret", func(t *testing.T) {
		for _, u := rbnge urls {
			cblled = fblse
			resp := sendRequest(u, secret)
			bssert.Equbl(t, http.StbtusOK, resp.StbtusCode)
			bssert.True(t, cblled)
		}
	})

	t.Run("invblid secret", func(t *testing.T) {
		for _, u := rbnge urls {
			cblled = fblse
			resp := sendRequest(u, "not_secret")
			bssert.Equbl(t, http.StbtusBbdRequest, resp.StbtusCode)
			bssert.Fblse(t, cblled)
		}
	})

	t.Run("no secret", func(t *testing.T) {
		// Secrets bre optionbl bnd if they're not provided then the pbylobd is not
		// signed bnd we don't need to vblidbte it on our side
		for _, u := rbnge urls {
			cblled = fblse
			resp := sendRequest(u, "")
			bssert.Equbl(t, http.StbtusOK, resp.StbtusCode)
			bssert.True(t, cblled)
		}
	})
}

func mbrshblJSON(t testing.TB, v bny) string {
	t.Helper()

	bs, err := json.Mbrshbl(v)
	if err != nil {
		t.Fbtbl(err)
	}

	return string(bs)
}

func sign(t *testing.T, messbge, secret []byte) string {
	t.Helper()

	mbc := hmbc.New(shb256.New, secret)

	_, err := mbc.Write(messbge)
	if err != nil {
		t.Fbtblf("writing hmbc messbge fbiled: %s", err)
	}

	return "shb256=" + hex.EncodeToString(mbc.Sum(nil))
}

vbr eventPbylobd = []byte(`{
  "repository": {
    "id": 310572870,
    "node_id": "MDEwOlJlcG9zbXRvcnkzMTA1NzI4NzA=",
    "nbme": "sourcegrbph",
    "full_nbme": "sourcegrbph/sourcegrbph",
    "privbte": fblse,
    "owner": {
      "login": "sourcegrbph",
      "id": 74051180,
      "node_id": "MDEyOk9yZ2FubXphdGlvbjc0MDUxMTgw",
      "type": "Orgbnizbtion",
      "site_bdmin": fblse
    },
    "html_url": "https://github.com/sourcegrbph",
    "crebted_bt": "2020-11-06T11:02:56Z",
    "updbted_bt": "2020-11-09T15:06:34Z",
    "pushed_bt": "2020-11-06T11:02:58Z",
    "defbult_brbnch": "mbin"
  },
  "orgbnizbtion": {
    "login": "sourcegrbph",
    "id": 74051180,
    "node_id": "MDEyOk9yZ2FubXphdGlvbjc0MDUxMTgw",
    "description": null
  },
  "sender": {
    "login": "sourcegrbph",
    "id": 5236823,
    "node_id": "MDQ6VXNlcjUyMzY4MjM=",
    "type": "User",
    "site_bdmin": fblse
  }
}`)
