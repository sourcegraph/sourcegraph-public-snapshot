package streaming

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// Stats contains fields that should be returned by all funcs
// that contribute to the overall search result set.
type Stats struct {
	// IsLimitHit is true if we do not have all results that match the query.
	IsLimitHit bool

	// Repos that were matched by the repo-related filters. This should only
	// be set once by search, when we have resolved Repos.
	Repos *search.Repos

	// Status is a RepoStatusMap of repository search statuses.
	Status search.RepoStatusMap

	// ExcludedForks is the count of excluded forked repos because the search
	// query doesn't apply to them, but that we want to know about.
	ExcludedForks int

	// ExcludedArchived is the count of excluded archived repos because the
	// search query doesn't apply to them, but that we want to know about.
	ExcludedArchived int

	// IsIndexUnavailable is true if indexed search was unavailable.
	IsIndexUnavailable bool
}

// update updates c with the other data, deduping as necessary. It modifies c but
// does not modify other.
func (c *Stats) Update(other *Stats) {
	if other == nil {
		return
	}

	c.IsLimitHit = c.IsLimitHit || other.IsLimitHit
	c.IsIndexUnavailable = c.IsIndexUnavailable || other.IsIndexUnavailable

	if c.Repos == nil && other.Repos.Len() > 0 {
		c.Repos = search.NewRepos()
	}

	other.Repos.ForEach(func(r *types.RepoName, revs search.RevSpecs) error {
		c.Repos.Add(r, revs...)
		return nil
	})

	c.Status.Union(&other.Status)

	c.ExcludedForks = c.ExcludedForks + other.ExcludedForks
	c.ExcludedArchived = c.ExcludedArchived + other.ExcludedArchived
}

// Zero returns true if stats is empty. IE calling Update will result in no
// change.
func (c *Stats) Zero() bool {
	if c == nil {
		return true
	}

	return !(c.IsLimitHit ||
		c.Repos.Len() > 0 ||
		c.Status.Len() > 0 ||
		c.ExcludedForks > 0 ||
		c.ExcludedArchived > 0 ||
		c.IsIndexUnavailable)
}

func (c *Stats) String() string {
	if c == nil {
		return "Stats{}"
	}

	parts := []string{
		fmt.Sprintf("status=%s", c.Status.String()),
	}
	nums := []struct {
		name string
		n    int
	}{
		{"repos", c.Repos.Len()},
		{"excludedForks", c.ExcludedForks},
		{"excludedArchived", c.ExcludedArchived},
	}
	for _, p := range nums {
		if p.n != 0 {
			parts = append(parts, fmt.Sprintf("%s=%d", p.name, p.n))
		}
	}
	if c.IsLimitHit {
		parts = append(parts, "limitHit")
	}
	if c.IsIndexUnavailable {
		parts = append(parts, "indexUnavailable")
	}

	return "Stats{" + strings.Join(parts, " ") + "}"
}

// Equal provides custom comparison which is used by go-cmp
func (c *Stats) Equal(other *Stats) bool {
	return reflect.DeepEqual(c, other)
}

// Deref returns the zero-valued stats if its receiver is nil
func (c *Stats) Deref() Stats {
	if c != nil {
		return *c
	}
	return Stats{}
}
