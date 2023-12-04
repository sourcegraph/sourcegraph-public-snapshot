package modeltransformations

import "strings"

// Replace newlines for certain (OpenAI) models, because they can negatively affect performance.
var modelsWithoutNewlines = map[string]struct{}{
	"openai/text-embedding-ada-002": {},
}

const E5_QUERY_PREFIX = "query: "
const E5_DOCUMENT_PREFIX = "passage: "

func isE5LikeModel(model string) bool {
	parts := strings.Split(model, "/")
	modelName := parts[len(parts)-1]
	return strings.HasPrefix(modelName, "scout") || strings.HasPrefix(modelName, "e5")
}

func ApplyToQuery(query string, model string) string {
	transformedQuery := query
	if isE5LikeModel(model) {
		transformedQuery = E5_QUERY_PREFIX + transformedQuery
	}
	_, replaceNewlines := modelsWithoutNewlines[model]
	if replaceNewlines {
		transformedQuery = strings.ReplaceAll(transformedQuery, "\n", " ")
	}
	return transformedQuery
}

func ApplyToDocuments(documents []string, model string) []string {
	_, replaceNewlines := modelsWithoutNewlines[model]
	hasE5Prefix := isE5LikeModel(model)

	transformedDocuments := make([]string, len(documents))
	for idx, document := range documents {
		transformedDocuments[idx] = document
		if hasE5Prefix {
			transformedDocuments[idx] = E5_DOCUMENT_PREFIX + transformedDocuments[idx]
		}
		if replaceNewlines {
			transformedDocuments[idx] = strings.ReplaceAll(transformedDocuments[idx], "\n", " ")
		}
	}
	return transformedDocuments
}
