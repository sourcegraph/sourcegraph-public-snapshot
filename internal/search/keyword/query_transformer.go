package keyword

import (
	"strings"

	"github.com/go-enry/go-enry/v2"
	"github.com/kljensen/snowball"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
)

type keywordQuery struct {
	query    query.Basic
	patterns []string
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

func removeStringAtIndex(s []string, index int) []string {
	ret := make([]string, 0, len(s)-1)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}

func nodeToPatternsAndParameters(rootNode query.Node) ([]string, []query.Parameter) {
	operator, ok := rootNode.(query.Operator)
	if !ok {
		return nil, nil
	}

	patterns := []string{}
	parameters := []query.Parameter{
		// Force search backend to return all results
		{Field: query.FieldCount, Value: "all"},
		// Only search file content
		{Field: query.FieldType, Value: "file"},
	}
	seenLangParameter := false

	switch operator.Kind {
	case query.And:
		for _, operand := range operator.Operands {
			switch op := operand.(type) {
			case query.Operator:
				if op.Kind == query.Concat {
					patterns = append(patterns, concatNodeToPatterns(op)...)
				}
			case query.Parameter:
				if op.Field != query.FieldCount && op.Field != query.FieldCase && op.Field != query.FieldType {
					parameters = append(parameters, op)
				}
				if op.Field == query.FieldLang {
					seenLangParameter = true
				}
			case query.Pattern:
				patterns = append(patterns, op.Value)
			}
		}
	case query.Concat:
		patterns = concatNodeToPatterns(operator)
	}

	// Check if any of the patterns can be substituted as a lang: filter
	if !seenLangParameter {
		langPatternIdx := -1
		for idx, pattern := range patterns {
			langAlias, ok := enry.GetLanguageByAlias(pattern)
			if ok {
				parameters = append(parameters, query.Parameter{Field: query.FieldLang, Value: langAlias})
				langPatternIdx = idx
				break
			}
		}
		if langPatternIdx >= 0 {
			patterns = removeStringAtIndex(patterns, langPatternIdx)
		}
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
		patternLowerCase := strings.ToLower(pattern)

		if stopWords.Has(patternLowerCase) {
			continue
		}
		add(patternLowerCase)

		stemmed, err := snowball.Stem(patternLowerCase, "english", false)
		if err != nil {
			continue
		}
		add(stemmed)
	}

	return transformedPatterns
}

func queryStringToKeywordQuery(queryString string) (*keywordQuery, error) {
	rawParseTree, err := query.Parse(queryString, query.SearchTypeStandard)
	if err != nil {
		return nil, err
	}

	if len(rawParseTree) != 1 {
		return nil, nil
	}

	patterns, parameters := nodeToPatternsAndParameters(rawParseTree[0])

	transformedPatterns := transformPatterns(patterns)
	if len(transformedPatterns) == 0 {
		return nil, nil
	}

	nodes := []query.Node{}
	for _, p := range parameters {
		nodes = append(nodes, p)
	}

	patternNodes := make([]query.Node, 0, len(transformedPatterns))
	for _, p := range transformedPatterns {
		patternNodes = append(patternNodes, query.Pattern{Value: p})
	}
	nodes = append(nodes, query.NewOperator(patternNodes, query.Or)...)

	newNodes, err := query.Sequence(query.For(query.SearchTypeStandard))(nodes)
	if err != nil {
		return nil, err
	}

	newBasic, err := query.ToBasicQuery(newNodes)
	if err != nil {
		return nil, err
	}

	return &keywordQuery{newBasic, transformedPatterns}, nil
}

func basicQueryToKeywordQuery(basicQuery query.Basic) (*keywordQuery, error) {
	return queryStringToKeywordQuery(query.StringHuman(basicQuery.ToParseTree()))
}
