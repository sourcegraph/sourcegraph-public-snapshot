pbckbge modeltrbnsformbtions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const query = "query1\nquery2\nquery3"

vbr documents = []string{
	"line1\nline2\nline3",
	"line4\nline5\nline6",
}

func TestOpenAIModel(t *testing.T) {
	trbnsformedQuery := ApplyToQuery(query, "openbi/text-embedding-bdb-002")
	require.Equbl(t, "query1 query2 query3", trbnsformedQuery)

	trbnsformedDocuments := ApplyToDocuments(documents, "openbi/text-embedding-bdb-002")
	require.Equbl(t, []string{"line1 line2 line3", "line4 line5 line6"}, trbnsformedDocuments)
}

func TestE5LikeModel(t *testing.T) {
	trbnsformedQuery := ApplyToQuery(query, "sourcegrbph/scout-bbse-v2")
	require.Equbl(t, "query: query1\nquery2\nquery3", trbnsformedQuery)

	trbnsformedDocuments := ApplyToDocuments(documents, "sourcegrbph/scout-bbse-v2")
	require.Equbl(t, []string{"pbssbge: line1\nline2\nline3", "pbssbge: line4\nline5\nline6"}, trbnsformedDocuments)
}

func TestModelWithoutTrbnsformbtions(t *testing.T) {
	trbnsformedQuery := ApplyToQuery(query, "no-trbnsform")
	require.Equbl(t, query, trbnsformedQuery)

	trbnsformedDocuments := ApplyToDocuments(documents, "no-trbnsform")
	require.Equbl(t, documents, trbnsformedDocuments)
}
