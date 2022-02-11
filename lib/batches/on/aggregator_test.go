package on

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAggregator(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		agg := NewRepoRevisionAggregator()

		revs := agg.Revisions()
		assert.Len(t, revs, 0)
	})

	t.Run("all queries", func(t *testing.T) {
		agg := NewRepoRevisionAggregator()

		r := agg.NewRuleRevisions(RepositoryRuleTypeQuery)
		r.AddRepoRevision("foo", "old-revision")
		r.AddRepoRevision("bar", "bar-revision")
		r = agg.NewRuleRevisions(RepositoryRuleTypeQuery)
		r.AddRepoRevision("foo", "new-revision")

		revs := agg.Revisions()
		// Ordering is not guaranteed, so we use ElementsMatch here and below.
		assert.ElementsMatch(t, []Revision{"bar-revision", "new-revision"}, revs)
	})

	t.Run("all explicit", func(t *testing.T) {
		agg := NewRepoRevisionAggregator()

		r := agg.NewRuleRevisions(RepositoryRuleTypeExplicit)
		r.AddRepoRevision("foo", "old-revision")
		r.AddRepoRevision("bar", "bar-revision")
		r = agg.NewRuleRevisions(RepositoryRuleTypeExplicit)
		r.AddRepoRevision("foo", "new-revision")

		revs := agg.Revisions()
		assert.ElementsMatch(t, []Revision{"bar-revision", "new-revision"}, revs)
	})

	t.Run("explicit after query", func(t *testing.T) {
		agg := NewRepoRevisionAggregator()

		r := agg.NewRuleRevisions(RepositoryRuleTypeQuery)
		r.AddRepoRevision("foo", "old-revision")
		r.AddRepoRevision("bar", "bar-revision")
		r = agg.NewRuleRevisions(RepositoryRuleTypeExplicit)
		r.AddRepoRevision("foo", "new-revision")

		revs := agg.Revisions()
		assert.ElementsMatch(t, []Revision{"bar-revision", "new-revision"}, revs)
	})

	t.Run("explicit before query", func(t *testing.T) {
		agg := NewRepoRevisionAggregator()

		r := agg.NewRuleRevisions(RepositoryRuleTypeExplicit)
		r.AddRepoRevision("foo", "old-revision")
		r.AddRepoRevision("bar", "bar-revision")
		r = agg.NewRuleRevisions(RepositoryRuleTypeQuery)
		r.AddRepoRevision("foo", "new-revision")

		revs := agg.Revisions()
		assert.ElementsMatch(t, []Revision{"bar-revision", "old-revision"}, revs)
	})

	t.Run("explicit sandwiched by queries", func(t *testing.T) {
		agg := NewRepoRevisionAggregator()

		r := agg.NewRuleRevisions(RepositoryRuleTypeQuery)
		r.AddRepoRevision("foo", "old-revision")
		r.AddRepoRevision("bar", "bar-revision")
		r = agg.NewRuleRevisions(RepositoryRuleTypeExplicit)
		r.AddRepoRevision("foo", "explicit-revision")
		r = agg.NewRuleRevisions(RepositoryRuleTypeQuery)
		r.AddRepoRevision("foo", "new-revision")

		revs := agg.Revisions()
		assert.ElementsMatch(t, []Revision{"bar-revision", "explicit-revision"}, revs)
	})

}
