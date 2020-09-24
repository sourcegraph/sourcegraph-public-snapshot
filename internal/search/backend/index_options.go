package backend

import (
	"encoding/json"
	"sort"

	"github.com/google/zoekt"
	"github.com/sourcegraph/sourcegraph/schema"
)

// zoektIndexOptions are options which change what we index for a
// repository. Everytime a repository is indexed by zoekt this structure is
// fetched. See getIndexOptions in the zoekt codebase.
//
// We only specify a subset of the fields.
type zoektIndexOptions struct {
	// LargeFiles is a slice of glob patterns where matching file paths should
	// be indexed regardless of their size. The pattern syntax can be found
	// here: https://golang.org/pkg/path/filepath/#Match.
	LargeFiles []string

	// Symbols if true will make zoekt index the output of ctags.
	Symbols bool

	// Branches is a slice of branches to index.
	Branches []zoekt.RepositoryBranch `json:",omitempty"`
}

// GetIndexOptions returns a json blob for consumption by
// sourcegraph-zoekt-indexserver. It is for repoName based on site settings c.
//
// getVersion is used to resolve revisions for repoName. If it fails, the
// error is returned. If the revision is missing, an empty string should be
// returned rather than an error.
func GetIndexOptions(c *schema.SiteConfiguration, repoName string, getVersion func(branch string) (string, error)) ([]byte, error) {
	o := &zoektIndexOptions{
		LargeFiles: c.SearchLargeFiles,
		Symbols:    getBoolPtr(c.SearchIndexSymbolsEnabled, true),
	}

	// Backwards compatibility for older Zoekt
	if repoName == "" {
		return json.Marshal(o)
	}

	// Set of branch names. Always index HEAD
	branches := map[string]struct{}{"HEAD": {}}

	if c.ExperimentalFeatures != nil {
		for _, vc := range c.ExperimentalFeatures.VersionContexts {
			for _, rev := range vc.Revisions {
				if rev.Repo == repoName && rev.Rev != "" {
					branches[rev.Rev] = struct{}{}
				}
			}
		}

		for _, rev := range c.ExperimentalFeatures.SearchIndexBranches[repoName] {
			branches[rev] = struct{}{}
		}
	}

	for branch := range branches {
		v, err := getVersion(branch)
		if err != nil {
			return nil, err
		}

		// If we failed to resolve a branch, skip it
		if v == "" {
			continue
		}

		o.Branches = append(o.Branches, zoekt.RepositoryBranch{
			Name:    branch,
			Version: v,
		})
	}

	sort.Slice(o.Branches, func(i, j int) bool {
		a, b := o.Branches[i].Name, o.Branches[j].Name
		// Zoekt treats first branch as default branch, so put HEAD first
		if a == "HEAD" || b == "HEAD" {
			return a == "HEAD"
		}
		return a < b
	})

	// If the first branch is not HEAD, do not index anything. This should
	// not happen, since HEAD should always exist if other branches exist.
	if len(o.Branches) == 0 || o.Branches[0].Name != "HEAD" {
		o.Branches = nil
	}

	// Zoekt has a limit of 64 branches
	if len(o.Branches) > 64 {
		o.Branches = o.Branches[:64]
	}

	return json.Marshal(o)
}

func getBoolPtr(b *bool, default_ bool) bool {
	if b == nil {
		return default_
	}
	return *b
}
