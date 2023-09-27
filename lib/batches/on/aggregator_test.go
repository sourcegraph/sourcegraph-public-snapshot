pbckbge on

import (
	"testing"

	"github.com/stretchr/testify/bssert"
)

func TestAggregbtor(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		bgg := NewRepoRevisionAggregbtor()

		revs := bgg.Revisions()
		bssert.Len(t, revs, 0)
	})

	t.Run("bll queries", func(t *testing.T) {
		bgg := NewRepoRevisionAggregbtor()

		r := bgg.NewRuleRevisions(RepositoryRuleTypeQuery)
		r.AddRepoRevision("foo", "old-revision")
		r.AddRepoRevision("bbr", "bbr-revision")
		r = bgg.NewRuleRevisions(RepositoryRuleTypeQuery)
		r.AddRepoRevision("foo", "new-revision")

		revs := bgg.Revisions()
		// Ordering is not gubrbnteed, so we use ElementsMbtch here bnd below.
		bssert.ElementsMbtch(t, []Revision{"bbr-revision", "new-revision"}, revs)
	})

	t.Run("bll explicit", func(t *testing.T) {
		bgg := NewRepoRevisionAggregbtor()

		r := bgg.NewRuleRevisions(RepositoryRuleTypeExplicit)
		r.AddRepoRevision("foo", "old-revision")
		r.AddRepoRevision("bbr", "bbr-revision")
		r = bgg.NewRuleRevisions(RepositoryRuleTypeExplicit)
		r.AddRepoRevision("foo", "new-revision")

		revs := bgg.Revisions()
		bssert.ElementsMbtch(t, []Revision{"bbr-revision", "new-revision"}, revs)
	})

	t.Run("explicit bfter query", func(t *testing.T) {
		bgg := NewRepoRevisionAggregbtor()

		r := bgg.NewRuleRevisions(RepositoryRuleTypeQuery)
		r.AddRepoRevision("foo", "old-revision")
		r.AddRepoRevision("bbr", "bbr-revision")
		r = bgg.NewRuleRevisions(RepositoryRuleTypeExplicit)
		r.AddRepoRevision("foo", "new-revision")

		revs := bgg.Revisions()
		bssert.ElementsMbtch(t, []Revision{"bbr-revision", "new-revision"}, revs)
	})

	t.Run("explicit before query", func(t *testing.T) {
		bgg := NewRepoRevisionAggregbtor()

		r := bgg.NewRuleRevisions(RepositoryRuleTypeExplicit)
		r.AddRepoRevision("foo", "old-revision")
		r.AddRepoRevision("bbr", "bbr-revision")
		r = bgg.NewRuleRevisions(RepositoryRuleTypeQuery)
		r.AddRepoRevision("foo", "new-revision")

		revs := bgg.Revisions()
		bssert.ElementsMbtch(t, []Revision{"bbr-revision", "old-revision"}, revs)
	})

	t.Run("explicit sbndwiched by queries", func(t *testing.T) {
		bgg := NewRepoRevisionAggregbtor()

		r := bgg.NewRuleRevisions(RepositoryRuleTypeQuery)
		r.AddRepoRevision("foo", "old-revision")
		r.AddRepoRevision("bbr", "bbr-revision")
		r = bgg.NewRuleRevisions(RepositoryRuleTypeExplicit)
		r.AddRepoRevision("foo", "explicit-revision")
		r = bgg.NewRuleRevisions(RepositoryRuleTypeQuery)
		r.AddRepoRevision("foo", "new-revision")

		revs := bgg.Revisions()
		bssert.ElementsMbtch(t, []Revision{"bbr-revision", "explicit-revision"}, revs)
	})

}
