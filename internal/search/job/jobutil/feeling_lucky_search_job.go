package jobutil

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"regexp/syntax"
	"strings"

	"github.com/go-enry/go-enry/v2"
	"github.com/opentracing/opentracing-go/log"
	"gonum.org/v1/gonum/stat/combin"

	"github.com/sourcegraph/sourcegraph/internal/search"
	alertobserver "github.com/sourcegraph/sourcegraph/internal/search/alert"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/limits"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewFeelingLuckySearchJob creates generators for opportunistic search queries
// that apply various rules, transforming the original input plan into various
// queries that alter its interpretation (e.g., search literally for quotes or
// not, attempt to search the pattern as a regexp, and so on). There is no
// random choice when applying rules.
func NewFeelingLuckySearchJob(initialJob job.Job, newJob newJob, plan query.Plan) *FeelingLuckySearchJob {
	generators := make([]next, 0, len(plan))
	for _, b := range plan {
		generators = append(generators, NewGenerator(b, rulesNarrow, rulesWiden))
	}

	newGeneratedJob := func(autoQ *autoQuery) job.Job {
		child, err := newJob(autoQ.query)
		if err != nil {
			return nil
		}

		notifier := &notifier{autoQuery: autoQ}

		return &generatedSearchJob{
			Child:           child,
			NewNotification: notifier.New,
		}
	}

	return &FeelingLuckySearchJob{
		initialJob:      initialJob,
		generators:      generators,
		newGeneratedJob: newGeneratedJob,
	}
}

type newJob func(query.Basic) (job.Job, error)

// FeelingLuckySearchJob represents a lucky search. Note `newGeneratedJob`
// returns a job given an autoQuery. It is a function so that generated queries
// can be composed at runtime (with auto queries that dictate runtime control
// flow) with static inputs (search inputs), while not exposing static inputs.
type FeelingLuckySearchJob struct {
	initialJob      job.Job
	generators      []next
	newGeneratedJob func(*autoQuery) job.Job
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

// rule represents a transformation function on a Basic query. Transformation
// cannot fail: either they apply in sequence and produce a valid, non-nil,
// Basic query, or they do not apply, in which case they return nil. See the
// `unquotePatterns` rule for an example.
type rule struct {
	description string
	transform   []func(query.Basic) *query.Basic
}

type transform []func(query.Basic) *query.Basic

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

// regexpPatterns converts literal patterns into regular expression patterns.
// The conversion is a heuristic and happens based on whether the pattern has
// indicative regular expression metasyntax. It would be overly aggressive to
// convert patterns containing _any_ potential metasyntax, since a pattern like
// my.config.yaml contains two `.` (match any character in regexp).
func regexpPatterns(b query.Basic) *query.Basic {
	rawParseTree, err := query.Parse(query.StringHuman(b.ToParseTree()), query.SearchTypeStandard)
	if err != nil {
		return nil
	}

	// we decide to interpret patterns as regular expressions if the number of
	// significant metasyntax operators exceed this threshold
	METASYNTAX_THRESHOLD := 2

	// countMetaSyntax counts the number of significant regular expression
	// operators in string when it is interpreted as a regular expression. A
	// rough map of operators to syntax can be found here:
	// https://sourcegraph.com/github.com/golang/go@bf5898ef53d1693aa572da0da746c05e9a6f15c5/-/blob/src/regexp/syntax/regexp.go?L116-244
	var countMetaSyntax func([]*syntax.Regexp) int
	countMetaSyntax = func(res []*syntax.Regexp) int {
		count := 0
		for _, r := range res {
			switch r.Op {
			case
				// operators that are weighted 0 on their own
				syntax.OpAnyCharNotNL,
				syntax.OpAnyChar,
				syntax.OpNoMatch,
				syntax.OpEmptyMatch,
				syntax.OpLiteral,
				syntax.OpCapture,
				syntax.OpConcat:
				count += countMetaSyntax(r.Sub)
			case
				// operators that are weighted 1 on their own
				syntax.OpCharClass,
				syntax.OpBeginLine,
				syntax.OpEndLine,
				syntax.OpBeginText,
				syntax.OpEndText,
				syntax.OpWordBoundary,
				syntax.OpNoWordBoundary,
				syntax.OpAlternate:
				count += countMetaSyntax(r.Sub) + 1

			case
				// quantifiers *, +, ?, {...} on metasyntax like
				// `.` or `(...)` are weighted 2. If the
				// quantifier applies to other syntax like
				// literals (not metasyntax) it's weighted 1.
				syntax.OpStar,
				syntax.OpPlus,
				syntax.OpQuest,
				syntax.OpRepeat:
				switch r.Sub[0].Op {
				case
					syntax.OpAnyChar,
					syntax.OpAnyCharNotNL,
					syntax.OpCapture:
					count += countMetaSyntax(r.Sub) + 2
				default:
					count += countMetaSyntax(r.Sub) + 1
				}
			}
		}
		return count
	}

	changed := false
	newParseTree := query.MapPattern(rawParseTree, func(value string, negated bool, annotation query.Annotation) query.Node {
		if annotation.Labels.IsSet(query.Regexp) {
			return query.Pattern{
				Value:      value,
				Negated:    negated,
				Annotation: annotation,
			}
		}

		re, err := syntax.Parse(value, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
		if err != nil {
			return query.Pattern{
				Value:      value,
				Negated:    negated,
				Annotation: annotation,
			}
		}

		count := countMetaSyntax([]*syntax.Regexp{re})
		if count < METASYNTAX_THRESHOLD {
			return query.Pattern{
				Value:      value,
				Negated:    negated,
				Annotation: annotation,
			}
		}

		changed = true
		annotation.Labels.Unset(query.Literal)
		annotation.Labels.Set(query.Regexp)
		return query.Pattern{
			Value:      value,
			Negated:    negated,
			Annotation: annotation,
		}
	})

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

var lookup = map[string]struct{}{
	"github.com": {},
	"gitlab.com": {},
}

// patternToCodeHostFilters checks if a pattern contains a code host URL and
// extracts the org/repo/branch and path and lifts these to filters, as
// applicable.
func patternToCodeHostFilters(v string, negated bool) *[]query.Node {
	if !strings.HasPrefix(v, "https://") {
		// normalize v with https:// prefix.
		v = "https://" + v
	}

	u, err := url.Parse(v)
	if err != nil {
		return nil
	}

	domain := strings.TrimPrefix(u.Host, "www.")
	if _, ok := lookup[domain]; !ok {
		return nil
	}

	var value string
	path := strings.Trim(u.Path, "/")
	pathElems := strings.Split(path, "/")
	if len(pathElems) == 0 {
		value = regexp.QuoteMeta(domain)
		value = fmt.Sprintf("^%s", value)
		return &[]query.Node{
			query.Parameter{
				Field:      query.FieldRepo,
				Value:      value,
				Negated:    negated,
				Annotation: query.Annotation{},
			}}
	} else if len(pathElems) == 1 {
		value = regexp.QuoteMeta(domain)
		value = fmt.Sprintf("^%s/%s", value, strings.Join(pathElems, "/"))
		return &[]query.Node{
			query.Parameter{
				Field:      query.FieldRepo,
				Value:      value,
				Negated:    negated,
				Annotation: query.Annotation{},
			}}
	} else if len(pathElems) == 2 {
		value = regexp.QuoteMeta(domain)
		value = fmt.Sprintf("^%s/%s$", value, strings.Join(pathElems, "/"))
		return &[]query.Node{
			query.Parameter{
				Field:      query.FieldRepo,
				Value:      value,
				Negated:    negated,
				Annotation: query.Annotation{},
			}}
	} else if len(pathElems) == 4 && (pathElems[2] == "tree" || pathElems[2] == "commit") {
		repoValue := regexp.QuoteMeta(domain)
		repoValue = fmt.Sprintf("^%s/%s$", repoValue, strings.Join(pathElems[:2], "/"))
		revision := pathElems[3]
		return &[]query.Node{
			query.Parameter{
				Field:      query.FieldRepo,
				Value:      repoValue,
				Negated:    negated,
				Annotation: query.Annotation{},
			},
			query.Parameter{
				Field:      query.FieldRev,
				Value:      revision,
				Negated:    negated,
				Annotation: query.Annotation{},
			},
		}
	} else if len(pathElems) >= 5 {
		repoValue := regexp.QuoteMeta(domain)
		repoValue = fmt.Sprintf("^%s/%s$", repoValue, strings.Join(pathElems[:2], "/"))

		revision := pathElems[3]

		pathValue := strings.Join(pathElems[4:], "/")
		pathValue = regexp.QuoteMeta(pathValue)

		if pathElems[2] == "blob" {
			pathValue = fmt.Sprintf("^%s$", pathValue)
		} else if pathElems[2] == "tree" {
			pathValue = fmt.Sprintf("^%s", pathValue)
		} else {
			// We don't know what this is.
			return nil
		}

		return &[]query.Node{
			query.Parameter{
				Field:      query.FieldRepo,
				Value:      repoValue,
				Negated:    negated,
				Annotation: query.Annotation{},
			},
			query.Parameter{
				Field:      query.FieldRev,
				Value:      revision,
				Negated:    negated,
				Annotation: query.Annotation{},
			},
			query.Parameter{
				Field:      query.FieldFile,
				Value:      pathValue,
				Negated:    negated,
				Annotation: query.Annotation{},
			},
		}
	}

	return nil
}

// patternsToCodeHostFilters converts patterns to `repo` or `path` filters if they
// can be interpreted as URIs.
func patternsToCodeHostFilters(b query.Basic) *query.Basic {
	rawPatternTree, err := query.Parse(query.StringHuman([]query.Node{b.Pattern}), query.SearchTypeStandard)
	if err != nil {
		return nil
	}

	filterParams := []query.Node{}
	changed := false
	newParseTree := query.MapPattern(rawPatternTree, func(value string, negated bool, annotation query.Annotation) query.Node {
		if params := patternToCodeHostFilters(value, negated); params != nil {
			changed = true
			filterParams = append(filterParams, *params...)
			// Collect the param and delete pattern. We're going to
			// add those parameters after. We can't map patterns
			// in-place because that might create parameters in
			// concat nodes.
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

	newParseTree = query.NewOperator(append(newParseTree, filterParams...), query.And) // Reduce with NewOperator to obtain valid partitioning.
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

var rulesNarrow = []rule{
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
	{
		description: "expand URL to filters",
		transform:   transform{patternsToCodeHostFilters},
	},
}

var rulesWiden = []rule{
	{
		description: "patterns as regular expressions",
		transform:   transform{regexpPatterns},
	},
	{
		description: "AND patterns together",
		transform:   transform{unorderedPatterns},
	},
}

type cg = combin.CombinationGenerator

type PHASE int

const (
	ONE PHASE = iota + 1
	TWO
	THREE
)

// NewComboGenerator returns a generator for queries produces by a combination
// of rules on a seed query. The generator has a strategy over two kinds of rule
// sets: narrowing and widening rules. You can read more below, but if you don't
// care about this and just want to apply rules sequentially, simply pass in
// only `widen` rules and pass in an empty `narrow` rule set. This will mean
// your queries are just generated by successively applying rules in order of
// the `widen` rule set. To get more sophisticated generation behavior, read on.
//
// This generator understands two kinds of rules:
//
// - narrowing rules (roughly, rules that we expect make a query more specific, and reduces the result set size)
// - widening rules (roughly, rules that we expect make a query more general, and increases the result set size).
//
// A concrete example of a narrowing rule might be: `go parse` -> `lang:go
// parse`. This since we restrict the subset of files to search for `parse` to
// Go files only.
//
// A concrete example of a widening rule might be: `a b` -> `a OR b`. This since
// the `OR` expression is more general and will typically find more results than
// the string `a b`.
//
// The way the generator applies narrowing and widening rules has three phases,
// executed in order. The phases work like this:
//
// PHASE ONE: The generator strategy tries to first apply _all narrowing_ rules,
// and then successively reduces the number of rules that it attempts to apply
// by one. This strategy is useful when we try the most aggressive
// interpretation of a query subject to rules first, and gradually loosen the
// number of rules and interpretation. Roughly, PHASE ONE can be thought of as
// trying to maximize applying "for all" rules on the narrow rule set.
//
// PHASE TWO: The generator performs PHASE ONE generation, generating
// combinations of narrow rules, and then additionally _adds_ the first widening
// rule to each narrowing combination. It continues iterating along the list of
// widening rules, appending them to each narrowing combination until the
// iteration of widening rules is exhausted. Roughly, PHASE TWO can be thought
// of as trying to maximize applying "for all" rules in the narrow rule set
// while widening them by applying, in order, "there exists" rules in the widen
// rule set.
//
// PHASE THREE: The generator only applies widening rules in order without any
// narrowing rules. Roughly, PHASE THREE can be thought of as an ordered "there
// exists" application over widen rules.
//
// To avoid spending time on generator invalid combinations, the generator
// prunes the initial rule set to only those rules that do successively apply
// individually to the seed query.
func NewGenerator(seed query.Basic, narrow, widen []rule) next {
	narrow = pruneRules(seed, narrow)
	widen = pruneRules(seed, widen)
	num := len(narrow)

	// the iterator state `n` stores:
	// - phase, the current generation phase based on progress
	// - k, the size of the selection in the narrow set to apply
	// - cg, an iterator producing the next sequence of rules for the current value of `k`.
	// - w, the index of the widen rule to apply (-1 if empty)
	var n func(phase PHASE, k int, c *cg, w int) next
	n = func(phase PHASE, k int, c *cg, w int) next {
		var transform transform
		var descriptions []string
		var generated *query.Basic

		narrowing_exhausted := k == 0
		widening_active := w != -1
		widening_exhausted := widening_active && w == len(widen)

		switch phase {
		case THREE:
			if widening_exhausted {
				// Base case: we exhausted the set of narrow
				// rules (if any) and we've attempted every
				// widen rule with the sets of narrow rules.
				return func() (*autoQuery, next) { return nil, nil }
			}

			transform = append(transform, widen[w].transform...)
			descriptions = append(descriptions, widen[w].description)
			w += 1 // advance to next widening rule.

		case TWO:
			if widening_exhausted {
				// Start phase THREE: apply only widening rules.
				return n(THREE, 0, nil, 0)
			}

			if narrowing_exhausted && !widening_exhausted {
				// Continue widening: We've exhausted the sets of narrow
				// rules for the current widen rule, but we're not done
				// yet: there are still more widen rules to try. So
				// increment w by 1.
				c = combin.NewCombinationGenerator(num, num)
				w += 1 // advance to next widening rule.
				return n(phase, num, c, w)
			}

			if !c.Next() {
				// Reduce narrow set size.
				k -= 1
				c = combin.NewCombinationGenerator(num, k)
				return n(phase, k, c, w)
			}

			for _, idx := range c.Combination(nil) {
				transform = append(transform, narrow[idx].transform...)
				descriptions = append(descriptions, narrow[idx].description)
			}

			// Compose narrow rules with a widen rule.
			transform = append(transform, widen[w].transform...)
			descriptions = append(descriptions, widen[w].description)

		case ONE:
			if narrowing_exhausted && !widening_active {
				// Start phase TWO: apply widening with
				// narrowing rules. We've exhausted the sets of
				// narrow rules, but have not attempted to
				// compose them with any widen rules. Compose
				// them with widen rules by initializing w to 0.
				cg := combin.NewCombinationGenerator(num, num)
				return n(TWO, num, cg, 0)
			}

			if !c.Next() {
				// Reduce narrow set size.
				k -= 1
				c = combin.NewCombinationGenerator(num, k)
				return n(phase, k, c, w)
			}

			for _, idx := range c.Combination(nil) {
				transform = append(transform, narrow[idx].transform...)
				descriptions = append(descriptions, narrow[idx].description)
			}
		}

		generated = applyTransformation(seed, transform)
		if generated == nil {
			// Rule does not apply, go to next rule.
			return n(phase, k, c, w)
		}

		q := autoQuery{
			description: strings.Join(descriptions, " ⚬ "),
			query:       *generated,
		}

		return func() (*autoQuery, next) {
			return &q, n(phase, k, c, w)
		}
	}

	if len(narrow) == 0 {
		return n(THREE, 0, nil, 0)
	}

	cg := combin.NewCombinationGenerator(num, num)
	return n(ONE, num, cg, -1)
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
