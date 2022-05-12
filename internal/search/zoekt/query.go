package zoekt

import (
	"github.com/go-enry/go-enry/v2"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	zoekt "github.com/google/zoekt/query"
)

func QueryToZoektQuery(b query.Basic, resultTypes result.Types, feat *search.Features, typ search.IndexedRequestType) (q zoekt.Q, err error) {
	isCaseSensitive := b.IsCaseSensitive()

	if b.Pattern != nil {
		q, err = toZoektPattern(
			b.Pattern,
			isCaseSensitive,
			resultTypes.Has(result.TypeFile),
			resultTypes.Has(result.TypePath),
			typ,
		)
		if err != nil {
			return nil, err
		}
	}

	// Handle file: and -file: filters.
	filesInclude, filesExclude := b.IncludeExcludeValues(query.FieldFile)
	// Handle lang: and -lang: filters.
	langInclude, langExclude := b.IncludeExcludeValues(query.FieldLang)
	filesInclude = append(filesInclude, mapSlice(langInclude, query.LangToFileRegexp)...)
	filesExclude = append(filesExclude, mapSlice(langExclude, query.LangToFileRegexp)...)
	filesReposMustInclude, filesReposMustExclude := b.IncludeExcludeValues(query.FieldRepoHasFile)

	var and []zoekt.Q
	if q != nil {
		and = append(and, q)
	}

	// zoekt also uses regular expressions for file paths
	// TODO PathPatternsAreCaseSensitive
	// TODO whitespace in file path patterns?
	for _, i := range filesInclude {
		q, err := FileRe(i, isCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, q)
	}
	if len(filesExclude) > 0 {
		q, err := FileRe(query.UnionRegExps(filesExclude), isCaseSensitive)
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
	for _, i := range filesReposMustInclude {
		q, err := FileRe(i, isCaseSensitive)
		if err != nil {
			return nil, err
		}
		and = append(and, &zoekt.Type{Type: zoekt.TypeRepo, Child: q})
	}
	for _, i := range filesReposMustExclude {
		q, err := FileRe(i, isCaseSensitive)
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
	if len(langInclude) > 0 && feat.ContentBasedLangFilters {
		or := &zoekt.Or{}
		for _, lang := range langInclude {
			lang, _ = enry.GetLanguageByAlias(lang) // Invariant: lang is valid.
			or.Children = append(or.Children, &zoekt.Language{Language: lang})
		}
		and = append(and, or)
	}

	return zoekt.Simplify(zoekt.NewAnd(and...)), nil
}

func toZoektPattern(
	expression query.Node, isCaseSensitive, patternMatchesContent, patternMatchesPath bool, typ search.IndexedRequestType) (zoekt.Q, error) {
	var fold func(node query.Node) (zoekt.Q, error)
	fold = func(node query.Node) (zoekt.Q, error) {
		switch n := node.(type) {
		case query.Operator:
			children := make([]zoekt.Q, 0, len(n.Operands))
			for _, op := range n.Operands {
				child, err := fold(op)
				if err != nil {
					return nil, err
				}
				children = append(children, child)
			}
			switch n.Kind {
			case query.Or:
				return &zoekt.Or{Children: children}, nil
			case query.And:
				return &zoekt.And{Children: children}, nil
			default:
				// unreachable
				return nil, errors.Errorf("broken invariant: don't know what to do with node %T in toZoektPattern", node)
			}
		case query.Pattern:
			var q zoekt.Q
			var err error

			fileNameOnly := patternMatchesPath && !patternMatchesContent
			contentOnly := !patternMatchesPath && patternMatchesContent

			pattern := n.Value
			if n.Annotation.Labels.IsSet(query.Literal) {
				pattern = regexp.QuoteMeta(pattern)
			}

			q, err = parseRe(pattern, fileNameOnly, contentOnly, isCaseSensitive)
			if err != nil {
				return nil, err
			}

			if typ == search.SymbolRequest && q != nil {
				// Tell zoekt q must match on symbols
				q = &zoekt.Symbol{
					Expr: q,
				}
			}

			if n.Negated {
				q = &zoekt.Not{Child: q}
			}
			return q, nil
		}
		// unreachable
		return nil, errors.Errorf("broken invariant: don't know what to do with node %T in toZoektPattern", node)
	}

	q, err := fold(expression)
	if err != nil {
		return nil, err
	}

	return q, nil
}

func mapSlice(values []string, f func(string) string) []string {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = f(v)
	}
	return result
}
