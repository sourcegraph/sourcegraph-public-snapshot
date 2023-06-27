package modeltransformations

import "strings"

// Replace newlines for certain (OpenAI) models, because they can negatively affect performance.
var modelsWithoutNewlines = map[string]struct{}{
	"openai/text-embedding-ada-002": {},
}

const E5_QUERY_PREFIX = "query: "
const E5_DOCUMENT_PREFIX = "passage: "

func isE5LikeModel(model string) bool {
	return strings.HasPrefix(model, "sourcegraph/scout") || strings.HasPrefix(model, "sourcegraph/e5")
}

func ApplyModelTransformationsForQuery(query string, model string) string {
	transformedQuery := query
	if isE5LikeModel(model) {
		transformedQuery = E5_QUERY_PREFIX + query
	}
	_, replaceNewlines := modelsWithoutNewlines[model]
	if replaceNewlines {
		transformedQuery = strings.ReplaceAll(transformedQuery, "\n", " ")
	}
	return transformedQuery
}

func ApplyModelTransformationsForDocuments(documents []string, model string) []string {
	_, replaceNewlines := modelsWithoutNewlines[model]
	hasE5Prefix := isE5LikeModel(model)

	transformedDocuments := documents
	for idx, text := range transformedDocuments {
		if hasE5Prefix {
			transformedDocuments[idx] = E5_DOCUMENT_PREFIX + text
		}
		if replaceNewlines {
			transformedDocuments[idx] = strings.ReplaceAll(text, "\n", " ")
		}
	}
	return transformedDocuments
}
