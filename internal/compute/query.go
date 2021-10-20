package compute

import (
	"fmt"
	"regexp"

	"github.com/sourcegraph/sourcegraph/internal/search/query"

	"github.com/cockroachdb/errors"
)

type Query interface {
	String() string
	node()
}

func (MatchOnly) node()            {}
func (ReplaceInPlace) node()       {}
func (ReplaceWithSeparator) node() {}

type MatchOnly struct {
	MatchPattern MatchPattern
	Parameters   []query.Parameter
}

type ReplaceInPlace struct {
	MatchPattern   MatchPattern
	ReplacePattern string
	Parameters     []query.Parameter
}

type ReplaceWithSeparator struct {
	MatchPattern   MatchPattern
	ReplacePattern string
	Separator      string
	Parameters     []query.Parameter
}

func (n MatchOnly) String() string {
	return fmt.Sprintf("Match only: %s", n.MatchPattern.String())
}

func (n ReplaceInPlace) String() string {
	return fmt.Sprintf("Replace in place: %s -> %s", n.MatchPattern.String(), n.ReplacePattern)
}

func (n ReplaceWithSeparator) String() string {
	return fmt.Sprintf("Replace with separator: %s -> %s separator: %s", n.MatchPattern.String(), n.ReplacePattern, n.Separator)
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

func extractPattern(basic query.Basic) (string, error) {
	if basic.Pattern == nil {
		return "", errors.New("compute endpoint expects nonempty pattern")
	}
	var err error
	var pattern string
	seen := false
	query.VisitPattern([]query.Node{basic.Pattern}, func(value string, negated bool, _ query.Annotation) {
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
		pattern = value
		seen = true
	})
	if err != nil {
		return "", err
	}
	return pattern, nil
}

func toRegexpPattern(value string) (*regexp.Regexp, error) {
	rp, err := regexp.Compile(value)
	if err != nil {
		return nil, errors.Wrap(err, "regular expression is not valid for compute endpoint")
	}
	return rp, nil
}

func toComputeQuery(plan query.Plan) (Query, error) {
	if len(plan) != 1 {
		return nil, errors.New("compute endpoint only supports one search pattern currently ('and' or 'or' operators are not supported yet)")
	}
	pattern, err := extractPattern(plan[0])
	if err != nil {
		return nil, err
	}
	rp, err := toRegexpPattern(pattern)
	if err != nil {
		return nil, err
	}
	return &MatchOnly{MatchPattern: &Regexp{Value: rp}, Parameters: plan[0].Parameters}, nil
}

func Parse(q string) (Query, error) {
	plan, err := query.Pipeline(query.Init(q, query.SearchTypeRegex))
	if err != nil {
		return nil, err
	}
	return toComputeQuery(plan)
}
