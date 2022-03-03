package backend

import (
	"bytes"
	"encoding/json"
	"sort"

	"github.com/google/zoekt"
	"github.com/grafana/regexp"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/schema"
)

// zoektIndexOptions are options which change what we index for a
// repository. Everytime a repository is indexed by zoekt this structure is
// fetched. See getIndexOptions in the zoekt codebase.
//
// We only specify a subset of the fields.
type zoektIndexOptions struct {
	// Name is the Repository Name.
	Name string

	// RepoID is the Sourcegraph Repository ID.
	RepoID int32

	// Public is true if the repository is public and does not require auth
	// filtering.
	Public bool

	// Fork is true if the repository is a fork.
	Fork bool

	// Archived is true if the repository is archived.
	Archived bool

	// LargeFiles is a slice of glob patterns where matching file paths should
	// be indexed regardless of their size. The pattern syntax can be found
	// here: https://golang.org/pkg/path/filepath/#Match.
	LargeFiles []string

	// Symbols if true will make zoekt index the output of ctags.
	Symbols bool

	// Branches is a slice of branches to index.
	Branches []zoekt.RepositoryBranch `json:",omitempty"`

	// Priority indicates ranking in results, higher first.
	Priority float64 `json:",omitempty"`

	// Error if non-empty indicates the request failed for the repo.
	Error string `json:",omitempty"`
}

// RepoIndexOptions are the options used by GetIndexOptions for a specific
// repository.
type RepoIndexOptions struct {
	// Name is the Repository Name.
	Name string

	// RepoID is the Sourcegraph Repository ID.
	RepoID int32

	// Public is true if the repository is public and does not require auth
	// filtering.
	Public bool

	// Priority indicates ranking in results, higher first.
	Priority float64

	// Fork is true if the repository is a fork.
	Fork bool

	// Archived is true if the repository is archived.
	Archived bool

	// GetVersion is used to resolve revisions for a repo. If it fails, the
	// error is encoded in the body. If the revision is missing, an empty
	// string should be returned rather than an error.
	GetVersion func(branch string) (string, error)
}

type getRepoIndexOptsFn func(repoID int32) (*RepoIndexOptions, error)

// GetIndexOptions returns a json blob for consumption by
// sourcegraph-zoekt-indexserver. It is for repos based on site settings c.
func GetIndexOptions(
	c *schema.SiteConfiguration,
	getRepoIndexOptions getRepoIndexOptsFn,
	getSearchContextRevisions func(repoID int32) ([]string, error),
	repos ...int32,
) []byte {
	// Limit concurrency to 32 to avoid too many active network requests and
	// strain on gitserver (as ported from zoekt-sourcegraph-indexserver). In
	// future we want a more intelligent global limit based on scale.
	sema := make(chan struct{}, 32)
	results := make([][]byte, len(repos))
	getSiteConfigRevisions := siteConfigRevisionsRuleFunc(c)

	for i := range repos {
		sema <- struct{}{}
		go func(i int) {
			defer func() { <-sema }()
			results[i] = getIndexOptions(c, repos[i], getRepoIndexOptions, getSearchContextRevisions, getSiteConfigRevisions)
		}(i)
	}

	// Wait for jobs to finish (acquire full semaphore)
	for i := 0; i < cap(sema); i++ {
		sema <- struct{}{}
	}

	return bytes.Join(results, []byte{'\n'})
}

func getIndexOptions(
	c *schema.SiteConfiguration,
	repoID int32,
	getRepoIndexOptions func(repoID int32) (*RepoIndexOptions, error),
	getSearchContextRevisions func(repoID int32) ([]string, error),
	getSiteConfigRevisions revsRuleFunc,
) []byte {
	opts, err := getRepoIndexOptions(repoID)
	if err != nil {
		return marshal(&zoektIndexOptions{Error: err.Error()})
	}

	o := &zoektIndexOptions{
		Name:       opts.Name,
		RepoID:     opts.RepoID,
		Public:     opts.Public,
		Priority:   opts.Priority,
		Fork:       opts.Fork,
		Archived:   opts.Archived,
		LargeFiles: c.SearchLargeFiles,
		Symbols:    getBoolPtr(c.SearchIndexSymbolsEnabled, true),
	}

	// Set of branch names. Always index HEAD
	branches := map[string]struct{}{"HEAD": {}}

	// Add all branches that are referenced by search.index.branches and search.index.revisions.
	if getSiteConfigRevisions != nil {
		for _, rev := range getSiteConfigRevisions(opts) {
			branches[rev] = struct{}{}
		}
	}

	// Add all branches that are referenced by search contexts
	revs, err := getSearchContextRevisions(opts.RepoID)
	if err != nil {
		return marshal(&zoektIndexOptions{Error: err.Error()})
	}
	for _, rev := range revs {
		branches[rev] = struct{}{}
	}

	// empty string means HEAD which is already in the set. Rather than
	// sanitize all inputs, just adjust the set before we start resolving.
	delete(branches, "")

	for branch := range branches {
		v, err := opts.GetVersion(branch)
		if err != nil {
			return marshal(&zoektIndexOptions{Error: err.Error()})
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

	return marshal(o)
}

type revsRuleFunc func(*RepoIndexOptions) (revs []string)

func siteConfigRevisionsRuleFunc(c *schema.SiteConfiguration) revsRuleFunc {
	if c == nil || c.ExperimentalFeatures == nil {
		return nil
	}

	rules := make([]revsRuleFunc, 0, len(c.ExperimentalFeatures.SearchIndexRevisions))
	for _, rule := range c.ExperimentalFeatures.SearchIndexRevisions {
		rule := rule
		switch {
		case rule.Name != "":
			namePattern, err := regexp.Compile(rule.Name)
			if err != nil {
				log15.Error("error compiling regex from search.index.revisions", "regex", rule.Name, "err", err)
				continue
			}

			rules = append(rules, func(o *RepoIndexOptions) []string {
				if !namePattern.MatchString(o.Name) {
					return nil
				}
				return rule.Revisions
			})
		}
	}

	return func(o *RepoIndexOptions) (matched []string) {
		cfg := c.ExperimentalFeatures

		if len(cfg.SearchIndexBranches) != 0 {
			matched = append(matched, cfg.SearchIndexBranches[o.Name]...)
		}

		for _, rule := range rules {
			matched = append(matched, rule(o)...)
		}

		return matched
	}
}

func getBoolPtr(b *bool, default_ bool) bool {
	if b == nil {
		return default_
	}
	return *b
}

func marshal(o *zoektIndexOptions) []byte {
	b, _ := json.Marshal(o)
	return b
}
