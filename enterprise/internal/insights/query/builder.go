package query

import (
	searchquery "github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// WithDefaults builds a Sourcegraph query from a base input query setting default fields if they are not specified
// in the base query. For example an input query of `repo:myrepo test` might be provided a default `archived:no`,
// and the result would be generated as `repo:myrepo test archive:no`. This preserves the semantics of the original query
// by fully parsing and reconstructing the tree, and does **not** overwrite user supplied values for the default fields.
func WithDefaults(inputQuery string, defaults searchquery.Parameters) (string, error) {
	plan, err := searchquery.Pipeline(searchquery.Init(inputQuery, searchquery.SearchTypeRegex))
	if err != nil {
		return "", errors.Wrap(err, "Pipeline")
	}
	modified := make(searchquery.Plan, 0, len(plan))

	for _, basic := range plan {
		p := make(searchquery.Parameters, 0, len(basic.Parameters)+len(defaults))

		for _, defaultParam := range defaults {
			if !basic.Parameters.Exists(defaultParam.Field) {
				p = append(p, defaultParam)
			}
		}
		p = append(p, basic.Parameters...)
		modified = append(modified, basic.MapParameters(p))
	}

	return searchquery.StringHuman(modified.ToParseTree()), nil
}
