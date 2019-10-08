package a8n

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/pkg/a8n"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/schema"
)

var update = flag.Bool("update", false, "update testdata")

func TestCalcCounts(t *testing.T) {
	ghChangeset := loadGithubChangeset(t, "sourcegraph/sourcegraph", "5834")

	// Date when PR #5834 was created: "2019-10-02T14:49:31Z"
	// We start exactly one day earlier
	start := parseJSONTime(t, "2019-10-01T14:49:31Z")
	// Date when PR #5834 was merged:  "2019-10-07T13:13:45Z"
	end := parseJSONTime(t, "2019-10-07T13:13:45Z")

	have, err := CalcCounts(start, end, ghChangeset)
	if err != nil {
		return
	}

	want := []*ChangesetCounts{
		{Time: end.Add(5 * -24 * time.Hour), Total: 0, Open: 0},
		{Time: end.Add(4 * -24 * time.Hour), Total: 1, Open: 1},
		{Time: end.Add(3 * -24 * time.Hour), Total: 1, Open: 1},
		{Time: end.Add(2 * -24 * time.Hour), Total: 1, Open: 1},
		{Time: end.Add(1 * -24 * time.Hour), Total: 1, Open: 1},
		{Time: end, Total: 1, Closed: 1, Merged: 1},
	}

	if !reflect.DeepEqual(have, want) {
		t.Errorf("wrong counts listed. diff=%s", cmp.Diff(have, want))
	}
}

func loadGithubChangeset(t testing.TB, repoWithOwner, number string) *a8n.Changeset {
	file := fmt.Sprintf("pullrequest-%s-%s.json", strings.Replace(repoWithOwner, "/", "-", -1), number)
	path := filepath.Join("testdata/fixtures/", file)

	cs := &a8n.Changeset{ExternalID: number}

	if *update {
		cf := httpcli.NewFactory(githubProxyRedirectMiddleware)

		githubExtSvc := &repos.ExternalService{
			Kind: "GITHUB",
			Config: marshalJSON(t, &schema.GitHubConnection{
				Url:   "https://github.com",
				Token: os.Getenv("GITHUB_TOKEN"),
			}),
		}

		githubSrc, err := repos.NewGithubSource(githubExtSvc, cf)
		if err != nil {
			t.Fatal(t)
		}

		repoChangeset := &repos.Changeset{
			Repo:      &repos.Repo{Metadata: &github.Repository{NameWithOwner: repoWithOwner}},
			Changeset: cs,
		}
		ctx := context.Background()

		err = githubSrc.LoadChangesets(ctx, repoChangeset)
		if err != nil {
			t.Fatalf("failed to load changeset: %s", err)
		}

		data, err := json.MarshalIndent(cs.Metadata, " ", " ")
		if err != nil {
			t.Fatal(err)
		}

		if err = ioutil.WriteFile(path, data, 0640); err != nil {
			t.Fatalf("failed to write changeset file %q: %s", path, err)
		}

		return cs
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read changeset file %q: %s", path, err)
	}

	var meta *github.PullRequest
	if err = json.Unmarshal(data, &meta); err != nil {
		t.Fatalf("failed to unmarshal changeset: %s", err)
	}

	cs.Metadata = meta

	return cs
}

func parseJSONTime(t testing.TB, ts string) time.Time {
	t.Helper()

	timestamp, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		t.Fatal(err)
	}

	return timestamp
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

func marshalJSON(t testing.TB, v interface{}) string {
	t.Helper()

	bs, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	return string(bs)
}
