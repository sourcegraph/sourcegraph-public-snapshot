package query

import (
	"regexp/syntax" //nolint:depguard
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/lib/errors"
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
}

var DefaultPredicateRegistry = PredicateRegistry{
	FieldRepo: {
		"contains":              func() Predicate { return &RepoContainsPredicate{} },
		"contains.file":         func() Predicate { return &RepoContainsFilePredicate{} },
		"contains.content":      func() Predicate { return &RepoContainsContentPredicate{} },
		"contains.commit.after": func() Predicate { return &RepoContainsCommitAfterPredicate{} },
		"dependencies":          func() Predicate { return &RepoDependenciesPredicate{} },
		"deps":                  func() Predicate { return &RepoDependenciesPredicate{} },
		"dependents":            func() Predicate { return &RepoDependentsPredicate{} },
		"revdeps":               func() Predicate { return &RepoDependentsPredicate{} },
		"has.description":       func() Predicate { return &RepoHasDescriptionPredicate{} },
	},
	FieldFile: {
		"contains.content": func() Predicate { return &FileContainsContentPredicate{} },
		"contains":         func() Predicate { return &FileContainsContentPredicate{} },
		"has.owner":        func() Predicate { return &FileHasOwnerPredicate{} },
	},
}

// PredicateTable is a lookup map of one or more predicate names that resolve to the Predicate type.
type PredicateTable map[string]func() Predicate

// PredicateRegistry is a lookup map of predicate tables associated with all fields.
type PredicateRegistry map[string]PredicateTable

// Get returns a predicate for the given field with the given name. It assumes
// it exists, and panics otherwise.
func (pr PredicateRegistry) Get(field, name string) Predicate {
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
	predicateRegexp = regexp.MustCompile(`^(?P<name>[a-z\.]+)\((?s:(?P<params>.*))\)$`)
	nameIndex       = predicateRegexp.SubexpIndex("name")
	paramsIndex     = predicateRegexp.SubexpIndex("params")
)

// ParsePredicate returns the name and value of syntax conforming to
// name(value). It assumes this syntax is already validated prior. If not, it
// panics.
func ParseAsPredicate(value string) (name, params string) {
	match := predicateRegexp.FindStringSubmatch(value)
	if match == nil {
		panic("Invariant broken: attempt to parse a predicate value " + value + " which appears to have not been properly validated")
	}
	name = match[nameIndex]
	params = match[paramsIndex]
	return name, params
}

// EmptyPredicate is a noop value that satisfies the Predicate interface.
type EmptyPredicate struct{}

func (EmptyPredicate) Field() string            { return "" }
func (EmptyPredicate) Name() string             { return "" }
func (EmptyPredicate) ParseParams(string) error { return nil }

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
			if _, err := regexp.Compile(v.Value); err != nil {
				return errors.Errorf("`contains` predicate has invalid `file` argument: %w", err)
			}
			f.File = v.Value
		case "content":
			if f.Content != "" {
				return errors.New("cannot specify content multiple times")
			}
			if _, err := regexp.Compile(v.Value); err != nil {
				return errors.Errorf("`contains` predicate has invalid `content` argument: %w", err)
			}
			f.Content = v.Value
		default:
			return errors.Errorf("unsupported option %q", v.Field)
		}
	case Pattern:
		return errors.Errorf(`prepend 'file:' or 'content:' to "%s" to search repositories containing files or content respectively.`, v.Value)
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
		return errors.Errorf("unsupported node type %T", n)
	}
	return nil
}

func (f *RepoContainsPredicate) Field() string { return FieldRepo }
func (f *RepoContainsPredicate) Name() string  { return "contains" }

/* repo:contains.content(pattern) */

type RepoContainsContentPredicate struct {
	Pattern string
}

func (f *RepoContainsContentPredicate) ParseParams(params string) error {
	if _, err := regexp.Compile(params); err != nil {
		return errors.Errorf("contains.content argument: %w", err)
	}
	if params == "" {
		return errors.Errorf("contains.content argument should not be empty")
	}
	f.Pattern = params
	return nil
}

func (f *RepoContainsContentPredicate) Field() string { return FieldRepo }
func (f *RepoContainsContentPredicate) Name() string  { return "contains.content" }

/* repo:contains.file(pattern) */

type RepoContainsFilePredicate struct {
	Pattern string
}

func (f *RepoContainsFilePredicate) ParseParams(params string) error {
	if _, err := regexp.Compile(params); err != nil {
		return errors.Errorf("contains.file argument: %w", err)
	}
	if params == "" {
		return errors.Errorf("contains.file argument should not be empty")
	}
	f.Pattern = params
	return nil
}

func (f *RepoContainsFilePredicate) Field() string { return FieldRepo }
func (f *RepoContainsFilePredicate) Name() string  { return "contains.file" }

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

// RepoDependenciesPredicate represents the `repo:dependencies(regex@rev)` predicate,
// which filters to repos that are dependencies of the repos matching the given of regex.
type RepoDependenciesPredicate struct {
	RepoRev    string
	Transitive bool
}

var emptyRepoDependencies = errors.New("no pattern to match a repository in repo:dependencies predicate parameter")

func (f *RepoDependenciesPredicate) ParseParams(params string) (err error) {
	for _, elem := range strings.Fields(params) {
		if trimmed := strings.TrimPrefix(elem, "transitive:"); trimmed != elem {
			f.Transitive = parseYesNoOnly(trimmed) == Yes
		} else {
			re := elem
			if n := strings.LastIndex(re, "@"); n > 0 {
				re = re[:n]
			}

			if re == "" {
				return emptyRepoDependencies
			}

			_, err = syntax.Parse(re, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
			if err != nil {
				return errors.Errorf("invalid repo:dependencies predicate parameter %q: %v", re, err)
			}

			f.RepoRev = elem
		}
	}

	if f.RepoRev == "" {
		return emptyRepoDependencies
	}

	return nil
}

func (f *RepoDependenciesPredicate) Field() string { return FieldRepo }
func (f *RepoDependenciesPredicate) Name() string  { return "dependencies" }

// RepoDependentsPredicate represents the `repo:dependents(regex@rev)`
// predicate, which filters to repos that depend on the repos matching the
// given of regex.
type RepoDependentsPredicate struct{}

func (f *RepoDependentsPredicate) ParseParams(params string) (err error) {
	re := params
	if n := strings.LastIndex(params, "@"); n > 0 {
		re = re[:n]
	}

	if re == "" {
		return errors.Errorf("empty repo:dependents predicate parameter %q", params)
	}

	_, err = syntax.Parse(re, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
	if err != nil {
		return errors.Errorf("invalid repo:dependents predicate parameter %q: %v", params, err)
	}

	return nil
}

func (f *RepoDependentsPredicate) Field() string { return FieldRepo }
func (f *RepoDependentsPredicate) Name() string  { return "dependents" }

/* repo:has.description(...) */

type RepoHasDescriptionPredicate struct {
	Pattern string
}

func (f *RepoHasDescriptionPredicate) ParseParams(params string) (err error) {
	if _, err := regexp.Compile(params); err != nil {
		return errors.Errorf("invalid repo:has.description() argument: %w", err)
	}
	if len(params) == 0 {
		return errors.New("empty repo:has.description() predicate parameter")
	}
	f.Pattern = params
	return nil
}

func (f *RepoHasDescriptionPredicate) Field() string { return FieldRepo }
func (f *RepoHasDescriptionPredicate) Name() string  { return "has.description" }

/* file:contains.content(pattern) */

type FileContainsContentPredicate struct {
	Pattern string
}

func (f *FileContainsContentPredicate) ParseParams(params string) error {
	if _, err := regexp.Compile(params); err != nil {
		return errors.Errorf("file:contains.content argument: %w", err)
	}
	if params == "" {
		return errors.Errorf("file:contains.content argument should not be empty")
	}
	f.Pattern = params
	return nil
}

func (f FileContainsContentPredicate) Field() string { return FieldFile }
func (f FileContainsContentPredicate) Name() string  { return "contains.content" }

/* file:has.owner(pattern) */

type FileHasOwnerPredicate struct {
	Owner string
}

func (f *FileHasOwnerPredicate) ParseParams(params string) error {
	if params == "" {
		return errors.Errorf("file:has.owner argument should not be empty")
	}
	f.Owner = params
	return nil
}

func (f FileHasOwnerPredicate) Field() string { return FieldFile }
func (f FileHasOwnerPredicate) Name() string  { return "has.owner" }

// nonPredicateRepos returns the repo nodes in a query that aren't predicates,
// respecting parameters that determine repo results.
func nonPredicateRepos(q Basic) []Node {
	var res []Node
	VisitParameter(q.ToParseTree(), func(field, value string, negated bool, ann Annotation) {
		if ann.Labels.IsSet(IsPredicate) {
			// Skip predicates
			return
		}
		switch field {
		case
			FieldRepo,
			FieldContext,
			FieldIndex,
			FieldFork,
			FieldArchived,
			FieldVisibility,
			FieldCase:
			res = append(res, Parameter{
				Field:      field,
				Value:      value,
				Negated:    negated,
				Annotation: ann,
			})
		}
	})
	return res
}
