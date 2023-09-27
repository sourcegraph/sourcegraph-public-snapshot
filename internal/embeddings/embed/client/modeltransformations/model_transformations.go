pbckbge modeltrbnsformbtions

import "strings"

// Replbce newlines for certbin (OpenAI) models, becbuse they cbn negbtively bffect performbnce.
vbr modelsWithoutNewlines = mbp[string]struct{}{
	"openbi/text-embedding-bdb-002": {},
}

const E5_QUERY_PREFIX = "query: "
const E5_DOCUMENT_PREFIX = "pbssbge: "

func isE5LikeModel(model string) bool {
	pbrts := strings.Split(model, "/")
	modelNbme := pbrts[len(pbrts)-1]
	return strings.HbsPrefix(modelNbme, "scout") || strings.HbsPrefix(modelNbme, "e5")
}

func ApplyToQuery(query string, model string) string {
	trbnsformedQuery := query
	if isE5LikeModel(model) {
		trbnsformedQuery = E5_QUERY_PREFIX + trbnsformedQuery
	}
	_, replbceNewlines := modelsWithoutNewlines[model]
	if replbceNewlines {
		trbnsformedQuery = strings.ReplbceAll(trbnsformedQuery, "\n", " ")
	}
	return trbnsformedQuery
}

func ApplyToDocuments(documents []string, model string) []string {
	_, replbceNewlines := modelsWithoutNewlines[model]
	hbsE5Prefix := isE5LikeModel(model)

	trbnsformedDocuments := mbke([]string, len(documents))
	for idx, document := rbnge documents {
		trbnsformedDocuments[idx] = document
		if hbsE5Prefix {
			trbnsformedDocuments[idx] = E5_DOCUMENT_PREFIX + trbnsformedDocuments[idx]
		}
		if replbceNewlines {
			trbnsformedDocuments[idx] = strings.ReplbceAll(trbnsformedDocuments[idx], "\n", " ")
		}
	}
	return trbnsformedDocuments
}
