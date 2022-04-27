package query

import (
	"fmt"
	"strconv"
	"time"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/search/limits"
)

type ExpectedOperand struct {
	Msg string
}

func (e *ExpectedOperand) Error() string {
	return e.Msg
}

type UnsupportedError struct {
	Msg string
}

func (e *UnsupportedError) Error() string {
	return e.Msg
}

type SearchType int

const (
	SearchTypeRegex SearchType = iota
	SearchTypeLiteral
	SearchTypeStructural
)

func (s SearchType) String() string {
	switch s {
	case SearchTypeRegex:
		return "regex"
	case SearchTypeLiteral:
		return "literal"
	case SearchTypeStructural:
		return "structural"
	default:
		return fmt.Sprintf("unknown{%d}", s)
	}
}

// A query plan represents a set of disjoint queries for the search engine to
// execute. The result of executing a plan is the union of individual query results.
type Plan []Basic

// ToParseTree models a plan as a parse tree of an Or-expression on plan queries.
func (p Plan) ToParseTree() Q {
	nodes := make([]Node, 0, len(p))
	for _, basic := range p {
		operands := basic.ToParseTree()
		nodes = append(nodes, newOperator(operands, And)...)
	}
	return Q(newOperator(nodes, Or))
}

// Basic represents a leaf expression to evaluate in our search engine. A basic
// query comprises:
//   (1) a single search pattern expression, which may contain
//       'and' or 'or' operators; and
//   (2) parameters that scope the evaluation of search
//       patterns (e.g., to repos, files, etc.).
type Basic struct {
	Pattern    Node
	Parameters []Parameter
}

func (b Basic) ToParseTree() Q {
	var nodes []Node
	for _, n := range b.Parameters {
		nodes = append(nodes, Node(n))
	}
	if b.Pattern == nil {
		return nodes
	}
	nodes = append(nodes, b.Pattern)
	if hoisted, err := Hoist(nodes); err == nil {
		return hoisted
	}
	return nodes
}

// MapPattern returns a copy of a basic query with updated pattern.
func (b Basic) MapPattern(pattern Node) Basic {
	return Basic{Parameters: b.Parameters, Pattern: pattern}
}

// MapParameters returns a copy of a basic query with updated parameters.
func (b Basic) MapParameters(parameters []Parameter) Basic {
	return Basic{Parameters: parameters, Pattern: b.Pattern}
}

// Count returns the string value of the "count:" field. Returns empty string if none.
func (b Basic) Count() (count *int) {
	VisitField(ToNodes(b.Parameters), FieldCount, func(value string, _ bool, _ Annotation) {
		c, err := strconv.Atoi(value)
		if err != nil {
			panic(fmt.Sprintf("Value %q for count cannot be parsed as an int", value))
		}
		count = &c
	})
	return count
}

// GetTimeout returns the time.Duration value from the `timeout:` field.
func (b Basic) GetTimeout() *time.Duration {
	var timeout *time.Duration
	VisitField(ToNodes(b.Parameters), FieldTimeout, func(value string, _ bool, _ Annotation) {
		t, err := time.ParseDuration(value)
		if err != nil {
			panic(fmt.Sprintf("Value %q for timeout cannot be parsed as an duration: %s", value, err))
		}
		timeout = &t
	})
	return timeout
}

// MapCount returns a copy of a basic query with a count parameter set.
func (b Basic) MapCount(count int) Basic {
	parameters := MapParameter(ToNodes(b.Parameters), func(field, value string, negated bool, annotation Annotation) Node {
		if field == "count" {
			value = strconv.FormatInt(int64(count), 10)
		}
		return Parameter{Field: field, Value: value, Negated: negated, Annotation: annotation}
	})
	return Basic{Parameters: toParameters(parameters), Pattern: b.Pattern}
}

func (b Basic) String() string {
	return fmt.Sprintf("%s %s", Q(ToNodes(b.Parameters)).String(), Q([]Node{b.Pattern}).String())
}

func (b Basic) StringHuman() string {
	return fmt.Sprintf("%s %s", StringHuman(ToNodes(b.Parameters)), StringHuman([]Node{b.Pattern}))
}

func (b Basic) VisitParameter(field string, f func(value string, negated bool, annotation Annotation)) {
	for _, p := range b.Parameters {
		if p.Field == field {
			f(p.Value, p.Negated, p.Annotation)
		}
	}
}

// HasPatternLabel returns whether a pattern atom has a specified label.
func (b Basic) HasPatternLabel(label labels) bool {
	if b.Pattern == nil {
		return false
	}
	if _, ok := b.Pattern.(Pattern); !ok {
		// Basic query is not atomic.
		return false
	}
	annot := b.Pattern.(Pattern).Annotation
	return annot.Labels.IsSet(label)
}

func (b Basic) IsLiteral() bool {
	return b.HasPatternLabel(Literal)
}

func (b Basic) IsRegexp() bool {
	return b.HasPatternLabel(Regexp)
}

func (b Basic) IsStructural() bool {
	return b.HasPatternLabel(Structural)
}

// FindParameter calls f on parameters matching field in b.
func (b Basic) FindParameter(field string, f func(value string, negated bool, annotation Annotation)) {
	for _, p := range b.Parameters {
		if p.Field == field {
			f(p.Value, p.Negated, p.Annotation)
			break
		}
	}
}

// FindValue returns the first value of a parameter matching field in b. It
// doesn't inspect whether the field is negated.
func (b Basic) FindValue(field string) (value string) {
	var found string
	b.FindParameter(field, func(v string, _ bool, _ Annotation) {
		found = v
	})
	return found
}

func (b Basic) IsCaseSensitive() bool {
	return Q(ToNodes(b.Parameters)).IsCaseSensitive()
}

func (b Basic) Index() YesNoOnly {
	v := Q(ToNodes(b.Parameters)).yesNoOnlyValue(FieldIndex)
	if v == nil {
		return Yes
	}
	return *v
}

func (b Basic) Fork() *YesNoOnly {
	return Q(ToNodes(b.Parameters)).yesNoOnlyValue(FieldFork)
}

func (b Basic) Archived() *YesNoOnly {
	return Q(ToNodes(b.Parameters)).yesNoOnlyValue(FieldArchived)
}

func (b Basic) Repositories() (repos, excludeRepos []string) {
	return Q(ToNodes(b.Parameters)).Repositories()
}

func (b Basic) Visibility() RepoVisibility {
	visibilityStr := b.FindValue(FieldVisibility)
	return ParseVisibility(visibilityStr)
}

// PatternString returns the simple string pattern of a basic query. It assumes
// there is only on pattern atom.
func (b Basic) PatternString() string {
	if p, ok := b.Pattern.(Pattern); ok {
		if b.IsLiteral() {
			// Escape regexp meta characters if this pattern should be treated literally.
			return regexp.QuoteMeta(p.Value)
		} else {
			return p.Value
		}
	}
	return ""
}

func (b Basic) IsEmptyPattern() bool {
	if b.Pattern == nil {
		return true
	}
	if p, ok := b.Pattern.(Pattern); ok {
		return p.Value == ""
	}
	return false
}

// IncludeExcludeValues partitions multiple values of a field into positive
// (include) and negated (exclude) values.
func (b Basic) IncludeExcludeValues(field string) (include, exclude []string) {
	b.VisitParameter(field, func(v string, negated bool, _ Annotation) {
		if negated {
			exclude = append(exclude, v)
		} else {
			include = append(include, v)
		}
	})
	return include, exclude
}

// Exists returns whether a parameter exists in the query (whether negated or not).
func (b Basic) Exists(field string) bool {
	found := false
	b.VisitParameter(field, func(_ string, _ bool, _ Annotation) {
		found = true
	})
	return found
}

func (b Basic) Dependencies() (dependencies []string) {
	VisitPredicate(b.ToParseTree(), func(field, name, value string) {
		if field == FieldRepo && (name == "dependencies" || name == "deps") {
			dependencies = append(dependencies, value)
		}
	})
	return dependencies
}

func (b Basic) MaxResults(defaultLimit int) int {
	if count := b.Count(); count != nil {
		return *count
	}

	if defaultLimit != 0 {
		return defaultLimit
	}

	return limits.DefaultMaxSearchResults
}

// A query is a tree of Nodes. We choose the type name Q so that external uses like query.Q do not stutter.
type Q []Node

func (q Q) String() string {
	return toString(q)
}

func (q Q) RegexpPatterns(field string) (values, negatedValues []string) {
	VisitField(q, field, func(visitedValue string, negated bool, _ Annotation) {
		if negated {
			negatedValues = append(negatedValues, visitedValue)
		} else {
			values = append(values, visitedValue)
		}
	})
	return values, negatedValues
}

func (q Q) StringValues(field string) (values, negatedValues []string) {
	VisitField(q, field, func(visitedValue string, negated bool, _ Annotation) {
		if negated {
			negatedValues = append(negatedValues, visitedValue)
		} else {
			values = append(values, visitedValue)
		}
	})
	return values, negatedValues
}

func (q Q) StringValue(field string) (value, negatedValue string) {
	VisitField(q, field, func(visitedValue string, negated bool, _ Annotation) {
		if negated {
			negatedValue = visitedValue
		} else {
			value = visitedValue
		}
	})
	return value, negatedValue
}

func (q Q) Index() YesNoOnly {
	v := q.yesNoOnlyValue(FieldIndex)
	if v == nil {
		return Yes
	}
	return *v
}

func (q Q) Exists(field string) bool {
	found := false
	VisitField(q, field, func(_ string, _ bool, _ Annotation) {
		found = true
	})
	return found
}

func (q Q) BoolValue(field string) bool {
	result := false
	VisitField(q, field, func(value string, _ bool, _ Annotation) {
		result, _ = parseBool(value) // err was checked during parsing and validation.
	})
	return result
}

func (q Q) Count() *int {
	var count *int
	VisitField(q, FieldCount, func(value string, _ bool, _ Annotation) {
		c, err := strconv.Atoi(value)
		if err != nil {
			panic(fmt.Sprintf("Value %q for count cannot be parsed as an int: %s", value, err))
		}
		count = &c
	})
	return count
}

func (q Q) Archived() *YesNoOnly {
	return q.yesNoOnlyValue(FieldArchived)
}

func (q Q) Fork() *YesNoOnly {
	return q.yesNoOnlyValue(FieldFork)
}

func (q Q) yesNoOnlyValue(field string) *YesNoOnly {
	var res *YesNoOnly
	VisitField(q, field, func(value string, _ bool, _ Annotation) {
		yno := ParseYesNoOnly(value)
		if yno == Invalid {
			panic(fmt.Sprintf("Invalid value %q for field %q", value, field))
		}
		res = &yno
	})
	return res
}

func (q Q) Timeout() *time.Duration {
	var timeout *time.Duration
	VisitField(q, FieldTimeout, func(value string, _ bool, _ Annotation) {
		t, err := time.ParseDuration(value)
		if err != nil {
			panic(fmt.Sprintf("Value %q for timeout cannot be parsed as an duration: %s", value, err))
		}
		timeout = &t
	})
	return timeout
}

func (q Q) IsCaseSensitive() bool {
	return q.BoolValue("case")
}

func (q Q) Repositories() (repos []string, negatedRepos []string) {
	VisitField(q, FieldRepo, func(value string, negated bool, a Annotation) {
		if a.Labels.IsSet(IsPredicate) {
			return
		}

		if negated {
			negatedRepos = append(negatedRepos, value)
		} else {
			repos = append(repos, value)
		}
	})
	return repos, negatedRepos
}

func (q Q) Dependencies() (dependencies []string) {
	VisitPredicate(q, func(field, name, value string) {
		if field == FieldRepo && (name == "dependencies" || name == "deps") {
			dependencies = append(dependencies, value)
		}
	})
	return dependencies
}

func (q Q) MaxResults(defaultLimit int) int {
	if q == nil {
		return 0
	}

	if count := q.Count(); count != nil {
		return *count
	}

	if defaultLimit != 0 {
		return defaultLimit
	}

	return limits.DefaultMaxSearchResults
}
