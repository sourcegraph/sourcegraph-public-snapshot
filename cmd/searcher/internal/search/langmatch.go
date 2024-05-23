package search

import (
	"slices"

	"github.com/sourcegraph/sourcegraph/internal/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/languages"
)

// langMatcher checks whether a file matches the 'include' and 'exclude
// language filters
type langMatcher interface {
	// Matches checks whether a file's language matches. It accepts a callback
	// to fetch the content to avoid loading content when it's not needed. It
	// returns whether the file matches, plus the detected language when available.
	Matches(path string, getContent func() ([]byte, error)) (bool, string)
}

type allLangMatcher struct{}

func (am *allLangMatcher) Matches(_ string, _ func() ([]byte, error)) (bool, string) {
	return true, ""
}

// enryLangMatcher uses go-ery to check whether files match the language
// filters. IncludeLangs and ExcludeLangs are expected to be valid languages,
// like from the output of enry.GetLanguagesByAlias.
type enryLangMatcher struct {
	IncludeLangs []string
	ExcludeLangs []string
	OnlyExcludes bool
}

func (em *enryLangMatcher) Matches(path string, getContent func() ([]byte, error)) (bool, string) {
	// We use Sourcegraph's wrapper around enry because it supports lazily fetching
	// content and contains some optimizations for ambiguous extensions.
	langs, err := languages.GetLanguages(path, getContent)

	// In practice err will always be nil, because we never error when fetching content
	if err != nil {
		return false, ""
	}

	// It's fine if file has no detected language, as long as there are no include filters
	if len(langs) == 0 {
		return em.OnlyExcludes, ""
	}

	// Choose the most likely language
	lang := langs[0]
	for _, includeLang := range em.IncludeLangs {
		if lang != includeLang {
			return false, ""
		}
	}
	return !slices.Contains(em.ExcludeLangs, lang), lang
}

func toLangMatcher(p *protocol.PatternInfo) langMatcher {
	// If there are no language filters, avoid all path and content checks.
	if len(p.IncludeLangs)+len(p.ExcludeLangs) == 0 {
		return &allLangMatcher{}
	}

	return &enryLangMatcher{
		IncludeLangs: p.IncludeLangs,
		ExcludeLangs: p.ExcludeLangs,
		OnlyExcludes: len(p.IncludeLangs) == 0,
	}
}
