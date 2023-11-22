package modeltransformations

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const query = "query1\nquery2\nquery3"

var documents = []string{
	"line1\nline2\nline3",
	"line4\nline5\nline6",
}

func TestOpenAIModel(t *testing.T) {
	transformedQuery := ApplyToQuery(query, "openai/text-embedding-ada-002")
	require.Equal(t, "query1 query2 query3", transformedQuery)

	transformedDocuments := ApplyToDocuments(documents, "openai/text-embedding-ada-002")
	require.Equal(t, []string{"line1 line2 line3", "line4 line5 line6"}, transformedDocuments)
}

func TestE5LikeModel(t *testing.T) {
	transformedQuery := ApplyToQuery(query, "sourcegraph/scout-base-v2")
	require.Equal(t, "query: query1\nquery2\nquery3", transformedQuery)

	transformedDocuments := ApplyToDocuments(documents, "sourcegraph/scout-base-v2")
	require.Equal(t, []string{"passage: line1\nline2\nline3", "passage: line4\nline5\nline6"}, transformedDocuments)
}

func TestModelWithoutTransformations(t *testing.T) {
	transformedQuery := ApplyToQuery(query, "no-transform")
	require.Equal(t, query, transformedQuery)

	transformedDocuments := ApplyToDocuments(documents, "no-transform")
	require.Equal(t, documents, transformedDocuments)
}
