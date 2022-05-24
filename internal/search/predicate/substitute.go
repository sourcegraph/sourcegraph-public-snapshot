package predicate

import (
	"context"
	"sync"

	"github.com/grafana/regexp"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrNoResults = errors.New("no results returned for predicate")

// Expand takes a query plan, and replaces any predicates with their expansion. The returned plan
// is guaranteed to be predicate-free.
func Expand(ctx context.Context, clients job.RuntimeClients, inputs *run.SearchInputs, oldPlan query.Plan) (_ query.Plan, err error) {
	tr, ctx := trace.New(ctx, "ExpandPredicates", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	var (
		mu      sync.Mutex
		newPlan = make(query.Plan, 0, len(oldPlan))
	)
	g, ctx := errgroup.WithContext(ctx)

	for _, q := range oldPlan {
		q := q
		g.Go(func() error {
			predicatePlan, err := Substitute(q, func(plan query.Plan) (result.Matches, error) {
				predicateJob, err := jobutil.NewPlanJob(inputs, plan)
				if err != nil {
					return nil, err
				}

				agg := streaming.NewAggregatingStream()
				_, err = predicateJob.Run(ctx, clients, agg)
				if err != nil {
					return nil, err
				}

				return agg.Results, nil
			})
			if errors.Is(err, ErrNoResults) {
				// The predicate has no results, so neither will this basic query
				return nil
			}
			if err != nil {
				// Fail if predicate processing fails.
				return err
			}

			mu.Lock()
			defer mu.Unlock()

			if predicatePlan != nil {
				// If the predicate generated a new plan, use that
				newPlan = append(newPlan, predicatePlan...)
			} else {
				// Otherwise, just use the original basic query
				newPlan = append(newPlan, q)
			}
			return nil
		})
	}

	return newPlan, g.Wait()
}

// Substitute replaces predicates that generate plans and substitutes them to create a new Plan.
func Substitute(q query.Basic, evaluate func(query.Plan) (result.Matches, error)) (query.Plan, error) {
	var topErr error
	success := false
	newQ := query.MapParameter(q.ToParseTree(), func(field, value string, neg bool, ann query.Annotation) query.Node {
		orig := query.Parameter{
			Field:      field,
			Value:      value,
			Negated:    neg,
			Annotation: ann,
		}

		if !ann.Labels.IsSet(query.IsPredicate) {
			return orig
		}

		if topErr != nil {
			return orig
		}

		name, params := query.ParseAsPredicate(value)
		predicate := query.DefaultPredicateRegistry.Get(field, name)
		predicate.ParseParams(params)
		plan, err := predicate.Plan(q)
		if err != nil {
			topErr = err
			return nil
		}
		if plan == nil {
			return orig
		}
		matches, err := evaluate(plan)
		if err != nil {
			topErr = err
			return nil
		}

		var nodes []query.Node
		switch predicate.Field() {
		case query.FieldRepo:
			nodes, err = searchResultsToRepoNodes(matches)
			if err != nil {
				topErr = err
				return nil
			}
		case query.FieldFile:
			nodes, err = searchResultsToFileNodes(matches)
			if err != nil {
				topErr = err
				return nil
			}
		default:
			topErr = errors.Errorf("unsupported predicate result type %q", predicate.Field())
			return nil
		}

		// If no results are returned, we need to return a sentinel error rather
		// than an empty expansion because an empty expansion means "everything"
		// rather than "nothing".
		if len(nodes) == 0 {
			topErr = ErrNoResults
			return nil
		}

		// A predicate was successfully evaluated and has results.
		success = true

		// No need to return an operator for only one result
		if len(nodes) == 1 {
			return nodes[0]
		}

		return query.Operator{
			Kind:     query.Or,
			Operands: nodes,
		}
	})

	if topErr != nil || !success {
		return nil, topErr
	}
	return query.BuildPlan(newQ), nil
}

// searchResultsToRepoNodes converts a set of search results into repository nodes
// such that they can be used to replace a repository predicate
func searchResultsToRepoNodes(matches []result.Match) ([]query.Node, error) {
	nodes := make([]query.Node, 0, len(matches))
	for _, match := range matches {
		repoMatch, ok := match.(*result.RepoMatch)
		if !ok {
			return nil, errors.Errorf("expected type %T, but got %T", &result.RepoMatch{}, match)
		}

		repoFieldValue := "^" + regexp.QuoteMeta(string(repoMatch.Name)) + "$"
		if repoMatch.Rev != "" {
			repoFieldValue += "@" + repoMatch.Rev
		}

		nodes = append(nodes, query.Parameter{
			Field: query.FieldRepo,
			Value: repoFieldValue,
		})
	}

	return nodes, nil
}

// searchResultsToFileNodes converts a set of search results into repo/file nodes so that they
// can replace a file predicate
func searchResultsToFileNodes(matches []result.Match) ([]query.Node, error) {
	nodes := make([]query.Node, 0, len(matches))
	for _, match := range matches {
		fileMatch, ok := match.(*result.FileMatch)
		if !ok {
			return nil, errors.Errorf("expected type %T, but got %T", &result.FileMatch{}, match)
		}

		repoFieldValue := "^" + regexp.QuoteMeta(string(fileMatch.Repo.Name)) + "$"
		if fileMatch.InputRev != nil {
			repoFieldValue += "@" + *fileMatch.InputRev
		}

		// We create AND nodes to match both the repo and the file at the same time so
		// we don't get files of the same name from different repositories.
		nodes = append(nodes, query.Operator{
			Kind: query.And,
			Operands: []query.Node{
				query.Parameter{
					Field: query.FieldRepo,
					Value: repoFieldValue,
				},
				query.Parameter{
					Field: query.FieldFile,
					Value: "^" + regexp.QuoteMeta(fileMatch.Path) + "$",
				},
			},
		})
	}

	return nodes, nil
}
