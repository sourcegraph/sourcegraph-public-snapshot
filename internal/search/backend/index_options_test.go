package backend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/zoekt"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGetIndexOptions(t *testing.T) {
	const (
		REPO = int32(iota + 1)
		FOO
		NOT_IN_VERSION_CONTEXT
		PRIORITY
		PUBLIC
		FORK
		ARCHIVED
	)

	name := func(repo int32) string {
		return fmt.Sprintf("repo-%.2d", repo)
	}

	withBranches := func(c schema.SiteConfiguration, repo int32, branches ...string) schema.SiteConfiguration {
		if c.ExperimentalFeatures == nil {
			c.ExperimentalFeatures = &schema.ExperimentalFeatures{}
		}
		if c.ExperimentalFeatures.SearchIndexBranches == nil {
			c.ExperimentalFeatures.SearchIndexBranches = map[string][]string{}
		}
		b := c.ExperimentalFeatures.SearchIndexBranches
		b[name(repo)] = append(b[name(repo)], branches...)
		return c
	}

	type caseT struct {
		name              string
		conf              schema.SiteConfiguration
		searchContextRevs []string
		repo              int32
		want              zoektIndexOptions
	}

	cases := []caseT{{
		name: "default",
		conf: schema.SiteConfiguration{},
		repo: REPO,
		want: zoektIndexOptions{
			RepoID:  1,
			Name:    "repo-01",
			Symbols: true,
			Branches: []zoekt.RepositoryBranch{
				{Name: "HEAD", Version: "!HEAD"},
			},
		},
	}, {
		name: "public",
		conf: schema.SiteConfiguration{},
		repo: PUBLIC,
		want: zoektIndexOptions{
			RepoID:  5,
			Name:    "repo-05",
			Public:  true,
			Symbols: true,
			Branches: []zoekt.RepositoryBranch{
				{Name: "HEAD", Version: "!HEAD"},
			},
		},
	}, {
		name: "fork",
		conf: schema.SiteConfiguration{},
		repo: FORK,
		want: zoektIndexOptions{
			RepoID:  6,
			Name:    "repo-06",
			Fork:    true,
			Symbols: true,
			Branches: []zoekt.RepositoryBranch{
				{Name: "HEAD", Version: "!HEAD"},
			},
		},
	}, {
		name: "archived",
		conf: schema.SiteConfiguration{},
		repo: ARCHIVED,
		want: zoektIndexOptions{
			RepoID:   7,
			Name:     "repo-07",
			Archived: true,
			Symbols:  true,
			Branches: []zoekt.RepositoryBranch{
				{Name: "HEAD", Version: "!HEAD"},
			},
		},
	}, {
		name: "nosymbols",
		conf: schema.SiteConfiguration{
			SearchIndexSymbolsEnabled: boolPtr(false),
		},
		repo: REPO,
		want: zoektIndexOptions{
			RepoID: 1,
			Name:   "repo-01",
			Branches: []zoekt.RepositoryBranch{
				{Name: "HEAD", Version: "!HEAD"},
			},
		},
	}, {
		name: "largefiles",
		conf: schema.SiteConfiguration{
			SearchLargeFiles: []string{"**/*.jar", "*.bin"},
		},
		repo: REPO,
		want: zoektIndexOptions{
			RepoID:     1,
			Name:       "repo-01",
			Symbols:    true,
			LargeFiles: []string{"**/*.jar", "*.bin"},
			Branches: []zoekt.RepositoryBranch{
				{Name: "HEAD", Version: "!HEAD"},
			},
		},
	}, {
		name: "conf index branches",
		conf: withBranches(schema.SiteConfiguration{}, REPO, "a", "", "b"),
		repo: REPO,
		want: zoektIndexOptions{
			RepoID:  1,
			Name:    "repo-01",
			Symbols: true,
			Branches: []zoekt.RepositoryBranch{
				{Name: "HEAD", Version: "!HEAD"},
				{Name: "a", Version: "!a"},
				{Name: "b", Version: "!b"},
			},
		},
	}, {
		name: "conf index revisions",
		conf: schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{
			SearchIndexRevisions: []*schema.SearchIndexRevisionsRule{
				{Name: "repo-.*", Revisions: []string{"a"}},
			},
		}},
		repo: REPO,
		want: zoektIndexOptions{
			RepoID:  1,
			Name:    "repo-01",
			Symbols: true,
			Branches: []zoekt.RepositoryBranch{
				{Name: "HEAD", Version: "!HEAD"},
				{Name: "a", Version: "!a"},
			},
		},
	}, {
		name: "conf index revisions and branches",
		conf: schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{
			SearchIndexBranches: map[string][]string{
				"repo-01": {"a", "b"},
			},
			SearchIndexRevisions: []*schema.SearchIndexRevisionsRule{
				{Name: "repo-.*", Revisions: []string{"a", "c"}},
			},
		}},
		repo: REPO,
		want: zoektIndexOptions{
			RepoID:  1,
			Name:    "repo-01",
			Symbols: true,
			Branches: []zoekt.RepositoryBranch{
				{Name: "HEAD", Version: "!HEAD"},
				{Name: "a", Version: "!a"},
				{Name: "b", Version: "!b"},
				{Name: "c", Version: "!c"},
			},
		},
	}, {
		name:              "with search context revisions",
		conf:              schema.SiteConfiguration{},
		repo:              REPO,
		searchContextRevs: []string{"rev1", "rev2"},
		want: zoektIndexOptions{
			RepoID:  1,
			Name:    "repo-01",
			Symbols: true,
			Branches: []zoekt.RepositoryBranch{
				{Name: "HEAD", Version: "!HEAD"},
				{Name: "rev1", Version: "!rev1"},
				{Name: "rev2", Version: "!rev2"},
			},
		},
	}, {
		name: "with a priority value",
		conf: schema.SiteConfiguration{},
		repo: PRIORITY,
		want: zoektIndexOptions{
			RepoID:  4,
			Name:    "repo-04",
			Symbols: true,
			Branches: []zoekt.RepositoryBranch{
				{Name: "HEAD", Version: "!HEAD"},
			},
			Priority: 10,
		},
	}}

	{
		// Generate case for no more than than 64 branches
		var branches []string
		for i := 0; i < 100; i++ {
			branches = append(branches, fmt.Sprintf("%.2d", i))
		}
		want := []zoekt.RepositoryBranch{{Name: "HEAD", Version: "!HEAD"}}
		for i := 0; i < 63; i++ {
			want = append(want, zoekt.RepositoryBranch{
				Name:    fmt.Sprintf("%.2d", i),
				Version: fmt.Sprintf("!%.2d", i),
			})
		}
		cases = append(cases, caseT{
			name: "limit branches",
			conf: withBranches(schema.SiteConfiguration{}, REPO, branches...),
			repo: REPO,
			want: zoektIndexOptions{
				RepoID:   1,
				Name:     "repo-01",
				Symbols:  true,
				Branches: want,
			},
		})
	}

	var getRepoIndexOptions getRepoIndexOptsFn = func(repo int32) (*RepoIndexOptions, error) {
		var priority float64
		if repo == PRIORITY {
			priority = 10
		}
		return &RepoIndexOptions{
			RepoID:   repo,
			Name:     name(repo),
			Public:   repo == PUBLIC,
			Fork:     repo == FORK,
			Archived: repo == ARCHIVED,
			Priority: priority,
			GetVersion: func(branch string) (string, error) {
				return "!" + branch, nil
			},
		}, nil
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			getSearchContextRevisions := func(int32) ([]string, error) { return tc.searchContextRevs, nil }

			b := GetIndexOptions(&tc.conf, getRepoIndexOptions, getSearchContextRevisions, tc.repo)

			var got zoektIndexOptions
			if err := json.Unmarshal(b, &got); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatal("mismatch (-want, +got):\n", diff)
			}
		})
	}
}

func TestGetIndexOptions_getVersion(t *testing.T) {
	conf := schema.SiteConfiguration{}
	getSearchContextRevs := func(int32) ([]string, error) { return []string{"b1", "b2"}, nil }

	boom := errors.New("boom")
	cases := []struct {
		name    string
		f       func(string) (string, error)
		want    []zoekt.RepositoryBranch
		wantErr string
	}{{
		name: "error",
		f: func(_ string) (string, error) {
			return "", boom
		},
		wantErr: "boom",
	}, {
		// no HEAD means we don't index anything. This leads to zoekt having
		// an empty index.
		name: "no HEAD",
		f: func(branch string) (string, error) {
			if branch == "HEAD" {
				return "", nil
			}
			return "!" + branch, nil
		},
		want: nil,
	}, {
		name: "no branch",
		f: func(branch string) (string, error) {
			if branch == "b1" {
				return "", nil
			}
			return "!" + branch, nil
		},
		want: []zoekt.RepositoryBranch{
			{Name: "HEAD", Version: "!HEAD"},
			{Name: "b2", Version: "!b2"},
		},
	}, {
		name: "all",
		f: func(branch string) (string, error) {
			return "!" + branch, nil
		},
		want: []zoekt.RepositoryBranch{
			{Name: "HEAD", Version: "!HEAD"},
			{Name: "b1", Version: "!b1"},
			{Name: "b2", Version: "!b2"},
		},
	}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			getRepoIndexOptions := func(repo int32) (*RepoIndexOptions, error) {
				return &RepoIndexOptions{
					GetVersion: tc.f,
				}, nil
			}

			b := GetIndexOptions(&conf, getRepoIndexOptions, getSearchContextRevs, 1)

			var got zoektIndexOptions
			if err := json.Unmarshal(b, &got); err != nil {
				t.Fatal(err)
			}

			if got.Error != tc.wantErr {
				t.Fatalf("expected error %v, got body %s and error %v", tc.wantErr, b, got.Error)
			}
			if tc.wantErr != "" {
				return
			}

			if diff := cmp.Diff(tc.want, got.Branches); diff != "" {
				t.Fatal("mismatch (-want, +got):\n", diff)
			}
		})
	}
}

func TestGetIndexOptions_batch(t *testing.T) {
	isError := func(repo int32) bool {
		return repo%20 == 0
	}
	var (
		repos []int32
		want  []zoektIndexOptions
	)
	for repo := int32(1); repo < 100; repo++ {
		repos = append(repos, repo)
		if isError(repo) {
			want = append(want, zoektIndexOptions{Error: "error"})
		} else {
			want = append(want, zoektIndexOptions{
				Symbols: true,
				Branches: []zoekt.RepositoryBranch{
					{Name: "HEAD", Version: fmt.Sprintf("!HEAD-%d", repo)},
				},
			})
		}
	}
	getRepoIndexOptions := func(repo int32) (*RepoIndexOptions, error) {
		return &RepoIndexOptions{
			GetVersion: func(branch string) (string, error) {
				if isError(repo) {
					return "", errors.New("error")
				}
				return fmt.Sprintf("!%s-%d", branch, repo), nil
			},
		}, nil
	}

	getSearchContextRevs := func(int32) ([]string, error) { return nil, nil }

	b := GetIndexOptions(&schema.SiteConfiguration{}, getRepoIndexOptions, getSearchContextRevs, repos...)
	dec := json.NewDecoder(bytes.NewReader(b))
	got := make([]zoektIndexOptions, len(repos))
	for i := range repos {
		if err := dec.Decode(&got[i]); err != nil {
			t.Fatal(err)
		}
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatal("mismatch (-want, +got):\n", diff)
	}
}

func boolPtr(b bool) *bool {
	return &b
}
