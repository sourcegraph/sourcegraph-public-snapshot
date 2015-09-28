package sqlwrapper

import (
	"regexp"
	"strings"
)

var getSQLVerbRe *regexp.Regexp = regexp.MustCompile(
	`(?is:[[:space:]]*(SELECT|INSERT|UPDATE|DELETE)[[:space:]]+.*)`)

type QueryType int

const (
	Select QueryType = iota
	Insert
	Update
	Delete
	Unknown
)

var kLowerVerbMap = map[string]QueryType{
	"select": Select,
	"insert": Insert,
	"update": Update,
	"delete": Delete,
}

type parsedQuery struct {
	Query     string
	QueryType QueryType
}

// A minimal SQL parser that extracts the verb from the query string.
func parseQuery(query string) *parsedQuery {
	rval := &parsedQuery{
		Query: query,
	}

	queryType := Unknown
	if matches := getSQLVerbRe.FindStringSubmatch(query); len(matches) == 2 {
		lowercaseVerb := strings.ToLower(matches[1])
		var found bool
		queryType, found = kLowerVerbMap[lowercaseVerb]
		if !found {
			queryType = Unknown
		}
	}
	rval.QueryType = queryType

	return rval
}
