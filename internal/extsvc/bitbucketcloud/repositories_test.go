package bitbucketcloud

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestClient_Repo(t *testing.T) {
	// WHEN UPDATING: ensure the token in use can read
	// https://bitbucket.org/sourcegraph-testing/sourcegraph/.

	ctx := context.Background()
	c := newTestClient(t)

	t.Run("valid repo", func(t *testing.T) {
		repo, err := c.Repo(ctx, "sourcegraph-testing", "sourcegraph")
		assert.NotNil(t, repo)
		assert.Nil(t, err)
		assertGolden(t, repo)
	})

	t.Run("invalid repo", func(t *testing.T) {
		repo, err := c.Repo(ctx, "sourcegraph-testing", "does-not-exist")
		assert.Nil(t, repo)
		assert.NotNil(t, err)
		assert.True(t, errcode.IsNotFound(err))
	})
}

func TestClient_Repos(t *testing.T) {
	// WHEN UPDATING: ensure the token in use can read
	// https://bitbucket.org/sourcegraph-testing/sourcegraph/ and
	// https://bitbucket.org/sourcegraph-testing/src-cli/.
	cli := newTestClient(t)

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
			Links: RepoLinks{
				Clone: CloneLinks{
					{Href: "https://sourcegraph-testing@bitbucket.org/sourcegraph-testing/src-cli.git", Name: "https"},
					{Href: "git@bitbucket.org:sourcegraph-testing/src-cli.git", Name: "ssh"},
				},
				HTML: Link{Href: "https://bitbucket.org/sourcegraph-testing/src-cli"},
			},
			ForkPolicy: ForkPolicyNoPublic,
			Owner: &Account{
				Links: Links{
					"avatar": Link{Href: "https://secure.gravatar.com/avatar/f964dc31564db8243e952bdaeabbe884?d=https%3A%2F%2Favatar-management--avatars.us-west-2.prod.public.atl-paas.net%2Finitials%2FST-2.png"},
					"html":   Link{Href: "https://bitbucket.org/%7B4b85b785-1433-4092-8512-20302f4a03be%7D/"},
					"self":   Link{Href: "https://api.bitbucket.org/2.0/users/%7B4b85b785-1433-4092-8512-20302f4a03be%7D"},
				},
				Nickname:    "Sourcegraph Testing",
				DisplayName: "Sourcegraph Testing",
				UUID:        "{4b85b785-1433-4092-8512-20302f4a03be}",
			},
		},
		"sourcegraph": {
			Slug:      "sourcegraph",
			Name:      "sourcegraph",
			FullName:  "sourcegraph-testing/sourcegraph",
			UUID:      "{f46afc56-15a7-4579-9429-1b9329ad4c09}",
			SCM:       "git",
			IsPrivate: true,
			Links: RepoLinks{
				Clone: CloneLinks{
					{Href: "https://sourcegraph-testing@bitbucket.org/sourcegraph-testing/sourcegraph.git", Name: "https"},
					{Href: "git@bitbucket.org:sourcegraph-testing/sourcegraph.git", Name: "ssh"},
				},
				HTML: Link{Href: "https://bitbucket.org/sourcegraph-testing/sourcegraph"},
			},
			ForkPolicy: ForkPolicyNoPublic,
			Owner: &Account{
				Links: Links{
					"avatar": Link{Href: "https://secure.gravatar.com/avatar/f964dc31564db8243e952bdaeabbe884?d=https%3A%2F%2Favatar-management--avatars.us-west-2.prod.public.atl-paas.net%2Finitials%2FST-2.png"},
					"html":   Link{Href: "https://bitbucket.org/%7B4b85b785-1433-4092-8512-20302f4a03be%7D/"},
					"self":   Link{Href: "https://api.bitbucket.org/2.0/users/%7B4b85b785-1433-4092-8512-20302f4a03be%7D"},
				},
				Nickname:    "Sourcegraph Testing",
				DisplayName: "Sourcegraph Testing",
				UUID:        "{4b85b785-1433-4092-8512-20302f4a03be}",
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

			repos, next, err := cli.Repos(tc.ctx, tc.page, tc.account, nil)
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

func TestClient_ForkRepository(t *testing.T) {
	// WHEN UPDATING: set the repository name below to an unused repository
	// within the sourcegraph-testing account. (This probably just means you
	// need to increment the number.) This will be used as the target for a fork
	// of https://bitbucket.org/sourcegraph-testing/src-cli/.

	repo := "src-cli-fork-00"

	ctx := context.Background()
	c := newTestClient(t)

	// Get the current user for use in the actual fork calls (as a workspace).
	user, err := c.CurrentUser(ctx)
	assert.Nil(t, err)
	workspace := ForkInputWorkspace(user.Username)

	// Get the upstream repo.
	upstream, err := c.Repo(ctx, "sourcegraph-testing", "src-cli")
	assert.Nil(t, err)

	t.Run("success", func(t *testing.T) {
		fork, err := c.ForkRepository(ctx, upstream, ForkInput{
			Name:      &repo,
			Workspace: workspace,
		})
		assert.Nil(t, err)
		assert.NotNil(t, fork)
		assert.Equal(t, repo, fork.Slug)
		assert.Equal(t, user.Username+"/"+repo, fork.FullName)
		assert.Equal(t, fork.Parent.FullName, upstream.FullName)
		assertGolden(t, fork)
	})

	t.Run("failure", func(t *testing.T) {
		// This looks a bit weird, but it's basically a patch around the fact
		// that we need to test the case where a name isn't given, but we don't
		// have a reliable upstream that we can fork to test that. So we'll make
		// sure that the request is valid, and that we get the error we expect
		// back from Bitbucket.
		fork, err := c.ForkRepository(ctx, upstream, ForkInput{Workspace: workspace})
		assert.Nil(t, fork)
		assert.NotNil(t, err)

		he := &httpError{}
		if ok := errors.As(err, &he); !ok {
			t.Fatal("could not extract httpError from error")
		}
		assert.Contains(t, he.Body, "Repository with this Slug and Owner already exists.")
	})
}

func TestRepo_Namespace(t *testing.T) {
	for name, tc := range map[string]struct {
		input   string
		want    string
		wantErr bool
	}{
		"empty string": {
			input:   "",
			want:    "",
			wantErr: true,
		},
		"no slash": {
			input:   "foo",
			want:    "",
			wantErr: true,
		},
		"one slash": {
			input:   "foo/bar",
			want:    "foo",
			wantErr: false,
		},
		"multiple slashes": {
			input:   "foo/bar/quux",
			want:    "foo",
			wantErr: false,
		},
	} {
		t.Run(name, func(t *testing.T) {
			repo := &Repo{FullName: tc.input}
			have, haveErr := repo.Namespace()
			if tc.wantErr {
				assert.Empty(t, have)
				assert.NotNil(t, haveErr)
			} else {
				assert.Nil(t, haveErr)
				assert.Equal(t, tc.want, have)
			}
		})
	}
}
