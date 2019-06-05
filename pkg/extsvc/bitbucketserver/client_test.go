package bitbucketserver_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/pkg/httptestutil"
)

var update = flag.Bool("update", false, "update testdata")

func TestUserFilters(t *testing.T) {
	for _, tc := range []struct {
		name string
		fs   bitbucketserver.UserFilters
		qry  url.Values
	}{
		{
			name: "last one wins",
			fs: bitbucketserver.UserFilters{
				{Filter: "admin"},
				{Filter: "tomas"}, // Last one wins
			},
			qry: url.Values{"filter": []string{"tomas"}},
		},
		{
			name: "filters can be combined",
			fs: bitbucketserver.UserFilters{
				{Filter: "admin"},
				{Group: "admins"},
			},
			qry: url.Values{
				"filter": []string{"admin"},
				"group":  []string{"admins"},
			},
		},
		{
			name: "permissions",
			fs: bitbucketserver.UserFilters{
				{
					Permission: bitbucketserver.PermissionFilter{
						Root:       bitbucketserver.PermProjectAdmin,
						ProjectKey: "ORG",
					},
				},
				{
					Permission: bitbucketserver.PermissionFilter{
						Root:           bitbucketserver.PermRepoWrite,
						ProjectKey:     "ORG",
						RepositorySlug: "foo",
					},
				},
			},
			qry: url.Values{
				"permission.1":                []string{"PROJECT_ADMIN"},
				"permission.1.projectKey":     []string{"ORG"},
				"permission.2":                []string{"REPO_WRITE"},
				"permission.2.projectKey":     []string{"ORG"},
				"permission.2.repositorySlug": []string{"foo"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			have := make(url.Values)
			tc.fs.EncodeTo(have)
			if want := tc.qry; !reflect.DeepEqual(have, want) {
				t.Error(cmp.Diff(have, want))
			}
		})
	}
}

func TestClient_Users(t *testing.T) {
	cli, save := newClient(t, "Users")
	defer save()

	timeout, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	for _, tc := range []struct {
		name    string
		ctx     context.Context
		page    *bitbucketserver.PageToken
		filters []bitbucketserver.UserFilter
		next    *bitbucketserver.PageToken
		err     string
	}{
		{
			name: "timeout",
			ctx:  timeout,
			err:  "context deadline exceeded",
		},
		{
			name: "pagination: first page",
			page: &bitbucketserver.PageToken{Limit: 1},
			next: &bitbucketserver.PageToken{
				Size:          1,
				Limit:         1,
				NextPageStart: 1,
			},
		},
		{
			name: "pagination: last page",
			page: &bitbucketserver.PageToken{
				Size:          1,
				Limit:         1,
				NextPageStart: 1,
			},
			next: &bitbucketserver.PageToken{
				Size:       1,
				Start:      1,
				Limit:      1,
				IsLastPage: true,
			},
		},
		{
			name:    "filter by substring match in username, name and email address",
			page:    &bitbucketserver.PageToken{Limit: 1000},
			filters: []bitbucketserver.UserFilter{{Filter: "Doe"}}, // matches "John Doe" in name
			next: &bitbucketserver.PageToken{
				Size:       1,
				Limit:      1000,
				IsLastPage: true,
			},
		},
		{
			name:    "filter by group",
			page:    &bitbucketserver.PageToken{Limit: 1000},
			filters: []bitbucketserver.UserFilter{{Group: "admins"}},
			next: &bitbucketserver.PageToken{
				Size:       1,
				Limit:      1000,
				IsLastPage: true,
			},
		},
		{
			name: "filter by multiple ANDed permissions",
			page: &bitbucketserver.PageToken{Limit: 1000},
			filters: []bitbucketserver.UserFilter{
				{
					Permission: bitbucketserver.PermissionFilter{
						Root: bitbucketserver.PermSysAdmin,
					},
				},
				{
					Permission: bitbucketserver.PermissionFilter{
						Root:           bitbucketserver.PermRepoRead,
						ProjectKey:     "ORG",
						RepositorySlug: "foo",
					},
				},
			},
			next: &bitbucketserver.PageToken{
				Size:       1,
				Limit:      1000,
				IsLastPage: true,
			},
		},
		{
			name: "multiple filters are ANDed",
			page: &bitbucketserver.PageToken{Limit: 1000},
			filters: []bitbucketserver.UserFilter{
				{
					Filter: "admin",
				},
				{
					Permission: bitbucketserver.PermissionFilter{
						Root:           bitbucketserver.PermRepoRead,
						ProjectKey:     "ORG",
						RepositorySlug: "foo",
					},
				},
			},
			next: &bitbucketserver.PageToken{
				Size:       1,
				Limit:      1000,
				IsLastPage: true,
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

			users, next, err := cli.Users(tc.ctx, tc.page, tc.filters...)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if have, want := next, tc.next; !reflect.DeepEqual(have, want) {
				t.Error(cmp.Diff(have, want))
			}

			if err != nil {
				return
			}

			bs, err := json.MarshalIndent(users, "", "  ")
			if err != nil {
				t.Fatalf("failed to marshal users: %s", err)
			}

			path := fmt.Sprintf("testdata/golden/Users-%s.json", normalize(tc.name))
			if *update {
				if err = ioutil.WriteFile(path, bs, 0640); err != nil {
					t.Fatalf("failed to update golden file %q: %s", path, err)
				}
				t.Skipf("Updated %s successfully. Skipping.", path)
			}

			golden, err := ioutil.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read golden file %q: %s", path, err)
			}

			if have, want := string(bs), string(golden); have != want {
				dmp := diffmatchpatch.New()
				diffs := dmp.DiffMain(have, want, false)
				t.Error(dmp.DiffPrettyText(diffs))
			}
		})
	}
}

func newClient(t testing.TB, name string) (*bitbucketserver.Client, func()) {
	t.Helper()

	cassete := filepath.Join("testdata/vcr/", normalize(name))
	rec, err := httptestutil.NewRecorder(cassete, *update)
	if err != nil {
		t.Fatal(err)
	}

	hc, err := httpcli.NewFactory(nil, httptestutil.NewRecorderOpt(rec)).Doer()
	if err != nil {
		t.Fatal(err)
	}

	instanceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instanceURL == "" {
		instanceURL = "http://localhost:7990"
	}

	u, err := url.Parse(instanceURL)
	if err != nil {
		t.Fatal(err)
	}

	cli := bitbucketserver.NewClient(u, hc)
	cli.Token = os.Getenv("BITBUCKET_SERVER_TOKEN")

	return cli, func() {
		if err := rec.Stop(); err != nil {
			t.Errorf("failed to update test data: %s", err)
		}
	}
}

var normalizer = regexp.MustCompile("[^A-Za-z0-9-]+")

func normalize(path string) string {
	return normalizer.ReplaceAllLiteralString(path, "-")
}
