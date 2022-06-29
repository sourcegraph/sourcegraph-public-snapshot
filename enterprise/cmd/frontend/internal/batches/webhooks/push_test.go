package webhooks

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	rp "github.com/sourcegraph/sourcegraph/cmd/repo-updater/repoupdater"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/syncer"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	syncwebhooks "github.com/sourcegraph/sourcegraph/internal/repos/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Main function to handle the
// "/enqueue-repo-update" endpoint
func testGitHubPushHooks(db database.DB, userID int32) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()

		svc := types.ExternalService{
			Kind: extsvc.KindGitHub,
			Config: `{
				"URL": "https://github.com",
				"Token": "secret-token"
				}`,
		}

		repo := types.Repo{
			ID:   1,
			Name: "github.com/sourcegraph/sourcegraph",
			ExternalRepo: api.ExternalRepoSpec{
				ID:          "hi mom",
				ServiceType: extsvc.TypeGitHub,
				ServiceID:   "https://github.com",
			},
			Metadata: new(github.Repository),
		}

		initStore := func(db database.DB) repos.Store {
			store := repos.NewStore(logtest.Scoped(t), db)
			if err := store.ExternalServiceStore().Upsert(ctx, &svc); err != nil {
				t.Fatal(err)
			}
			if err := store.RepoStore().Create(ctx, &repo); err != nil {
				t.Fatal(err)
			}
			return store
		}

		type testCase struct {
			name string
			repo api.RepoName
			res  *protocol.RepoUpdateResponse
			err  string
			init func(database.DB) repos.Store
		}

		tcs := []testCase{
			{
				name: "existing repo",
				repo: repo.Name,
				init: initStore,
				res: &protocol.RepoUpdateResponse{
					ID:   repo.ID,
					Name: string(repo.Name),
				},
			},
		}

		logger := logtest.Scoped(t)
		for _, tc := range tcs {
			tc := tc

			t.Run(tc.name, func(t *testing.T) {
				fmt.Println("Running", tc.name)
				sqlDB := dbtest.NewDB(t)
				store := tc.init(database.NewDB(sqlDB))
				s := &rp.Server{
					Logger:    logger,
					Store:     store,
					Scheduler: repos.NewUpdateScheduler(logger, db),
				}
				server := httptest.NewServer(s.Handler())
				defer server.Close()

				activatePushWebhook(db, userID, server.URL, t)

			})
		}
	}
}

// Helper function to create the hook
// and make the POST request
func activatePushWebhook(db database.DB, userID int32, url string, t *testing.T) {
	now := timeutil.Now()
	clock := func() time.Time { return now }

	ctx := context.Background()

	rcache.SetupForTest(t)

	ct.TruncateTables(t, db, "changeset_events", "changesets")

	cf, save := httptestutil.NewGitHubRecorderFactory(t, *update, "github-webhooks")
	defer save()

	secret := "secret"
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = "no-GITHUB_TOKEN-set"
	}
	repoStore := db.Repos()
	esStore := db.ExternalServices()
	extSvc := &types.ExternalService{
		Kind:        extsvc.KindGitHub,
		DisplayName: "GitHub",
		Config: ct.MarshalJSON(t, &schema.GitHubConnection{
			Url:      "https://github.com",
			Token:    token,
			Repos:    []string{"sourcegraph/sourcegraph"},
			Webhooks: []*schema.GitHubWebhook{{Org: "sourcegraph", Secret: secret}},
		}),
	}

	err := esStore.Upsert(ctx, extSvc)
	if err != nil {
		t.Fatal(t)
	}

	githubSrc, err := repos.NewGithubSource(db.ExternalServices(), extSvc, cf)
	if err != nil {
		t.Fatal(t)
	}

	githubRepo, err := githubSrc.GetRepo(ctx, "sourcegraph/sourcegraph") // yaml GET req
	if err != nil {
		t.Fatal(err)
	}
	err = repoStore.Create(ctx, githubRepo)
	if err != nil {
		t.Fatal(err)
	}

	s := store.NewWithClock(db, &observation.TestContext, nil, clock)
	sourcer := sources.NewSourcer(cf)

	spec := &btypes.BatchSpec{
		NamespaceUserID: userID,
		UserID:          userID,
	}
	if err := s.CreateBatchSpec(ctx, spec); err != nil {
		t.Fatal(err)
	}

	batchChange := &btypes.BatchChange{
		Name:            "Test batch changes",
		Description:     "Testing THE WEBHOOKS",
		CreatorID:       userID,
		NamespaceUserID: userID,
		LastApplierID:   userID,
		LastAppliedAt:   clock(),
		BatchSpecID:     spec.ID,
	}

	err = s.CreateBatchChange(ctx, batchChange)
	if err != nil {
		t.Fatal(err)
	}

	// NOTE: Your sample payload should apply to a PR with the number matching below
	changeset := &btypes.Changeset{
		RepoID:              githubRepo.ID,
		ExternalID:          "10156",
		ExternalServiceType: githubRepo.ExternalRepo.ServiceType,
		BatchChanges:        []btypes.BatchChangeAssoc{{BatchChangeID: batchChange.ID}},
	}

	err = s.CreateChangeset(ctx, changeset)
	if err != nil {
		t.Fatal(err)
	}

	// Set up mocks to prevent the diffstat computation from trying to
	// use a real gitserver, and so we can control what diff is used to
	// create the diffstat.
	state := ct.MockChangesetSyncState(&protocol.RepoInfo{
		Name: "repo",
		VCS:  protocol.VCSInfo{URL: "https://example.com/repo/"},
	})
	defer state.Unmock()

	src, err := sourcer.ForRepo(ctx, s, githubRepo)
	if err != nil {
		t.Fatal(err)
	}

	err = syncer.SyncChangeset(ctx, s, src, githubRepo, changeset)
	if err != nil {
		t.Fatal(err)
	}

	sc := syncwebhooks.SyncWebhook{}

	fixtureFiles, err := filepath.Glob("testdata/fixtures/webhooks/github/*.json")
	if err != nil {
		t.Fatal(err)
	}

	for _, fixtureFile := range fixtureFiles {
		_, name := path.Split(fixtureFile)
		name = strings.TrimSuffix(name, ".json")
		if name != "push-event" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			fmt.Println("name:", name)

			tc := loadWebhookTestCase(t, fixtureFile)
			fmt.Println("done loading tc")

			for _, event := range tc.Payloads {
				// fmt.Printf("Event:%+v\n", event)
				handler := webhooks.GitHubWebhook{
					ExternalServices: esStore,
				}
				fmt.Println("registering")
				sc.Register(&handler)

				u := url + "/enqueue-repo-update"
				syncwebhooks.Url = url

				req, err := http.NewRequest("POST", u, bytes.NewReader(event.Data))
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Set("X-Github-Event", event.PayloadType)
				req.Header.Set("X-Hub-Signature", sign(t, event.Data, []byte(secret)))

				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				resp := rec.Result()

				if resp.StatusCode != http.StatusOK {
					t.Fatalf("Non 200 code: %v", resp.StatusCode)
				}
			}

			if err != nil {
				t.Fatal(err)
			}

			// Overwrite and format test case
			if *update {
			}
		})
	}
}
