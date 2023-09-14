package query

import (
	"fmt"
	"strings"

	"github.com/grafana/regexp"
	"github.com/grafana/regexp/syntax"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Predicate interface {
	// Field is the name of the field that the predicate applies to.
	// For example, with `repo:contains.file`, Field returns "repo".
	Field() string

	// Name is the name of the predicate.
	// For example, with `repo:contains.file`, Name returns "contains.file".
	Name() string

	// Unmarshal parses the contents of the predicate arguments
	// into the predicate object.
	Unmarshal(params string, negated bool) error
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
		"has.key":               func() Predicate { return &RepoHasKeyPredicate{} },
		"has.meta":              func() Predicate { return &RepoHasMetaPredicate{} },
		"has.topic":             func() Predicate { return &RepoHasTopicPredicate{} },

		// Deprecated predicates
		"contains": func() Predicate { return &RepoContainsPredicate{} },
	},
	FieldFile: {
		"contains.content": func() Predicate { return &FileContainsContentPredicate{} },
		"has.content":      func() Predicate { return &FileContainsContentPredicate{} },
		"has.owner":        func() Predicate { return &FileHasOwnerPredicate{} },
		"has.contributor":  func() Predicate { return &FileHasContributorPredicate{} },
	},
}

type NegatedPredicateError struct {
	name string
}

func (e *NegatedPredicateError) Error() string {
	return fmt.Sprintf("search predicate %q does not support negation", e.name)
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

func (EmptyPredicate) Field() string { return "" }
func (EmptyPredicate) Name() string  { return "" }
func (EmptyPredicate) Unmarshal(_ string, negated bool) error {
	if negated {
		return &NegatedPredicateError{"empty"}
	}

	return nil
}

// RepoContainsFilePredicate represents the `repo:contains.file()` predicate, which filters to
// repos that contain a path and/or content. NOTE: this predicate still supports the deprecated
// syntax `repo:contains.file(name.go)` on a best-effort basis.
type RepoContainsFilePredicate struct {
	Path    string
	Content string
	Negated bool
}

func (f *RepoContainsFilePredicate) Unmarshal(params string, negated bool) error {
	nodes, err := Parse(params, SearchTypeRegex)
	if err != nil {
		return err
	}

	if err := f.parseNodes(nodes); err != nil {
		// If there's a parsing error, try falling back to the deprecated syntax `repo:contains.file(name.go)`.
		// Only attempt to fall back if there is a single pattern node, to avoid being too lenient.
		if len(nodes) != 1 {
			return err
		}

		pattern, ok := nodes[0].(Pattern)
		if !ok {
			return err
		}

		if _, err := syntax.Parse(pattern.Value, syntax.Perl); err != nil {
			return err
		}
		f.Path = pattern.Value
	}

	if f.Path == "" && f.Content == "" {
		return errors.New("one of path or content must be set")
	}

	f.Negated = negated
	return nil
}

func (f *RepoContainsFilePredicate) parseNodes(nodes []Node) error {
	for _, node := range nodes {
		if err := f.parseNode(node); err != nil {
			return err
		}
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
			if _, err := syntax.Parse(v.Value, syntax.Perl); err != nil {
				return errors.Errorf("`contains.file` predicate has invalid `path` argument: %w", err)
			}
			f.Path = v.Value
		case "content":
			if f.Content != "" {
				return errors.New("cannot specify content multiple times")
			}
			if _, err := syntax.Parse(v.Value, syntax.Perl); err != nil {
				return errors.Errorf("`contains.file` predicate has invalid `content` argument: %w", err)
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

func (f *RepoContainsFilePredicate) Field() string { return FieldRepo }
func (f *RepoContainsFilePredicate) Name() string  { return "contains.file" }

/* repo:contains.content(pattern) */

type RepoContainsContentPredicate struct {
	Pattern string
	Negated bool
}

func (f *RepoContainsContentPredicate) Unmarshal(params string, negated bool) error {
	if _, err := syntax.Parse(params, syntax.Perl); err != nil {
		return errors.Errorf("contains.content argument: %w", err)
	}
	if params == "" {
		return errors.Errorf("contains.content argument should not be empty")
	}
	f.Pattern = params
	f.Negated = negated
	return nil
}

func (f *RepoContainsContentPredicate) Field() string { return FieldRepo }
func (f *RepoContainsContentPredicate) Name() string  { return "contains.content" }

/* repo:contains.path(pattern) */

type RepoContainsPathPredicate struct {
	Pattern string
	Negated bool
}

func (f *RepoContainsPathPredicate) Unmarshal(params string, negated bool) error {
	if _, err := syntax.Parse(params, syntax.Perl); err != nil {
		return errors.Errorf("contains.path argument: %w", err)
	}
	if params == "" {
		return errors.Errorf("contains.path argument should not be empty")
	}
	f.Pattern = params
	f.Negated = negated
	return nil
}

func (f *RepoContainsPathPredicate) Field() string { return FieldRepo }
func (f *RepoContainsPathPredicate) Name() string  { return "contains.path" }

/* repo:contains.commit.after(...) */

type RepoContainsCommitAfterPredicate struct {
	TimeRef string
	Negated bool
}

func (f *RepoContainsCommitAfterPredicate) Unmarshal(params string, negated bool) error {
	f.TimeRef = params
	f.Negated = negated
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

func (f *RepoHasDescriptionPredicate) Unmarshal(params string, negated bool) (err error) {
	if negated {
		return &NegatedPredicateError{f.Field() + ":" + f.Name()}
	}

	if _, err := syntax.Parse(params, syntax.Perl); err != nil {
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

// DEPRECATED: Use "repo:has.meta({tag}:)" instead
type RepoHasTagPredicate struct {
	Key     string
	Negated bool
}

func (f *RepoHasTagPredicate) Unmarshal(params string, negated bool) (err error) {
	if len(params) == 0 {
		return errors.New("tag must be non-empty")
	}
	f.Key = params
	f.Negated = negated
	return nil
}

func (f *RepoHasTagPredicate) Field() string { return FieldRepo }
func (f *RepoHasTagPredicate) Name() string  { return "has.tag" }

type RepoHasMetaPredicate struct {
	Key     string
	Value   *string
	Negated bool
	KeyOnly bool
}

func (p *RepoHasMetaPredicate) Unmarshal(params string, negated bool) (err error) {
	scanLiteral := func(data string) (string, int, error) {
		if strings.HasPrefix(data, `"`) {
			return ScanDelimited([]byte(data), true, '"')
		}
		if strings.HasPrefix(data, `'`) {
			return ScanDelimited([]byte(data), true, '\'')
		}
		loc := strings.Index(data, ":")
		if loc >= 0 {
			return data[:loc], loc, nil
		}
		return data, len(data), nil
	}

	// Trim leading and trailing spaces in params
	params = strings.Trim(params, " \t")

	// Scan the possibly-quoted key
	key, advance, err := scanLiteral(params)
	if err != nil {
		return err
	}

	if len(key) == 0 {
		return errors.New("key cannot be empty")
	}

	params = params[advance:]

	keyOnly := false
	var value *string = nil
	if strings.HasPrefix(params, ":") {
		// Chomp the leading ":"
		params = params[len(":"):]

		// Scan the possibly-quoted value
		val, advance, err := scanLiteral(params)
		if err != nil {
			return err
		}
		params = params[advance:]

		// If we have more text after scanning both the key and the value,
		// that means someone tried to use a quoted string with data outside
		// the quotes.
		if len(params) != 0 {
			return errors.New("unexpected extra content")
		}
		if len(val) > 0 {
			value = &val
		}
	} else {
		keyOnly = true
	}

	p.Key = key
	p.KeyOnly = keyOnly
	p.Value = value
	p.Negated = negated
	return nil
}

func (p *RepoHasMetaPredicate) Field() string { return FieldRepo }
func (p *RepoHasMetaPredicate) Name() string  { return "has.meta" }

// DEPRECATED: Use "repo:has.meta({key:value})" instead
type RepoHasKVPPredicate struct {
	Key     string
	Value   string
	Negated bool
}

func (p *RepoHasKVPPredicate) Unmarshal(params string, negated bool) (err error) {
	scanLiteral := func(data string) (string, int, error) {
		if strings.HasPrefix(data, `"`) {
			return ScanDelimited([]byte(data), true, '"')
		}
		if strings.HasPrefix(data, `'`) {
			return ScanDelimited([]byte(data), true, '\'')
		}
		loc := strings.Index(data, ":")
		if loc >= 0 {
			return data[:loc], loc, nil
		}
		return data, len(data), nil
	}
	// Trim leading and trailing spaces in params
	params = strings.Trim(params, " \t")
	// Scan the possibly-quoted key
	key, advance, err := scanLiteral(params)
	if err != nil {
		return err
	}
	params = params[advance:]

	// Chomp the leading ":"
	if !strings.HasPrefix(params, ":") {
		return errors.New("expected params of the form key:value")
	}
	params = params[len(":"):]

	// Scan the possibly-quoted value
	value, advance, err := scanLiteral(params)
	if err != nil {
		return err
	}
	params = params[advance:]

	// If we have more text after scanning both the key and the value,
	// that means someone tried to use a quoted string with data outside
	// the quotes.
	if len(params) != 0 {
		return errors.New("unexpected extra content")
	}

	if len(key) == 0 {
		return errors.New("key cannot be empty")
	}

	p.Key = key
	p.Value = value
	p.Negated = negated
	return nil
}

func (p *RepoHasKVPPredicate) Field() string { return FieldRepo }
func (p *RepoHasKVPPredicate) Name() string  { return "has" }

// DEPRECATED: Use "repo:has.meta({key})" instead
type RepoHasKeyPredicate struct {
	Key     string
	Negated bool
}

func (p *RepoHasKeyPredicate) Unmarshal(params string, negated bool) (err error) {
	if len(params) == 0 {
		return errors.New("key must be non-empty")
	}
	p.Key = params
	p.Negated = negated
	return nil
}

func (p *RepoHasKeyPredicate) Field() string { return FieldRepo }
func (p *RepoHasKeyPredicate) Name() string  { return "has.key" }

type RepoHasTopicPredicate struct {
	Topic   string
	Negated bool
}

func (p *RepoHasTopicPredicate) Unmarshal(params string, negated bool) (err error) {
	if len(params) == 0 {
		return errors.New("topic must be non-empty")
	}
	p.Topic = params
	p.Negated = negated
	return nil
}

func (p *RepoHasTopicPredicate) Field() string { return FieldRepo }
func (p *RepoHasTopicPredicate) Name() string  { return "has.topic" }

// RepoContainsPredicate represents the `repo:contains(file:a content:b)` predicate.
// DEPRECATED: this syntax is deprecated in favor of `repo:contains.file`.
type RepoContainsPredicate struct {
	File    string
	Content string
	Negated bool
}

func (f *RepoContainsPredicate) Unmarshal(params string, negated bool) error {
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
	f.Negated = negated
	return nil
}

func (f *RepoContainsPredicate) parseNode(n Node) error {
	switch v := n.(type) {
	case Parameter:
		if v.Negated {
			return errors.New("the repo:contains() predicate does not currently support negated values")
		}
		switch strings.ToLower(v.Field) {
		case "file":
			if f.File != "" {
				return errors.New("cannot specify file multiple times")
			}
			if _, err := regexp.Compile(v.Value); err != nil {
				return errors.Errorf("the repo:contains() predicate has invalid `file` argument: %w", err)
			}
			f.File = v.Value
		case "content":
			if f.Content != "" {
				return errors.New("cannot specify content multiple times")
			}
			if _, err := regexp.Compile(v.Value); err != nil {
				return errors.Errorf("the repo:contains() predicate has invalid `content` argument: %w", err)
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

/* file:contains.content(pattern) */

type FileContainsContentPredicate struct {
	Pattern string
}

func (f *FileContainsContentPredicate) Unmarshal(params string, negated bool) error {
	if negated {
		return &NegatedPredicateError{f.Field() + ":" + f.Name()}
	}

	if _, err := syntax.Parse(params, syntax.Perl); err != nil {
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
	Owner   string
	Negated bool
}

func (f *FileHasOwnerPredicate) Unmarshal(params string, negated bool) error {
	f.Owner = params
	f.Negated = negated
	return nil
}

func (f FileHasOwnerPredicate) Field() string { return FieldFile }
func (f FileHasOwnerPredicate) Name() string  { return "has.owner" }

/* file:has.contributor(pattern) */

type FileHasContributorPredicate struct {
	Contributor string
	Negated     bool
}

func (f *FileHasContributorPredicate) Unmarshal(params string, negated bool) error {
	if _, err := syntax.Parse(params, syntax.Perl); err != nil {
		return errors.Errorf("the file:has.contributor() predicate has invalid argument: %w", err)
	}

	f.Contributor = params
	f.Negated = negated
	return nil
}

func (f FileHasContributorPredicate) Field() string { return FieldFile }
func (f FileHasContributorPredicate) Name() string  { return "has.contributor" }
