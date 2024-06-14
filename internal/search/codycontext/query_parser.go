package codycontext

import (
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const maxKeywords = 10

type contextQuery struct {
	symbolQuery  query.Basic
	symbols      []string
	keywordQuery query.Basic
	patterns     []string
}

func parseQuery(queryString string) (*contextQuery, error) {
	rawParseTree, err := query.Parse(queryString, query.SearchTypeStandard)
	if err != nil {
		return nil, err
	}

	if len(rawParseTree) != 1 {
		return nil, errors.New("The 'codycontext' patterntype does not support multiple clauses")
	}

	patterns, parameters := nodeToPatternsAndParameters(rawParseTree[0])

	symbols := findSymbols(patterns)
	keywords := append(expandQuery(queryString), transformPatterns(patterns)...)

	// To maintain decent latency, limit the number of keywords we search.
	if len(keywords) > maxKeywords {
		keywords = keywords[:maxKeywords]
	}

	symbolQuery, err := newBasicQuery(parameters, symbols)
	if err != nil {
		return nil, err
	}

	keywordQuery, err := newBasicQuery(parameters, keywords)
	if err != nil {
		return nil, err
	}

	return &contextQuery{symbolQuery, symbols, keywordQuery, keywords}, nil
}

func newBasicQuery(parameters []query.Parameter, patterns []string) (query.Basic, error) {
	var nodes []query.Node
	for _, p := range parameters {
		nodes = append(nodes, p)
	}

	patternNodes := make([]query.Node, 0, len(patterns))
	for _, p := range patterns {
		node := query.Pattern{Value: p}
		node.Annotation.Labels.Set(query.Literal)
		patternNodes = append(patternNodes, node)
	}

	if len(patternNodes) > 0 {
		nodes = append(nodes, query.NewOperator(patternNodes, query.Or)...)
	}

	newNodes, err := query.Sequence(query.For(query.SearchTypeStandard))(nodes)
	if err != nil {
		return query.Basic{}, err
	}

	basicQuery, err := query.ToBasicQuery(newNodes)
	if err != nil {
		return query.Basic{}, err
	}
	return basicQuery, nil
}

func concatNodeToPatterns(concat query.Operator) []string {
	patterns := make([]string, 0, len(concat.Operands))
	for _, operand := range concat.Operands {
		pattern, ok := operand.(query.Pattern)
		if ok {
			patterns = append(patterns, pattern.Value)
		}
	}
	return patterns
}

func nodeToPatternsAndParameters(rootNode query.Node) ([]string, []query.Parameter) {
	operator, ok := rootNode.(query.Operator)
	if !ok {
		return nil, nil
	}

	var patterns []string
	var parameters []query.Parameter

	switch operator.Kind {
	case query.And:
		for _, operand := range operator.Operands {
			switch op := operand.(type) {
			case query.Operator:
				if op.Kind == query.Concat {
					patterns = append(patterns, concatNodeToPatterns(op)...)
				}
			case query.Parameter:
				if op.Field == query.FieldContent {
					// Split any content field on white space into a set of patterns
					patterns = append(patterns, strings.Fields(op.Value)...)
				} else if op.Field != query.FieldCase && op.Field != query.FieldType {
					parameters = append(parameters, op)
				}
			case query.Pattern:
				patterns = append(patterns, op.Value)
			}
		}
	case query.Concat:
		patterns = concatNodeToPatterns(operator)
	}

	return patterns, parameters
}

// transformPatterns applies stops words and stemming. The returned slice
// contains the lowercased patterns and their stems minus the stop words.
func transformPatterns(patterns []string) []string {
	var transformedPatterns []string
	transformedPatternsSet := stringSet{}

	// To eliminate a possible source of non-determinism of search results, we
	// want transformPatterns to be a pure function. Hence we maintain a slice
	// of transformed patterns (transformedPatterns) in addition to
	// transformedPatternsSet.
	add := func(pattern string) {
		if transformedPatternsSet.Has(pattern) {
			return
		}
		transformedPatternsSet.Add(pattern)
		transformedPatterns = append(transformedPatterns, pattern)
	}

	for _, pattern := range patterns {
		pattern = strings.ToLower(pattern)
		pattern = removePunctuation(pattern)

		terms := tokenize(pattern)
		for _, term := range terms {
			if len(term) < 3 || isCommonTerm(term) {
				continue
			}

			term = stemTerm(term)
			add(term)
		}
	}

	return transformedPatterns
}

var namespaceRegex = regexp.MustCompile(`::|->`)
var camelCaseRegexp = regexp.MustCompile(`[a-z][A-Z]`)

// findSymbols extracts patterns that look like symbols. It uses the following heuristics:
//   - If the pattern contains an underscore or camelCase, it's probably a symbol
//   - If the pattern contains a namespace marker, each component is probably a symbol
//   - If the pattern contains a dot, split but still check if the components look like symbols.
//     This avoids overmatching in cases where the dot just represents a period.
func findSymbols(patterns []string) []string {
	var symbols []string
	for _, pattern := range patterns {

		split := namespaceRegex.Split(pattern, -1)
		if len(split) > 1 {
			symbols = append(symbols, split...)
		} else if strings.Contains(pattern, ".") {
			for _, split := range strings.Split(pattern, ".") {
				if isLikelySymbol(split) {
					symbols = append(symbols, split)
				}
			}
		} else if isLikelySymbol(pattern) {
			symbols = append(symbols, pattern)
		}
	}

	return symbols
}

func isLikelySymbol(pattern string) bool {
	return strings.Contains(pattern, "_") || camelCaseRegexp.MatchString(pattern)
}

var projectSignifiers = []string{
	"code",
	"codebase",
	"library",
	"module",
	"package",
	"program",
	"project",
	"repo",
	"repository",
}

var questionSignifiers = []string{
	"what",
	"how",
	"describe",
	"explain",
}

func needsReadmeContext(input string) bool {
	loweredInput := strings.ToLower(input)

	var containsQuestionSignifier bool
	for _, signifier := range questionSignifiers {
		if strings.Contains(loweredInput, signifier) {
			containsQuestionSignifier = true
			break
		}
	}
	if !containsQuestionSignifier {
		return false
	}

	for _, signifier := range projectSignifiers {
		if strings.Contains(loweredInput, signifier) {
			return true
		}
	}

	return false
}

const (
	kwReadme string = "readme"
)

// expandQuery returns a slice of keywords that likely relate to the query but
// are not explicitly mentioned in it.
func expandQuery(queryString string) []string {
	var keywords []string
	if needsReadmeContext(queryString) {
		keywords = append(keywords, kwReadme)
	}

	return keywords
}
