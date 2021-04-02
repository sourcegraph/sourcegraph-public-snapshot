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
		"contains": func() Predicate {
			return &RepoContainsPredicate{}
		},
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
	predicateRegexp = regexp.MustCompile(`^(?P<name>[a-z]+)\((?P<params>.*)\)$`)
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
		switch v := node.(type) {
		case Parameter:
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
			if f.Content != "" {
				return errors.New("cannot specify content multiple times")
			}
			f.Content = v.Value
		default:
			return fmt.Errorf("unsupported node type %T", node)
		}
	}

	if f.File == "" && f.Content == "" {
		return errors.New("one of file or content must be set")
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
			Value: f.Content,
		})
	}

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
