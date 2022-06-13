package query

import "github.com/sourcegraph/sourcegraph/lib/errors"

/*
Query processing involves multiple steps to produce a query to evaluate.

To unify multiple concerns, query processing is abstracted to a sequence of
steps that entail parsing, validity checking, transformation, and conditional
processing logic driven by external options.
*/

// A step performs a transformation on nodes, which may fail.
type step func([]Node) ([]Node, error)

// A pass is a step that never fails.
type pass func([]Node) []Node

// Sequence sequences zero or more steps to create a single step.
func Sequence(steps ...step) step {
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

func identity(nodes []Node) ([]Node, error) {
	return nodes, nil
}

// With returns step if enabled is true. Use it to compose a pipeline that
// conditionally run steps.
func With(enabled bool, step step) step {
	if !enabled {
		return identity
	}
	return step
}

// SubstituteSearchContexts substitutes terms of the form `context:contextValue`
// for entire queries, like (repo:foo or repo:bar or repo:baz). It relies on a
// lookup function, which should return the query string for some
// `contextValue`.
func SubstituteSearchContexts(lookupQueryString func(contextValue string) (string, error)) step {
	return func(nodes []Node) ([]Node, error) {
		var errs error
		substitutedContext := MapField(nodes, FieldContext, func(value string, negated bool, ann Annotation) Node {
			queryString, err := lookupQueryString(value)
			if err != nil {
				errs = errors.Append(errs, err)
				return nil
			}

			if queryString == "" {
				return Parameter{
					Value:      value,
					Field:      FieldContext,
					Negated:    negated,
					Annotation: ann,
				}
			}

			query, err := ParseRegexp(queryString)
			if err != nil {
				errs = errors.Append(errs, err)
				return nil
			}
			return Operator{Kind: And, Operands: query}
		})

		return substitutedContext, errs
	}
}

// For runs processing steps for a given search type. This includes
// normalization, substitution for whitespace, and pattern labeling.
func For(searchType SearchType) step {
	var processType step
	switch searchType {
	case SearchTypeLiteralDefault:
		processType = succeeds(substituteConcat(space))
	case SearchTypeRegex:
		processType = succeeds(escapeParensHeuristic, substituteConcat(fuzzyRegexp))
	case SearchTypeStructural:
		processType = succeeds(labelStructural, ellipsesForHoles, substituteConcat(space))
	case SearchTypeLucky:
		processType = succeeds(substituteConcat(space))
	}
	normalize := succeeds(LowercaseFieldNames, SubstituteAliases(searchType), SubstituteCountAll)
	return Sequence(normalize, processType)
}

// Init creates a step from an input string and search type. It parses the
// initial input string.
func Init(in string, searchType SearchType) step {
	parser := func([]Node) ([]Node, error) {
		return Parse(in, searchType)
	}
	return Sequence(parser, For(searchType))
}

// InitLiteral is Init where SearchType is Literal.
func InitLiteral(in string) step {
	return Init(in, SearchTypeLiteralDefault)
}

// InitRegexp is Init where SearchType is Regex.
func InitRegexp(in string) step {
	return Init(in, SearchTypeRegex)
}

// InitStructural is Init where SearchType is Structural.
func InitStructural(in string) step {
	return Init(in, SearchTypeStructural)
}

func Run(step step) ([]Node, error) {
	return step(nil)
}

func ValidatePlan(plan Plan) error {
	for _, basic := range plan {
		if err := validate(basic.ToParseTree()); err != nil {
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

// Pipeline processes zero or more steps to produce a query. The first step must
// be Init, otherwise this function is a no-op.
func Pipeline(steps ...step) (Plan, error) {
	nodes, err := Sequence(steps...)(nil)
	if err != nil {
		return nil, err
	}

	plan := BuildPlan(nodes)
	if err := ValidatePlan(plan); err != nil {
		return nil, err
	}
	plan = MapPlan(plan, ConcatRevFilters)
	return plan, nil
}
