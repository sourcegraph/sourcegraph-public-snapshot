package query

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type Predicate interface {
	// Field is the name of the field that the predicate applies to.
	// For example, with `file:contains()`, Field returns "file".
	Field() string

	// Name is the name of the predicate.
	// For example, with `file:contains()`, Name returns "contains".
	Name() string

	// ParseParams parses the contents of the predicate arguments
	// into the predicate object.
	ParseParams(string) error

	// Plan generates a plan of (possibly multiple) queries to execute the
	// behavior of a predicate in a query Q.
	Plan(parent Basic) (Plan, error)
}

var DefaultPredicateRegistry = predicateRegistry{
	FieldRepo: {
		"contains":              func() Predicate { return &RepoContainsPredicate{} },
		"contains.file":         func() Predicate { return &RepoContainsFilePredicate{} },
		"contains.content":      func() Predicate { return &RepoContainsContentPredicate{} },
		"contains.commit.after": func() Predicate { return &RepoContainsCommitAfterPredicate{} },
	},
}

type predicateRegistry map[string]map[string]func() Predicate

// Get returns a predicate for the given field with the given name. It assumes
// it exists, and panics otherwise.
func (pr predicateRegistry) Get(field, name string) Predicate {
	fieldPredicates, ok := pr[field]
	if !ok {
		panic("predicate lookup for " + field + " is invalid")
	}
	newPredicateFunc, ok := fieldPredicates[name]
	if !ok {
		panic("predicate lookup for " + name + " on " + field + " is invalid")
	}
	return newPredicateFunc()
}

var (
	predicateRegexp = regexp.MustCompile(`^(?P<name>[a-z\.]+)\((?P<params>.*)\)$`)
	nameIndex       = predicateRegexp.SubexpIndex("name")
	paramsIndex     = predicateRegexp.SubexpIndex("params")
)

// ParsePredicate returns the name and value of syntax conforming to
// name(value). It assumes this syntax is already validated prior. If not, it
// panics.
func ParseAsPredicate(value string) (name, params string) {
	match := predicateRegexp.FindStringSubmatch(value)
	if match == nil {
		panic("Invariant broken: attempt to parse a predicate value " + value + " that has not been validated")
	}
	name = match[nameIndex]
	params = match[paramsIndex]
	return name, params
}

// RepoContainsPredicate represents the `repo:contains()` predicate,
// which filters to repos that contain either a file or content
type RepoContainsPredicate struct {
	File    string
	Content string
}

func (f *RepoContainsPredicate) ParseParams(params string) error {
	nodes, err := Parse(params, SearchTypeRegex)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		if err := f.parseNode(node); err != nil {
			return err
		}
	}

	if f.File == "" && f.Content == "" {
		return errors.New("one of file or content must be set")
	}

	return nil
}

func (f *RepoContainsPredicate) parseNode(n Node) error {
	switch v := n.(type) {
	case Parameter:
		if v.Negated {
			return errors.New("predicates do not currently support negated values")
		}
		switch strings.ToLower(v.Field) {
		case "file":
			if f.File != "" {
				return errors.New("cannot specify file multiple times")
			}
			f.File = v.Value
		case "content":
			if f.Content != "" {
				return errors.New("cannot specify content multiple times")
			}
			f.Content = v.Value
		default:
			return fmt.Errorf("unsupported option %q", v.Field)
		}
	case Pattern:
		return fmt.Errorf(`prepend 'file:' or 'content:' to "%s" to search repositories containing files or content respectively.`, v.Value)
	case Operator:
		if v.Kind == Or {
			return errors.New("predicates do not currently support 'or' queries")
		}
		for _, operand := range v.Operands {
			if err := f.parseNode(operand); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unsupported node type %T", n)
	}
	return nil
}

func (f *RepoContainsPredicate) Field() string { return FieldRepo }
func (f *RepoContainsPredicate) Name() string  { return "contains" }
func (f *RepoContainsPredicate) Plan(parent Basic) (Plan, error) {
	nodes := make([]Node, 0, 3)
	nodes = append(nodes, Parameter{
		Field: FieldSelect,
		Value: "repo",
	}, Parameter{
		Field: FieldCount,
		Value: "99999",
	})

	if f.File != "" {
		nodes = append(nodes, Parameter{
			Field: FieldFile,
			Value: f.File,
		})
	}

	if f.Content != "" {
		nodes = append(nodes, Pattern{
			Value:      f.Content,
			Annotation: Annotation{Labels: Regexp},
		})
	}

	nodes = append(nodes, nonPredicateRepos(parent)...)
	return ToPlan(Dnf(nodes))
}

/* repo:contains.content(pattern) */

type RepoContainsContentPredicate struct {
	Pattern string
}

func (f *RepoContainsContentPredicate) ParseParams(params string) error {
	if _, err := regexp.Compile(params); err != nil {
		return fmt.Errorf("contains.content argument: %w", err)
	}
	if params == "" {
		return fmt.Errorf("contains.content argument should not be empty")
	}
	f.Pattern = params
	return nil
}

func (f *RepoContainsContentPredicate) Field() string { return FieldRepo }
func (f *RepoContainsContentPredicate) Name() string  { return "contains.content" }
func (f *RepoContainsContentPredicate) Plan(parent Basic) (Plan, error) {
	contains := RepoContainsPredicate{File: "", Content: f.Pattern}
	return contains.Plan(parent)
}

/* repo:contains.file(pattern) */

type RepoContainsFilePredicate struct {
	Pattern string
}

func (f *RepoContainsFilePredicate) ParseParams(params string) error {
	if _, err := regexp.Compile(params); err != nil {
		return fmt.Errorf("contains.file argument: %w", err)
	}
	if params == "" {
		return fmt.Errorf("contains.file argument should not be empty")
	}
	f.Pattern = params
	return nil
}

func (f *RepoContainsFilePredicate) Field() string { return FieldRepo }
func (f *RepoContainsFilePredicate) Name() string  { return "contains.file" }
func (f *RepoContainsFilePredicate) Plan(parent Basic) (Plan, error) {
	contains := RepoContainsPredicate{File: f.Pattern, Content: ""}
	return contains.Plan(parent)
}

/* repo:contains.commit.after(...) */

type RepoContainsCommitAfterPredicate struct {
	TimeRef string
}

func (f *RepoContainsCommitAfterPredicate) ParseParams(params string) error {
	f.TimeRef = params
	return nil
}

func (f RepoContainsCommitAfterPredicate) Field() string { return FieldRepo }
func (f RepoContainsCommitAfterPredicate) Name() string {
	return "contains.commit.after"
}
func (f *RepoContainsCommitAfterPredicate) Plan(parent Basic) (Plan, error) {
	nodes := make([]Node, 0, 3)
	nodes = append(nodes, Parameter{
		Field: FieldCount,
		Value: "99999",
	}, Parameter{
		Field: FieldRepoHasCommitAfter,
		Value: f.TimeRef,
	})

	nodes = append(nodes, nonPredicateRepos(parent)...)
	return ToPlan(Dnf(nodes))
}

// nonPredicateRepos returns the repo nodes in a query that aren't predicates
func nonPredicateRepos(q Basic) []Node {
	var res []Node
	VisitField(q.ToParseTree(), FieldRepo, func(value string, negated bool, ann Annotation) {
		if ann.Labels.IsSet(IsPredicate) {
			// Skip predicates
			return
		}

		res = append(res, Parameter{
			Field:      FieldRepo,
			Value:      value,
			Negated:    negated,
			Annotation: ann,
		})
	})
	return res
}
