package smartsearch

import (
	"fmt"
	"net/url"
	"regexp/syntax" //nolint:depguard // using the grafana fork of regexp clashes with zoekt, which uses the std regexp/syntax.
	"strings"

	"github.com/go-enry/go-enry/v2"
	"github.com/grafana/regexp"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

// rule represents a transformation function on a Basic query. Transformation
// cannot fail: either they apply in sequence and produce a valid, non-nil,
// Basic query, or they do not apply, in which case they return nil. See the
// `unquotePatterns` rule for an example.
type rule struct {
	description string
	transform   []transform
}

type transform func(query.Basic) *query.Basic

var rulesNarrow = []rule{
	{
		description: "unquote patterns",
		transform:   []transform{unquotePatterns},
	},
	{
		description: "apply search type for pattern",
		transform:   []transform{typePatterns},
	},
	{
		description: "apply language filter for pattern",
		transform:   []transform{langPatterns},
	},
	{
		description: "apply symbol select for pattern",
		transform:   []transform{symbolPatterns},
	},
	{
		description: "expand URL to filters",
		transform:   []transform{patternsToCodeHostFilters},
	},
	{
		description: "rewrite repo URLs",
		transform:   []transform{rewriteRepoFilter},
	},
}

var rulesWiden = []rule{
	{
		description: "patterns as regular expressions",
		transform:   []transform{regexpPatterns},
	},
	{
		description: "AND patterns together",
		transform:   []transform{unorderedPatterns},
	},
}

// unquotePatterns is a rule that unquotes all patterns in the input query (it
// removes quotes, and honors escape sequences inside quoted values).
func unquotePatterns(b query.Basic) *query.Basic {
	// Go back all the way to the raw tree representation :-). We just parse
	// the string as regex, since parsing with regex annotates quoted
	// patterns.
	rawParseTree, err := query.Parse(query.StringHuman(b.ToParseTree()), query.SearchTypeRegex)
	if err != nil {
		return nil
	}

	changed := false // track whether we've successfully changed any pattern, which means this rule applies.
	newParseTree := query.MapPattern(rawParseTree, func(value string, negated bool, annotation query.Annotation) query.Node {
		if annotation.Labels.IsSet(query.Quoted) && !annotation.Labels.IsSet(query.IsAlias) {
			changed = true
			annotation.Labels.Unset(query.Quoted)
			annotation.Labels.Set(query.Literal)
			return query.Pattern{
				Value:      value,
				Negated:    negated,
				Annotation: annotation,
			}
		}
		return query.Pattern{
			Value:      value,
			Negated:    negated,
			Annotation: annotation,
		}
	})

	if !changed {
		// No unquoting happened, so we don't run the search.
		return nil
	}

	newNodes, err := query.Sequence(query.For(query.SearchTypeStandard))(newParseTree)
	if err != nil {
		return nil
	}

	newBasic, err := query.ToBasicQuery(newNodes)
	if err != nil {
		return nil
	}

	return &newBasic
}

// regexpPatterns converts literal patterns into regular expression patterns.
// The conversion is a heuristic and happens based on whether the pattern has
// indicative regular expression metasyntax. It would be overly aggressive to
// convert patterns containing _any_ potential metasyntax, since a pattern like
// my.config.yaml contains two `.` (match any character in regexp).
func regexpPatterns(b query.Basic) *query.Basic {
	rawParseTree, err := query.Parse(query.StringHuman(b.ToParseTree()), query.SearchTypeStandard)
	if err != nil {
		return nil
	}

	// we decide to interpret patterns as regular expressions if the number of
	// significant metasyntax operators exceed this threshold
	METASYNTAX_THRESHOLD := 2

	// countMetaSyntax counts the number of significant regular expression
	// operators in string when it is interpreted as a regular expression. A
	// rough map of operators to syntax can be found here:
	// https://sourcegraph.com/github.com/golang/go@bf5898ef53d1693aa572da0da746c05e9a6f15c5/-/blob/src/regexp/syntax/regexp.go?L116-244
	var countMetaSyntax func([]*syntax.Regexp) int
	countMetaSyntax = func(res []*syntax.Regexp) int {
		count := 0
		for _, r := range res {
			switch r.Op {
			case
				// operators that are weighted 0 on their own
				syntax.OpAnyCharNotNL,
				syntax.OpAnyChar,
				syntax.OpNoMatch,
				syntax.OpEmptyMatch,
				syntax.OpLiteral,
				syntax.OpConcat:
				count += countMetaSyntax(r.Sub)
			case
				// operators that are weighted 1 on their own
				syntax.OpCharClass,
				syntax.OpBeginLine,
				syntax.OpEndLine,
				syntax.OpBeginText,
				syntax.OpEndText,
				syntax.OpWordBoundary,
				syntax.OpNoWordBoundary,
				syntax.OpAlternate:
				count += countMetaSyntax(r.Sub) + 1

			case
				// quantifiers *, +, ?, {...} on metasyntax like
				// `.` or `(...)` are weighted 2. If the
				// quantifier applies to other syntax like
				// literals (not metasyntax) it's weighted 1.
				syntax.OpStar,
				syntax.OpPlus,
				syntax.OpQuest,
				syntax.OpRepeat:
				switch r.Sub[0].Op {
				case
					syntax.OpAnyChar,
					syntax.OpAnyCharNotNL,
					syntax.OpCapture:
					count += countMetaSyntax(r.Sub) + 2
				default:
					count += countMetaSyntax(r.Sub) + 1
				}
			case
				// capture groups over an alternate like (a|b)
				// are weighted one. All other capture groups
				// are weighted zero on their own because parens
				// are very common in code.
				syntax.OpCapture:
				switch r.Sub[0].Op {
				case syntax.OpAlternate:
					count += countMetaSyntax(r.Sub) + 1
				default:
					count += countMetaSyntax(r.Sub)
				}
			}
		}
		return count
	}

	changed := false
	newParseTree := query.MapPattern(rawParseTree, func(value string, negated bool, annotation query.Annotation) query.Node {
		if annotation.Labels.IsSet(query.Regexp) {
			return query.Pattern{
				Value:      value,
				Negated:    negated,
				Annotation: annotation,
			}
		}

		re, err := syntax.Parse(value, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
		if err != nil {
			return query.Pattern{
				Value:      value,
				Negated:    negated,
				Annotation: annotation,
			}
		}

		count := countMetaSyntax([]*syntax.Regexp{re})
		if count < METASYNTAX_THRESHOLD {
			return query.Pattern{
				Value:      value,
				Negated:    negated,
				Annotation: annotation,
			}
		}

		changed = true
		annotation.Labels.Unset(query.Literal)
		annotation.Labels.Set(query.Regexp)
		return query.Pattern{
			Value:      value,
			Negated:    negated,
			Annotation: annotation,
		}
	})

	if !changed {
		return nil
	}

	newNodes, err := query.Sequence(query.For(query.SearchTypeStandard))(newParseTree)
	if err != nil {
		return nil
	}

	newBasic, err := query.ToBasicQuery(newNodes)
	if err != nil {
		return nil
	}

	return &newBasic
}

// UnorderedPatterns generates a query that interprets all recognized patterns
// as unordered terms (`and`-ed terms). The implementation detail is that we
// simply map all `concat` nodes (after a raw parse) to `and` nodes. This works
// because parsing maintains the invariant that `concat` nodes only ever have
// pattern children.
func unorderedPatterns(b query.Basic) *query.Basic {
	rawParseTree, err := query.Parse(query.StringHuman(b.ToParseTree()), query.SearchTypeStandard)
	if err != nil {
		return nil
	}

	newParseTree, changed := mapConcat(rawParseTree)
	if !changed {
		return nil
	}

	newNodes, err := query.Sequence(query.For(query.SearchTypeStandard))(newParseTree)
	if err != nil {
		return nil
	}

	newBasic, err := query.ToBasicQuery(newNodes)
	if err != nil {
		return nil
	}

	return &newBasic
}

func mapConcat(q []query.Node) ([]query.Node, bool) {
	mapped := make([]query.Node, 0, len(q))
	changed := false
	for _, node := range q {
		if n, ok := node.(query.Operator); ok {
			if n.Kind != query.Concat {
				// recurse
				operands, newChanged := mapConcat(n.Operands)
				mapped = append(mapped, query.Operator{
					Kind:     n.Kind,
					Operands: operands,
				})
				changed = changed || newChanged
				continue
			}
			// no need to recurse: `concat` nodes only have patterns.
			mapped = append(mapped, query.Operator{
				Kind:     query.And,
				Operands: n.Operands,
			})
			changed = true
			continue
		}
		mapped = append(mapped, node)
	}
	return mapped, changed
}

var symbolTypes = map[string]string{
	"function":       "function",
	"func":           "function",
	"module":         "module",
	"namespace":      "namespace",
	"package":        "package",
	"class":          "class",
	"method":         "method",
	"property":       "property",
	"field":          "field",
	"constructor":    "constructor",
	"interface":      "interface",
	"variable":       "variable",
	"var":            "variable",
	"constant":       "constant",
	"const":          "constant",
	"string":         "string",
	"number":         "number",
	"boolean":        "boolean",
	"bool":           "boolean",
	"array":          "array",
	"object":         "object",
	"key":            "key",
	"enum":           "enum-member",
	"struct":         "struct",
	"type-parameter": "type-parameter",
}

func symbolPatterns(b query.Basic) *query.Basic {
	rawPatternTree, err := query.Parse(query.StringHuman([]query.Node{b.Pattern}), query.SearchTypeStandard)
	if err != nil {
		return nil
	}

	changed := false
	var symbolType string // store the first pattern that matches a recognized symbol type.
	isNegated := false
	newPattern := query.MapPattern(rawPatternTree, func(value string, negated bool, annotation query.Annotation) query.Node {
		symbolAlias, ok := symbolTypes[value]
		if !ok || changed {
			return query.Pattern{
				Value:      value,
				Negated:    negated,
				Annotation: annotation,
			}
		}
		changed = true
		symbolType = symbolAlias
		isNegated = negated
		// remove this node
		return nil
	})

	if !changed {
		return nil
	}

	selectParam := query.Parameter{
		Field:      query.FieldSelect,
		Value:      fmt.Sprintf("symbol.%s", symbolType),
		Negated:    isNegated,
		Annotation: query.Annotation{},
	}
	symbolParam := query.Parameter{
		Field:      query.FieldType,
		Value:      "symbol",
		Negated:    false,
		Annotation: query.Annotation{},
	}

	var pattern query.Node
	if len(newPattern) > 0 {
		// Process concat nodes
		nodes, err := query.Sequence(query.For(query.SearchTypeStandard))(newPattern)
		if err != nil {
			return nil
		}
		pattern = nodes[0] // guaranteed root at first node
	}

	return &query.Basic{
		Parameters: append(b.Parameters, selectParam, symbolParam),
		Pattern:    pattern,
	}
}

type repoFilterReplacement struct {
	match   *regexp.Regexp
	replace string
}

var repoFilterReplacements = []repoFilterReplacement{
	{
		match:   regexp.MustCompile(`^(?:https?:\/\/)github\.com\/([^\/]+)\/([^\/\?#]+)(?:.+)?$`),
		replace: "^github.com/$1/$2$",
	},
}

func rewriteRepoFilter(b query.Basic) *query.Basic {
	newParams := make([]query.Parameter, 0, len(b.Parameters))
	anyParamChanged := false
	for _, param := range b.Parameters {
		if param.Field != "repo" {
			newParams = append(newParams, param)
			continue
		}

		changed := false
		for _, replacer := range repoFilterReplacements {
			if replacer.match.MatchString(param.Value) {
				newParams = append(newParams, query.Parameter{
					Field:      param.Field,
					Value:      replacer.match.ReplaceAllString(param.Value, replacer.replace),
					Negated:    param.Negated,
					Annotation: param.Annotation,
				})
				changed = true
				anyParamChanged = true
				break
			}
		}
		if !changed {
			newParams = append(newParams, param)
		}
	}
	if !anyParamChanged {
		return nil
	}
	newQuery := b.MapParameters(newParams)
	return &newQuery
}

func langPatterns(b query.Basic) *query.Basic {
	rawPatternTree, err := query.Parse(query.StringHuman([]query.Node{b.Pattern}), query.SearchTypeStandard)
	if err != nil {
		return nil
	}

	changed := false
	var lang string // store the first pattern that matches a recognized language.
	isNegated := false
	newPattern := query.MapPattern(rawPatternTree, func(value string, negated bool, annotation query.Annotation) query.Node {
		langAlias, ok := enry.GetLanguageByAlias(value)
		if !ok || changed {
			return query.Pattern{
				Value:      value,
				Negated:    negated,
				Annotation: annotation,
			}
		}
		changed = true
		lang = langAlias
		isNegated = negated
		// remove this node
		return nil
	})

	if !changed {
		return nil
	}

	langParam := query.Parameter{
		Field:      query.FieldLang,
		Value:      lang,
		Negated:    isNegated,
		Annotation: query.Annotation{},
	}

	var pattern query.Node
	if len(newPattern) > 0 {
		// Process concat nodes
		nodes, err := query.Sequence(query.For(query.SearchTypeStandard))(newPattern)
		if err != nil {
			return nil
		}
		pattern = nodes[0] // guaranteed root at first node
	}

	return &query.Basic{
		Parameters: append(b.Parameters, langParam),
		Pattern:    pattern,
	}
}

func typePatterns(b query.Basic) *query.Basic {
	rawPatternTree, err := query.Parse(query.StringHuman([]query.Node{b.Pattern}), query.SearchTypeStandard)
	if err != nil {
		return nil
	}

	changed := false
	var typ string // store the first pattern that matches a recognized `type:`.
	newPattern := query.MapPattern(rawPatternTree, func(value string, negated bool, annotation query.Annotation) query.Node {
		if changed {
			return query.Pattern{
				Value:      value,
				Negated:    negated,
				Annotation: annotation,
			}
		}

		switch strings.ToLower(value) {
		case "symbol", "commit", "diff", "path":
			typ = value
			changed = true
			// remove this node
			return nil
		}

		return query.Pattern{
			Value:      value,
			Negated:    negated,
			Annotation: annotation,
		}
	})

	if !changed {
		return nil
	}

	typParam := query.Parameter{
		Field:      query.FieldType,
		Value:      typ,
		Negated:    false,
		Annotation: query.Annotation{},
	}

	var pattern query.Node
	if len(newPattern) > 0 {
		// Process concat nodes
		nodes, err := query.Sequence(query.For(query.SearchTypeStandard))(newPattern)
		if err != nil {
			return nil
		}
		pattern = nodes[0] // guaranteed root at first node
	}

	return &query.Basic{
		Parameters: append(b.Parameters, typParam),
		Pattern:    pattern,
	}
}

var lookup = map[string]struct{}{
	"github.com": {},
	"gitlab.com": {},
}

// patternToCodeHostFilters checks if a pattern contains a code host URL and
// extracts the org/repo/branch and path and lifts these to filters, as
// applicable.
func patternToCodeHostFilters(v string, negated bool) *[]query.Node {
	if !strings.HasPrefix(v, "https://") {
		// normalize v with https:// prefix.
		v = "https://" + v
	}

	u, err := url.Parse(v)
	if err != nil {
		return nil
	}

	domain := strings.TrimPrefix(u.Host, "www.")
	if _, ok := lookup[domain]; !ok {
		return nil
	}

	var value string
	path := strings.Trim(u.Path, "/")
	pathElems := strings.Split(path, "/")
	if len(pathElems) == 0 {
		value = regexp.QuoteMeta(domain)
		value = fmt.Sprintf("^%s", value)
		return &[]query.Node{
			query.Parameter{
				Field:      query.FieldRepo,
				Value:      value,
				Negated:    negated,
				Annotation: query.Annotation{},
			}}
	} else if len(pathElems) == 1 {
		value = regexp.QuoteMeta(domain)
		value = fmt.Sprintf("^%s/%s", value, strings.Join(pathElems, "/"))
		return &[]query.Node{
			query.Parameter{
				Field:      query.FieldRepo,
				Value:      value,
				Negated:    negated,
				Annotation: query.Annotation{},
			}}
	} else if len(pathElems) == 2 {
		value = regexp.QuoteMeta(domain)
		value = fmt.Sprintf("^%s/%s$", value, strings.Join(pathElems, "/"))
		return &[]query.Node{
			query.Parameter{
				Field:      query.FieldRepo,
				Value:      value,
				Negated:    negated,
				Annotation: query.Annotation{},
			}}
	} else if len(pathElems) == 4 && (pathElems[2] == "tree" || pathElems[2] == "commit") {
		repoValue := regexp.QuoteMeta(domain)
		repoValue = fmt.Sprintf("^%s/%s$", repoValue, strings.Join(pathElems[:2], "/"))
		revision := pathElems[3]
		return &[]query.Node{
			query.Parameter{
				Field:      query.FieldRepo,
				Value:      repoValue,
				Negated:    negated,
				Annotation: query.Annotation{},
			},
			query.Parameter{
				Field:      query.FieldRev,
				Value:      revision,
				Negated:    negated,
				Annotation: query.Annotation{},
			},
		}
	} else if len(pathElems) >= 5 {
		repoValue := regexp.QuoteMeta(domain)
		repoValue = fmt.Sprintf("^%s/%s$", repoValue, strings.Join(pathElems[:2], "/"))

		revision := pathElems[3]

		pathValue := strings.Join(pathElems[4:], "/")
		pathValue = regexp.QuoteMeta(pathValue)

		if pathElems[2] == "blob" {
			pathValue = fmt.Sprintf("^%s$", pathValue)
		} else if pathElems[2] == "tree" {
			pathValue = fmt.Sprintf("^%s", pathValue)
		} else {
			// We don't know what this is.
			return nil
		}

		return &[]query.Node{
			query.Parameter{
				Field:      query.FieldRepo,
				Value:      repoValue,
				Negated:    negated,
				Annotation: query.Annotation{},
			},
			query.Parameter{
				Field:      query.FieldRev,
				Value:      revision,
				Negated:    negated,
				Annotation: query.Annotation{},
			},
			query.Parameter{
				Field:      query.FieldFile,
				Value:      pathValue,
				Negated:    negated,
				Annotation: query.Annotation{},
			},
		}
	}

	return nil
}

// patternsToCodeHostFilters converts patterns to `repo` or `path` filters if they
// can be interpreted as URIs.
func patternsToCodeHostFilters(b query.Basic) *query.Basic {
	rawPatternTree, err := query.Parse(query.StringHuman([]query.Node{b.Pattern}), query.SearchTypeStandard)
	if err != nil {
		return nil
	}

	filterParams := []query.Node{}
	changed := false
	newParseTree := query.MapPattern(rawPatternTree, func(value string, negated bool, annotation query.Annotation) query.Node {
		if params := patternToCodeHostFilters(value, negated); params != nil {
			changed = true
			filterParams = append(filterParams, *params...)
			// Collect the param and delete pattern. We're going to
			// add those parameters after. We can't map patterns
			// in-place because that might create parameters in
			// concat nodes.
			return nil
		}

		return query.Pattern{
			Value:      value,
			Negated:    negated,
			Annotation: annotation,
		}
	})

	if !changed {
		return nil
	}

	newParseTree = query.NewOperator(append(newParseTree, filterParams...), query.And) // Reduce with NewOperator to obtain valid partitioning.
	newNodes, err := query.Sequence(query.For(query.SearchTypeStandard))(newParseTree)
	if err != nil {
		return nil
	}

	newBasic, err := query.ToBasicQuery(newNodes)
	if err != nil {
		return nil
	}

	return &newBasic
}
