package repos

import (
	"strconv"
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/bytesize"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// excludeFunc takes either a generic object and returns true if the repo should be excluded. In
// the case of repo sourcing it will take a repository name, ID, or the repo itself as input.
type excludeFunc func(input any) bool

// excludeBuilder builds an excludeFunc.
type excludeBuilder struct {
	exact    map[string]struct{}
	patterns []*regexp.Regexp
	generic  []excludeFunc
	err      error
}

// Exact will case-insensitively exclude the string name.
func (e *excludeBuilder) Exact(name string) {
	if e.exact == nil {
		e.exact = map[string]struct{}{}
	}
	if name == "" {
		return
	}
	e.exact[strings.ToLower(name)] = struct{}{}
}

// Pattern will exclude strings matching the regex pattern.
func (e *excludeBuilder) Pattern(pattern string) {
	if pattern == "" {
		return
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		e.err = err
		return
	}
	e.patterns = append(e.patterns, re)
}

// Generic registers the passed in exclude function that will be used to determine whether a repo
// should be excluded.
func (e *excludeBuilder) Generic(ef excludeFunc) {
	if ef == nil {
		return
	}
	e.generic = append(e.generic, ef)
}

// Build will return an excludeFunc based on the previous calls to Exact, Pattern, and
// Generic.
func (e *excludeBuilder) Build() (excludeFunc, error) {
	return func(input any) bool {
		if inputString, ok := input.(string); ok {
			if _, ok := e.exact[strings.ToLower(inputString)]; ok {
				return true
			}

			for _, re := range e.patterns {
				if re.MatchString(inputString) {
					return true
				}
			}
		} else {
			for _, ef := range e.generic {
				if ef(input) {
					return true
				}
			}
		}

		return false
	}, e.err
}

func buildGitHubExcludeRule(rule *schema.ExcludedGitHubRepo) (excludeFunc, error) {
	var fns []gitHubExcludeFunc
	if rule.Stars != "" {
		fn, err := buildStarsConstraintsExcludeFn(rule.Stars)
		if err != nil {
			return nil, err
		}
		fns = append(fns, fn)
	}

	if rule.Size != "" {
		fn, err := buildSizeConstraintsExcludeFn(rule.Size)
		if err != nil {
			return nil, err
		}
		fns = append(fns, fn)
	}

	return func(repo any) bool {
		githubRepo, ok := repo.(github.Repository)
		if !ok {
			return false
		}

		// We're AND'ing the functions together. If one of them does NOT exclude
		// the repository, then we don't exclude it.
		for _, fn := range fns {
			excluded := fn(githubRepo)
			if !excluded {
				return false
			}
		}

		return true
	}, nil
}

type gitHubExcludeFunc func(github.Repository) bool

var starsConstraintRegex = regexp.MustCompile(`([<>=]{1,2})\s*(\d+)`)
var sizeConstraintRegex = regexp.MustCompile(`([<>=]{1,2})\s*(\d+\s*\w+)`)

func buildStarsConstraintsExcludeFn(constraint string) (gitHubExcludeFunc, error) {
	matches := starsConstraintRegex.FindStringSubmatch(constraint)
	if matches == nil {
		return nil, errors.Newf("invalid stars constraint format: %q", constraint)
	}

	operator, err := newOperator(matches[1])
	if err != nil {
		return nil, errors.Wrapf(err, "failed to evaluate stars constraint")
	}

	count, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, err
	}

	return func(r github.Repository) bool {
		return operator.Eval(r.StargazerCount, count)
	}, nil
}

func buildSizeConstraintsExcludeFn(constraint string) (gitHubExcludeFunc, error) {
	sizeMatch := sizeConstraintRegex.FindStringSubmatch(constraint)
	if sizeMatch == nil {
		return nil, errors.Newf("invalid size constraint format: %q", constraint)
	}

	operator, err := newOperator(sizeMatch[1])
	if err != nil {
		return nil, errors.Wrapf(err, "failed to evaluate size constraint")
	}

	size, err := bytesize.Parse(sizeMatch[2])
	if err != nil {
		return nil, err
	}

	return func(r github.Repository) bool {
		return operator.Eval(int(r.SizeBytes()), int(size))
	}, nil
}

type operator string

const (
	opLess           operator = "<"
	opLessOrEqual    operator = "<="
	opGreater        operator = ">"
	opGreaterOrEqual operator = ">="
)

func newOperator(input string) (operator, error) {
	if input != "<" && input != "<=" && input != ">" && input != ">=" {
		return "", errors.Newf("invalid operator %q", input)
	}
	return operator(input), nil
}

func (o operator) Eval(left, right int) bool {
	switch o {
	case "<":
		return left < right
	case ">":
		return left > right
	case "<=":
		return left <= right
	case ">=":
		return left >= right
	default:
		// I wish Go had enums
		panic(errors.Newf("unknown operator: %q", o))
	}
}
