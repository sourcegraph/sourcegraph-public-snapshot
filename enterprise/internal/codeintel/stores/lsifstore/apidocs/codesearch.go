package apidocs

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/keegancsmith/sqlf"
)

// Lexemes splits the string into lexemes, each will be any contiguous section of Unicode digits,
// numbers, and letters. All other unicode runes, such as punctuation, are considered their own
// individual lexemes - and spaces are removed and considered boundaries.
func Lexemes(s string) []string {
	var (
		lexemes       []string
		currentLexeme []rune
	)
	for _, r := range s {
		if unicode.IsDigit(r) || unicode.IsNumber(r) || unicode.IsLetter(r) {
			currentLexeme = append(currentLexeme, r)
			continue
		}
		if len(currentLexeme) > 0 {
			lexemes = append(lexemes, string(currentLexeme))
			currentLexeme = currentLexeme[:0]
		}
		if !unicode.IsSpace(r) {
			lexemes = append(lexemes, string(r))
		}
	}
	if len(currentLexeme) > 0 {
		lexemes = append(lexemes, string(currentLexeme))
	}
	return lexemes
}

// TextSearchVector constructs an ordered tsvector from the given string.
//
// Postgres' built in to_tsvector configurations (`simple`, `english`, etc.) work well for human
// language search but for multiple reasons produces tsvectors that are not appropriate for our
// use case of code search.
//
// By default, tsvectors perform word deduplication and normalization of words (Rat -> rat for
// example.) They also get sorted alphabetically:
//
// ```
// SELECT 'a fat cat sat on a mat and ate a fat rat'::tsvector;
//                       tsvector
// ----------------------------------------------------
//  'a' 'and' 'ate' 'cat' 'fat' 'mat' 'on' 'rat' 'sat'
// ```
//
// In the context of general document search, this doesn't matter. But in our context of API docs
// search, the order in which words (in the general computing sense) appear matters. For example,
// when searching `mux.router` it's important we match (package mux, symbol router) and not
// (package router, symbol mux).
//
// Another critical reason to_tsvector's configurations are not suitable for codes search is that
// they explicitly drop most punctuation (excluding periods) and don't split words between periods:
//
// ```
// select to_tsvector('foo::bar mux.Router const Foo* = Bar<T<X>>');
//                  to_tsvector
// ----------------------------------------------
//  'bar':2,6 'const':4 'foo':1,5 'mux.router':3
// ```
//
// Luckily, Postgres allows one to construct tsvectors manually by providing a sorted list of lexemes
// with optional integer _positions_ and weights, or:
//
// ```
// SELECT $$'foo':1 '::':2 'bar':3 'mux':4 'Router':5 'const':6 'Foo':7 '*':8 '=':9 'Bar':10 '<':11 'T':12 '<':13 'X':14 '>':15 '>':16$$::tsvector;
//                                                       tsvector
// --------------------------------------------------------------------------------------------------------------------
//  '*':8 '::':2 '<':11,13 '=':9 '>':15,16 'Bar':10 'Foo':7 'Router':5 'T':12 'X':14 'bar':3 'const':6 'foo':1 'mux':4
// ```
//
// Ordered, case-sensitive, punctuation-inclusive tsquery matches against the tsvector are then possible.
//
// For example, a query for `bufio.` would then match a tsvector ("bufio", ".", "Reader", ".", "writeBuf"):
//
// ```
// SELECT $$'bufio':1 '.':2 'Reader':3 '.':4 'writeBuf':5$$::tsvector @@ tsquery_phrase($$'bufio'$$::tsquery, $$'.'$$::tsquery) AS matches;
//  matches
// ---------
//  t
// ```
//
func TextSearchVector(s string) string {
	// We need to emit a string in the Postgres tsvector format, roughly:
	//
	//     lexeme1:1 lexeme2:2 lexeme3:3
	//
	lexemes := Lexemes(s)
	pairs := make([]string, 0, len(lexemes))
	for i, lexeme := range lexemes {
		pairs = append(pairs, fmt.Sprintf("%s:%v", lexeme, i+1))
	}
	return strings.Join(pairs, " ")
}

// TextSearchRank returns an SQL expression of the form:
//
// 	(ts_rank_cd(...) + ts_rank_cd(...)...)
//
// Which determines the effective rank of the query string relative to the given tsvector column name.
//
// subStringMatches indicates whether or not lexemes should be substring-matched, e.g. if "htt" should
// match "http". This applies only at a lexeme level. The value should match the value provided to
// TextSearchQuery when composing the tsquery for ranking to match properly.
func TextSearchRank(columnName, query string, subStringMatches bool) *sqlf.Query {
	terms := strings.Fields(query)
	if len(terms) == 0 {
		return sqlf.Sprintf("0") // lowest rank
	}

	var rankFunctions []*sqlf.Query
	for _, term := range terms {
		seq := lexemeSequence(Lexemes(term), subStringMatches, true, " <-> ")
		rankFunctions = append(rankFunctions, sqlf.Sprintf("ts_rank_cd("+columnName+", %s, 2)", seq))
	}
	return sqlf.Join(rankFunctions, "+")
}

// Composes a tsquery string which matches the lexemes for the given string in order, e.g.:
//
// 	gorilla:* <-> /:* <-> mux:*
//
// If subStringMatches is true, the tsquery `:*` operator ("match prefix") is applied.
//
// If substringMatchLastLexemeOnly is true, only the last lexeme has the tsquery `:*` operator
// applied.
//
// Note: This composing a tsquery _value_ and so using fmt.Sprintf instead of sqlf.Sprintf is
// correct here.
func lexemeSequence(lexemes []string, subStringMatches, substringMatchLastLexemeOnly bool, distance string) string {
	sequence := make([]string, 0, len(lexemes))
	for i, lexeme := range lexemes {
		if subStringMatches && (!substringMatchLastLexemeOnly || i == len(lexemes)-1) {
			sequence = append(sequence, fmt.Sprintf("%s:*", lexeme))
		} else {
			sequence = append(sequence, lexeme)
		}
	}
	return strings.Join(sequence, distance)
}

// TextSearchQuery returns an SQL expression of e.g. the form:
//
// 	(column_name @@ ... OR column_name @@ ... OR column_name @@ ...)
//
// Which can be used in a WHERE clause to match the given query string against the provided tsvector column
// name.
//
// subStringMatches indicates whether or not lexemes should be substring-matched, e.g. if "htt" should
// match "http". This applies only at a lexeme level. The value should match the value provided to
// TextSearchQuery when composing the tsquery for ranking to match properly.
func TextSearchQuery(columnName, query string, subStringMatches bool) *sqlf.Query {
	terms := strings.Fields(query)
	if len(terms) == 0 {
		return sqlf.Sprintf("false")
	}
	expressions := make([]*sqlf.Query, 0, len(terms))

	contiguousDisjointed := make([][]string, 0, len(terms)/2)
	currentDisjointed := []string{}
	for _, term := range terms {
		lexemes := Lexemes(term)

		// Firstly, for query terms like ["golang/go", "http.StatusNotFound"] we've got a pretty specific
		// thing we're looking for. Our terms are not disjointed, they're specific. We know this because
		// the number of lexemes in each term is >1 (["golang", "/", "go"], ["http", ".", "StatusNotFound"])
		//
		// In this case, we would like to emit an expression to match any of these terms' lexemes in
		// sequence ("golang/go" with no lexemes between, "http.StatusNotFound" with no lexemes
		// between)
		if len(lexemes) > 1 {
			sequence := lexemeSequence(lexemes, subStringMatches, true, " <-> ")
			expressions = append(expressions, sqlf.Sprintf(columnName+" @@ %s", sequence))

			if len(currentDisjointed) > 0 {
				contiguousDisjointed = append(contiguousDisjointed, currentDisjointed)
				currentDisjointed = currentDisjointed[:0]
			}
			continue
		}

		// Secondly, there may be query terms like ["http", "StatusNotFound"] where the user has
		// space-separated something (e.g. "http.StatusNotFound"), perhaps with desire for a more
		// fuzzy match. A query "json Decode" for example should match the search key "json.Decoder.Decode"
		// ideally.
		//
		// Note that the tsquery `foo <-> bar` matches foo _exactly_ followed by bar, and `foo <1> bar`
		// matches _exactly_ foo followed by one lexeme and then bar. There is no way in Postgres today
		// to specify a range of lexemes between, or a wildcard (unknown number of lexemes between). https://stackoverflow.com/a/59146601
		//
		// So, to support this fuzzier interpretation we can emit multiple expressions with multiple
		// distance operators between the lexemes.:
		//
		// 	json <-> Decode ("json", "Decode")
		// 	json <2> Decode ("json", any lexeme, "Decode")
		// 	json <4> Decode ("json", any lexeme, any lexeme, "Decode")
		// 	json <5> Decode ("json", any lexeme, any lexeme, any lexeme, "Decode")
		//
		// The choice of distances is as follows:
		//
		// "<->" (no distance) enables exact search terms like "http.StatusNotFound" to match lexemes ['http', '.', 'StatusNotFound']
		// "<2>" enables search terms like "http StatusNotFound" (missing period) to match lexemes ['http', '.', 'StatusNotFound']
		// "<4>" enables search terms like "Player Run" (missing "::") to match lexemes ['Player', ':', ':', 'Run]
		// "<5>" enables search terms like "json Decode" (missing "Decoder" and periods) to match lexemes ['json', '.', 'Decoder', '.', 'Decode']
		//
		// Note that more distance != better, the greater the distance the worse relevance of
		// results.
		//
		// Another problem when this happens is that these query terms are _too_ simple, and often
		// match too many records. For example the search query "http StatusNotFound" could match
		// every Go package with "http" in the name, "Router Error" could match every "Error" type,
		// etc. So the first thing we do is collect such disjointed simple terms, and later emit the
		// varying-distance expressions between the contiguous sets.
		currentDisjointed = append(currentDisjointed, lexemes...)
	}
	if len(currentDisjointed) > 0 {
		contiguousDisjointed = append(contiguousDisjointed, currentDisjointed)
	}

	distances := []string{" <-> ", " <2> ", " <4> ", " <5> "}
	for _, disjointedTermLexemes := range contiguousDisjointed {
		// If only one lexeme, it won't have multiple distances.
		if len(disjointedTermLexemes) == 1 {
			tsquery := lexemeSequence(disjointedTermLexemes, subStringMatches, false, "")
			expressions = append(expressions, sqlf.Sprintf(columnName+" @@ %s", tsquery))
		} else {
			for _, distance := range distances {
				tsquery := lexemeSequence(disjointedTermLexemes, subStringMatches, false, distance)
				expressions = append(expressions, sqlf.Sprintf(columnName+" @@ %s", tsquery))
			}
		}
	}
	return sqlf.Sprintf("(%s)", sqlf.Join(expressions, "OR"))
}

// RepoSearchQuery returns an SQL expression of e.g. the form:
//
// 	(column_name @@ ... OR column_name @@ ... OR column_name @@ ...)
//
// Which can be used in a WHERE clause to match any of the given query repositories against the
// provided tsvector column name.
//
// Repo search queries in practice have much stricter matching than TextSearchQuery, because the
// risks of matching a repository (and filtering results to just that repo, or few repos) are in
// practice worse. With text search queries, you want them to be a bit fuzzy. With repo search
// queries, you really only want to match repos if you're pretty sure that's what the user meant.
func RepoSearchQuery(columnName string, possibleRepoNames []string) *sqlf.Query {
	if len(possibleRepoNames) == 0 {
		return sqlf.Sprintf("false") // match no repo names
	}
	expressions := make([]*sqlf.Query, 0, len(possibleRepoNames))
	for _, repoName := range possibleRepoNames {
		expressions = append(expressions, sqlf.Sprintf(columnName+" @@ %s", lexemeSequence(Lexemes(repoName), false, false, " <-> ")))
	}
	return sqlf.Sprintf("(%s)", sqlf.Join(expressions, "OR"))
}

// Query describes an API docs search query.
type Query struct {
	// MetaTerms are the terms that should be matched against tags, repo names, and language name
	// metadata.
	MetaTerms string

	// MainTerms are the terms that should be matched against the search key and label.
	MainTerms string

	// PossibleRepos are extracted query terms that are possibly repositories. These are any query
	// terms separated with a slash. If metadata terms were separated, this will only contain
	// possible repos from the metadata terms. i.e.:
	//
	// 	golang/go: net/http
	//
	// Will have PossibleRepos = ["golang/go"], not ["golang/go", "net/http"].
	PossibleRepos []string

	// Whether or not lexemes should match substrings, e.g. if "gith.com/sourcegraph/source" should
	// be matched as "*gith*.*com*/*sourcegraph*/*source*" (* being a wildcard)
	SubStringMatches bool
}

// ParseQuery parses an API docs search query. We support two forms:
//
// 	<metadata>:<query terms>
// 	<query terms and metadata>
//
// That is, if the user jumbling all the query terms and metadata (repo names, tags, language name,
// etc.) into the query does not yield good results, they can prefix their query with metadata and
// a colon, for queries like:
//
// 	golang/go: net/http
// 	go package: mux
// 	go private variable my/repo: mux
//
// This is a stupid-simple form for people to get more specific results, and may be the only
// "advanced" syntax API docs search will ever support because we want it to have stupid simple
/// syntax.
func ParseQuery(query string) Query {
	q := Query{
		MetaTerms:        query,
		MainTerms:        query,
		SubStringMatches: true,
	}
	if i := strings.Index(query, ":"); i != -1 {
		q.MetaTerms = strings.TrimSpace(query[:i])
		q.MainTerms = strings.TrimSpace(query[i+len(":"):])
	}

	// Identify possible repos in the query terms.
	for _, term := range strings.Fields(q.MetaTerms) {
		if strings.Contains(term, "/") {
			q.PossibleRepos = append(q.PossibleRepos, term)
		}
	}
	return q
}
