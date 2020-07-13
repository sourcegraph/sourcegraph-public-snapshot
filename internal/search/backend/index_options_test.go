package backend

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/zoekt"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestGetIndexOptions(t *testing.T) {
	vc := parseVersionContext
	vcConf := func(contexts ...*schema.VersionContext) schema.SiteConfiguration {
		return schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{
				VersionContexts: contexts,
			},
		}
	}

	cases := []struct {
		name string
		conf schema.SiteConfiguration
		repo string
		want zoektIndexOptions
	}{{
		name: "default",
		conf: schema.SiteConfiguration{},
		repo: "",
		want: zoektIndexOptions{
			Symbols: true,
		},
	}, {
		name: "nosymbols",
		conf: schema.SiteConfiguration{
			SearchIndexSymbolsEnabled: boolPtr(false)},
		repo: "",
		want: zoektIndexOptions{},
	}, {
		name: "largefiles",
		conf: schema.SiteConfiguration{
			SearchLargeFiles: []string{"**/*.jar", "*.bin"},
		},
		repo: "",
		want: zoektIndexOptions{
			Symbols:    true,
			LargeFiles: []string{"**/*.jar", "*.bin"},
		},
	}, {
		name: "implicit HEAD",
		conf: vcConf(vc("foo", "repo@b", "repo@a"), vc("bar", "repo@c", "repo@a", "other@d")),
		repo: "repo",
		want: zoektIndexOptions{
			Symbols: true,
			Branches: []zoekt.RepositoryBranch{
				{Name: "HEAD", Version: "!HEAD"},
				{Name: "a", Version: "!a"},
				{Name: "b", Version: "!b"},
				{Name: "c", Version: "!c"},
			},
		},
	}, {
		name: "implicit HEAD not in vc",
		conf: vcConf(vc("foo", "repo@a")),
		repo: "not_in_version_context",
		want: zoektIndexOptions{
			Symbols: true,
			Branches: []zoekt.RepositoryBranch{{
				Name:    "HEAD",
				Version: "!HEAD",
			}},
		},
	}, {
		name: "explicit HEAD",
		conf: vcConf(vc("foo", "repo@HEAD", "repo@a")),
		repo: "repo",
		want: zoektIndexOptions{
			Symbols: true,
			Branches: []zoekt.RepositoryBranch{
				{Name: "HEAD", Version: "!HEAD"},
				{Name: "a", Version: "!a"},
			},
		},
	}, {
		// a revision can be the empty string, treat as HEAD
		name: "explicit HEAD empty",
		conf: vcConf(vc("foo", "repo", "repo@a")),
		repo: "repo",
		want: zoektIndexOptions{
			Symbols: true,
			Branches: []zoekt.RepositoryBranch{
				{Name: "HEAD", Version: "!HEAD"},
				{Name: "a", Version: "!a"},
			},
		},
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := GetIndexOptions(&tc.conf, tc.repo, func(branch string) (string, error) {
				return "!" + branch, nil
			})
			if err != nil {
				t.Fatal(err)
			}

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
	conf := schema.SiteConfiguration{
		ExperimentalFeatures: &schema.ExperimentalFeatures{
			VersionContexts: []*schema.VersionContext{
				parseVersionContext("foo", "repo@b1", "repo@b2"),
			},
		},
	}

	boom := errors.New("boom")
	cases := []struct {
		name    string
		f       func(string) (string, error)
		want    []zoekt.RepositoryBranch
		wantErr error
	}{{
		name: "error",
		f: func(_ string) (string, error) {
			return "", boom
		},
		wantErr: boom,
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
			b, err := GetIndexOptions(&conf, "repo", tc.f)
			if err != tc.wantErr {
				t.Fatalf("expected error %v, got body %s and error %v", tc.wantErr, b, err)
			}
			if tc.wantErr != nil {
				return
			}

			var got zoektIndexOptions
			if err := json.Unmarshal(b, &got); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(tc.want, got.Branches); diff != "" {
				t.Fatal("mismatch (-want, +got):\n", diff)
			}
		})
	}
}

func parseVersionContext(name string, repoRevStrs ...string) *schema.VersionContext {
	var repoRevs []*schema.VersionContextRevision
	for _, repo := range repoRevStrs {
		rev := ""
		if idx := strings.LastIndex(repo, "@"); idx > 0 {
			rev = repo[idx+1:]
			repo = repo[:idx]
		}
		repoRevs = append(repoRevs, &schema.VersionContextRevision{
			Repo: repo,
			Rev:  rev,
		})
	}
	return &schema.VersionContext{
		Name:      name,
		Revisions: repoRevs,
	}
}

func boolPtr(b bool) *bool {
	return &b
}
