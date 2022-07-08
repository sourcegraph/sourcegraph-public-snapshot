package jobutil

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-enry/go-enry/v2"
	"github.com/opentracing/opentracing-go/log"
	"gonum.org/v1/gonum/stat/combin"

	"github.com/sourcegraph/sourcegraph/internal/search"
	alertobserver "github.com/sourcegraph/sourcegraph/internal/search/alert"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/limits"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewFeelingLuckySearchJob creates generators for opportunistic search queries
// that apply various rules, transforming the original input plan into various
// queries that alter its interpretation (e.g., search literally for quotes or
// not, attempt to search the pattern as a regexp, and so on). There is no
// random choice when applying rules.
func NewFeelingLuckySearchJob(initialJob job.Job, inputs *run.SearchInputs, plan query.Plan) *FeelingLuckySearchJob {
	generators := make([]next, 0, len(plan))
	for _, b := range plan {
		generators = append(generators, NewGenerator(inputs, b, rules))
	}

	jobGenerator := &jobGenerator{SearchInputs: inputs}

	return &FeelingLuckySearchJob{
		initialJob:      initialJob,
		generators:      generators,
		newGeneratedJob: jobGenerator.New,
	}
}

// FeelingLuckySearchJob represents a lucky search. Note `newGeneratedJob`
// returns a job given an autoQuery. It is a function so that generated queries
// can be composed at runtime (with auto queries that dictate runtime control
// flow) with static inputs (search inputs), while not exposing static inputs.
type FeelingLuckySearchJob struct {
	initialJob      job.Job
	generators      []next
	newGeneratedJob func(*autoQuery) job.Job
}

// jobGenerator stores static values that should not be exposed to runtime
// concerns. jobGenerator exposes a method `New` for constructing jobs that
// require runtime information.
type jobGenerator struct {
	*run.SearchInputs
}

func (g *jobGenerator) New(autoQ *autoQuery) job.Job {
	j, err := NewGeneratedSearchJob(g.SearchInputs, autoQ)
	if err != nil {
		return nil
	}
	return j
}

func (f *FeelingLuckySearchJob) Run(ctx context.Context, clients job.RuntimeClients, parentStream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, parentStream, finish := job.StartSpan(ctx, parentStream, f)
	defer func() { finish(alert, err) }()

	stream := streaming.NewDedupingStream(parentStream)

	var maxAlerter search.MaxAlerter
	var errs errors.MultiError
	alert, err = f.initialJob.Run(ctx, clients, stream)
	if err != nil {
		return alert, err
	}
	maxAlerter.Add(alert)

	// Now start counting how many additional results we get from generated queries
	countedStream := streaming.NewResultCountingStream(stream)

	generated := &alertobserver.ErrLuckyQueries{ProposedQueries: []*search.ProposedQuery{}}
	var autoQ *autoQuery
	for _, next := range f.generators {
		for {
			autoQ, next = next()
			if autoQ == nil {
				if next == nil {
					// No query and generator is exhausted.
					break
				}
				continue
			}

			j := f.newGeneratedJob(autoQ)
			if j == nil {
				// Generated an invalid job with this query, just continue.
				continue
			}
			alert, err = j.Run(ctx, clients, countedStream)
			if countedStream.Count() >= limits.DefaultMaxSearchResultsStreaming {
				// We've sent additional results up to the maximum bound. Let's stop here.
				var lErr *alertobserver.ErrLuckyQueries
				if errors.As(err, &lErr) {
					generated.ProposedQueries = append(generated.ProposedQueries, lErr.ProposedQueries...)
				}
				if len(generated.ProposedQueries) > 0 {
					errs = errors.Append(errs, generated)
				}
				return maxAlerter.Alert, errs
			}

			var lErr *alertobserver.ErrLuckyQueries
			if errors.As(err, &lErr) {
				// collected generated queries, we'll add it after this loop is done running.
				generated.ProposedQueries = append(generated.ProposedQueries, lErr.ProposedQueries...)
			} else {
				errs = errors.Append(errs, err)
			}

			maxAlerter.Add(alert)

			if next == nil {
				break
			}
		}
	}

	if len(generated.ProposedQueries) > 0 {
		errs = errors.Append(errs, generated)
	}
	return maxAlerter.Alert, errs
}

func (f *FeelingLuckySearchJob) Name() string {
	return "FeelingLuckySearchJob"
}

func (f *FeelingLuckySearchJob) Fields(job.Verbosity) []log.Field { return nil }

func (f *FeelingLuckySearchJob) Children() []job.Describer {
	return []job.Describer{f.initialJob}
}

func (f *FeelingLuckySearchJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *f
	cp.initialJob = job.Map(f.initialJob, fn)
	return &cp
}

// autoQuery is an automatically generated query with associated data (e.g., description).
type autoQuery struct {
	description string
	query       query.Basic
}

// next is the continuation for the query generator.
type next func() (*autoQuery, next)

// NewGenerator creates a new generator using rules and a seed Basic query. It
// returns a `next` function which, when called, returns the next job
// generated, and the `next` continuation. A `nil` continuation means query
// generation is exhausted. Currently it implements a simple strategy that tries
// to apply rule transforms in order, and generates jobs for transforms that
// apply successfully.
func NewGenerator(inputs *run.SearchInputs, seed query.Basic, rules []rule) next {
	var n func(i int) next // i keeps track of rule index in the continuation
	n = func(i int) next {
		if i >= len(rules) {
			return func() (*autoQuery, next) { return nil, nil }
		}

		return func() (*autoQuery, next) {
			generated := applyTransformation(seed, rules[i].transform)
			if generated == nil {
				// Rule doesn't apply, go to next rule.
				return nil, n(i + 1)
			}

			q := autoQuery{
				description: rules[i].description,
				query:       *generated,
			}

			return &q, n(i + 1)
		}
	}

	return n(0)
}

// rule represents a transformation function on a Basic query. Transformation
// cannot fail: either they apply in sequence and produce a valid, non-nil,
// Basic query, or they do not apply, in which case they return nil. See the
// `unquotePatterns` rule for an example.
type rule struct {
	description string
	transform   []func(query.Basic) *query.Basic
}

type transform []func(query.Basic) *query.Basic

// rules is an ordered list of rules. Each item represents one possible query
// production, if the sequence of the transformation functions associated with
// this item apply successfully. Example:
//
// If we have input query B0 and a list with two items like this:
//
// {
//   "first rule list"  : [ R1, R2 ],
//   "second rule list" : [ R2 ],
// }
//
// Then:
//
// - If both entries apply, we output [ B1, B2 ] where B1 is generated
//   from applying R1 then R2, and B2 is generated from just applying R2.
// - If only the first item applies, R1 then R2, we get the output [ B1 ].
// - If only the second item applies, R2 on its own, we get the output [ B2 ].
var rules = []rule{
	{
		description: "unquote patterns",
		transform:   transform{unquotePatterns},
	},
	{
		description: "apply search type and language filter for patterns",
		transform:   transform{typePatterns, langPatterns},
	},
	{
		description: "apply search type for pattern",
		transform:   transform{typePatterns},
	},
	{
		description: "apply language filter for pattern",
		transform:   transform{langPatterns},
	},
	{
		description: "apply search type and language filter for patterns with AND patterns",
		transform:   transform{typePatterns, langPatterns, unorderedPatterns},
	},
	{
		description: "apply search type with AND patterns",
		transform:   transform{typePatterns, unorderedPatterns},
	},
	{
		description: "apply language filter with AND patterns",
		transform:   transform{langPatterns, unorderedPatterns},
	},
	{
		description: "AND patterns together",
		transform:   transform{unorderedPatterns},
	},
}

// applyTransformation applies a transformation on `b`. If any function does not apply, it returns nil.
func applyTransformation(b query.Basic, transform transform) *query.Basic {
	for _, apply := range transform {
		res := apply(b)
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

	newNodes, err := query.Sequence(query.For(query.SearchTypeStandard))(newParseTree)
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
	rawParseTree, err := query.Parse(query.StringHuman(b.ToParseTree()), query.SearchTypeStandard)
	if err != nil {
		return nil
	}

	newParseTree, changed := mapConcat(rawParseTree)
	if !changed {
		return nil
	}

	newNodes, err := query.Sequence(query.For(query.SearchTypeStandard))(newParseTree)
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

func langPatterns(b query.Basic) *query.Basic {
	rawPatternTree, err := query.Parse(query.StringHuman([]query.Node{b.Pattern}), query.SearchTypeStandard)
	if err != nil {
		return nil
	}

	changed := false
	var lang string // store the first pattern that matches a recognized language.
	isNegated := false
	newPattern := query.MapPattern(rawPatternTree, func(value string, negated bool, annotation query.Annotation) query.Node {
		langAlias, ok := enry.GetLanguageByAlias(value)
		if !ok || changed {
			return query.Pattern{
				Value:      value,
				Negated:    negated,
				Annotation: annotation,
			}
		}
		changed = true
		lang = langAlias
		isNegated = negated
		// remove this node
		return nil
	})

	if !changed {
		return nil
	}

	langParam := query.Parameter{
		Field:      query.FieldLang,
		Value:      lang,
		Negated:    isNegated,
		Annotation: query.Annotation{},
	}

	var pattern query.Node
	if len(newPattern) > 0 {
		// Process concat nodes
		nodes, err := query.Sequence(query.For(query.SearchTypeStandard))(newPattern)
		if err != nil {
			return nil
		}
		pattern = nodes[0] // guaranteed root at first node
	}

	return &query.Basic{
		Parameters: append(b.Parameters, langParam),
		Pattern:    pattern,
	}
}

func typePatterns(b query.Basic) *query.Basic {
	rawPatternTree, err := query.Parse(query.StringHuman([]query.Node{b.Pattern}), query.SearchTypeStandard)
	if err != nil {
		return nil
	}

	changed := false
	var typ string // store the first pattern that matches a recognized `type:`.
	newPattern := query.MapPattern(rawPatternTree, func(value string, negated bool, annotation query.Annotation) query.Node {
		if changed {
			return query.Pattern{
				Value:      value,
				Negated:    negated,
				Annotation: annotation,
			}
		}

		switch strings.ToLower(value) {
		case "symbol", "commit", "diff", "path":
			typ = value
			changed = true
			// remove this node
			return nil
		}

		return query.Pattern{
			Value:      value,
			Negated:    negated,
			Annotation: annotation,
		}
	})

	if !changed {
		return nil
	}

	typParam := query.Parameter{
		Field:      query.FieldType,
		Value:      typ,
		Negated:    false,
		Annotation: query.Annotation{},
	}

	var pattern query.Node
	if len(newPattern) > 0 {
		// Process concat nodes
		nodes, err := query.Sequence(query.For(query.SearchTypeStandard))(newPattern)
		if err != nil {
			return nil
		}
		pattern = nodes[0] // guaranteed root at first node
	}

	return &query.Basic{
		Parameters: append(b.Parameters, typParam),
		Pattern:    pattern,
	}
}

func NewGeneratedSearchJob(inputs *run.SearchInputs, autoQ *autoQuery) (job.Job, error) {
	child, err := NewBasicJob(inputs, autoQ.query)
	if err != nil {
		return nil, err
	}

	notifier := &notifier{autoQuery: autoQ}

	return &generatedSearchJob{
		Child:           child,
		NewNotification: notifier.New,
	}, nil
}

// generatedSearchJob represents a generated search at run time. Note
// `NewNotification` returns the query notifications (encoded as error) given
// the result count of the job. It is a function so that notifications can be
// composed at runtime (with result counts) with static inputs (query string),
// while not exposing static inputs.
type generatedSearchJob struct {
	Child           job.Job
	NewNotification func(count int) error
}

// notifier stores static values that should not be exposed to runtime concerns.
// notifier exposes a method `New` for constructing notifications that require
// runtime information.
type notifier struct {
	*autoQuery
}

func (n *notifier) New(count int) error {
	var resultCountString string
	if count == limits.DefaultMaxSearchResultsStreaming {
		resultCountString = fmt.Sprintf("%d+ results", count)
	} else if count == 1 {
		resultCountString = fmt.Sprintf("1 result")
	} else {
		resultCountString = fmt.Sprintf("%d additional results", count)
	}

	return &alertobserver.ErrLuckyQueries{
		ProposedQueries: []*search.ProposedQuery{{
			Description: fmt.Sprintf("%s (%s)", n.description, resultCountString),
			Query:       query.StringHuman(n.query.ToParseTree()),
			PatternType: query.SearchTypeLucky,
		}},
	}
}

func (g *generatedSearchJob) Run(ctx context.Context, clients job.RuntimeClients, parentStream streaming.Sender) (*search.Alert, error) {
	stream := streaming.NewResultCountingStream(parentStream)
	alert, err := g.Child.Run(ctx, clients, stream)
	resultCount := stream.Count()
	if resultCount == 0 {
		return nil, nil
	}

	if ctx.Err() != nil {
		notification := g.NewNotification(resultCount)
		return alert, errors.Append(err, notification)
	}

	notification := g.NewNotification(resultCount)
	if err != nil {
		return alert, errors.Append(err, notification)
	}

	return alert, notification
}

func (g *generatedSearchJob) Name() string {
	return "GeneratedSearchJob"
}

func (g *generatedSearchJob) Children() []job.Describer { return []job.Describer{g.Child} }

func (g *generatedSearchJob) Fields(job.Verbosity) []log.Field { return nil }

func (g *generatedSearchJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *g
	cp.Child = job.Map(g.Child, fn)
	return &cp
}

var rulesMaxSet = []rule{
	{
		description: "unquote patterns",
		transform:   transform{unquotePatterns},
	},
	{
		description: "apply search type for pattern",
		transform:   transform{typePatterns},
	},
	{
		description: "apply language filter for pattern",
		transform:   transform{langPatterns},
	},
}

type cg = *combin.CombinationGenerator

// NewComboGenerator returns a generator for queries produces by a combination
// of rules on a seed query. The generator strategy tries to first apply _all_
// rules, and then successively reduces the number of rules that it attempts to
// apply by one. This strategy is useful when we try the most aggressive
// interpretation of a query subject to rules first, and gradually loosen the
// number of rules and interpretation. To avoid spending time on generator
// invalid combinations, the generator prunes the initial rule set to only those
// rules that do successively apply individually to the seed query.
func NewComboGenerator(seed query.Basic, rules []rule) next {
	rules = pruneRules(seed, rules)
	// the iterator state `n` stores:
	// - k, the number of rules to try and apply.
	// - cg, an iterator producing the next sequence of rules for the current value of `k`.
	var n func(k int, cg cg) next
	n = func(k int, cg cg) next {
		if k == 0 {
			return func() (*autoQuery, next) { return nil, nil }
		}

		return func() (*autoQuery, next) {
			if cg.Next() {
				var transform transform
				var descriptions []string
				for _, idx := range cg.Combination(nil) {
					transform = append(transform, rules[idx].transform...)
					descriptions = append(descriptions, rules[idx].description)
				}
				generated := applyTransformation(seed, transform)
				if generated == nil {
					// Rule does not apply, go to next rule.
					continuation := n(k, cg)
					return continuation()
				}

				q := autoQuery{
					description: strings.Join(descriptions, " and "),
					query:       *generated,
				}

				return &q, n(k, cg)
			}
			k -= 1
			cg = combin.NewCombinationGenerator(len(rules), k)
			continuation := n(k, cg)
			return continuation()
		}
	}

	num := len(rules)
	cg := combin.NewCombinationGenerator(num, num)
	return n(num, cg)
}

// pruneRules produces a minimum set of rules that apply successfully on the seed query.
func pruneRules(seed query.Basic, rules []rule) []rule {
	applies := make([]rule, 0, len(rules))
	for _, r := range rules {
		g := applyTransformation(seed, r.transform)
		if g == nil {
			continue
		}
		applies = append(applies, r)
	}
	return applies
}
