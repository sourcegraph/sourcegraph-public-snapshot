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
// 	ts_rank_cd(...) + ts_rank_cd(...)...
//
// Which determines the effective rank of the query string relative to the given tsvector column name.
//
// subStringMatches indicates whether or not lexemes should be substring-matched, e.g. if "htt" should
// match "http". This applies only at a lexeme level. The value should match the value provided to
// TextSearchQuery when composing the tsquery for ranking to match properly.
func TextSearchRank(columnName, query string, subStringMatches bool) *sqlf.Query {
	var rankFunctions []*sqlf.Query
	for _, term := range strings.Fields(query) {
		seq := lexemeSequence(term, subStringMatches)
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
// Note: This composing a tsquery _value_ and so using fmt.Sprintf instead of sqlf.Sprintf is
// correct here.
func lexemeSequence(s string, subStringMatches bool) string {
	lexemes := Lexemes(s)
	sequence := make([]string, 0, len(lexemes))
	for _, lexeme := range lexemes {
		if subStringMatches {
			sequence = append(sequence, fmt.Sprintf("%s:*", lexeme))
		} else {
			sequence = append(sequence, lexeme)
		}
	}
	return strings.Join(sequence, " <-> ")
}

// TextSearchQuery returns an SQL expression of e.g. the form:
//
// 	column_name @@ ... OR column_name @@ ... OR column_name @@ ...
//
// Which can be used in a WHERE clause to match the given query string against the provided tsvector column
// name.
//
// subStringMatches indicates whether or not lexemes should be substring-matched, e.g. if "htt" should
// match "http". This applies only at a lexeme level. The value should match the value provided to
// TextSearchQuery when composing the tsquery for ranking to match properly.
func TextSearchQuery(columnName, query string, subStringMatches bool) *sqlf.Query {
	// For every term in the query string, produce the lexeme sequence that would match it. e.g.
	// "gorilla/mux Router" -> [`gorilla:* <-> /:* <-> mux:*`, `Router:*`]
	terms := strings.Fields(query)
	termLexemeSequences := make([]string, 0, len(terms))
	for _, term := range terms {
		termLexemeSequences = append(termLexemeSequences, lexemeSequence(term, subStringMatches))
	}

	// Build expressions that would match all the query terms in sequence, with some distance of
	// lexemes between them. Note that the tsquery `foo <-> bar` matches foo _exactly_ followed by
	// bar, and `foo <1> bar` matches _exactly_ foo followed by one lexeme and then bar. There is
	// no way in Postgres today to specify a range of lexemes between, or a wildcard (unknown number
	// of lexemes between). https://stackoverflow.com/a/59146601
	//
	// "<->" (no distance) enables exact search terms like "http.StatusNotFound" to match lexemes ['http', '.', 'StatusNotFound']
	// "<2>" enables search terms like "http StatusNotFound" (missing period) to match lexemes ['http', '.', 'StatusNotFound']
	// "<4>" enables search terms like "json Decode" (missing "Decoder") to match lexemes ['json', 'Decoder', 'Decode']
	// "<5>" enables search terms like "Player Run" (missing "::") to match lexemes ['Player', ':', ':', 'Run]
	//
	// Note that more distance != better, the greater the distance the worse relevance of results.
	distances := []string{" <-> ", " <2> ", " <4> ", " <5> "}
	expressions := make([]*sqlf.Query, 0, len(distances))
	for _, distance := range distances {
		expressions = append(expressions, sqlf.Sprintf(columnName+" @@ %s", strings.Join(termLexemeSequences, distance)))
	}
	return sqlf.Join(expressions, "OR")
}

// Query describes an API docs search query.
type Query struct {
	// MetaTerms are the terms that should be matched against tags, repo names, and language name
	// metadata.
	MetaTerms string

	// MainTerms are the terms that should be matched against the search key and label.
	MainTerms string

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
	return q
}
