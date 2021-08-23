package query

import "github.com/cockroachdb/errors"

/*
Query processing involves multiple steps to produce a query to evaluate.

To unify multiple concerns, query processing is abstracted to a sequence of
steps that entail parsing, validity checking, transformation, and conditional
processing logic driven by external options.
*/

// A step performs a transformation on nodes, which may fail.
type step func(nodes []Node) ([]Node, error)

// A pass is a step that never fails.
type pass func(nodes []Node) []Node

// sequence sequences zero or more steps to create a single step.
func sequence(steps ...step) step {
	return func(nodes []Node) ([]Node, error) {
		var err error
		for _, step := range steps {
			nodes, err = step(nodes)
			if err != nil {
				return nil, err
			}
		}
		return nodes, nil
	}
}

// succeeds converts a sequence of passes into a single step.
func succeeds(passes ...pass) step {
	return func(nodes []Node) ([]Node, error) {
		for _, pass := range passes {
			nodes = pass(nodes)
		}
		return nodes, nil
	}
}

// With returns step if enabled is true. Use it to compose a pipeline that
// conditionally run steps.
func With(enabled bool, step step) step {
	if !enabled {
		return identity
	}
	return step
}

// normalize is a normalization step that applies to all search queries. It:
// (1) ensures fields are lowercase so we can reference them consistently internally.
// (2) labels any `content:` values so that its interpreted according to `searchType`.
// (3) substitutes `count:all` for a numeric value.
func normalize(searchType SearchType) step {
	return succeeds(LowercaseFieldNames, SubstituteAliases(searchType), SubstituteCountAll)
}

// For runs processing steps for a given search type. This includes
// normalization, substitution for whitespace, and pattern labeling.
func For(searchType SearchType) step {
	var processType step
	switch searchType {
	case SearchTypeLiteral:
		processType = succeeds(substituteConcat(space))
	case SearchTypeRegex:
		processType = succeeds(escapeParensHeuristic, substituteConcat(fuzzyRegexp))
	case SearchTypeStructural:
		processType = succeeds(labelStructural, ellipsesForHoles, substituteConcat(space))
	}
	return sequence(normalize(searchType), processType)
}

// Init creates a step from an input string and search type. It parses the
// initial input string.
func Init(in string, searchType SearchType) step {
	parser := func([]Node) ([]Node, error) {
		return Parse(in, searchType)
	}
	return sequence(parser, For(searchType))
}

// InitLiteral is Init where SearchType is Literal.
func InitLiteral(in string) step {
	return Init(in, SearchTypeLiteral)
}

// InitRegexp is Init where SearchType is Regex.
func InitRegexp(in string) step {
	return Init(in, SearchTypeRegex)
}

// InitStructural is Init where SearchType is Structural.
func InitStructural(in string) step {
	return Init(in, SearchTypeStructural)
}

// InitRegexpStrict is an initializer step for Compute queries.
// It differs from SearchType queries in:
// (1) validating that this is a regular expression search.
// (2) always treating regular expression case sensitively unless `case` is specified.
// (3) not interpolating `.*` (fuzzyRegexp) between regular expression patterns.
// (4) requiring strictly correct regular expression syntax (no parentheses heuristics).
func InitRegexpStrict(in string) step {
	checkPatternType := func([]Node) ([]Node, error) {
		literalTerms, err := Parse(in, SearchTypeLiteral)
		if err != nil {
			return nil, err
		}
		var visitErr error
		VisitField(literalTerms, FieldPatternType, func(value string, _ bool, _ Annotation) {
			if value == "structural" {
				visitErr = errors.New("patterntype:structural is not supported for this kind of query")
			}
		})
		return nil, visitErr
	}
	parse := func([]Node) ([]Node, error) {
		return Parse(in, SearchTypeRegex)
	}
	makeCaseSensitive := func(nodes []Node) ([]Node, error) {
		caseSet := exists(nodes, func(node Node) bool {
			if p, ok := node.(Parameter); ok && p.Field == FieldCase {
				return true
			}
			return false
		})
		if caseSet {
			return nodes, nil
		}
		return newOperator(append(nodes,
			Parameter{
				Field:      "case",
				Value:      "yes",
				Negated:    false,
				Annotation: Annotation{},
			}), And), nil
	}
	return sequence(checkPatternType, parse, makeCaseSensitive, normalize(SearchTypeRegex), succeeds(substituteConcat(space)))
}

func Run(step step) ([]Node, error) {
	return step(nil)
}

func Validate(disjuncts [][]Node) error {
	for _, disjunct := range disjuncts {
		if err := validate(disjunct); err != nil {
			return err
		}
	}
	return nil
}

// A BasicPass is a transformation on Basic queries.
type BasicPass func(Basic) Basic

// MapPlan applies a conversion to all Basic queries in a plan. It expects a
// valid plan. guarantee transformation succeeds.
func MapPlan(plan Plan, pass BasicPass) Plan {
	updated := make([]Basic, 0, len(plan))
	for _, query := range plan {
		updated = append(updated, pass(query))
	}
	return Plan(updated)
}

func ToPlan(disjuncts [][]Node) (Plan, error) {
	plan := make([]Basic, 0, len(disjuncts))
	for _, disjunct := range disjuncts {
		basic, err := ToBasicQuery(disjunct)
		if err != nil {
			return nil, err
		}
		plan = append(plan, basic)
	}
	return plan, nil
}

// Pipeline processes zero or more steps to produce a query. The first step must
// be Init, otherwise this function is a no-op.
func Pipeline(steps ...step) (Plan, error) {
	nodes, err := sequence(steps...)(nil)
	if err != nil {
		return nil, err
	}

	disjuncts := Dnf(nodes)
	if err := Validate(disjuncts); err != nil {
		return nil, err
	}

	plan, err := ToPlan(disjuncts)
	if err != nil {
		return nil, err
	}
	plan = MapPlan(plan, ConcatRevFilters)
	return plan, nil
}
