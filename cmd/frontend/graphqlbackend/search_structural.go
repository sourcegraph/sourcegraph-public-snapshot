package graphqlbackend

import (
	"regexp"
	"regexp/syntax"
	"strings"

	zoektquery "github.com/google/zoekt/query"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

func splitOnHolesPattern() string {
	word := `\w+`
	whitespaceAndOptionalWord := `[ ]+(` + word + `)?`
	holeAnything := `:\[` + word + `\]`
	holeAlphanum := `:\[\[` + word + `\]\]`
	holeWithPunctuation := `:\[` + word + `\.\]`
	holeWithNewline := `:\[` + word + `\\n\]`
	holeWhitespace := `:\[` + whitespaceAndOptionalWord + `\]`
	return strings.Join([]string{
		holeAnything,
		holeAlphanum,
		holeWithPunctuation,
		holeWithNewline,
		holeWhitespace,
	}, "|")
}

var matchHoleRegexp = lazyregexp.New(splitOnHolesPattern())

// Converts comby a structural pattern to a Zoekt regular expression query. It
// converts whitespace in the pattern so that content across newlines can be
// matched in the index. As an incomplete approximation, we use the regex
// pattern .*? to scan ahead.
// Example:
// "ParseInt(:[args]) if err != nil" -> "ParseInt(.*)\s+if\s+err!=\s+nil"
func StructuralPatToRegexpQuery(pattern string) (zoektquery.Q, error) {
	substrings := matchHoleRegexp.Split(pattern, -1)
	var children []zoektquery.Q
	var pieces []string
	for _, s := range substrings {
		piece := regexp.QuoteMeta(s)
		onMatchWhitespace := lazyregexp.New(`[\s]+`)
		piece = onMatchWhitespace.ReplaceAllLiteralString(piece, `[\s]+`)
		pieces = append(pieces, piece)
	}

	if len(pieces) == 0 {
		return &zoektquery.Const{Value: true}, nil
	}
	rs := "(" + strings.Join(pieces, ")(.|\\s)*?(") + ")"
	re, _ := syntax.Parse(rs, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
	children = append(children, &zoektquery.Regexp{
		Regexp:        re,
		CaseSensitive: true,
		Content:       true,
	})
	return &zoektquery.And{Children: children}, nil
}

func StructuralPatToQuery(pattern string) (zoektquery.Q, error) {
	regexpQuery, err := StructuralPatToRegexpQuery(pattern)
	if err != nil {
		return nil, err
	}
	return &zoektquery.Or{Children: []zoektquery.Q{regexpQuery}}, nil
}
