package backend

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/zoekt"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestReposSubset(t *testing.T) {
	var indexed map[string][]types.MinimalRepo
	index := &Indexers{
		Map: prefixMap([]string{"foo", "bar", "baz.fully.qualified:80"}),
		Indexed: func(ctx context.Context, k string) zoekt.ReposMap {
			set := zoekt.ReposMap{}
			if indexed == nil {
				return set
			}
			for _, s := range indexed[k] {
				set[uint32(s.ID)] = zoekt.MinimalRepoListEntry{HasSymbols: true}
			}
			return set
		},
	}

	repos := make(map[string]types.MinimalRepo)
	getRepos := func(names ...string) (rs []types.MinimalRepo) {
		for _, name := range names {
			r, ok := repos[name]
			if !ok {
				r = types.MinimalRepo{
					ID:   api.RepoID(rand.Int31()),
					Name: api.RepoName(name),
				}
				repos[name] = r
			}
			rs = append(rs, r)
		}
		return rs
	}

	cases := []struct {
		name     string
		hostname string
		indexed  map[string][]types.MinimalRepo
		repos    []types.MinimalRepo
		want     []types.MinimalRepo
		errS     string
	}{{
		name:     "bad hostname",
		hostname: "bam",
		errS:     "hostname \"bam\" not found",
	}, {
		name:     "all",
		hostname: "foo",
		repos:    getRepos("foo-1", "foo-2", "foo-3"),
		want:     getRepos("foo-1", "foo-2", "foo-3"),
	}, {
		name:     "none",
		hostname: "bar",
		repos:    getRepos("foo-1", "foo-2", "foo-3"),
		want:     []types.MinimalRepo{},
	}, {
		name:     "subset",
		hostname: "foo",
		repos:    getRepos("foo-2", "bar-1", "foo-1", "foo-3"),
		want:     getRepos("foo-2", "foo-1", "foo-3"),
	}, {
		name:     "qualified",
		hostname: "baz.fully.qualified",
		repos:    getRepos("baz.fully.qualified:80-1", "baz.fully.qualified:80-2", "foo-1"),
		want:     getRepos("baz.fully.qualified:80-1", "baz.fully.qualified:80-2"),
	}, {
		name:     "unqualified",
		hostname: "baz",
		repos:    getRepos("baz.fully.qualified:80-1", "baz.fully.qualified:80-2", "foo-1"),
		want:     getRepos("baz.fully.qualified:80-1", "baz.fully.qualified:80-2"),
	}, {
		name:     "drop",
		hostname: "foo",
		indexed: map[string][]types.MinimalRepo{
			"foo": getRepos("foo-1", "foo-drop", "bar-drop", "bar-keep"),
			"bar": getRepos("foo-1", "bar-drop"),
		},
		repos: getRepos("foo-1", "foo-2", "foo-3", "bar-drop", "bar-keep"),
		want:  getRepos("foo-1", "foo-2", "foo-3", "bar-keep"),
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			indexed = tc.indexed
			got, err := index.ReposSubset(ctx, tc.hostname, index.Indexed(ctx, tc.hostname), tc.repos)
			if tc.errS != "" {
				got := fmt.Sprintf("%v", err)
				if !strings.Contains(got, tc.errS) {
					t.Fatalf("error %q does not contain %q", got, tc.errS)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if !cmp.Equal(tc.want, got) {
				t.Errorf("reposSubset mismatch (-want +got):\n%s", cmp.Diff(tc.want, got))
			}
		})
	}
}

func TestFindEndpoint(t *testing.T) {
	cases := []struct {
		name      string
		hostname  string
		endpoints []string
		want      string
		errS      string
	}{{
		name:      "empty",
		hostname:  "",
		endpoints: []string{"foo.internal", "bar.internal"},
		errS:      "hostname \"\" not found",
	}, {
		name:      "empty endpoints",
		hostname:  "foo",
		endpoints: []string{},
		errS:      "hostname \"foo\" not found",
	}, {
		name:      "bad prefix",
		hostname:  "foo",
		endpoints: []string{"foobar", "barfoo"},
		errS:      "hostname \"foo\" not found",
	}, {
		name:      "bad port",
		hostname:  "foo",
		endpoints: []string{"foo:80", "foo.internal"},
		errS:      "hostname \"foo\" matches multiple",
	}, {
		name:      "multiple",
		hostname:  "foo",
		endpoints: []string{"foo.internal", "foo.external"},
		errS:      "hostname \"foo\" matches multiple",
	}, {
		name:      "exact multiple",
		hostname:  "foo",
		endpoints: []string{"foo", "foo.internal"},
		errS:      "hostname \"foo\" matches multiple",
	}, {
		name:      "exact",
		hostname:  "foo",
		endpoints: []string{"foo", "bar"},
		want:      "foo",
	}, {
		name:      "prefix",
		hostname:  "foo",
		endpoints: []string{"foo.internal", "bar.internal"},
		want:      "foo.internal",
	}, {
		name:      "prefix with bad",
		hostname:  "foo",
		endpoints: []string{"foo.internal", "foobar.internal"},
		want:      "foo.internal",
	}, {
		name:      "port prefix",
		hostname:  "foo",
		endpoints: []string{"foo.internal:80", "bar.internal:80"},
		want:      "foo.internal:80",
	}, {
		name:      "port exact",
		hostname:  "foo.internal",
		endpoints: []string{"foo.internal:80", "bar.internal:80"},
		want:      "foo.internal:80",
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := findEndpoint(tc.endpoints, tc.hostname)
			if tc.errS != "" {
				got := fmt.Sprintf("%v", err)
				if !strings.Contains(got, tc.errS) {
					t.Fatalf("error %q does not contain %q", got, tc.errS)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}

			if tc.want != got {
				t.Errorf("findEndpoint got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFilterReposPendingIndexing(t *testing.T) {
	cases := []struct {
		name  string
		state map[string][]string
		want  []string
	}{
		{
			name: "draining, with duplicates",
			// map of host -> indexed repos. All repos are assigned to host "target".
			state: map[string][]string{
				"target": {"target_repo1", "target_repo2"},
				"drain":  {"target_repo1", "target_repo2", "target_repo3"},
			},
			want: []string{"target_repo3"},
		},
		{
			name: "draining, no duplicates",
			state: map[string][]string{
				"target": {"target_repo1", "target_repo2"},
				"drain":  {"target_repo3"},
			},
			want: []string{"target_repo3"},
		},
		{
			name: "not drained",
			state: map[string][]string{
				"target": {},
				"drain":  {"target_repo1", "target_repo2", "target_repo3", "target_repo4"},
			},
			want: []string{"target_repo1", "target_repo2", "target_repo3", "target_repo4"},
		},
		{
			name: "fully drained",
			state: map[string][]string{
				"target": {"target_repo1", "target_repo2", "target_repo3", "target_repo4"},
				"drain":  {},
			},
			want: []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			m := make(map[string]types.MinimalRepo)

			c := &Indexers{
				Map: prefixMap([]string{"target", "drain"}),
				Indexed: func(ctx context.Context, host string) zoekt.ReposMap {
					set := zoekt.ReposMap{}
					for _, s := range tc.state[host] {
						set[uint32(m[s].ID)] = zoekt.MinimalRepoListEntry{}
					}
					return set
				},
			}

			var repos []types.MinimalRepo
			for _, v := range tc.state {
				for _, repo := range v {
					r := types.MinimalRepo{
						ID:   api.RepoID(rand.Int31()),
						Name: api.RepoName(repo),
					}
					m[repo] = r
					repos = append(repos, r)
				}
			}

			indexed := make(zoekt.ReposMap)
			for _, v := range tc.state["drain"] {
				indexed[uint32(m[v].ID)] = zoekt.MinimalRepoListEntry{}
			}

			got, err := c.filterReposPendingIndexing(ctx, indexed, repos)
			require.NoError(t, err)

			gotRepos := []string{}
			for _, r := range got {
				gotRepos = append(gotRepos, string(r.Name))
			}
			if diff := cmp.Diff(tc.want, gotRepos); diff != "" {
				t.Fatalf("filterReposPendingIndexing mismatch (-want +got):\n%s", diff)
			}
		})

	}
}

// prefixMap assigns keys to values if the value is a prefix of key.
type prefixMap []string

func (m prefixMap) Endpoints() ([]string, error) {
	return m, nil
}

func (m prefixMap) Get(k string) (string, error) {
	for _, v := range m {
		if strings.HasPrefix(k, v) {
			return v, nil
		}
	}
	return "", nil
}
