package zoekt

import (
	"github.com/go-enry/go-enry/v2"
	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/searchcontexts"
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

func IsGlobal(op search.RepoOptions) bool {
	// We do not do global searches if a repo: filter was specified. I
	// (@camdencheek) could not find any documentation or historical reasons
	// for why this is, so I'm going to speculate here for future wanderers.
	//
	// If a user specifies a single repo, that repo may or may not be indexed
	// but we still want to search it. A Zoekt search will not tell us that a
	// search returned no results because the repo filtered to was unindexed,
	// it will just return no results.
	//
	// Additionally, if a user specifies a repo: filter, they are likely
	// targeting only a few repos, so the benefits of running a filtered global
	// search vs just paging over the few repos that match the query are
	// probably do not outweigh the cost of potentially skipping unindexed
	// repos.
	//
	// We see this assumption break down with filters like `repo:github.com/`
	// or `repo:.*`, in which case a global search would be much faster than
	// paging through all the repos.
	if len(op.RepoFilters) > 0 {
		return false
	}

	// Zoekt does not know about repo descriptions, so we depend on the
	// database to handle this filter.
	if len(op.DescriptionPatterns) > 0 {
		return false
	}

	// If a search context is specified, we do not know ahead of time whether
	// the repos in the context are indexed and we need to go through the repo
	// resolution process.
	if !searchcontexts.IsGlobalSearchContextSpec(op.SearchContextSpec) {
		return false
	}

	// repo:has.commit.after() is handled during the repo resolution step,
	// and we cannot depend on Zoekt for this information.
	if op.CommitAfter != "" {
		return false
	}

	// There should be no cursors when calling this, but if there are that
	// means we're already paginating. Cursors should probably not live on this
	// struct since they are an implementation detail of pagination.
	if len(op.Cursors) > 0 {
		return false
	}

	// If indexed search is explicitly disabled, that implicitly means global
	// search is also disabled since global search means Zoekt.
	if op.UseIndex == query.No {
		return false
	}

	// For now, we handle all repo:has.file and repo:has.content during repo
	// pagination. Zoekt can handle this, so we should push this down to Zoekt
	// and allow global search with these filters.
	if len(op.HasFileContent) > 0 {
		return false
	}

	// All the fields not mentioned above can be handled by Zoekt global search.
	// Listing them here for posterity:
	// - MinusRepoFilters
	// - CaseSensitiveRepoFilters
	// - Visibility
	// - Limit
	// - ForkSet
	// - NoForks
	// - OnlyForks
	// - OnlyCloned
	// - ArchivedSet
	// - NoArchived
	// - OnlyArchived
	return true
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
