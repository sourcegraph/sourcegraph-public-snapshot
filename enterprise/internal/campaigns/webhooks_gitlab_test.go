package campaigns

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab/webhooks"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/schema"
)

func testGitLabWebhook(db *sql.DB, userID int32) func(*testing.T) {
	return func(t *testing.T) {
		ctx := context.Background()

		t.Run("ServeHTTP", func(t *testing.T) {
			t.Run("missing external service", func(t *testing.T) {
				store, rstore, clock := gitLabTestSetup(t, db)
				h := NewGitLabWebhook(store, rstore, clock.now)

				u := extsvc.WebhookURL(extsvc.TypeGitLab, 12345, "https://example.com/")
				req, err := http.NewRequest("POST", u, nil)
				if err != nil {
					t.Fatal(err)
				}

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				if have, want := rec.Result().StatusCode, http.StatusUnauthorized; have != want {
					t.Errorf("unexpected status code: have %d; want %d", have, want)
				}
			})

			t.Run("invalid external service", func(t *testing.T) {
				store, rstore, clock := gitLabTestSetup(t, db)
				h := NewGitLabWebhook(store, rstore, clock.now)

				u := strings.ReplaceAll(extsvc.WebhookURL(extsvc.TypeGitLab, 12345, "https://example.com/"), "12345", "foo")
				req, err := http.NewRequest("POST", u, nil)
				if err != nil {
					t.Fatal(err)
				}

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				resp := rec.Result()
				if have, want := resp.StatusCode, http.StatusInternalServerError; have != want {
					t.Errorf("unexpected status code: have %d; want %d", have, want)
				}
				assertBodyIncludes(t, resp.Body, "getting external service")
			})

			t.Run("malformed external service", func(t *testing.T) {
				store, rstore, clock := gitLabTestSetup(t, db)
				h := NewGitLabWebhook(store, rstore, clock.now)
				es := createGitLabExternalService(t, ctx, rstore)

				es.Config = "invalid JSON"
				if err := rstore.UpsertExternalServices(ctx, es); err != nil {
					t.Fatal(err)
				}

				u := extsvc.WebhookURL(extsvc.TypeGitLab, es.ID, "https://example.com/")
				req, err := http.NewRequest("POST", u, nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add(webhooks.TokenHeaderName, "not a valid secret")

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				resp := rec.Result()
				if have, want := resp.StatusCode, http.StatusInternalServerError; have != want {
					t.Errorf("unexpected status code: have %d; want %d", have, want)
				}
				assertBodyIncludes(t, resp.Body, "validating the shared secret")
			})

			t.Run("missing secret", func(t *testing.T) {
				store, rstore, clock := gitLabTestSetup(t, db)
				h := NewGitLabWebhook(store, rstore, clock.now)
				es := createGitLabExternalService(t, ctx, rstore)

				u := extsvc.WebhookURL(extsvc.TypeGitLab, es.ID, "https://example.com/")
				req, err := http.NewRequest("POST", u, nil)
				if err != nil {
					t.Fatal(err)
				}

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				resp := rec.Result()
				if have, want := resp.StatusCode, http.StatusUnauthorized; have != want {
					t.Errorf("unexpected status code: have %d; want %d", have, want)
				}
				assertBodyIncludes(t, resp.Body, "shared secret is incorrect")
			})

			t.Run("incorrect secret", func(t *testing.T) {
				store, rstore, clock := gitLabTestSetup(t, db)
				h := NewGitLabWebhook(store, rstore, clock.now)
				es := createGitLabExternalService(t, ctx, rstore)

				u := extsvc.WebhookURL(extsvc.TypeGitLab, es.ID, "https://example.com/")
				req, err := http.NewRequest("POST", u, nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add(webhooks.TokenHeaderName, "not a valid secret")

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				resp := rec.Result()
				if have, want := resp.StatusCode, http.StatusUnauthorized; have != want {
					t.Errorf("unexpected status code: have %d; want %d", have, want)
				}
				assertBodyIncludes(t, resp.Body, "shared secret is incorrect")
			})

			t.Run("missing body", func(t *testing.T) {
				store, rstore, clock := gitLabTestSetup(t, db)
				h := NewGitLabWebhook(store, rstore, clock.now)
				es := createGitLabExternalService(t, ctx, rstore)

				u := extsvc.WebhookURL(extsvc.TypeGitLab, es.ID, "https://example.com/")
				req, err := http.NewRequest("POST", u, nil)
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add(webhooks.TokenHeaderName, "secret")

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				resp := rec.Result()
				if have, want := resp.StatusCode, http.StatusBadRequest; have != want {
					t.Errorf("unexpected status code: have %d; want %d", have, want)
				}
				assertBodyIncludes(t, resp.Body, "missing request body")
			})

			t.Run("malformed body", func(t *testing.T) {
				store, rstore, clock := gitLabTestSetup(t, db)
				h := NewGitLabWebhook(store, rstore, clock.now)
				es := createGitLabExternalService(t, ctx, rstore)

				u := extsvc.WebhookURL(extsvc.TypeGitLab, es.ID, "https://example.com/")
				req, err := http.NewRequest("POST", u, bytes.NewBufferString("invalid JSON"))
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add(webhooks.TokenHeaderName, "secret")

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				resp := rec.Result()
				if have, want := resp.StatusCode, http.StatusInternalServerError; have != want {
					t.Errorf("unexpected status code: have %d; want %d", have, want)
				}
				assertBodyIncludes(t, resp.Body, "unmarshalling payload")
			})

			t.Run("invalid body", func(t *testing.T) {
				store, rstore, clock := gitLabTestSetup(t, db)
				h := NewGitLabWebhook(store, rstore, clock.now)
				es := createGitLabExternalService(t, ctx, rstore)

				u := extsvc.WebhookURL(extsvc.TypeGitLab, es.ID, "https://example.com/")
				body := marshalJSON(t, &webhooks.EventCommon{
					ObjectKind: "unknown",
				})
				req, err := http.NewRequest("POST", u, bytes.NewBufferString(body))
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add(webhooks.TokenHeaderName, "secret")

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				resp := rec.Result()
				if have, want := resp.StatusCode, http.StatusNotImplemented; have != want {
					t.Errorf("unexpected status code: have %d; want %d", have, want)
				}
				assertBodyIncludes(t, resp.Body, "unknown object kind")
			})

			// The valid tests below are pretty "happy path": specific unit
			// tests for the utility methods on GitLabWebhook are below. We're
			// mostly just testing the routing here, since these are ServeHTTP
			// tests; however, these also act as integration tests. (Which,
			// considering they're ultimately invoked from TestIntegration,
			// seems fair.)

			t.Run("valid merge request approval events", func(t *testing.T) {
				for _, action := range []string{"approved", "unapproved"} {
					t.Run(action, func(t *testing.T) {
						store, rstore, clock := gitLabTestSetup(t, db)
						h := NewGitLabWebhook(store, rstore, clock.now)
						es := createGitLabExternalService(t, ctx, rstore)
						repo := createGitLabRepo(t, ctx, rstore, es)
						changeset := createGitLabChangeset(t, ctx, store, repo)
						body := createMergeRequestPayload(t, repo, changeset, "approved")

						u := extsvc.WebhookURL(extsvc.TypeGitLab, es.ID, "https://example.com/")
						req, err := http.NewRequest("POST", u, bytes.NewBufferString(body))
						if err != nil {
							t.Fatal(err)
						}
						req.Header.Add(webhooks.TokenHeaderName, "secret")

						repoupdater.MockEnqueueChangesetSync = func(ctx context.Context, ids []int64) error {
							if diff := cmp.Diff(ids, []int64{changeset.ID}); diff != "" {
								t.Errorf("unexpected changeset ID: %s", diff)
							}
							return nil
						}
						defer func() { repoupdater.MockEnqueueChangesetSync = nil }()

						rec := httptest.NewRecorder()
						h.ServeHTTP(rec, req)

						resp := rec.Result()
						if have, want := resp.StatusCode, http.StatusNoContent; have != want {
							t.Errorf("unexpected status code: have %d; want %d", have, want)
						}
					})
				}
			})

			t.Run("valid merge request state change events", func(t *testing.T) {
				for action, want := range map[string]campaigns.ChangesetEventKind{
					"close":  campaigns.ChangesetEventKindGitLabClosed,
					"merge":  campaigns.ChangesetEventKindGitLabMerged,
					"reopen": campaigns.ChangesetEventKindGitLabReopened,
				} {
					t.Run(action, func(t *testing.T) {
						store, rstore, clock := gitLabTestSetup(t, db)
						h := NewGitLabWebhook(store, rstore, clock.now)
						es := createGitLabExternalService(t, ctx, rstore)
						repo := createGitLabRepo(t, ctx, rstore, es)
						changeset := createGitLabChangeset(t, ctx, store, repo)
						body := createMergeRequestPayload(t, repo, changeset, action)

						u := extsvc.WebhookURL(extsvc.TypeGitLab, es.ID, "https://example.com/")
						req, err := http.NewRequest("POST", u, bytes.NewBufferString(body))
						if err != nil {
							t.Fatal(err)
						}
						req.Header.Add(webhooks.TokenHeaderName, "secret")

						rec := httptest.NewRecorder()
						h.ServeHTTP(rec, req)

						resp := rec.Result()
						if have, want := resp.StatusCode, http.StatusNoContent; have != want {
							t.Errorf("unexpected status code: have %d; want %d", have, want)
						}

						// Verify that the changeset event was upserted.
						assertChangesetEventForChangeset(t, ctx, store, changeset, want)
					})
				}
			})

			t.Run("valid pipeline events", func(t *testing.T) {
				store, rstore, clock := gitLabTestSetup(t, db)
				h := NewGitLabWebhook(store, rstore, clock.now)
				es := createGitLabExternalService(t, ctx, rstore)
				repo := createGitLabRepo(t, ctx, rstore, es)
				changeset := createGitLabChangeset(t, ctx, store, repo)
				body := createPipelinePayload(t, repo, changeset, gitlab.Pipeline{
					ID:     123,
					Status: gitlab.PipelineStatusSuccess,
				})

				u := extsvc.WebhookURL(extsvc.TypeGitLab, es.ID, "https://example.com/")
				req, err := http.NewRequest("POST", u, bytes.NewBufferString(body))
				if err != nil {
					t.Fatal(err)
				}
				req.Header.Add(webhooks.TokenHeaderName, "secret")

				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, req)

				resp := rec.Result()
				if have, want := resp.StatusCode, http.StatusNoContent; have != want {
					t.Errorf("unexpected status code: have %d; want %d", have, want)
				}

				assertChangesetEventForChangeset(t, ctx, store, changeset, campaigns.ChangesetEventKindGitLabPipeline)
			})
		})
	}
}

// nestedTx wraps an existing transaction and overrides its transaction methods
// to be no-ops. This allows us to have a master transaction used in tests that
// test functions that attempt to create and commit transactions: since
// PostgreSQL doesn't support nested transactions, we can still use the master
// transaction to manage the test database state without rollback/commit
// already performed errors.
//
// It would be theoretically possible to use savepoints to implement something
// resembling the semantics of a true nested transaction, but that's
// unnecessary for these tests.
type nestedTx struct {
	*sql.Tx
}

func (ntx *nestedTx) Rollback() error                                        { return nil }
func (ntx *nestedTx) Commit() error                                          { return nil }
func (ntx *nestedTx) BeginTx(ctx context.Context, opts *sql.TxOptions) error { return nil }

// gitLabTestSetup instantiates the stores and a clock for use within tests.
// Any changes made to the stores will be rolled back after the test is
// complete.
func gitLabTestSetup(t *testing.T, db *sql.DB) (*Store, repos.Store, clock) {
	c := &testClock{t: time.Now().UTC().Truncate(time.Microsecond)}
	tx := dbtest.NewTx(t, db)

	// Note that tx is wrapped in nestedTx to effectively neuter further use of
	// transactions within the test.
	return NewStoreWithClock(&nestedTx{tx}, c.now), repos.NewDBStore(tx, sql.TxOptions{}), c
}

func assertBodyIncludes(t *testing.T, r io.Reader, want string) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(body, []byte(want)) {
		t.Errorf("cannot find expected string in output: want: %s; have:\n%s", want, string(body))
	}
}

func assertChangesetEventForChangeset(t *testing.T, ctx context.Context, store *Store, changeset *campaigns.Changeset, want campaigns.ChangesetEventKind) {
	ces, _, err := store.ListChangesetEvents(ctx, ListChangesetEventsOpts{
		ChangesetIDs: []int64{changeset.ID},
		Limit:        100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ces) == 1 {
		ce := ces[0]

		if ce.ChangesetID != changeset.ID {
			t.Errorf("unexpected changeset ID: have %d; want %d", ce.ChangesetID, changeset.ID)
		}
		if ce.Kind != want {
			t.Errorf(
				"unexpected changeset event kind: have %v; want %v", ce.Kind, want)
		}
	} else {
		t.Errorf("unexpected number of changeset events; got %+v", ces)
	}
}

// createGitLabExternalService creates a mock GitLab service with a valid
// configuration, including the secrets "super" and "secret".
func createGitLabExternalService(t *testing.T, ctx context.Context, rstore repos.Store) *repos.ExternalService {
	es := &repos.ExternalService{
		Kind:        extsvc.KindGitLab,
		DisplayName: "gitlab",
		Config: marshalJSON(t, &schema.GitLabConnection{
			Url: "https://gitlab.com/",
			Webhooks: []*schema.GitLabWebhook{
				{Secret: "super"},
				{Secret: "secret"},
			},
		}),
	}
	if err := rstore.UpsertExternalServices(ctx, es); err != nil {
		t.Fatal(err)
	}

	return es
}

func createGitLabRepo(t *testing.T, ctx context.Context, rstore repos.Store, es *repos.ExternalService) *repos.Repo {
	repo := (&repos.Repo{
		Name: "gitlab.com/sourcegraph/test",
		URI:  "gitlab.com/sourcegraph/test",
		ExternalRepo: api.ExternalRepoSpec{
			ID:          "123",
			ServiceType: extsvc.TypeGitLab,
			ServiceID:   "https://gitlab.com/",
		},
	}).With(repos.Opt.RepoSources(es.URN()))
	if err := rstore.UpsertRepos(ctx, repo); err != nil {
		t.Fatal(err)
	}

	return repo
}

func createGitLabChangeset(t *testing.T, ctx context.Context, store *Store, repo *repos.Repo) *campaigns.Changeset {
	c := &campaigns.Changeset{
		RepoID:              repo.ID,
		ExternalID:          "1",
		ExternalServiceType: extsvc.TypeGitLab,
	}
	if err := store.CreateChangesets(ctx, c); err != nil {
		t.Fatal(err)
	}

	return c
}

func createMergeRequestPayload(t *testing.T, repo *repos.Repo, changeset *campaigns.Changeset, action string) string {
	cid, err := strconv.Atoi(changeset.ExternalID)
	if err != nil {
		t.Fatal(err)
	}

	pid, err := strconv.Atoi(repo.ExternalRepo.ID)
	if err != nil {
		t.Fatal(err)
	}

	// We use an untyped set of maps here because the webhooks package doesn't
	// export its internal mergeRequestEvent type that is used for
	// unmarshalling. (Which is fine; it's an implementation detail.)
	return marshalJSON(t, map[string]interface{}{
		"object_kind": "merge_request",
		"project": map[string]interface{}{
			"id": pid,
		},
		"object_attributes": map[string]interface{}{
			"iid":    cid,
			"action": action,
		},
	})
}

func createPipelinePayload(t *testing.T, repo *repos.Repo, changeset *campaigns.Changeset, pipeline gitlab.Pipeline) string {
	pid, err := strconv.Atoi(repo.ExternalRepo.ID)
	if err != nil {
		t.Fatal(err)
	}

	payload := &webhooks.PipelineEvent{
		EventCommon: webhooks.EventCommon{
			ObjectKind: "pipeline",
			Project: gitlab.ProjectCommon{
				ID: pid,
			},
		},
		Pipeline: pipeline,
	}

	if changeset != nil {
		cid, err := strconv.Atoi(changeset.ExternalID)
		if err != nil {
			t.Fatal(err)
		}

		payload.MergeRequest = &gitlab.MergeRequest{
			IID: gitlab.ID(cid),
		}
	}

	return marshalJSON(t, payload)
}
