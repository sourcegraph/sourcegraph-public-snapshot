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

	// Branches is a slice of branches to index. If empty it will be
	// HEAD. These will be resolved, so you can pass in tags/refs/commits.
	//
	// Indexing multiple branches is still experimental. As such this should
	// only be set if an admin has opted into it.
	Branches []zoekt.RepositoryBranch `json:",omitempty"`
}

// GetIndexOptions returns a json blob for consumption by
// sourcegraph-zoekt-indexserver. It is for repoName based on site settings c.
//
// getVersion is used to resolve revisions for repoName. If it fails, the
// error is returned.
func GetIndexOptions(c *schema.SiteConfiguration, repoName string, getVersion func(branch string) (string, error)) ([]byte, error) {
	o := &zoektIndexOptions{
		LargeFiles: c.SearchLargeFiles,
		Symbols:    getBoolPtr(c.SearchIndexSymbolsEnabled, true),
	}

	// Only set Branches if we have VersionContexts set. Using the presence as
	// a feature flag for multi-branch indexing.
	if repoName != "" && c.ExperimentalFeatures != nil && len(c.ExperimentalFeatures.VersionContexts) > 0 {
		// Set of branch names. Always index HEAD
		branches := map[string]struct{}{"HEAD": {}}
		for _, vc := range c.ExperimentalFeatures.VersionContexts {
			for _, rev := range vc.Revisions {
				if rev.Repo == repoName && rev.Rev != "" {
					branches[rev.Rev] = struct{}{}
				}
			}
		}

		for branch := range branches {
			v, err := getVersion(branch)
			if err != nil {
				return nil, err
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
	}

	return json.Marshal(o)
}

func getBoolPtr(b *bool, default_ bool) bool {
	if b == nil {
		return default_
	}
	return *b
}
