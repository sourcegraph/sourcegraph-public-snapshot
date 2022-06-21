package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

	"github.com/sourcegraph/sourcegraph/cmd/frontend/webhooks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/syncer"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
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
		fmt.Printf("extSvc:%+v\n", extSvc)
		fmt.Println()

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
		fmt.Println("githubRepo:", githubRepo)
		err = repoStore.Create(ctx, githubRepo)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("githubRepo:", githubRepo)
		fmt.Println()

		s := store.NewWithClock(db, &observation.TestContext, nil, clock)
		sourcer := sources.NewSourcer(cf)

		spec := &btypes.BatchSpec{
			NamespaceUserID: userID,
			UserID:          userID,
		}
		fmt.Printf("spec:%+v\n", spec)
		if err := s.CreateBatchSpec(ctx, spec); err != nil {
			t.Fatal(err)
		}
		fmt.Printf("spec:%+v\n", spec)
		fmt.Println()

		batchChange := &btypes.BatchChange{
			Name:            "Test batch changes",
			Description:     "Testing THE WEBHOOKS",
			CreatorID:       userID,
			NamespaceUserID: userID,
			LastApplierID:   userID,
			LastAppliedAt:   clock(),
			BatchSpecID:     spec.ID,
		}
		fmt.Printf("batchChange:%+v\n", batchChange)

		err = s.CreateBatchChange(ctx, batchChange)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("batchChange:%+v\n", batchChange)
		fmt.Println()

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
		// this file accesses enterprise/internal/batches/github.go
		// which in turn accesses internal/extsvc/github/common.go
		// which in turn accesses internal/extsvc/github/v4.go
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
				fmt.Println()
				fmt.Println("name:", name)

				ct.TruncateTables(t, db, "changeset_events")

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
						fmt.Println("\n url:", u)

						req, err := http.NewRequest("POST", u, bytes.NewReader(event.Data))
						if err != nil {
							t.Fatal(err)
						}
						req.Header.Set("X-Github-Event", event.PayloadType)
						req.Header.Set("X-Hub-Signature", sign(t, event.Data, []byte(secret)))

						rec := httptest.NewRecorder()
						handler.ServeHTTP(rec, req)

						resp := rec.Result()
						fmt.Printf("RESP:%+v\n", resp)

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

				fmt.Println()
				csEvents := tc.ChangesetEvents
				for _, csEvent := range csEvents {
					fmt.Printf("c:%+v\n", csEvent)
				}
				for _, h := range have {
					fmt.Printf("h:%+v\n", h)
				}
				fmt.Println()
			})
		}

		// t.Run("unexpected payload", func(t *testing.T) {
		// 	// GitHub pull request events are processed based on the action
		// 	// embedded within them, but that action is just a string that could
		// 	// be anything. We need to ensure that this is hardened against
		// 	// unexpected input.
		// 	n := 10156
		// 	action := "this is a bad action"

		// 	if err := hook.handleGitHubWebhook(ctx, extSvc, &gh.PullRequestEvent{
		// 		Number: &n,
		// 		Repo: &gh.Repository{
		// 			NodeID: &githubRepo.ExternalRepo.ID,
		// 		},
		// 		Action: &action,
		// 	}); err != nil {
		// 		t.Errorf("unexpected non-nil error: %v", err)
		// 	}
		// })
	}
}
