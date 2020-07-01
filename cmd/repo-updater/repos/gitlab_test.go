package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/testutil"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestProjectQueryToURL(t *testing.T) {
	tests := []struct {
		projectQuery string
		perPage      int
		expURL       string
		expErr       error
	}{{
		projectQuery: "?membership=true",
		perPage:      100,
		expURL:       "projects?membership=true&per_page=100",
	}, {
		projectQuery: "projects?membership=true",
		perPage:      100,
		expURL:       "projects?membership=true&per_page=100",
	}, {
		projectQuery: "groups/groupID/projects",
		perPage:      100,
		expURL:       "groups/groupID/projects?per_page=100",
	}, {
		projectQuery: "groups/groupID/projects?foo=bar",
		perPage:      100,
		expURL:       "groups/groupID/projects?foo=bar&per_page=100",
	}, {
		projectQuery: "",
		perPage:      100,
		expURL:       "projects?per_page=100",
	}, {
		projectQuery: "https://somethingelse.com/foo/bar",
		perPage:      100,
		expErr:       schemeOrHostNotEmptyErr,
	}}

	for _, test := range tests {
		t.Logf("Test case %+v", test)
		url, err := projectQueryToURL(test.projectQuery, test.perPage)
		if url != test.expURL {
			t.Errorf("expected %v, got %v", test.expURL, url)
		}
		if !reflect.DeepEqual(test.expErr, err) {
			t.Errorf("expected err %v, got %v", test.expErr, err)
		}
	}
}

func TestGitLabSource_GetRepo(t *testing.T) {
	testCases := []struct {
		name                 string
		projectWithNamespace string
		assert               func(*testing.T, *Repo)
		err                  string
	}{
		{
			name:                 "not found",
			projectWithNamespace: "foobarfoobarfoobar/please-let-this-not-exist",
			err:                  `unexpected response from GitLab API (https://gitlab.com/api/v4/projects/foobarfoobarfoobar%2Fplease-let-this-not-exist): HTTP error status 404`,
		},
		{
			name:                 "found",
			projectWithNamespace: "gitlab-org/gitaly",
			assert: func(t *testing.T, have *Repo) {
				t.Helper()

				want := &Repo{
					Name:        "gitlab.com/gitlab-org/gitaly",
					Description: "Gitaly is a Git RPC service for handling all the git calls made by GitLab",
					URI:         "gitlab.com/gitlab-org/gitaly",
					ExternalRepo: api.ExternalRepoSpec{
						ID:          "2009901",
						ServiceType: "gitlab",
						ServiceID:   "https://gitlab.com/",
					},
					Sources: map[string]*SourceInfo{
						"extsvc:gitlab:0": {
							ID:       "extsvc:gitlab:0",
							CloneURL: "https://gitlab.com/gitlab-org/gitaly.git",
						},
					},
					Metadata: &gitlab.Project{
						ProjectCommon: gitlab.ProjectCommon{
							ID:                2009901,
							PathWithNamespace: "gitlab-org/gitaly",
							Description:       "Gitaly is a Git RPC service for handling all the git calls made by GitLab",
							WebURL:            "https://gitlab.com/gitlab-org/gitaly",
							HTTPURLToRepo:     "https://gitlab.com/gitlab-org/gitaly.git",
							SSHURLToRepo:      "git@gitlab.com:gitlab-org/gitaly.git",
						},
						Visibility: "",
						Archived:   false,
					},
				}

				if !reflect.DeepEqual(have, want) {
					t.Errorf("response: %s", cmp.Diff(have, want))
				}
			},
			err: "<nil>",
		},
	}

	for _, tc := range testCases {
		tc := tc
		tc.name = "GITLAB-DOT-COM/" + tc.name

		t.Run(tc.name, func(t *testing.T) {
			// The GitLabSource uses the gitlab.Client under the hood, which
			// uses rcache, a caching layer that uses Redis.
			// We need to clear the cache before we run the tests
			rcache.SetupForTest(t)

			cf, save := newClientFactory(t, tc.name)
			defer save(t)

			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			svc := &ExternalService{
				Kind: extsvc.KindGitLab,
				Config: marshalJSON(t, &schema.GitLabConnection{
					Url: "https://gitlab.com",
				}),
			}

			gitlabSrc, err := NewGitLabSource(svc, cf)
			if err != nil {
				t.Fatal(err)
			}

			repo, err := gitlabSrc.GetRepo(context.Background(), tc.projectWithNamespace)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Errorf("error:\nhave: %q\nwant: %q", have, want)
			}

			if tc.assert != nil {
				tc.assert(t, repo)
			}
		})
	}
}

func TestGitLabSource_makeRepo(t *testing.T) {
	b, err := ioutil.ReadFile(filepath.Join("testdata", "gitlab-repos.json"))
	if err != nil {
		t.Fatal(err)
	}
	var repos []*gitlab.Project
	if err := json.Unmarshal(b, &repos); err != nil {
		t.Fatal(err)
	}

	svc := ExternalService{ID: 1, Kind: extsvc.KindGitLab}

	tests := []struct {
		name   string
		schmea *schema.GitLabConnection
	}{
		{
			name: "simple",
			schmea: &schema.GitLabConnection{
				Url: "https://gitlab.com",
			},
		}, {
			name: "ssh",
			schmea: &schema.GitLabConnection{
				Url:        "https://gitlab.com",
				GitURLType: "ssh",
			},
		}, {
			name: "path-pattern",
			schmea: &schema.GitLabConnection{
				Url:                   "https://gitlab.com",
				RepositoryPathPattern: "gl/{pathWithNamespace}",
			},
		},
	}
	for _, test := range tests {
		test.name = "GitLabSource_makeRepo_" + test.name
		t.Run(test.name, func(t *testing.T) {
			lg := log15.New()
			lg.SetHandler(log15.DiscardHandler())

			s, err := newGitLabSource(&svc, test.schmea, nil)
			if err != nil {
				t.Fatal(err)
			}

			var got []*Repo
			for _, r := range repos {
				got = append(got, s.makeRepo(r))
			}

			testutil.AssertGolden(t, "testdata/golden/"+test.name, update(test.name), got)
		})
	}
}

// TestGitLabSource_ChangesetSource tests the various Changeset functions that
// implement the ChangesetSource interface.
func TestGitLabSource_ChangesetSource(t *testing.T) {
	// Set up some common values for use in subtests.
	c := &Changeset{
		Changeset: &campaigns.Changeset{},
		Repo:      &Repo{Metadata: &gitlab.Project{}},
		HeadRef:   "refs/heads/head",
		BaseRef:   "refs/heads/base",
		Title:     "title",
		Body:      "description",
	}
	ctx := context.Background()
	mr := &gitlab.MergeRequest{
		ID:           1,
		IID:          2,
		ProjectID:    3,
		Title:        "title",
		Description:  "description",
		SourceBranch: "head",
		TargetBranch: "base",
	}
	s := &GitLabSource{
		client: gitlab.NewClientProvider(&url.URL{}, &panicDoer{}).GetClient(),
	}

	testCommonParams := func(client *gitlab.Client, context context.Context, project *gitlab.Project) {
		if client != s.client {
			t.Errorf("unexpected GitLabSource client: have %+v; want %+v", client, s.client)
		}
		if context != ctx {
			t.Errorf("unexpected context: have %+v; want %+v", context, ctx)
		}
		if project != c.Repo.Metadata.(*gitlab.Project) {
			t.Errorf("unexpected Project: have %+v; want %+v", project, c.Repo.Metadata)
		}
	}

	// Helpers for creating mocks. Note that unmock() must be called to remove
	// the mocks (usually in a defer).
	mockCreateMergeRequest := func(mr *gitlab.MergeRequest, err error) *int {
		invocations := 0

		gitlab.MockCreateMergeRequest = func(client *gitlab.Client, ctx context.Context, project *gitlab.Project, opts gitlab.CreateMergeRequestOpts) (*gitlab.MergeRequest, error) {
			invocations++
			testCommonParams(client, ctx, project)

			if want := "head"; opts.SourceBranch != want {
				t.Errorf("unexpected SourceBranch: have %s; want %s", opts.SourceBranch, want)
			}
			if want := "base"; opts.TargetBranch != want {
				t.Errorf("unexpected TargetBranch: have %s; want %s", opts.TargetBranch, want)
			}

			return mr, err
		}

		return &invocations
	}

	mockGetMergeRequest := func(expected gitlab.ID, mr *gitlab.MergeRequest, err error) *int {
		invocations := 0

		gitlab.MockGetMergeRequest = func(client *gitlab.Client, ctx context.Context, project *gitlab.Project, iid gitlab.ID) (*gitlab.MergeRequest, error) {
			invocations++
			testCommonParams(client, ctx, project)
			if expected != iid {
				t.Errorf("invalid IID: have %d; want %d", iid, expected)
			}
			return mr, err
		}

		return &invocations
	}

	mockGetOpenMergeRequestByRefs := func(mr *gitlab.MergeRequest, err error) *int {
		invocations := 0

		gitlab.MockGetOpenMergeRequestByRefs = func(client *gitlab.Client, ctx context.Context, project *gitlab.Project, source, target string) (*gitlab.MergeRequest, error) {
			testCommonParams(client, ctx, project)
			return mr, err
		}

		return &invocations
	}

	mockUpdateMergeRequest := func(updated *gitlab.MergeRequest, err error) *int {
		invocations := 0

		gitlab.MockUpdateMergeRequest = func(client *gitlab.Client, ctx context.Context, project *gitlab.Project, mrIn *gitlab.MergeRequest, opts gitlab.UpdateMergeRequestOpts) (*gitlab.MergeRequest, error) {
			invocations++
			testCommonParams(client, ctx, project)
			if mr != mrIn {
				t.Errorf("unexpected MergeRequest: have %+v; want %+v", mrIn, mr)
			}
			return updated, err
		}

		return &invocations
	}

	unmock := func() {
		gitlab.MockCreateMergeRequest = nil
		gitlab.MockGetMergeRequest = nil
		gitlab.MockGetOpenMergeRequestByRefs = nil
		gitlab.MockUpdateMergeRequest = nil
	}

	t.Run("CreateChangeset", func(t *testing.T) {
		t.Run("invalid metadata", func(t *testing.T) {
			defer func() { _ = recover() }()

			_, _ = s.CreateChangeset(ctx, &Changeset{
				Repo: &Repo{
					Metadata: struct{}{},
				},
			})
			t.Error("invalid metadata did not panic")
		})

		t.Run("error from CreateMergeRequest", func(t *testing.T) {
			inner := errors.New("foo")
			mockCreateMergeRequest(nil, inner)
			defer unmock()

			exists, have := s.CreateChangeset(ctx, c)
			if exists {
				t.Errorf("unexpected exists value: %v", exists)
			}
			if !errors.Is(have, inner) {
				t.Errorf("error does not include inner error: have %+v; want %+v", have, inner)
			}
		})

		t.Run("error from GetOpenMergeRequestByRefs", func(t *testing.T) {
			inner := errors.New("foo")
			mockCreateMergeRequest(nil, gitlab.ErrMergeRequestAlreadyExists)
			mockGetOpenMergeRequestByRefs(nil, inner)
			defer unmock()

			exists, have := s.CreateChangeset(ctx, c)
			if !exists {
				t.Errorf("unexpected exists value: %v", exists)
			}
			if !errors.Is(have, inner) {
				t.Errorf("error does not include inner error: have %+v; want %+v", have, inner)
			}
		})

		t.Run("merge request already exists", func(t *testing.T) {
			mockCreateMergeRequest(nil, gitlab.ErrMergeRequestAlreadyExists)
			mockGetOpenMergeRequestByRefs(mr, nil)
			defer unmock()

			exists, err := s.CreateChangeset(ctx, c)
			if !exists {
				t.Errorf("unexpected exists value: %v", exists)
			}
			if err != nil {
				t.Errorf("unexpected non-nil err: %+v", err)
			}

			if c.Changeset.Metadata != mr {
				t.Errorf("unexpected metadata: have %+v; want %+v", c.Changeset.Metadata, mr)
			}
		})

		t.Run("merge request is new", func(t *testing.T) {
			mockCreateMergeRequest(mr, nil)
			defer unmock()

			exists, err := s.CreateChangeset(ctx, c)
			if exists {
				t.Errorf("unexpected exists value: %v", exists)
			}
			if err != nil {
				t.Errorf("unexpected non-nil err: %+v", err)
			}

			if c.Changeset.Metadata != mr {
				t.Errorf("unexpected metadata: have %+v; want %+v", c.Changeset.Metadata, mr)
			}
		})
	})

	t.Run("CloseChangeset", func(t *testing.T) {
		t.Run("invalid metadata", func(t *testing.T) {
			defer func() { _ = recover() }()

			_ = s.CloseChangeset(ctx, &Changeset{
				Repo: &Repo{
					Metadata: struct{}{},
				},
			})
			t.Error("invalid metadata did not panic")
		})

		t.Run("error from UpdateMergeRequest", func(t *testing.T) {
			inner := errors.New("foo")
			mockUpdateMergeRequest(nil, inner)
			defer unmock()

			have := s.CloseChangeset(ctx, c)
			if !errors.Is(have, inner) {
				t.Errorf("error does not include inner error: have %+v; want %+v", have, inner)
			}
		})

		t.Run("success", func(t *testing.T) {
			want := &gitlab.MergeRequest{}
			mockUpdateMergeRequest(want, nil)
			defer unmock()

			if err := s.CloseChangeset(ctx, c); err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
		})
	})

	t.Run("LoadChangesets", func(t *testing.T) {
		cloneChangeset := func(c *Changeset) *Changeset {
			cc := *c
			inner := *c.Changeset
			cc.Changeset = &inner
			return &cc
		}

		a := cloneChangeset(c)
		a.Changeset.ExternalID = "42"
		b := cloneChangeset(c)
		b.Changeset.ExternalID = "42"
		cs := []*Changeset{a, b}

		t.Run("invalid metadata", func(t *testing.T) {
			defer func() { _ = recover() }()

			_ = s.LoadChangesets(ctx, []*Changeset{{
				Repo: &Repo{Metadata: struct{}{}},
			}}...)
			t.Error("invalid metadata did not panic")
		})

		t.Run("error from ParseInt", func(t *testing.T) {
			if err := s.LoadChangesets(ctx, []*Changeset{{
				Changeset: &campaigns.Changeset{ExternalID: "foo"},
				Repo:      &Repo{Metadata: &gitlab.Project{}},
			}}...); err == nil {
				t.Error("invalid ExternalID did not result in an error")
			}
		})

		t.Run("error from GetMergeRequest", func(t *testing.T) {
			inner := errors.New("foo")
			mockGetMergeRequest(42, nil, inner)
			defer unmock()

			if have := s.LoadChangesets(ctx, cs...); !errors.Is(have, inner) {
				t.Errorf("error does not include inner error: have %+v; want %+v", have, inner)
			}
		})

		t.Run("success", func(t *testing.T) {
			inv := mockGetMergeRequest(42, &gitlab.MergeRequest{}, nil)
			defer unmock()

			if err := s.LoadChangesets(ctx, cs...); err != nil {
				t.Errorf("unexpected error: %+v", err)
			}
			if want := 2; *inv != want {
				t.Errorf("unexpected number of GetMergeRequest invocations: have %d; want %d", *inv, want)
			}
		})
	})
}

type panicDoer struct{}

func (d *panicDoer) Do(r *http.Request) (*http.Response, error) {
	panic("this function should not be called; a mock must be missing")
}
