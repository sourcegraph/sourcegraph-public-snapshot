package on

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAggregator(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		agg := NewAggregator()

		revs := agg.Revisions()
		assert.Len(t, revs, 0)
	})

	t.Run("all queries", func(t *testing.T) {
		agg := NewAggregator()

		result := agg.NewRuleResult(RepositoryRuleTypeQuery)
		result.AddRepoRevision("foo", "old-revision")
		result.AddRepoRevision("bar", "bar-revision")
		result = agg.NewRuleResult(RepositoryRuleTypeQuery)
		result.AddRepoRevision("foo", "new-revision")

		revs := agg.Revisions()
		// Ordering is not guaranteed, so we use ElementsMatch here and below.
		assert.ElementsMatch(t, []Revision{"bar-revision", "new-revision"}, revs)
	})

	t.Run("all explicit", func(t *testing.T) {
		agg := NewAggregator()

		result := agg.NewRuleResult(RepositoryRuleTypeExplicit)
		result.AddRepoRevision("foo", "old-revision")
		result.AddRepoRevision("bar", "bar-revision")
		result = agg.NewRuleResult(RepositoryRuleTypeExplicit)
		result.AddRepoRevision("foo", "new-revision")

		revs := agg.Revisions()
		assert.ElementsMatch(t, []Revision{"bar-revision", "new-revision"}, revs)
	})

	t.Run("explicit after query", func(t *testing.T) {
		agg := NewAggregator()

		result := agg.NewRuleResult(RepositoryRuleTypeQuery)
		result.AddRepoRevision("foo", "old-revision")
		result.AddRepoRevision("bar", "bar-revision")
		result = agg.NewRuleResult(RepositoryRuleTypeExplicit)
		result.AddRepoRevision("foo", "new-revision")

		revs := agg.Revisions()
		assert.ElementsMatch(t, []Revision{"bar-revision", "new-revision"}, revs)
	})

	t.Run("explicit before query", func(t *testing.T) {
		agg := NewAggregator()

		result := agg.NewRuleResult(RepositoryRuleTypeExplicit)
		result.AddRepoRevision("foo", "old-revision")
		result.AddRepoRevision("bar", "bar-revision")
		result = agg.NewRuleResult(RepositoryRuleTypeQuery)
		result.AddRepoRevision("foo", "new-revision")

		revs := agg.Revisions()
		assert.ElementsMatch(t, []Revision{"bar-revision", "old-revision"}, revs)
	})

	t.Run("explicit sandwiched by queries", func(t *testing.T) {
		agg := NewAggregator()

		result := agg.NewRuleResult(RepositoryRuleTypeQuery)
		result.AddRepoRevision("foo", "old-revision")
		result.AddRepoRevision("bar", "bar-revision")
		result = agg.NewRuleResult(RepositoryRuleTypeExplicit)
		result.AddRepoRevision("foo", "explicit-revision")
		result = agg.NewRuleResult(RepositoryRuleTypeQuery)
		result.AddRepoRevision("foo", "new-revision")

		revs := agg.Revisions()
		assert.ElementsMatch(t, []Revision{"bar-revision", "explicit-revision"}, revs)
	})

}
