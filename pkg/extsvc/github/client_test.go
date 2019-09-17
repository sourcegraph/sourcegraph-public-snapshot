package github

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/pkg/errors"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/pkg/httptestutil"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
)

func TestUnmarshal(t *testing.T) {
	type result struct {
		FieldA string
		FieldB string
	}
	cases := map[string]string{
		// Valid
		`[]`:                                  "",
		`[{"FieldA": "hi"}]`:                  "",
		`[{"FieldA": "hi", "FieldB": "bye"}]`: "",

		// Error
		`[[]]`:            `graphql: cannot unmarshal at offset 2: before "[["; after "]]": json: cannot unmarshal array into Go value of type github.result`,
		`[{"FieldA": 1}]`: `graphql: cannot unmarshal at offset 13: before "[{\"FieldA\": 1"; after "}]": json: cannot unmarshal number`,
	}
	// Large body
	repeated := strings.Repeat(`{"FieldA": "hi", "FieldB": "bye"},`, 100)
	cases[fmt.Sprintf(`[%s {"FieldA": 1}, %s]`, repeated, repeated[:len(repeated)-1])] = `graphql: cannot unmarshal at offset 3414: before ", \"FieldB\": \"bye\"},{\"FieldA\": \"hi\", \"FieldB\": \"bye\"},{\"FieldA\": \"hi\", \"FieldB\": \"bye\"}, {\"FieldA\": 1"; after "}, {\"FieldA\": \"hi\", \"FieldB\": \"bye\"},{\"FieldA\": \"hi\", \"FieldB\": \"bye\"},{\"FieldA\": \"hi\", \"FieldB\": \"b": json: cannot unmarshal number`

	for data, errStr := range cases {
		var a []result
		var b []result
		errA := json.Unmarshal([]byte(data), &a)
		errB := unmarshal([]byte(data), &b)

		if len(data) > 50 {
			data = data[:50] + "..."
		}

		if !reflect.DeepEqual(a, b) {
			t.Errorf("Expected the same result unmarshalling %v\na: %v\nb: %v", data, a, b)
		}
		if !reflect.DeepEqual(errA, errors.Cause(errB)) {
			t.Errorf("Expected the same underlying error unmarshalling %v\na: %v\nb: %v", data, errA, errB)
		}
		got := ""
		if errB != nil {
			got = errB.Error()
		}
		if !strings.HasPrefix(got, errStr) {
			t.Errorf("Unexpected error message %v\ngot:  %s\nwant: %s", data, got, errStr)
		}
	}
}

func TestNewRepoCache_GitHubDotCom(t *testing.T) {
	url, _ := url.Parse("https://www.github.com")
	token := "asdf"

	// github.com caches should:
	// (1) use githubProxyURL for the prefix hash rather than the given url
	// (2) have a TTL of 10 minutes
	key := sha256.Sum256([]byte(token + ":" + githubProxyURL.String()))
	prefix := "gh_repo:" + base64.URLEncoding.EncodeToString(key[:])
	got := NewRepoCache(url, token, "", 0)
	want := rcache.NewWithTTL(prefix, 600)
	if *got != *want {
		t.Errorf("TestNewRepoCache_GitHubDotCom: got %#v, want %#v", *got, *want)
	}
}

func TestNewRepoCache_GitHubEnterprise(t *testing.T) {
	url, _ := url.Parse("https://www.sourcegraph.com")
	token := "asdf"

	// GitHub Enterprise caches should:
	// (1) use the given URL for the prefix hash
	// (2) have a TTL of 30 seconds
	key := sha256.Sum256([]byte(token + ":" + url.String()))
	prefix := "gh_repo:" + base64.URLEncoding.EncodeToString(key[:])
	got := NewRepoCache(url, token, "", 0)
	want := rcache.NewWithTTL(prefix, 30)
	if *got != *want {
		t.Errorf("TestNewRepoCache_GitHubEnterprise: got %#v, want %#v", *got, *want)
	}
}

var update = flag.Bool("update", false, "update testdata")

func TestClient_LoadPullRequests(t *testing.T) {
	cli, save := newClient(t, "LoadPullRequests")
	defer save()

	for i, tc := range []struct {
		name string
		ctx  context.Context
		prs  []*PullRequest
		err  string
	}{
		{
			name: "non-existing-repo",
			prs:  []*PullRequest{{RepoWithOwner: "whoisthis/sourcegraph", Number: 5550}},
			err:  "error in GraphQL response: Could not resolve to a Repository with the name 'sourcegraph'.",
		},
		{
			name: "non-existing-pr",
			prs:  []*PullRequest{{RepoWithOwner: "sourcegraph/sourcegraph", Number: 0}},
			err:  "error in GraphQL response: Could not resolve to a PullRequest with the number of 0.",
		},
		{
			name: "success",
			prs: []*PullRequest{
				{RepoWithOwner: "sourcegraph/sourcegraph", Number: 5550},
				{RepoWithOwner: "sourcegraph/sourcegraph", Number: 5551},
				{RepoWithOwner: "tsenart/vegeta", Number: 50},
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Background()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			err := cli.LoadPullRequests(tc.ctx, tc.prs...)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil {
				return
			}

			data, err := json.MarshalIndent(tc.prs, " ", " ")
			if err != nil {
				t.Fatal(err)
			}

			path := "testdata/golden/LoadPullRequests-" + strconv.Itoa(i)
			if *update {
				if err = ioutil.WriteFile(path, data, 0640); err != nil {
					t.Fatalf("failed to update golden file %q: %s", path, err)
				}
			}

			golden, err := ioutil.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read golden file %q: %s", path, err)
			}

			if have, want := string(data), string(golden); have != want {
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(have, want, false)
				t.Error(dmp.DiffPrettyText(diffs))
			}
		})
	}
}

func newClient(t testing.TB, name string) (*Client, func()) {
	t.Helper()

	cassete := filepath.Join("testdata/vcr/", strings.Replace(name, " ", "-", -1))
	rec, err := httptestutil.NewRecorder(cassete, *update, func(i *cassette.Interaction) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	mw := httpcli.NewMiddleware(githubProxyRedirectMiddleware)

	hc, err := httpcli.NewFactory(mw, httptestutil.NewRecorderOpt(rec)).Doer()
	if err != nil {
		t.Fatal(err)
	}

	uri, err := url.Parse("https://github.com")
	if err != nil {
		t.Fatal(err)
	}

	cli := NewClient(
		uri,
		os.Getenv("GITHUB_TOKEN"),
		hc,
	)

	return cli, func() {
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
