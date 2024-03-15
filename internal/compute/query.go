package compute

import (
	"fmt"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Query struct {
	Command    Command
	Parameters []query.Node
}

func (q Query) String() string {
	if len(q.Parameters) == 0 {
		return fmt.Sprintf("Command: `%s`", q.Command.String())
	}
	return fmt.Sprintf("Command: `%s`, Parameters: `%s`",
		q.Command.String(),
		query.StringHuman(q.Parameters))
}

func (q Query) ToSearchQuery() (string, error) {
	pattern := q.Command.ToSearchPattern()
	expression := []query.Node{
		query.Operator{
			Kind:     query.And,
			Operands: append(q.Parameters, query.Pattern{Value: pattern}),
		},
	}
	return query.StringHuman(expression), nil
}

type MatchPattern interface {
	pattern()
	String() string
}

func (Regexp) pattern() {}
func (Comby) pattern()  {}

type Regexp struct {
	Value *regexp.Regexp
}

type Comby struct {
	Value string
}

func (p Regexp) String() string {
	return p.Value.String()
}

func (p Comby) String() string {
	return p.Value
}

func extractPattern(basic *query.Basic) (*query.Pattern, error) {
	if basic.Pattern == nil {
		return nil, errors.New("compute endpoint expects nonempty pattern")
	}
	var err error
	var pattern *query.Pattern
	seen := false
	query.VisitPattern([]query.Node{basic.Pattern}, func(value string, negated bool, annotation query.Annotation) {
		if err != nil {
			return
		}
		if negated {
			err = errors.New("compute endpoint expects a nonnegated pattern")
			return
		}
		if seen {
			err = errors.New("compute endpoint only supports one search pattern currently ('and' or 'or' operators are not supported yet)")
			return
		}
		pattern = &query.Pattern{Value: value, Annotation: annotation}
		seen = true
	})
	if err != nil {
		return nil, err
	}
	return pattern, nil
}

func toRegexpPattern(value string) (MatchPattern, error) {
	rp, err := regexp.Compile(value)
	if err != nil {
		return nil, errors.Wrap(err, "compute endpoint")
	}
	return &Regexp{Value: rp}, nil
}

var ComputePredicateRegistry = query.PredicateRegistry{
	query.FieldContent: {
		"replace":            func() query.Predicate { return query.EmptyPredicate{} },
		"replace.regexp":     func() query.Predicate { return query.EmptyPredicate{} },
		"replace.structural": func() query.Predicate { return query.EmptyPredicate{} },
		"output":             func() query.Predicate { return query.EmptyPredicate{} },
		"output.regexp":      func() query.Predicate { return query.EmptyPredicate{} },
		"output.structural":  func() query.Predicate { return query.EmptyPredicate{} },
		"output.extra":       func() query.Predicate { return query.EmptyPredicate{} },
	},
}

func parseContentPredicate(pattern *query.Pattern) (string, string, bool) {
	if !pattern.Annotation.Labels.IsSet(query.IsContent) {
		// pattern is not set via `content:`, so it cannot be a replace command.
		return "", "", false
	}
	value, _, ok := query.ScanPredicate("content", []byte(pattern.Value), ComputePredicateRegistry)
	if !ok {
		return "", "", false
	}
	name, args := query.ParseAsPredicate(value)
	return name, args, true
}

var arrowSyntax = lazyregexp.New(`\s*->\s*`)

func parseArrowSyntax(args string) (string, string, error) {
	parts := arrowSyntax.Split(args, 2)
	if len(parts) != 2 {
		return "", "", errors.New("invalid arrow statement, no left and right hand sides of `->`")
	}
	return parts[0], parts[1], nil
}

func parseReplace(q *query.Basic) (Command, bool, error) {
	pattern, err := extractPattern(q)
	if err != nil {
		return nil, false, err
	}

	name, args, ok := parseContentPredicate(pattern)
	if !ok {
		return nil, false, nil
	}
	left, right, err := parseArrowSyntax(args)
	if err != nil {
		return nil, false, err
	}

	var matchPattern MatchPattern
	switch name {
	case "replace", "replace.regexp":
		var err error
		matchPattern, err = toRegexpPattern(left)
		if err != nil {
			return nil, false, errors.Wrap(err, "replace command")
		}
	case "replace.structural":
		// structural search doesn't do any match pattern validation
		matchPattern = &Comby{Value: left}
	default:
		// unrecognized name
		return nil, false, nil
	}

	return &Replace{
		SearchPattern:  matchPattern,
		ReplacePattern: right,
	}, true, nil
}

func parseOutput(q *query.Basic) (Command, bool, error) {
	pattern, err := extractPattern(q)
	if err != nil {
		return nil, false, err
	}

	name, args, ok := parseContentPredicate(pattern)
	if !ok {
		return nil, false, nil
	}
	left, right, err := parseArrowSyntax(args)
	if err != nil {
		return nil, false, err
	}

	var matchPattern MatchPattern
	switch name {
	case "output", "output.regexp", "output.extra":
		var err error
		matchPattern, err = toRegexpPattern(left)
		if err != nil {
			return nil, false, errors.Wrap(err, "output command")
		}
	case "output.structural":
		// structural search doesn't do any match pattern validation
		matchPattern = &Comby{Value: left}

	default:
		// unrecognized name
		return nil, false, nil
	}

	var typeValue string
	query.VisitField(q.ToParseTree(), query.FieldType, func(value string, _ bool, _ query.Annotation) {
		typeValue = value
	})

	var selector string
	query.VisitField(q.ToParseTree(), query.FieldSelect, func(value string, _ bool, _ query.Annotation) {
		selector = value
	})

	// The default separator is newline and cannot be changed currently.
	return &Output{
		SearchPattern: matchPattern,
		OutputPattern: right,
		Separator:     "\n",
		TypeValue:     typeValue,
		Selector:      selector,
		Kind:          name,
	}, true, nil
}

func parseMatchOnly(q *query.Basic) (Command, bool, error) {
	pattern, err := extractPattern(q)
	if err != nil {
		return nil, false, err
	}

	sp, err := toRegexpPattern(pattern.Value)
	if err != nil {
		return nil, false, err
	}

	cp := sp
	if !q.IsCaseSensitive() {
		cp, err = toRegexpPattern("(?i:" + pattern.Value + ")")
		if err != nil {
			return nil, false, err
		}
	}

	return &MatchOnly{SearchPattern: sp, ComputePattern: cp}, true, nil
}

type commandParser func(pattern *query.Basic) (Command, bool, error)

// first returns the first parser that succeeds at parsing a command from a pattern.
func first(parsers ...commandParser) commandParser {
	return func(q *query.Basic) (Command, bool, error) {
		for _, parse := range parsers {
			command, ok, err := parse(q)
			if err != nil {
				return nil, false, err
			}
			if ok {
				return command, true, nil
			}
		}
		return nil, false, errors.Errorf("could not parse valid compute command from query %s", q)
	}
}

var parseCommand = first(
	parseReplace,
	parseOutput,
	parseMatchOnly,
)

func toComputeQuery(plan query.Plan) (*Query, error) {
	if len(plan) < 1 {
		return nil, errors.New("compute endpoint can't do anything with empty query")
	}

	command, _, err := parseCommand(&plan[0])
	if err != nil {
		return nil, err
	}

	parameters := query.MapPattern(plan.ToQ(), func(_ string, _ bool, _ query.Annotation) query.Node {
		// remove the pattern node.
		return nil
	})
	return &Query{
		Parameters: parameters,
		Command:    command,
	}, nil
}

func Parse(q string) (*Query, error) {
	parseTree, err := query.ParseRegexp(q)
	if err != nil {
		return nil, err
	}
	seenPatterns := 0
	query.VisitPattern(parseTree, func(_ string, _ bool, _ query.Annotation) {
		seenPatterns += 1
	})

	if seenPatterns > 1 {
		return nil, errors.New("compute endpoint cannot currently support expressions in patterns containing 'and', 'or', 'not' (or negation) right now!")
	}

	plan, err := query.Pipeline(query.InitRegexp(q))
	if err != nil {
		return nil, err
	}
	return toComputeQuery(plan)
}
