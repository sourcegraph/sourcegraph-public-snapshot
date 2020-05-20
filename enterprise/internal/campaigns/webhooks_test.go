package campaigns

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/schema"
)

var update = flag.Bool("update", false, "update testdata")

// Run from integration_test.go
func testGitHubWebhook(db *sql.DB, userID int32) func(*testing.T) {
	return func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Microsecond)
		clock := func() time.Time { return now }

		ctx := context.Background()

		rcache.SetupForTest(t)

		truncateTables(t, db, "changeset_jobs", "changeset_events", "changesets")

		cf, save := newGithubClientFactory(t, "github-webhooks")
		defer save()

		secret := "secret"
		repoStore := repos.NewDBStore(db, sql.TxOptions{})
		extSvc := &repos.ExternalService{
			Kind:        "GITHUB",
			DisplayName: "GitHub",
			Config: marshalJSON(t, &schema.GitHubConnection{
				Url:      "https://github.com",
				Token:    os.Getenv("GITHUB_TOKEN"),
				Repos:    []string{"sourcegraph/sourcegraph"},
				Webhooks: []*schema.GitHubWebhook{{Org: "sourcegraph", Secret: secret}},
			}),
		}

		err := repoStore.UpsertExternalServices(ctx, extSvc)
		if err != nil {
			t.Fatal(t)
		}

		githubSrc, err := repos.NewGithubSource(extSvc, cf, nil)
		if err != nil {
			t.Fatal(t)
		}

		githubRepo, err := githubSrc.GetRepo(ctx, "sourcegraph/sourcegraph")
		if err != nil {
			t.Fatal(err)
		}

		err = repoStore.UpsertRepos(ctx, githubRepo)
		if err != nil {
			t.Fatal(err)
		}

		store := NewStoreWithClock(db, clock)

		campaign := &campaigns.Campaign{
			Name:            "Test campaign",
			Description:     "Testing THE WEBHOOKS",
			AuthorID:        userID,
			NamespaceUserID: userID,
		}

		err = store.CreateCampaign(ctx, campaign)
		if err != nil {
			t.Fatal(err)
		}

		// NOTE: Your sample payload should apply to a PR with the number matching below
		changesets := []*campaigns.Changeset{
			{
				RepoID:              githubRepo.ID,
				ExternalID:          "10156",
				ExternalServiceType: githubRepo.ExternalRepo.ServiceType,
				CampaignIDs:         []int64{campaign.ID},
			},
		}

		err = store.CreateChangesets(ctx, changesets...)
		if err != nil {
			t.Fatal(err)
		}

		err = SyncChangesets(ctx, repoStore, store, cf, changesets...)
		if err != nil {
			t.Fatal(err)
		}

		hook := NewGitHubWebhook(store, repoStore, clock)

		fixtureFiles, err := filepath.Glob("testdata/fixtures/webhooks/github/*.json")
		if err != nil {
			t.Fatal(err)
		}

		for _, fixtureFile := range fixtureFiles {
			_, name := path.Split(fixtureFile)
			name = strings.TrimSuffix(name, ".json")
			t.Run(name, func(t *testing.T) {
				truncateTables(t, db, "changeset_events")

				tc := loadWebhookTestCase(t, fixtureFile)

				// Send all events twice to ensure we are idempotent
				for i := 0; i < 2; i++ {
					for _, event := range tc.Payloads {
						u := extsvc.WebhookURL(github.ServiceType, extSvc.ID, "https://example.com/")

						req, err := http.NewRequest("POST", u, bytes.NewReader(event.Data))
						if err != nil {
							t.Fatal(err)
						}
						req.Header.Set("X-Github-Event", event.PayloadType)
						req.Header.Set("X-Hub-Signature", sign(t, event.Data, []byte(secret)))

						rec := httptest.NewRecorder()
						hook.ServeHTTP(rec, req)
						resp := rec.Result()

						if resp.StatusCode != http.StatusOK {
							t.Fatalf("Non 200 code: %v", resp.StatusCode)
						}
					}
				}

				have, _, err := store.ListChangesetEvents(ctx, ListChangesetEventsOpts{Limit: -1})
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
					err = ioutil.WriteFile(fixtureFile, data, 0666)
					if err != nil {
						t.Fatal(err)
					}
				}

				opts := []cmp.Option{
					cmpopts.IgnoreFields(campaigns.ChangesetEvent{}, "CreatedAt"),
					cmpopts.IgnoreFields(campaigns.ChangesetEvent{}, "UpdatedAt"),
				}
				if diff := cmp.Diff(tc.ChangesetEvents, have, opts...); diff != "" {
					t.Error(diff)
				}

			})
		}
	}
}

// Run from integration_test.go
func testBitbucketWebhook(db *sql.DB, userID int32) func(*testing.T) {
	return func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Microsecond)
		clock := func() time.Time { return now }

		ctx := context.Background()

		rcache.SetupForTest(t)

		truncateTables(t, db, "changeset_jobs", "changeset_events", "changesets")

		cf, save := newGithubClientFactory(t, "bitbucket-webhooks")
		defer save()

		secret := "secret"
		repoStore := repos.NewDBStore(db, sql.TxOptions{})
		extSvc := &repos.ExternalService{
			Kind:        "BITBUCKETSERVER",
			DisplayName: "Bitbucket",
			Config: marshalJSON(t, &schema.BitbucketServerConnection{
				Url:   "https://bitbucket.sgdev.org",
				Token: os.Getenv("BITBUCKET_SERVER_TOKEN"),
				Repos: []string{"SOUR/automation-testing"},
				Webhooks: &schema.Webhooks{
					Secret: secret,
				},
			}),
		}

		err := repoStore.UpsertExternalServices(ctx, extSvc)
		if err != nil {
			t.Fatal(t)
		}

		bitbucketSource, err := repos.NewBitbucketServerSource(extSvc, cf, nil)
		if err != nil {
			t.Fatal(t)
		}

		bitbucketRepo, err := getSingleRepo(ctx, bitbucketSource, "bitbucket.sgdev.org/SOUR/automation-testing")
		if err != nil {
			t.Fatal(err)
		}

		if bitbucketRepo == nil {
			t.Fatal("repo not found")
		}

		err = repoStore.UpsertRepos(ctx, bitbucketRepo)
		if err != nil {
			t.Fatal(err)
		}

		store := NewStoreWithClock(db, clock)

		campaign := &campaigns.Campaign{
			Name:            "Test campaign",
			Description:     "Testing THE WEBHOOKS",
			AuthorID:        userID,
			NamespaceUserID: userID,
		}

		err = store.CreateCampaign(ctx, campaign)
		if err != nil {
			t.Fatal(err)
		}

		changesets := []*campaigns.Changeset{
			{
				RepoID:              bitbucketRepo.ID,
				ExternalID:          "69",
				ExternalServiceType: bitbucketRepo.ExternalRepo.ServiceType,
				CampaignIDs:         []int64{campaign.ID},
			},
			{
				RepoID:              bitbucketRepo.ID,
				ExternalID:          "19",
				ExternalServiceType: bitbucketRepo.ExternalRepo.ServiceType,
				CampaignIDs:         []int64{campaign.ID},
			},
		}

		err = store.CreateChangesets(ctx, changesets...)
		if err != nil {
			t.Fatal(err)
		}

		err = SyncChangesets(ctx, repoStore, store, cf, changesets...)
		if err != nil {
			t.Fatal(err)
		}

		hook := NewBitbucketServerWebhook(store, repoStore, clock, "testhook")

		fixtureFiles, err := filepath.Glob("testdata/fixtures/webhooks/bitbucketserver/*.json")
		if err != nil {
			t.Fatal(err)
		}

		for _, fixtureFile := range fixtureFiles {
			_, name := path.Split(fixtureFile)
			name = strings.TrimSuffix(name, ".json")
			t.Run(name, func(t *testing.T) {
				truncateTables(t, db, "changeset_events")

				tc := loadWebhookTestCase(t, fixtureFile)

				// Send all events twice to ensure we are idempotent
				for i := 0; i < 2; i++ {
					for _, event := range tc.Payloads {
						u := extsvc.WebhookURL(bitbucketserver.ServiceType, extSvc.ID, "https://example.com/")

						req, err := http.NewRequest("POST", u, bytes.NewReader(event.Data))
						if err != nil {
							t.Fatal(err)
						}
						req.Header.Set("X-Event-Key", event.PayloadType)
						req.Header.Set("X-Hub-Signature", sign(t, event.Data, []byte(secret)))

						rec := httptest.NewRecorder()
						hook.ServeHTTP(rec, req)
						resp := rec.Result()

						if resp.StatusCode != http.StatusOK {
							t.Fatalf("Non 200 code: %v", resp.StatusCode)
						}
					}
				}

				have, _, err := store.ListChangesetEvents(ctx, ListChangesetEventsOpts{Limit: -1})
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
					err = ioutil.WriteFile(fixtureFile, data, 0666)
					if err != nil {
						t.Fatal(err)
					}
				}

				opts := []cmp.Option{
					cmpopts.IgnoreFields(campaigns.ChangesetEvent{}, "CreatedAt"),
					cmpopts.IgnoreFields(campaigns.ChangesetEvent{}, "UpdatedAt"),
				}
				if diff := cmp.Diff(tc.ChangesetEvents, have, opts...); diff != "" {
					t.Error(diff)
				}

			})
		}
	}
}

func getSingleRepo(ctx context.Context, bitbucketSource *repos.BitbucketServerSource, name string) (*repos.Repo, error) {
	repoChan := make(chan repos.SourceResult)
	go func() {
		bitbucketSource.ListRepos(ctx, repoChan)
		close(repoChan)
	}()

	var bitbucketRepo *repos.Repo
	for result := range repoChan {
		if result.Err != nil {
			return nil, result.Err
		}
		if result.Repo == nil {
			continue
		}
		if result.Repo.Name == name {
			bitbucketRepo = result.Repo
		}
	}

	return bitbucketRepo, nil
}

type webhookTestCase struct {
	Payloads []struct {
		PayloadType string          `json:"payload_type"`
		Data        json.RawMessage `json:"data"`
	} `json:"payloads"`
	ChangesetEvents []*campaigns.ChangesetEvent `json:"changeset_events"`
}

func loadWebhookTestCase(t testing.TB, path string) webhookTestCase {
	t.Helper()

	bs, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var tc webhookTestCase
	if err := json.Unmarshal(bs, &tc); err != nil {
		t.Fatal(err)
	}
	for i, ev := range tc.ChangesetEvents {
		meta, err := campaigns.NewChangesetEventMetadata(ev.Kind)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(ev.Metadata)
		if err != nil {
			t.Fatal(err)
		}
		err = json.Unmarshal(raw, &meta)
		if err != nil {
			t.Fatal(err)
		}
		tc.ChangesetEvents[i].Metadata = meta
	}

	return tc
}

func TestBitbucketWebhookUpsert(t *testing.T) {
	testCases := []struct {
		name    string
		con     *schema.BitbucketServerConnection
		secrets map[int64]string
		expect  []string
	}{
		{
			name: "No existing secret",
			con: &schema.BitbucketServerConnection{
				Plugin: &schema.BitbucketServerPlugin{
					Permissions: "",
					Webhooks: &schema.BitbucketServerPluginWebhooks{
						Secret: "secret",
					},
				},
			},
			secrets: map[int64]string{},
			expect:  []string{"POST"},
		},
		{
			name: "existing secret matches",
			con: &schema.BitbucketServerConnection{
				Plugin: &schema.BitbucketServerPlugin{
					Permissions: "",
					Webhooks: &schema.BitbucketServerPluginWebhooks{
						Secret: "secret",
					},
				},
			},
			secrets: map[int64]string{
				1: "secret",
			},
			expect: []string{},
		},
		{
			name: "existing secret does not match matches",
			con: &schema.BitbucketServerConnection{
				Plugin: &schema.BitbucketServerPlugin{
					Permissions: "",
					Webhooks: &schema.BitbucketServerPluginWebhooks{
						Secret: "secret",
					},
				},
			},
			secrets: map[int64]string{
				1: "old",
			},
			expect: []string{"POST"},
		},
		{
			name: "secret removed",
			con: &schema.BitbucketServerConnection{
				Plugin: &schema.BitbucketServerPlugin{
					Permissions: "",
					Webhooks: &schema.BitbucketServerPluginWebhooks{
						Secret: "",
					},
				},
			},
			secrets: map[int64]string{
				1: "old",
			},
			expect: []string{"DELETE"},
		},
		{
			name: "secret removed, no history",
			con: &schema.BitbucketServerConnection{
				Plugin: &schema.BitbucketServerPlugin{
					Permissions: "",
					Webhooks: &schema.BitbucketServerPluginWebhooks{
						Secret: "",
					},
				},
			},
			secrets: map[int64]string{},
			expect:  []string{"DELETE"},
		},
		{
			name: "secret removed, with history",
			con: &schema.BitbucketServerConnection{
				Plugin: &schema.BitbucketServerPlugin{
					Permissions: "",
					Webhooks: &schema.BitbucketServerPluginWebhooks{
						Secret: "",
					},
				},
			},
			secrets: map[int64]string{
				1: "",
			},
			expect: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := new(requestRecorder)
			h := NewBitbucketServerWebhook(nil, nil, time.Now, "testhook")
			h.secrets = tc.secrets
			h.httpClient = rec

			err := h.syncWebhook(1, tc.con, "http://example.com/")
			if err != nil {
				t.Fatal(err)
			}
			methods := make([]string, len(rec.requests))
			for i := range rec.requests {
				methods[i] = rec.requests[i].Method
			}
			if diff := cmp.Diff(tc.expect, methods); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

type requestRecorder struct {
	requests []*http.Request
}

func (r *requestRecorder) Do(req *http.Request) (*http.Response, error) {
	r.requests = append(r.requests, req)
	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}, nil
}

func sign(t *testing.T, message, secret []byte) string {
	t.Helper()

	mac := hmac.New(sha256.New, secret)

	_, err := mac.Write(message)
	if err != nil {
		t.Fatalf("writing hmac message failed: %s", err)
	}

	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func marshalJSON(t testing.TB, v interface{}) string {
	t.Helper()

	bs, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	return string(bs)
}

func newGithubClientFactory(t testing.TB, name string) (*httpcli.Factory, func()) {
	t.Helper()

	cassete := filepath.Join("testdata/vcr/", strings.Replace(name, " ", "-", -1))

	rec, err := httptestutil.NewRecorder(cassete, *update, func(i *cassette.Interaction) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	mw := httpcli.NewMiddleware(githubProxyRedirectMiddleware)

	hc := httpcli.NewFactory(mw, httptestutil.NewRecorderOpt(rec))

	return hc, func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to update test data: %s", err)
		}
	}
}

func githubProxyRedirectMiddleware(cli httpcli.Doer) httpcli.Doer {
	return httpcli.DoerFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Hostname() == "github-proxy" {
			req.URL.Host = "api.github.com"
			req.URL.Scheme = "https"
		}
		return cli.Do(req)
	})
}
