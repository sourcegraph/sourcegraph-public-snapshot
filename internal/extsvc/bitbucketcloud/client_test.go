package bitbucketcloud

import (
	"context"
	"flag"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

var update = flag.Bool("update", false, "update testdata")

func TestClient_Repos(t *testing.T) {
	cli, save := NewTestClient(t, "Repos", *update)
	defer save()

	timeout, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	repos := map[string]*Repo{
		"src-cli": {
			Slug:      "src-cli",
			Name:      "src-cli",
			FullName:  "sourcegraph-testing/src-cli",
			UUID:      "{b090a669-ac7b-44cd-9610-02d027cb39f3}",
			SCM:       "git",
			IsPrivate: true,
			Links: Links{
				Clone: CloneLinks{
					{"https://sourcegraph-testing@bitbucket.org/sourcegraph-testing/src-cli.git", "https"},
					{"git@bitbucket.org:sourcegraph-testing/src-cli.git", "ssh"},
				},
				HTML: Link{"https://bitbucket.org/sourcegraph-testing/src-cli"},
			},
		},
		"sourcegraph": {
			Slug:      "sourcegraph",
			Name:      "sourcegraph",
			FullName:  "sourcegraph-testing/sourcegraph",
			UUID:      "{f46afc56-15a7-4579-9429-1b9329ad4c09}",
			SCM:       "git",
			IsPrivate: true,
			Links: Links{
				Clone: CloneLinks{
					{"https://sourcegraph-testing@bitbucket.org/sourcegraph-testing/sourcegraph.git", "https"},
					{"git@bitbucket.org:sourcegraph-testing/sourcegraph.git", "ssh"},
				},
				HTML: Link{"https://bitbucket.org/sourcegraph-testing/sourcegraph"},
			},
		},
	}

	for _, tc := range []struct {
		name    string
		ctx     context.Context
		page    *PageToken
		account string
		repos   []*Repo
		next    *PageToken
		err     string
	}{
		{
			name: "timeout",
			ctx:  timeout,
			err:  "context deadline exceeded",
		},
		{
			name:    "pagination: first page",
			page:    &PageToken{Pagelen: 1},
			account: "sourcegraph-testing",
			repos:   []*Repo{repos["src-cli"]},
			next: &PageToken{
				Size:    2,
				Page:    1,
				Pagelen: 1,
				Next:    "https://api.bitbucket.org/2.0/repositories/sourcegraph-testing?pagelen=1&page=2",
			},
		},
		{
			name: "pagination: last page",
			page: &PageToken{
				Pagelen: 1,
				Next:    "https://api.bitbucket.org/2.0/repositories/sourcegraph-testing?pagelen=1&page=2",
			},
			account: "sourcegraph-testing",
			repos:   []*Repo{repos["sourcegraph"]},
			next: &PageToken{
				Size:    2,
				Page:    2,
				Pagelen: 1,
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

			repos, next, err := cli.Repos(tc.ctx, tc.page, tc.account)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if have, want := next, tc.next; !reflect.DeepEqual(have, want) {
				t.Error(cmp.Diff(have, want))
			}

			if have, want := repos, tc.repos; !reflect.DeepEqual(have, want) {
				t.Error(cmp.Diff(have, want))
			}
		})
	}
}
