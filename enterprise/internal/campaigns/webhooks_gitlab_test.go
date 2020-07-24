package campaigns

import (
	"bytes"
	"context"
	"database/sql"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab/webhooks"
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
				assertBodyIncludes(t, resp.Body, "unknown event type")
			})

			// TODO: valid event types.
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

// createGitLabExternalService creates a mock GitLab service with a valid
// configuration, including the secrets "super" and "secret".
func createGitLabExternalService(t *testing.T, ctx context.Context, rstore repos.Store) *repos.ExternalService {
	es := &repos.ExternalService{
		Kind:        extsvc.KindGitLab,
		DisplayName: "gitlab",
		Config: marshalJSON(t, &schema.GitLabConnection{
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
