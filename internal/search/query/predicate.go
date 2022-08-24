package query

import (
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
		"contains.file":         func() Predicate { return &RepoContainsFilePredicate{} },
		"has.file":              func() Predicate { return &RepoContainsFilePredicate{} },
		"contains.path":         func() Predicate { return &RepoContainsPathPredicate{} },
		"has.path":              func() Predicate { return &RepoContainsPathPredicate{} },
		"contains.content":      func() Predicate { return &RepoContainsContentPredicate{} },
		"has.content":           func() Predicate { return &RepoContainsContentPredicate{} },
		"contains.commit.after": func() Predicate { return &RepoContainsCommitAfterPredicate{} },
		"has.commit.after":      func() Predicate { return &RepoContainsCommitAfterPredicate{} },
		"has.description":       func() Predicate { return &RepoHasDescriptionPredicate{} },
		"has.tag":               func() Predicate { return &RepoHasTagPredicate{} },
		"has":                   func() Predicate { return &RepoHasKVPPredicate{} },
	},
	FieldFile: {
		"contains.content": func() Predicate { return &FileContainsContentPredicate{} },
		"has.content":      func() Predicate { return &FileContainsContentPredicate{} },
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

// RepoContainsFilePredicate represents the `repo:contains.file()` predicate,
// which filters to repos that contain a path and/or content
type RepoContainsFilePredicate struct {
	Path    string
	Content string
}

func (f *RepoContainsFilePredicate) ParseParams(params string) error {
	nodes, err := Parse(params, SearchTypeRegex)
	if err != nil {
		return err
	}

	for _, node := range nodes {
		if err := f.parseNode(node); err != nil {
			return err
		}
	}

	if f.Path == "" && f.Content == "" {
		return errors.New("one of path or content must be set")
	}

	return nil
}

func (f *RepoContainsFilePredicate) parseNode(n Node) error {
	switch v := n.(type) {
	case Parameter:
		if v.Negated {
			return errors.New("predicates do not currently support negated values")
		}
		switch strings.ToLower(v.Field) {
		case "path":
			if f.Path != "" {
				return errors.New("cannot specify path multiple times")
			}
			if _, err := regexp.Compile(v.Value); err != nil {
				return errors.Errorf("`contains.file` predicate has invalid `path` argument: %w", err)
			}
			f.Path = v.Value
		case "content":
			if f.Content != "" {
				return errors.New("cannot specify content multiple times")
			}
			if _, err := regexp.Compile(v.Value); err != nil {
				return errors.Errorf("`contains.file` predicate has invalid `content` argument: %w", err)
			}
			f.Content = v.Value
		default:
			return errors.Errorf("unsupported option %q", v.Field)
		}
	case Pattern:
		return errors.Errorf(`prepend 'path:' or 'content:' to "%s" to search repositories containing path or content respectively.`, v.Value)
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

func (f *RepoContainsFilePredicate) Field() string { return FieldRepo }
func (f *RepoContainsFilePredicate) Name() string  { return "contains.file" }

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

/* repo:contains.path(pattern) */

type RepoContainsPathPredicate struct {
	Pattern string
}

func (f *RepoContainsPathPredicate) ParseParams(params string) error {
	if _, err := regexp.Compile(params); err != nil {
		return errors.Errorf("contains.path argument: %w", err)
	}
	if params == "" {
		return errors.Errorf("contains.path argument should not be empty")
	}
	f.Pattern = params
	return nil
}

func (f *RepoContainsPathPredicate) Field() string { return FieldRepo }
func (f *RepoContainsPathPredicate) Name() string  { return "contains.path" }

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

type RepoHasTagPredicate struct {
	Key string
}

func (f *RepoHasTagPredicate) ParseParams(params string) (err error) {
	if len(params) == 0 {
		return errors.New("tag must be non-empty")
	}
	f.Key = params
	return nil
}

func (f *RepoHasTagPredicate) Field() string { return FieldRepo }
func (f *RepoHasTagPredicate) Name() string  { return "has.tag" }

type RepoHasKVPPredicate struct {
	Key   string
	Value string
}

func (p *RepoHasKVPPredicate) ParseParams(params string) (err error) {
	split := strings.Split(params, ":")
	if len(split) != 2 || len(split[0]) == 0 || len(split[1]) == 0 {
		return errors.New("expected params in the form of key:value")
	}
	p.Key = split[0]
	p.Value = split[1]
	return nil
}

func (p *RepoHasKVPPredicate) Field() string { return FieldRepo }
func (p *RepoHasKVPPredicate) Name() string  { return "has" }

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
