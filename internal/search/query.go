package search

import "github.com/sourcegraph/sourcegraph/internal/search/query"

// Pipeline processes zero or more steps to produce a query. The first step must
// be Init, otherwise this function is a no-op.
func Pipeline(steps ...query.Step) (query.Plan, error) {
	nodes, err := query.Sequence(steps...)(nil)
	if err != nil {
		return nil, err
	}

	disjuncts := query.Dnf(nodes)
	if err := query.Validate(disjuncts); err != nil {
		return nil, err
	}

	plan, err := query.ToPlan(disjuncts)
	if err != nil {
		return nil, err
	}
	plan = query.MapPlan(plan, query.ConcatRevFilters)
	return plan, nil
}
