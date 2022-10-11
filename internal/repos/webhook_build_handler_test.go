package repos

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/repos/webhookworker"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestWebhookBuildHandle(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Background()

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		// We need this so that the config validation below passes even when we're
		// running against VCR data
		token = "faketoken"
	}

	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := NewStore(logger, db)
	esStore := store.ExternalServiceStore()
	repoStore := store.RepoStore()

	repo := &types.Repo{
		ID:       1,
		Name:     api.RepoName("ghe.sgdev.org/milton/test"),
		Metadata: &github.Repository{},
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "12345",
			ServiceID:   "https://ghe.sgdev.org",
			ServiceType: extsvc.TypeGitHub,
		},
	}
	if err := repoStore.Create(ctx, repo); err != nil {
		t.Fatal(err)
	}

	ghConn := &schema.GitHubConnection{
		Url:      "https://github.com",
		Token:    token,
		Repos:    []string{"milton/test"},
		Webhooks: []*schema.GitHubWebhook{{Org: "ghe.sgdev.org", Secret: "secret"}},
	}

	configData, err := json.Marshal(ghConn)
	if err != nil {
		t.Fatal(err)
	}

	config := string(configData)
	svc := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "TestService",
		Config:      extsvc.NewUnencryptedConfig(config),
	}
	if err := esStore.Upsert(ctx, svc); err != nil {
		t.Fatal(err)
	}

	job := &webhookworker.Job{
		RepoID:     int32(repo.ID),
		RepoName:   string(repo.Name),
		Org:        strings.Split(string(repo.Name), "/")[0],
		ExtSvcID:   svc.ID,
		ExtSvcKind: svc.Kind,
	}

	testName := "webhook-build-handler"
	cf, save := httptestutil.NewGitHubRecorderFactory(t, update(testName), testName)
	defer save()

	opts := []httpcli.Opt{}
	doer, err := cf.Doer(opts...)
	if err != nil {
		t.Fatal(err)
	}

	handler := newWebhookBuildHandler(store, doer)
	if err := handler.Handle(ctx, logger, job); err != nil {
		t.Fatal(err)
	}
}

func TestRandomHex(t *testing.T) {
	t.Run("calling twice gives different result", func(t *testing.T) {
		a, err := randomHex(10)
		if err != nil {
			t.Fatal(err)
		}
		b, err := randomHex(10)
		if err != nil {
			t.Fatal(err)
		}
		assert.NotEqual(t, a, b)
	})

	t.Run("length works as expected", func(t *testing.T) {
		for i := 0; i < 32; i++ {
			s, err := randomHex(i)
			if err != nil {
				t.Fatalf("randomHex for length %d: %v", i, err)
			}
			assert.Equal(t, i, len(s))
		}
	})
}
