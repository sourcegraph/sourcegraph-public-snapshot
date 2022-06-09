package jobutil

import (
	"context"

	"github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/alert"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

// NewFeelingLuckySearchJob generates opportunistic search queries by applying
// various rules in sequence, transforming the original input plan into various
// queries that alter its interpretation (e.g., search literally for quotes or
// not, attempt to search the pattern as a regexp, and so on).
//
// Generated queries return a resulting list of jobs. The application order of
// rules is deterministic. This means the query order, and therefore runtime
// execution, outputs the same search results for the same inputs. I.e., there
// is no random choice when applying a rule. The first job in this list
// propagates an alert-error that notifies consumers of valid generated queries
// for this plan.
func NewFeelingLuckySearchJob(inputs *run.SearchInputs, plan query.Plan) []job.Job {
	children := make([]job.Job, 0, len(plan))
	var generatedQueries []*search.ProposedQuery // Collects valid generated queries for alerting.
	for _, b := range plan {
		for description, rules := range rulesMap {
			generated := applyRules(b, rules...)
			if generated == nil {
				continue
			}

			child, err := NewBasicJob(inputs, *generated)
			if err != nil {
				return nil
			}

			children = append(children, child)
			generatedQueries = append(generatedQueries, &search.ProposedQuery{
				Description: description,
				Query:       query.StringHuman(generated.ToParseTree()),
				PatternType: query.SearchTypeLiteralDefault,
			})
		}
	}

	alertJob := &alternateQueriesAlertJob{AlternateQueries: generatedQueries}
	return append([]job.Job{alertJob}, children...)
}

// rule represents a transformation function on a Basic query. Applying rules
// cannot fail: either they apply and produce a valid, non-nil, Basic query, or
// they cannot apply, in which case they return nil. See the `unquotePatterns`
// rule for an example.
type rule func(query.Basic) *query.Basic

// rulesMap is an ordered map of rule lists. Each entry in the map represent one
// possible query production, if the sequence of the rules associated with this
// map's entry apply successfully. Example:
//
// If we have input query B0 and a map with two entries like this:
//
// {
//   "first rule list"  : [ R1, R2 ],
//   "second rule list" : [ R2 ],
// }
//
// Then:
//
// - If both entries apply, the map outputs [ B1, B2 ] where B1 is generated
//   from applying R1 then R2, and B2 is generated from just applying R2.
// - If only the first entry applies, R1 then R2, we get the output [ B1 ].
// - If only the second entry applies, R2 on its own, we get the output [ B2 ].
var rulesMap = map[string][]rule{
	"unquote patterns":      {unquotePatterns},
	"AND patterns together": {unorderedPatterns},
}

// applyRules applies every rule in sequence to `b`. If any rule does not apply, it returns nil.
func applyRules(b query.Basic, rules ...rule) *query.Basic {
	for _, rule := range rules {
		res := rule(b)
		if res == nil {
			return nil
		}
		b = *res
	}
	return &b
}

// unquotePatterns is a rule that unquotes all patterns in the input query (it
// removes quotes, and honors escape sequences inside quoted values).
func unquotePatterns(b query.Basic) *query.Basic {
	// Go back all the way to the raw tree representation :-). We just parse
	// the string as regex, since parsing with regex annotates quoted
	// patterns.
	rawParseTree, err := query.Parse(query.StringHuman(b.ToParseTree()), query.SearchTypeRegex)
	if err != nil {
		return nil
	}

	changed := false // track whether we've successfully changed any pattern, which means this rule applies.
	newParseTree := query.MapPattern(rawParseTree, func(value string, negated bool, annotation query.Annotation) query.Node {
		if annotation.Labels.IsSet(query.Quoted) {
			changed = true
			annotation.Labels.Unset(query.Quoted)
			annotation.Labels.Set(query.Literal)
			return query.Pattern{
				Value:      value,
				Negated:    negated,
				Annotation: annotation,
			}
		}
		return query.Pattern{
			Value:      value,
			Negated:    negated,
			Annotation: annotation,
		}
	})

	if !changed {
		// No unquoting happened, so we don't run the search.
		return nil
	}

	newNodes, err := query.Sequence(query.For(query.SearchTypeLiteralDefault))(newParseTree)
	if err != nil {
		return nil
	}

	newBasic, err := query.ToBasicQuery(newNodes)
	if err != nil {
		return nil
	}

	return &newBasic
}

// UnorderedPatterns generates a query that interprets all recognized patterns
// as unordered terms (`and`-ed terms). The implementation detail is that we
// simply map all `concat` nodes (after a raw parse) to `and` nodes. This works
// because parsing maintains the invariant that `concat` nodes only ever have
// pattern children.
func unorderedPatterns(b query.Basic) *query.Basic {
	rawParseTree, err := query.Parse(query.StringHuman(b.ToParseTree()), query.SearchTypeRegex)
	if err != nil {
		return nil
	}

	newParseTree, changed := mapConcat(rawParseTree)
	if !changed {
		return nil
	}

	newNodes, err := query.Sequence(query.For(query.SearchTypeLiteralDefault))(newParseTree)
	if err != nil {
		return nil
	}

	newBasic, err := query.ToBasicQuery(newNodes)
	if err != nil {
		return nil
	}

	return &newBasic
}

func mapConcat(q []query.Node) ([]query.Node, bool) {
	mapped := make([]query.Node, 0, len(q))
	changed := false
	for _, node := range q {
		if n, ok := node.(query.Operator); ok {
			if n.Kind != query.Concat {
				// recurse
				operands, newChanged := mapConcat(n.Operands)
				mapped = append(mapped, query.Operator{
					Kind:     n.Kind,
					Operands: operands,
				})
				changed = changed || newChanged
				continue
			}
			// no need to recurse: `concat` nodes only have patterns.
			mapped = append(mapped, query.Operator{
				Kind:     query.And,
				Operands: n.Operands,
			})
			changed = true
			continue
		}
		mapped = append(mapped, node)
	}
	return mapped, changed
}

// A job that propogates which queries were generated. Consumers can use this
// information (represented by an error value) into alerts, and can further
// display prominently in a client or user-facing app.
type alternateQueriesAlertJob struct {
	AlternateQueries []*search.ProposedQuery
}

func (g *alternateQueriesAlertJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (*search.Alert, error) {
	return nil, &alert.ErrLuckyQueries{ProposedQueries: g.AlternateQueries}
}

func (*alternateQueriesAlertJob) Name() string {
	return "AlternateQueriesAlertJob"
}

func (*alternateQueriesAlertJob) Tags() []log.Field {
	return []log.Field{}
}
