package a8n

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
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/google/go-cmp/cmp"
	gh "github.com/google/go-github/github"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/schema"
)

var update = flag.Bool("update", false, "update testdata")

// Ran in integration_test.go
func testGitHubWebhook(db *sql.DB) func(*testing.T) {
	return func(t *testing.T) {
		now := time.Now()
		clock := func() time.Time {
			return now.UTC().Truncate(time.Microsecond)
		}
		now = clock()

		ctx := context.Background()

		rcache.SetupForTest(t)

		cf, save := newGithubClientFactory(t, "github-webhooks")
		defer save()

		var userID int32
		err := db.QueryRow("INSERT INTO users (username) VALUES ('admin') RETURNING id").Scan(&userID)
		if err != nil {
			t.Fatal(err)
		}

		secret := "secret"
		repoStore := repos.NewDBStore(db, sql.TxOptions{})
		githubExtSvc := &repos.ExternalService{
			Kind:        "GITHUB",
			DisplayName: "GitHub",
			Config: marshalJSON(t, &schema.GitHubConnection{
				Url:      "https://github.com",
				Token:    os.Getenv("GITHUB_TOKEN"),
				Repos:    []string{"oklog/ulid"},
				Webhooks: []*schema.GitHubWebhook{{Org: "oklog", Secret: secret}},
			}),
		}

		err = repoStore.UpsertExternalServices(ctx, githubExtSvc)
		if err != nil {
			t.Fatal(t)
		}

		githubSrc, err := repos.NewGithubSource(githubExtSvc, cf)
		if err != nil {
			t.Fatal(t)
		}

		githubRepo, err := githubSrc.GetRepo(ctx, "oklog/ulid")
		if err != nil {
			t.Fatal(err)
		}

		err = repoStore.UpsertRepos(ctx, githubRepo)
		if err != nil {
			t.Fatal(err)
		}

		store := NewStoreWithClock(db, clock)

		campaign := &a8n.Campaign{
			Name:            "Test campaign",
			Description:     "Testing THE WEBHOOKS",
			AuthorID:        userID,
			NamespaceUserID: userID,
		}

		err = store.CreateCampaign(ctx, campaign)
		if err != nil {
			t.Fatal(err)
		}

		changesets := []*a8n.Changeset{
			{
				RepoID:              int32(githubRepo.ID),
				ExternalID:          "16",
				ExternalServiceType: githubRepo.ExternalRepo.ServiceType,
				CampaignIDs:         []int64{campaign.ID},
			},
		}

		err = store.CreateChangesets(ctx, changesets...)
		if err != nil {
			t.Fatal(err)
		}

		syncer := ChangesetSyncer{
			ReposStore:  repoStore,
			Store:       store,
			HTTPFactory: cf,
		}

		err = syncer.SyncChangesets(ctx, changesets...)
		if err != nil {
			t.Fatal(err)
		}

		_, err = db.Exec("DELETE FROM changeset_events")
		if err != nil {
			t.Fatal(err)
		}

		fs := loadFixtures(t)
		hook := &GitHubWebhook{Store: store, Repos: repoStore, Now: clock}

		issueComment := github.IssueComment{
			DatabaseID: 540540777,
			Author: github.Actor{
				AvatarURL: "https://avatars2.githubusercontent.com/u/67471?v=4",
				Login:     "tsenart",
				URL:       "https://api.github.com/users/tsenart",
			},
			Editor: &github.Actor{
				AvatarURL: "https://avatars2.githubusercontent.com/u/67471?v=4",
				Login:     "tsenart",
				URL:       "https://api.github.com/users/tsenart",
			},
			AuthorAssociation:   "CONTRIBUTOR",
			Body:                "A comment on an old event. Aaaand it was updated. Twice. Thrice. Four times even.",
			URL:                 "https://api.github.com/repos/oklog/ulid/issues/comments/540540777",
			CreatedAt:           parseTimestamp(t, "2019-10-10T12:06:54Z"),
			UpdatedAt:           parseTimestamp(t, "2019-10-10T12:15:20Z"),
			IncludesCreatedEdit: true,
		}

		events := []*a8n.ChangesetEvent{
			{
				ID:          7,
				ChangesetID: changesets[0].ID,
				Kind:        "github:commented",
				Key:         "540540777",
				CreatedAt:   now,
				UpdatedAt:   now,
				Metadata: func() interface{} {
					m := issueComment
					return &m
				}(),
			},
		}

		for _, tc := range []struct {
			name   string
			secret string
			event  event
			code   int
			want   []*a8n.ChangesetEvent
		}{
			{
				name:   "unauthorized",
				secret: "wrong-secret",
				event:  fs["issue_comment-edited"],
				code:   http.StatusUnauthorized,
				want:   []*a8n.ChangesetEvent{},
			},
			{
				name:   "non-existent-changeset",
				secret: secret,
				event: func() event {
					e := fs["issue_comment-edited"]
					clone := *(e.event.(*gh.IssueCommentEvent))
					issue := *clone.Issue
					clone.Issue = &issue
					nonExistingPRNumber := 999999
					issue.Number = &nonExistingPRNumber
					return event{name: e.name, event: &clone}
				}(),
				code: http.StatusOK,
				want: []*a8n.ChangesetEvent{},
			},
			{
				name:   "non-existent-changeset-event",
				secret: secret,
				event:  fs["issue_comment-edited"],
				code:   http.StatusOK,
				want:   events,
			},
			{
				name:   "existent-changeset-event",
				secret: secret,
				event: func() event {
					e := fs["issue_comment-edited"]
					clone := *(e.event.(*gh.IssueCommentEvent))
					comment := *clone.Comment
					clone.Comment = &comment
					body := "Foo bar"
					comment.Body = &body
					return event{name: e.name, event: &clone}
				}(),
				code: http.StatusOK,
				want: func() []*a8n.ChangesetEvent {
					m := issueComment
					m.Body = "Foo bar"
					e := events[0].Clone()
					e.Metadata = &m
					return []*a8n.ChangesetEvent{e}
				}(),
			},
		} {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				body, err := json.Marshal(tc.event.event)
				if err != nil {
					t.Fatal(err)
				}

				req, err := http.NewRequest("POST", "", bytes.NewReader(body))
				if err != nil {
					t.Fatal(err)
				}

				req.Header.Set("X-Github-Event", tc.event.name)
				req.Header.Set("X-Hub-Signature", sign(t, body, []byte(tc.secret)))

				rec := httptest.NewRecorder()
				hook.ServeHTTP(rec, req)
				resp := rec.Result()

				if tc.code != 0 && tc.code != resp.StatusCode {
					bs, err := httputil.DumpResponse(resp, true)
					if err != nil {
						t.Fatal(err)
					}

					t.Log(string(bs))
					t.Errorf("have status code %d, want %d", resp.StatusCode, tc.code)
				}

				have, _, err := store.ListChangesetEvents(ctx, ListChangesetEventsOpts{Limit: 1000})
				if err != nil {
					t.Fatal(err)
				}

				if diff := cmp.Diff(have, tc.want); diff != "" {
					t.Error(diff)
				}
			})
		}
	}
}

type event struct {
	name  string
	event interface{}
}

func loadFixtures(t testing.TB) map[string]event {
	t.Helper()

	matches, err := filepath.Glob("testdata/fixtures/*")
	if err != nil {
		t.Fatal(err)
	}

	fs := make(map[string]event, len(matches))
	for _, m := range matches {
		bs, err := ioutil.ReadFile(m)
		if err != nil {
			t.Fatal(err)
		}

		base := filepath.Base(m)
		name := strings.TrimSuffix(base, filepath.Ext(base))
		parts := strings.SplitN(name, "-", 2)

		if len(parts) != 2 {
			t.Fatalf("unexpected fixture file name format: %s", m)
		}

		ev, err := gh.ParseWebHook(parts[0], bs)
		if err != nil {
			t.Fatal(err)
		}

		fs[name] = event{name: parts[0], event: ev}
	}

	return fs
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

func parseTimestamp(t testing.TB, ts string) time.Time {
	t.Helper()

	timestamp, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t.Fatal(err)
	}

	return timestamp
}
