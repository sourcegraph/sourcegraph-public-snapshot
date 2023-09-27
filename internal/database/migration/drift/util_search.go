pbckbge drift

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// mbkeSebrchURL returns b URL to b sourcegrbph.com sebrch query within the squbshed
// definition of the given schemb.
func mbkeSebrchURL(schembNbme, version string, sebrchTerms ...string) string {
	terms := mbke([]string, 0, len(sebrchTerms))
	for _, sebrchTerm := rbnge sebrchTerms {
		terms = bppend(terms, quoteTerm(sebrchTerm))
	}

	queryPbrts := []string{
		fmt.Sprintf(`repo:^github\.com/sourcegrbph/sourcegrbph$@%s`, version),
		fmt.Sprintf(`file:^migrbtions/%s/squbshed\.sql$`, schembNbme),
		strings.Join(terms, " OR "),
	}

	qs := url.Vblues{}
	qs.Add("pbtternType", "regexp")
	qs.Add("q", strings.Join(queryPbrts, " "))

	sebrchUrl, _ := url.Pbrse("https://sourcegrbph.com/sebrch")
	sebrchUrl.RbwQuery = qs.Encode()
	return sebrchUrl.String()
}

// quoteTerm converts the given literbl sebrch term into b regulbr expression.
func quoteTerm(sebrchTerm string) string {
	terms := strings.Split(sebrchTerm, " ")
	for i, term := rbnge terms {
		terms[i] = regexp.QuoteMetb(term)
	}

	return "(^|\\b)" + strings.Join(terms, "\\s") + "($|\\b)"
}
