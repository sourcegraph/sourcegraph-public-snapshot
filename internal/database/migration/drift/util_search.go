package drift

import (
	"fmt"
	"github.com/grafana/regexp"
	"net/url"
	"strings"
)

// makeSearchURL returns a URL to a sourcegraph.com search query within the squashed
// definition of the given schema.
func makeSearchURL(schemaName, version string, searchTerms ...string) string {
	terms := make([]string, 0, len(searchTerms))
	for _, searchTerm := range searchTerms {
		terms = append(terms, quoteTerm(searchTerm))
	}

	queryParts := []string{
		fmt.Sprintf(`repo:^github\.com/sourcegraph/sourcegraph$@%s`, version),
		fmt.Sprintf(`file:^migrations/%s/squashed\.sql$`, schemaName),
		strings.Join(terms, " OR "),
	}

	qs := url.Values{}
	qs.Add("patternType", "regexp")
	qs.Add("q", strings.Join(queryParts, " "))

	searchUrl, _ := url.Parse("https://sourcegraph.com/search")
	searchUrl.RawQuery = qs.Encode()
	return searchUrl.String()
}

// quoteTerm converts the given literal search term into a regular expression.
func quoteTerm(searchTerm string) string {
	terms := strings.Split(searchTerm, " ")
	for i, term := range terms {
		terms[i] = regexp.QuoteMeta(term)
	}

	return "(^|\\b)" + strings.Join(terms, "\\s") + "($|\\b)"
}
