package search

import (
	"regexp"
	"regexp/syntax"
	"strconv"
	"strings"
	"time"

	"github.com/go-enry/go-enry/v2"
	"github.com/go-enry/go-enry/v2/data"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/query"

	zoekt "github.com/google/zoekt/query"
)

// UnionRegExps separates values with a | operator to create a string
// representing a union of regexp patterns.
func UnionRegExps(values []string) string {
	if len(values) == 0 {
		// As a regular expression, "()" and "" are equivalent so this
		// condition wouldn't ordinarily be needed to distinguish these
		// values. But, our internal search engine assumes that ""
		// implies "no regexp" (no values), while "()" implies "match
		// empty regexp" (all values) for file patterns.
		return ""
	}
	if len(values) == 1 {
		// Cosmetic format for regexp value, wherever this happens to be
		// pretty printed.
		return values[0]
	}
	return "(" + strings.Join(values, ")|(") + ")"
}

// filenamesFromLanguage is a map of language name to full filenames
// that are associated with it. This is different from extensions, because
// some languages (like Dockerfile) do not conventionally have an associated
// extension.
var filenamesFromLanguage = func() map[string][]string {
	res := make(map[string][]string, len(data.LanguagesByFilename))
	for filename, languages := range data.LanguagesByFilename {
		for _, language := range languages {
			res[language] = append(res[language], filename)
		}
	}
	return res
}()

// LangToFileRegexp converts a lang: parameter to its corresponding file
// patterns for file filters. The lang value must be valid, cf. validate.go
func LangToFileRegexp(lang string) string {
	lang, _ = enry.GetLanguageByAlias(lang) // Invariant: lang is valid.
	extensions := enry.GetLanguageExtensions(lang)
	patterns := make([]string, len(extensions))
	for i, e := range extensions {
		// Add `\.ext$` pattern to match files with the given extension.
		patterns[i] = regexp.QuoteMeta(e) + "$"
	}
	for _, filename := range filenamesFromLanguage[lang] {
		patterns = append(patterns, "^"+regexp.QuoteMeta(filename)+"$")
	}
	return UnionRegExps(patterns)
}

func mapSlice(values []string, f func(string) string) []string {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = f(v)
	}
	return result
}

func IncludeExcludeValues(q query.Basic, field string) (include, exclude []string) {
	q.VisitParameter(field, func(v string, negated bool, _ query.Annotation) {
		if negated {
			exclude = append(exclude, v)
		} else {
			include = append(include, v)
		}
	})
	return include, exclude
}

func count(q query.Basic, p Protocol) int {
	if count := q.GetCount(); count != "" {
		v, _ := strconv.Atoi(count) // Invariant: count is validated.
		return v
	}

	if q.IsStructural() {
		return DefaultMaxSearchResults
	}

	switch p {
	case Batch:
		return DefaultMaxSearchResults
	case Streaming:
		return DefaultMaxSearchResultsStreaming
	}
	panic("unreachable")
}

type Protocol int

const (
	Streaming Protocol = iota
	Batch
)

// ToTextPatternInfo converts a an atomic query to internal values that drive
// text search. An atomic query is a Basic query where the Pattern is either
// nil, or comprises only one Pattern node (hence, an atom, and not an
// expression). See TextPatternInfo for the values it computes and populates.
func ToTextPatternInfo(q query.Basic, p Protocol, transform query.BasicPass) *TextPatternInfo {
	q = transform(q)
	// Handle file: and -file: filters.
	filesInclude, filesExclude := IncludeExcludeValues(q, query.FieldFile)
	// Handle lang: and -lang: filters.
	langInclude, langExclude := IncludeExcludeValues(q, query.FieldLang)
	filesInclude = append(filesInclude, mapSlice(langInclude, LangToFileRegexp)...)
	filesExclude = append(filesExclude, mapSlice(langExclude, LangToFileRegexp)...)
	filesReposMustInclude, filesReposMustExclude := IncludeExcludeValues(q, query.FieldRepoHasFile)
	selector, _ := filter.SelectPathFromString(q.FindValue(query.FieldSelect)) // Invariant: select is validated
	count := count(q, p)

	// Ugly assumption: for a literal search, the IsRegexp member of
	// TextPatternInfo must be set true. The logic assumes that a literal
	// pattern is an escaped regular expression.
	isRegexp := q.IsLiteral() || q.IsRegexp()

	var pattern string
	if p, ok := q.Pattern.(query.Pattern); ok {
		if q.IsLiteral() {
			// Escape regexp meta characters if this pattern should be treated literally.
			pattern = regexp.QuoteMeta(p.Value)
		} else {
			pattern = p.Value
		}
	}

	if q.Pattern == nil {
		// For compatibility: A nil pattern implies isRegexp is set to
		// true. This has no effect on search logic.
		isRegexp = true
	}

	negated := false
	if p, ok := q.Pattern.(query.Pattern); ok {
		negated = p.Negated
	}

	return &TextPatternInfo{
		// Values dependent on pattern atom.
		IsRegExp:        isRegexp,
		IsStructuralPat: q.IsStructural(),
		IsCaseSensitive: q.IsCaseSensitive(),
		FileMatchLimit:  int32(count),
		Pattern:         pattern,
		IsNegated:       negated,

		// Values dependent on parameters.
		IncludePatterns:              filesInclude,
		ExcludePattern:               UnionRegExps(filesExclude),
		FilePatternsReposMustInclude: filesReposMustInclude,
		FilePatternsReposMustExclude: filesReposMustExclude,
		Languages:                    langInclude,
		PathPatternsAreCaseSensitive: q.IsCaseSensitive(),
		CombyRule:                    q.FindValue(query.FieldCombyRule),
		Index:                        q.Index(),
		Select:                       selector,
	}
}

func TimeoutDuration(b query.Basic) time.Duration {
	d := DefaultTimeout
	maxTimeout := time.Duration(SearchLimits(conf.Get()).MaxTimeoutSeconds) * time.Second
	timeout := b.GetTimeout()
	if timeout != nil {
		d = *timeout
	} else if b.GetCount() != "" {
		// If `count:` is set but `timeout:` is not explicitly set, use the max timeout
		d = maxTimeout
	}
	if d > maxTimeout {
		d = maxTimeout
	}
	return d
}

func fileRe(pattern string, queryIsCaseSensitive bool) (zoekt.Q, error) {
	return parseRe(pattern, true, false, queryIsCaseSensitive)
}

func noOpAnyChar(re *syntax.Regexp) {
	if re.Op == syntax.OpAnyChar {
		re.Op = syntax.OpAnyCharNotNL
	}
	for _, s := range re.Sub {
		noOpAnyChar(s)
	}
}

func parseRe(pattern string, filenameOnly bool, contentOnly bool, queryIsCaseSensitive bool) (zoekt.Q, error) {
	// these are the flags used by zoekt, which differ to searcher.
	re, err := syntax.Parse(pattern, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
	if err != nil {
		return nil, err
	}
	noOpAnyChar(re)
	// zoekt decides to use its literal optimization at the query parser
	// level, so we check if our regex can just be a literal.
	if re.Op == syntax.OpLiteral {
		return &zoekt.Substring{
			Pattern:       string(re.Rune),
			CaseSensitive: queryIsCaseSensitive,
			Content:       contentOnly,
			FileName:      filenameOnly,
		}, nil
	}
	return &zoekt.Regexp{
		Regexp:        re,
		CaseSensitive: queryIsCaseSensitive,
		Content:       contentOnly,
		FileName:      filenameOnly,
	}, nil
}

func QueryToZoektQuery(p *TextPatternInfo, feat *Features, typ IndexedRequestType) (zoekt.Q, error) {
	var and []zoekt.Q

	var q zoekt.Q
	var err error
	if p.IsRegExp {
		fileNameOnly := p.PatternMatchesPath && !p.PatternMatchesContent
		contentOnly := !p.PatternMatchesPath && p.PatternMatchesContent
		q, err = parseRe(p.Pattern, fileNameOnly, contentOnly, p.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
	} else {
		q = &zoekt.Substring{
			Pattern:       p.Pattern,
			CaseSensitive: p.IsCaseSensitive,

			FileName: true,
			Content:  true,
		}
	}

	if p.IsNegated {
		q = &zoekt.Not{Child: q}
	}

	if typ == SymbolRequest {
		// Tell zoekt q must match on symbols
		q = &zoekt.Symbol{
			Expr: q,
		}
	}

	and = append(and, q)

	// zoekt also uses regular expressions for file paths
	// TODO PathPatternsAreCaseSensitive
	// TODO whitespace in file path patterns?
	for _, i := range p.IncludePatterns {
		q, err := fileRe(i, p.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, q)
	}
	if p.ExcludePattern != "" {
		q, err := fileRe(p.ExcludePattern, p.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoekt.Not{Child: q})
	}

	// For conditionals that happen on a repo we can use type:repo queries. eg
	// (type:repo file:foo) (type:repo file:bar) will match all repos which
	// contain a filename matching "foo" and a filename matchinb "bar".
	//
	// Note: (type:repo file:foo file:bar) will only find repos with a
	// filename containing both "foo" and "bar".
	for _, i := range p.FilePatternsReposMustInclude {
		q, err := fileRe(i, p.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoekt.Type{Type: zoekt.TypeRepo, Child: q})
	}
	for _, i := range p.FilePatternsReposMustExclude {
		q, err := fileRe(i, p.IsCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoekt.Not{Child: &zoekt.Type{Type: zoekt.TypeRepo, Child: q}})
	}

	// Languages are already partially expressed with IncludePatterns, but Zoekt creates
	// more precise language metadata based on file contents analyzed by go-enry, so it's
	// useful to pass lang: queries down.
	//
	// Currently, negated lang queries create filename-based ExcludePatterns that cannot be
	// corrected by the more precise language metadata. If this is a problem, indexed search
	// queries should have a special query converter that produces *only* Language predicates
	// instead of filepatterns.
	if len(p.Languages) > 0 && feat.ContentBasedLangFilters {
		or := &zoekt.Or{}
		for _, lang := range p.Languages {
			lang, _ = enry.GetLanguageByAlias(lang) // Invariant: lang is valid.
			or.Children = append(or.Children, &zoekt.Language{Language: lang})
		}
		and = append(and, or)
	}

	return zoekt.Simplify(zoekt.NewAnd(and...)), nil
}
