package streaming

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
)

// Stats contains fields that should be returned by all funcs
// that contribute to the overall search result set.
type Stats struct {
	// IsLimitHit is true if we do not have all results that match the query.
	IsLimitHit bool

	// Repos that were matched by the repo-related filters.
	Repos map[api.RepoID]struct{}

	// Status is a RepoStatusMap of repository search statuses.
	Status search.RepoStatusMap

	// BackendsMissing is the number of search backends that failed to be
	// searched. This is due to it being unreachable. The most common reason
	// for this is during zoekt rollout.
	BackendsMissing int

	// ExcludedForks is the count of excluded forked repos because the search
	// query doesn't apply to them, but that we want to know about.
	ExcludedForks int

	// ExcludedArchived is the count of excluded archived repos because the
	// search query doesn't apply to them, but that we want to know about.
	ExcludedArchived int
}

// Update updates c with the other data, deduping as necessary. It modifies c but
// does not modify other.
func (c *Stats) Update(other *Stats) {
	if other == nil {
		return
	}

	c.IsLimitHit = c.IsLimitHit || other.IsLimitHit

	if c.Repos == nil && len(other.Repos) > 0 {
		c.Repos = make(map[api.RepoID]struct{}, len(other.Repos))
	}
	for id := range other.Repos {
		if _, ok := c.Repos[id]; !ok {
			c.Repos[id] = struct{}{}
		}
	}

	c.Status.Union(&other.Status)

	c.BackendsMissing += other.BackendsMissing
	c.ExcludedForks += other.ExcludedForks
	c.ExcludedArchived += other.ExcludedArchived
}

// Zero returns true if stats is empty. IE calling Update will result in no
// change.
func (c *Stats) Zero() bool {
	if c == nil {
		return true
	}

	return !(c.IsLimitHit ||
		len(c.Repos) > 0 ||
		c.Status.Len() > 0 ||
		c.BackendsMissing > 0 ||
		c.ExcludedForks > 0 ||
		c.ExcludedArchived > 0)
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
		{"repos", len(c.Repos)},
		{"backendsMissing", c.BackendsMissing},
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

	return "Stats{" + strings.Join(parts, " ") + "}"
}

// Equal provides custom comparison which is used by go-cmp
func (c *Stats) Equal(other *Stats) bool {
	return reflect.DeepEqual(c, other)
}
