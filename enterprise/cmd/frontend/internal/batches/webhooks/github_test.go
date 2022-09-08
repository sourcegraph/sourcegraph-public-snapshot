package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	gh "github.com/google/go-github/v43/github"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/syncer"
	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Run from integration_test.go
func testGitHubWebhook(db database.DB, userID int32) func(*testing.T) {
	return func(t *testing.T) {
		now := timeutil.Now()
		clock := func() time.Time { return now }

		ctx := context.Background()

		rcache.SetupForTest(t)

		bt.TruncateTables(t, db, "changeset_events", "changesets")

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
			Config: extsvc.NewUnencryptedConfig(bt.MarshalJSON(t, &schema.GitHubConnection{
				Url:      "https://github.com",
				Repos:    []string{"sourcegraph/sourcegraph"},
				Webhooks: []*schema.GitHubWebhook{{Org: "sourcegraph", Secret: secret}},
			})),
		}

		err := esStore.Upsert(ctx, extSvc)
		if err != nil {
			t.Fatal(t)
		}

		githubSrc, err := repos.NewGithubSource(ctx, logtest.Scoped(t), db.ExternalServices(), extSvc, cf)
		if err != nil {
			t.Fatal(t)
		}

		githubRepo, err := githubSrc.GetRepo(ctx, "sourcegraph/sourcegraph")
		if err != nil {
			t.Fatal(err)
		}

		err = repoStore.Create(ctx, githubRepo)
		if err != nil {
			t.Fatal(err)
		}

		s := store.NewWithClock(db, &observation.TestContext, nil, clock)
		if err := s.CreateSiteCredential(ctx, &btypes.SiteCredential{
			ExternalServiceType: githubRepo.ExternalRepo.ServiceType,
			ExternalServiceID:   githubRepo.ExternalRepo.ServiceID,
		},
			&auth.OAuthBearerTokenWithSSH{
				OAuthBearerToken: auth.OAuthBearerToken{Token: token},
			},
		); err != nil {
			t.Fatal(err)
		}
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
		state := bt.MockChangesetSyncState(&protocol.RepoInfo{
			Name: "repo",
			VCS:  protocol.VCSInfo{URL: "https://example.com/repo/"},
		})
		defer state.Unmock()

		src, err := sourcer.ForChangeset(ctx, s, changeset)
		if err != nil {
			t.Fatal(err)
		}
		err = syncer.SyncChangeset(ctx, s, src, githubRepo, changeset)
		if err != nil {
			t.Fatal(err)
		}

		hook := NewGitHubWebhook(s)

		fixtureFiles, err := filepath.Glob("testdata/fixtures/webhooks/github/*.json")
		if err != nil {
			t.Fatal(err)
		}

		for _, fixtureFile := range fixtureFiles {
			_, name := path.Split(fixtureFile)
			name = strings.TrimSuffix(name, ".json")
			t.Run(name, func(t *testing.T) {
				bt.TruncateTables(t, db, "changeset_events")

				tc := loadWebhookTestCase(t, fixtureFile)

				// Send all events twice to ensure we are idempotent
				for i := 0; i < 2; i++ {
					for _, event := range tc.Payloads {
						handler := webhooks.GitHubWebhook{
							ExternalServices: esStore,
						}
						hook.Register(&handler)

						u, err := extsvc.WebhookURL(extsvc.TypeGitHub, extSvc.ID, nil, "https://example.com/")
						if err != nil {
							t.Fatal(err)
						}

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
				}

				have, _, err := s.ListChangesetEvents(ctx, store.ListChangesetEventsOpts{})
				if err != nil {
					t.Fatal(err)
				}

				// Overwrite and format test case
				if *update {
					tc.ChangesetEvents = have
					data, err := json.MarshalIndent(tc, "  ", "  ")
					if err != nil {
						t.Fatal(err)
					}
					err = os.WriteFile(fixtureFile, data, 0666)
					if err != nil {
						t.Fatal(err)
					}
				}

				opts := []cmp.Option{
					cmpopts.IgnoreFields(btypes.ChangesetEvent{}, "CreatedAt"),
					cmpopts.IgnoreFields(btypes.ChangesetEvent{}, "UpdatedAt"),
				}
				if diff := cmp.Diff(tc.ChangesetEvents, have, opts...); diff != "" {
					t.Error(diff)
				}
			})
		}

		t.Run("unexpected payload", func(t *testing.T) {
			// GitHub pull request events are processed based on the action
			// embedded within them, but that action is just a string that could
			// be anything. We need to ensure that this is hardened against
			// unexpected input.
			n := 10156
			action := "this is a bad action"

			if err := hook.handleGitHubWebhook(ctx, extSvc, &gh.PullRequestEvent{
				Number: &n,
				Repo: &gh.Repository{
					NodeID: &githubRepo.ExternalRepo.ID,
				},
				Action: &action,
			}); err != nil {
				t.Errorf("unexpected non-nil error: %v", err)
			}
		})
	}
}
