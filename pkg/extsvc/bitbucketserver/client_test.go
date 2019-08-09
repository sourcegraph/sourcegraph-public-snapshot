package bitbucketserver

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_788(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
