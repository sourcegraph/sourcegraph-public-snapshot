package query

import (
	"fmt"
	"strconv"
	"strings"
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
	SearchTypeLucky
	SearchTypeStandard
	SearchTypeKeyword
)

func (s SearchType) String() string {
	switch s {
	case SearchTypeStandard:
		return "standard"
	case SearchTypeRegex:
		return "regex"
	case SearchTypeLiteral:
		return "literal"
	case SearchTypeStructural:
		return "structural"
	case SearchTypeLucky:
		return "lucky"
	case SearchTypeKeyword:
		return "keyword"
	default:
		return fmt.Sprintf("unknown{%d}", s)
	}
}

// A query is a tree of Nodes. We choose the type name Q so that external uses like query.Q do not stutter.
type Q []Node

func (q Q) String() string {
	return toString(q)
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
		yno := parseYesNoOnly(value)
		if yno == Invalid {
			panic(fmt.Sprintf("Invalid value %q for field %q", value, field))
		}
		res = &yno
	})
	return res
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

func (q Q) Dependents() (dependents []string) {
	VisitPredicate(q, func(field, name, value string) {
		if field == FieldRepo && (name == "dependents" || name == "revdeps") {
			dependents = append(dependents, value)
		}
	})
	return dependents
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

// A query plan represents a set of disjoint queries for the search engine to
// execute. The result of executing a plan is the union of individual query results.
type Plan []Basic

// ToQ models a plan as a parse tree of an Or-expression on plan queries.
func (p Plan) ToQ() Q {
	nodes := make([]Node, 0, len(p))
	for _, basic := range p {
		operands := basic.ToParseTree()
		nodes = append(nodes, NewOperator(operands, And)...)
	}
	return Q(NewOperator(nodes, Or))
}

// Basic represents a leaf expression to evaluate in our search engine. A basic
// query comprises:
//   (1) a single search pattern expression, which may contain
//       'and' or 'or' operators; and
//   (2) parameters that scope the evaluation of search
//       patterns (e.g., to repos, files, etc.).
type Basic struct {
	Parameters
	Pattern Node
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

// MapCount returns a copy of a basic query with a count parameter set.
func (b Basic) MapCount(count int) Basic {
	parameters := MapParameter(toNodes(b.Parameters), func(field, value string, negated bool, annotation Annotation) Node {
		if field == "count" {
			value = strconv.FormatInt(int64(count), 10)
		}
		return Parameter{Field: field, Value: value, Negated: negated, Annotation: annotation}
	})
	return Basic{Parameters: toParameters(parameters), Pattern: b.Pattern}
}

func (b Basic) String() string {
	return b.toString(func(nodes []Node) string {
		return Q(nodes).String()
	})
}

func (b Basic) StringHuman() string {
	return b.toString(StringHuman)
}

// toString is a helper for String and StringHuman
func (b Basic) toString(marshal func([]Node) string) string {
	param := marshal(toNodes(b.Parameters))
	if b.Pattern != nil {
		return param + " " + marshal([]Node{b.Pattern})
	}
	return param
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

// PatternString returns the simple string pattern of a basic query. It assumes
// there is only on pattern atom.
func (b Basic) PatternString() string {
	if b.Pattern == nil {
		return ""
	}
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

type Parameters []Parameter

// IncludeExcludeValues partitions multiple values of a field into positive
// (include) and negated (exclude) values.
func (p Parameters) IncludeExcludeValues(field string) (include, exclude []string) {
	VisitField(toNodes(p), field, func(v string, negated bool, ann Annotation) {
		if ann.Labels.IsSet(IsPredicate) {
			// Skip predicates
			return
		}

		if negated {
			exclude = append(exclude, v)
		} else {
			include = append(include, v)
		}
	})
	return include, exclude
}

// RepoHasFileContentArgs represents the args of any of the following predicates:
// - repo:contains.file(f)
// - repo:contains.content(c)
// - repo:contains(file:f content:c)
// - repohasfile:f
type RepoHasFileContentArgs struct {
	// At least one of these strings should be non-empty
	Path    string // optional
	Content string // optional
	Negated bool
}

func (p Parameters) RepoHasFileContent() (res []RepoHasFileContentArgs) {
	nodes := toNodes(p)
	VisitField(nodes, FieldRepoHasFile, func(v string, negated bool, _ Annotation) {
		res = append(res, RepoHasFileContentArgs{
			Path:    v,
			Negated: negated,
		})
	})

	VisitTypedPredicate(nodes, func(pred *RepoContainsPathPredicate, negated bool) {
		res = append(res, RepoHasFileContentArgs{
			Path:    pred.Pattern,
			Negated: negated,
		})
	})

	VisitTypedPredicate(nodes, func(pred *RepoContainsContentPredicate, negated bool) {
		res = append(res, RepoHasFileContentArgs{
			Content: pred.Pattern,
			Negated: negated,
		})
	})

	VisitTypedPredicate(nodes, func(pred *RepoContainsFilePredicate, negated bool) {
		res = append(res, RepoHasFileContentArgs{
			Path:    pred.Path,
			Content: pred.Content,
			Negated: negated,
		})
	})

	return res
}

func (p Parameters) FileContainsContent() (include []string) {
	VisitTypedPredicate(toNodes(p), func(pred *FileContainsContentPredicate, negated bool) {
		include = append(include, pred.Pattern)
	})
	return include
}

func (p Parameters) RepoContainsCommitAfter() (value string) {
	nodes := toNodes(p)

	// Look for values of repohascommitafter:
	value = p.FindValue(FieldRepoHasCommitAfter)

	// Look for values of repo:contains.commit.after()
	VisitTypedPredicate(nodes, func(pred *RepoContainsCommitAfterPredicate, _ bool) {
		value = pred.TimeRef
	})

	return value
}

type RepoKVPFilter struct {
	Key     string
	Value   *string
	Negated bool
}

func (p Parameters) RepoHasKVPs() (res []RepoKVPFilter) {
	VisitTypedPredicate(toNodes(p), func(pred *RepoHasKVPPredicate, negated bool) {
		res = append(res, RepoKVPFilter{
			Key:     pred.Key,
			Value:   &pred.Value,
			Negated: negated,
		})
	})

	VisitTypedPredicate(toNodes(p), func(pred *RepoHasTagPredicate, negated bool) {
		res = append(res, RepoKVPFilter{
			Key:     pred.Key,
			Negated: negated,
		})
	})

	return res
}

func (p Parameters) FileHasOwner() (include, exclude []string) {
	VisitTypedPredicate(toNodes(p), func(pred *FileHasOwnerPredicate, negated bool) {
		if negated {
			exclude = append(exclude, pred.Owner)
		} else {
			include = append(include, pred.Owner)
		}
	})

	return include, exclude
}

// Exists returns whether a parameter exists in the query (whether negated or not).
func (p Parameters) Exists(field string) bool {
	found := false
	VisitField(toNodes(p), field, func(_ string, _ bool, _ Annotation) {
		found = true
	})
	return found
}

func (p Parameters) RepoHasDescription() (descriptionPatterns []string) {
	VisitTypedPredicate(toNodes(p), func(pred *RepoHasDescriptionPredicate, _ bool) {
		split := strings.Split(pred.Pattern, " ")
		descriptionPatterns = append(descriptionPatterns, "(?:"+strings.Join(split, ").*?(?:")+")")
	})
	return descriptionPatterns
}

func (p Parameters) MaxResults(defaultLimit int) int {
	if count := p.Count(); count != nil {
		return *count
	}

	if defaultLimit != 0 {
		return defaultLimit
	}

	return limits.DefaultMaxSearchResults
}

// Count returns the string value of the "count:" field. Returns empty string if none.
func (p Parameters) Count() (count *int) {
	VisitField(toNodes(p), FieldCount, func(value string, _ bool, _ Annotation) {
		c, err := strconv.Atoi(value)
		if err != nil {
			panic(fmt.Sprintf("Value %q for count cannot be parsed as an int", value))
		}
		count = &c
	})
	return count
}

// GetTimeout returns the time.Duration value from the `timeout:` field.
func (p Parameters) GetTimeout() *time.Duration {
	var timeout *time.Duration
	VisitField(toNodes(p), FieldTimeout, func(value string, _ bool, _ Annotation) {
		t, err := time.ParseDuration(value)
		if err != nil {
			panic(fmt.Sprintf("Value %q for timeout cannot be parsed as an duration: %s", value, err))
		}
		timeout = &t
	})
	return timeout
}

func (p Parameters) VisitParameter(field string, f func(value string, negated bool, annotation Annotation)) {
	for _, parameter := range p {
		if parameter.Field == field {
			f(parameter.Value, parameter.Negated, parameter.Annotation)
		}
	}
}

func (p Parameters) boolValue(field string) bool {
	result := false
	VisitField(toNodes(p), field, func(value string, _ bool, _ Annotation) {
		result, _ = parseBool(value) // err was checked during parsing and validation.
	})
	return result
}

func (p Parameters) IsCaseSensitive() bool {
	return p.boolValue(FieldCase)
}

func (p Parameters) yesNoOnlyValue(field string) *YesNoOnly {
	var res *YesNoOnly
	VisitField(toNodes(p), field, func(value string, _ bool, _ Annotation) {
		yno := parseYesNoOnly(value)
		if yno == Invalid {
			panic(fmt.Sprintf("Invalid value %q for field %q", value, field))
		}
		res = &yno
	})
	return res
}

func (p Parameters) Index() YesNoOnly {
	v := p.yesNoOnlyValue(FieldIndex)
	if v == nil {
		return Yes
	}
	return *v
}

func (p Parameters) Fork() *YesNoOnly {
	return p.yesNoOnlyValue(FieldFork)
}

func (p Parameters) Archived() *YesNoOnly {
	return p.yesNoOnlyValue(FieldArchived)
}

func (p Parameters) Repositories() (repos []string, negatedRepos []string) {
	VisitField(toNodes(p), FieldRepo, func(value string, negated bool, a Annotation) {
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

func (p Parameters) Visibility() RepoVisibility {
	visibilityStr := p.FindValue(FieldVisibility)
	return ParseVisibility(visibilityStr)
}

// FindValue returns the first value of a parameter matching field in b. It
// doesn't inspect whether the field is negated.
func (p Parameters) FindValue(field string) (value string) {
	var found string
	p.FindParameter(field, func(v string, _ bool, _ Annotation) {
		found = v
	})
	return found
}

// FindParameter calls f on parameters matching field in b.
func (p Parameters) FindParameter(field string, f func(value string, negated bool, annotation Annotation)) {
	for _, parameter := range p {
		if parameter.Field == field {
			f(parameter.Value, parameter.Negated, parameter.Annotation)
			break
		}
	}
}

// Flat is a more restricted form of Basic that has exactly zero or one atomic
// pattern nodes.
type Flat struct {
	Parameters
	Pattern *Pattern
}

func (f *Flat) ToBasic() Basic {
	var pattern Node
	if f.Pattern != nil {
		pattern = *f.Pattern
	}
	return Basic{Parameters: f.Parameters, Pattern: pattern}
}
