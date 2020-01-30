package bitbucketcloud

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

func TestClient_Repos(t *testing.T) {
	cli, save := NewTestClient(t, "Repos", *update, &url.URL{Scheme: "https", Host: "api.bitbucket.org"})
	defer save()

	timeout, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	repos := map[string]*Repo{
		"mux": {
			Slug:      "mux",
			Name:      "mux",
			FullName:  "sglocal/mux",
			UUID:      "{e1e75436-05e6-4c38-8543-9c36ec26fad1}",
			SCM:       "git",
			IsPrivate: true,
			Links: Links{
				Clone: CloneLinks{
					{"https://Unknwon@bitbucket.org/sglocal/mux.git", "https"},
					{"git@bitbucket.org:sglocal/mux.git", "ssh"},
				},
				HTML: Link{"https://bitbucket.org/sglocal/mux"},
			},
		},
		"python-langserver": {
			Slug:      "python-langserver",
			Name:      "python-langserver",
			FullName:  "sglocal/python-langserver",
			UUID:      "{421b93e9-1f00-4054-8156-4d821d4a768b}",
			SCM:       "git",
			IsPrivate: false,
			Links: Links{
				Clone: CloneLinks{
					{"https://Unknwon@bitbucket.org/sglocal/python-langserver.git", "https"},
					{"git@bitbucket.org:sglocal/python-langserver.git", "ssh"},
				},
				HTML: Link{"https://bitbucket.org/sglocal/python-langserver"},
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
			account: "sglocal",
			repos:   []*Repo{repos["mux"]},
			next: &PageToken{
				Size:    3,
				Page:    1,
				Pagelen: 1,
				Next:    "https://api.bitbucket.org/2.0/repositories/sglocal?pagelen=1&page=2",
			},
		},
		{
			name: "pagination: last page",
			page: &PageToken{
				Pagelen: 1,
				Next:    "https://api.bitbucket.org/2.0/repositories/sglocal?pagelen=1&page=3",
			},
			repos: []*Repo{repos["python-langserver"]},
			next: &PageToken{
				Size:    3,
				Page:    3,
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
