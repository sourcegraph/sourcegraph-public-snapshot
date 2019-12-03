package bitbucketserver

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var update = flag.Bool("update", false, "update testdata")

func TestParseQueryStrings(t *testing.T) {
	for _, tc := range []struct {
		name string
		qs   []string
		vals url.Values
		err  string
	}{
		{
			name: "ignores query separator",
			qs:   []string{"?foo=bar&baz=boo"},
			vals: url.Values{"foo": {"bar"}, "baz": {"boo"}},
		},
		{
			name: "ignores query separator by itself",
			qs:   []string{"?"},
			vals: url.Values{},
		},
		{
			name: "perserves multiple values",
			qs:   []string{"?foo=bar&foo=baz", "foo=boo"},
			vals: url.Values{"foo": {"bar", "baz", "boo"}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == "" {
				tc.err = "<nil>"
			}

			vals, err := parseQueryStrings(tc.qs...)

			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if have, want := vals, tc.vals; !reflect.DeepEqual(have, want) {
				t.Error(cmp.Diff(have, want))
			}
		})
	}
}

func TestUserFilters(t *testing.T) {
	for _, tc := range []struct {
		name string
		fs   UserFilters
		qry  url.Values
	}{
		{
			name: "last one wins",
			fs: UserFilters{
				{Filter: "admin"},
				{Filter: "tomas"}, // Last one wins
			},
			qry: url.Values{"filter": []string{"tomas"}},
		},
		{
			name: "filters can be combined",
			fs: UserFilters{
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
			fs: UserFilters{
				{
					Permission: PermissionFilter{
						Root:       PermProjectAdmin,
						ProjectKey: "ORG",
					},
				},
				{
					Permission: PermissionFilter{
						Root:           PermRepoWrite,
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
	cli, save := NewTestClient(t, "Users", *update)
	defer save()

	timeout, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	users := map[string]*User{
		"admin": {
			Name:         "admin",
			EmailAddress: "tomas@sourcegraph.com",
			ID:           1,
			DisplayName:  "admin",
			Active:       true,
			Slug:         "admin",
			Type:         "NORMAL",
		},
		"john": {
			Name:         "john",
			EmailAddress: "john@mycorp.org",
			ID:           52,
			DisplayName:  "John Doe",
			Active:       true,
			Slug:         "john",
			Type:         "NORMAL",
		},
	}

	for _, tc := range []struct {
		name    string
		ctx     context.Context
		page    *PageToken
		filters []UserFilter
		users   []*User
		next    *PageToken
		err     string
	}{
		{
			name: "timeout",
			ctx:  timeout,
			err:  "context deadline exceeded",
		},
		{
			name:  "pagination: first page",
			page:  &PageToken{Limit: 1},
			users: []*User{users["admin"]},
			next: &PageToken{
				Size:          1,
				Limit:         1,
				NextPageStart: 1,
			},
		},
		{
			name: "pagination: last page",
			page: &PageToken{
				Size:          1,
				Limit:         1,
				NextPageStart: 1,
			},
			users: []*User{users["john"]},
			next: &PageToken{
				Size:       1,
				Start:      1,
				Limit:      1,
				IsLastPage: true,
			},
		},
		{
			name:    "filter by substring match in username, name and email address",
			page:    &PageToken{Limit: 1000},
			filters: []UserFilter{{Filter: "Doe"}}, // matches "John Doe" in name
			users:   []*User{users["john"]},
			next: &PageToken{
				Size:       1,
				Limit:      1000,
				IsLastPage: true,
			},
		},
		{
			name:    "filter by group",
			page:    &PageToken{Limit: 1000},
			filters: []UserFilter{{Group: "admins"}},
			users:   []*User{users["admin"]},
			next: &PageToken{
				Size:       1,
				Limit:      1000,
				IsLastPage: true,
			},
		},
		{
			name: "filter by multiple ANDed permissions",
			page: &PageToken{Limit: 1000},
			filters: []UserFilter{
				{
					Permission: PermissionFilter{
						Root: PermSysAdmin,
					},
				},
				{
					Permission: PermissionFilter{
						Root:           PermRepoRead,
						ProjectKey:     "ORG",
						RepositorySlug: "foo",
					},
				},
			},
			users: []*User{users["admin"]},
			next: &PageToken{
				Size:       1,
				Limit:      1000,
				IsLastPage: true,
			},
		},
		{
			name: "multiple filters are ANDed",
			page: &PageToken{Limit: 1000},
			filters: []UserFilter{
				{
					Filter: "admin",
				},
				{
					Permission: PermissionFilter{
						Root:           PermRepoRead,
						ProjectKey:     "ORG",
						RepositorySlug: "foo",
					},
				},
			},
			users: []*User{users["admin"]},
			next: &PageToken{
				Size:       1,
				Limit:      1000,
				IsLastPage: true,
			},
		},
		{
			name: "maximum 50 permission filters",
			page: &PageToken{Limit: 1000},
			filters: func() (fs UserFilters) {
				for i := 0; i < 51; i++ {
					fs = append(fs, UserFilter{
						Permission: PermissionFilter{
							Root: PermSysAdmin,
						},
					})
				}
				return fs
			}(),
			err: ErrUserFiltersLimit.Error(),
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

			if have, want := users, tc.users; !reflect.DeepEqual(have, want) {
				t.Error(cmp.Diff(have, want))
			}
		})
	}
}

func TestClient_LoadPullRequest(t *testing.T) {
	instanceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instanceURL == "" {
		instanceURL = "http://127.0.0.1:7990"
	}

	cli, save := NewTestClient(t, "PullRequests", *update)
	defer save()

	timeout, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	pr := &PullRequest{ID: 2}
	pr.ToRef.Repository.Slug = "vegeta"
	pr.ToRef.Repository.Project.Key = "SOUR"

	for _, tc := range []struct {
		name string
		ctx  context.Context
		pr   func() *PullRequest
		err  string
	}{
		{
			name: "timeout",
			pr:   func() *PullRequest { return pr },
			ctx:  timeout,
			err:  "context deadline exceeded",
		},
		{
			name: "repo not set",
			pr:   func() *PullRequest { return &PullRequest{ID: 2} },
			err:  "repository slug empty",
		},
		{
			name: "project not set",
			pr: func() *PullRequest {
				pr := &PullRequest{ID: 2}
				pr.ToRef.Repository.Slug = "vegeta"
				return pr
			},
			err: "project key empty",
		},
		{
			name: "non existing pr",
			pr: func() *PullRequest {
				pr := &PullRequest{ID: 9999}
				pr.ToRef.Repository.Slug = "vegeta"
				pr.ToRef.Repository.Project.Key = "SOUR"
				return pr
			},
			err: "Bitbucket API HTTP error: code=404 url=\"${INSTANCEURL}/rest/api/1.0/projects/SOUR/repos/vegeta/pull-requests/9999\" body=\"{\\\"errors\\\":[{\\\"context\\\":null,\\\"message\\\":\\\"Pull request 9999 does not exist in SOUR/vegeta.\\\",\\\"exceptionName\\\":\\\"com.atlassian.bitbucket.pull.NoSuchPullRequestException\\\"}]}\"",
		},
		{
			name: "non existing repo",
			pr: func() *PullRequest {
				pr := &PullRequest{ID: 9999}
				pr.ToRef.Repository.Slug = "invalidslug"
				pr.ToRef.Repository.Project.Key = "SOUR"
				return pr
			},
			err: "Bitbucket API HTTP error: code=404 url=\"${INSTANCEURL}/rest/api/1.0/projects/SOUR/repos/invalidslug/pull-requests/9999\" body=\"{\\\"errors\\\":[{\\\"context\\\":null,\\\"message\\\":\\\"Repository SOUR/invalidslug does not exist.\\\",\\\"exceptionName\\\":\\\"com.atlassian.bitbucket.repository.NoSuchRepositoryException\\\"}]}\"",
		},
		{
			name: "success",
			pr:   func() *PullRequest { return pr },
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
			tc.err = strings.ReplaceAll(tc.err, "${INSTANCEURL}", instanceURL)

			pr := tc.pr()
			err := cli.LoadPullRequest(tc.ctx, pr)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Fatalf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil || tc.err != "<nil>" {
				return
			}

			data, err := json.MarshalIndent(pr, " ", " ")
			if err != nil {
				t.Fatal(err)
			}

			path := "testdata/golden/LoadPullRequest-" + strings.Replace(tc.name, " ", "-", -1)
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

func TestClient_CreatePullRequest(t *testing.T) {
	instanceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instanceURL == "" {
		instanceURL = "http://127.0.0.1:7990"
	}

	timeout, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	pr := &PullRequest{}
	pr.Title = "This is a test PR"
	pr.Description = "This is a test PR. Feel free to ignore."
	pr.ToRef.Repository.Slug = "automation-testing"
	pr.ToRef.Repository.Project.Key = "SOUR"
	pr.ToRef.ID = "refs/heads/master"
	pr.FromRef.Repository.Slug = "automation-testing"
	pr.FromRef.Repository.Project.Key = "SOUR"
	pr.FromRef.ID = "refs/heads/test-pr-bbs-1"

	for _, tc := range []struct {
		name string
		ctx  context.Context
		pr   func() *PullRequest
		err  string
	}{
		{
			name: "timeout",
			pr:   func() *PullRequest { return pr },
			ctx:  timeout,
			err:  "context deadline exceeded",
		},
		{
			name: "ToRef repo not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.ToRef.Repository.Slug = ""
				return &pr
			},
			err: "ToRef repository slug empty",
		},
		{
			name: "ToRef project not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.ToRef.Repository.Project.Key = ""
				return &pr
			},
			err: "ToRef project key empty",
		},
		{
			name: "ToRef ID not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.ToRef.ID = ""
				return &pr
			},
			err: "ToRef id empty",
		},
		{
			name: "FromRef repo not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.FromRef.Repository.Slug = ""
				return &pr
			},
			err: "FromRef repository slug empty",
		},
		{
			name: "FromRef project not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.FromRef.Repository.Project.Key = ""
				return &pr
			},
			err: "FromRef project key empty",
		},
		{
			name: "FromRef ID not set",
			pr: func() *PullRequest {
				pr := *pr
				pr.FromRef.ID = ""
				return &pr
			},
			err: "FromRef id empty",
		},
		{
			name: "success",
			pr: func() *PullRequest {
				pr := *pr
				pr.FromRef.ID = "refs/heads/test-pr-bbs-3"
				return &pr
			},
		},
		{
			name: "pull request already exists",
			pr: func() *PullRequest {
				pr := *pr
				pr.FromRef.ID = "refs/heads/always-open-pr-bbs"
				return &pr
			},
			err: ErrAlreadyExists.Error(),
		},
		{
			name: "description includes GFM tasklist items",
			pr: func() *PullRequest {
				pr := *pr
				pr.FromRef.ID = "refs/heads/test-pr-bbs-17"
				pr.Description = "- [ ] One\n- [ ] Two\n"
				return &pr
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			name := "CreatePullRequest-" + strings.Replace(tc.name, " ", "-", -1)

			cli, save := NewTestClient(t, name, *update)
			defer save()

			if tc.ctx == nil {
				tc.ctx = context.Background()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}
			tc.err = strings.ReplaceAll(tc.err, "${INSTANCEURL}", instanceURL)

			pr := tc.pr()
			err := cli.CreatePullRequest(tc.ctx, pr)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Fatalf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil || tc.err != "<nil>" {
				return
			}

			data, err := json.MarshalIndent(pr, " ", " ")
			if err != nil {
				t.Fatal(err)
			}

			path := "testdata/golden/" + name
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

func TestClient_LoadPullRequestActivities(t *testing.T) {
	instanceURL := os.Getenv("BITBUCKET_SERVER_URL")
	if instanceURL == "" {
		instanceURL = "http://127.0.0.1:7990"
	}

	cli, save := NewTestClient(t, "PullRequestActivities", *update)
	defer save()

	timeout, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	pr := &PullRequest{ID: 2}
	pr.ToRef.Repository.Slug = "vegeta"
	pr.ToRef.Repository.Project.Key = "SOUR"

	for _, tc := range []struct {
		name string
		ctx  context.Context
		pr   func() *PullRequest
		err  string
	}{
		{
			name: "timeout",
			pr:   func() *PullRequest { return pr },
			ctx:  timeout,
			err:  "context deadline exceeded",
		},
		{
			name: "repo not set",
			pr:   func() *PullRequest { return &PullRequest{ID: 2} },
			err:  "repository slug empty",
		},
		{
			name: "project not set",
			pr: func() *PullRequest {
				pr := &PullRequest{ID: 2}
				pr.ToRef.Repository.Slug = "vegeta"
				return pr
			},
			err: "project key empty",
		},
		{
			name: "success",
			pr:   func() *PullRequest { return pr },
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
			tc.err = strings.ReplaceAll(tc.err, "${INSTANCEURL}", instanceURL)

			pr := tc.pr()
			err := cli.LoadPullRequestActivities(tc.ctx, pr)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Fatalf("error:\nhave: %q\nwant: %q", have, want)
			}

			if err != nil || tc.err != "<nil>" {
				return
			}

			data, err := json.MarshalIndent(pr, " ", " ")
			if err != nil {
				t.Fatal(err)
			}

			path := "testdata/golden/LoadPullRequestActivities-" + strings.Replace(tc.name, " ", "-", -1)
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
