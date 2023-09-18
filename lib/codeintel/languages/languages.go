package languages

import (
	"strings"

	"github.com/go-enry/go-enry/v2"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"golang.org/x/exp/slices"
)

// Make sure all names are lowercase here, since they are normalized
var enryLanguageMappings = map[string]string{
	"c#": "c_sharp",
}

func NormalizeLanguage(filetype string) string {
	normalized := strings.ToLower(filetype)
	if mapped, ok := enryLanguageMappings[normalized]; ok {
		normalized = mapped
	}

	return normalized
}

// GetLanguage returns the language for the given path and contents.
func GetLanguage(path, contents string) (lang string, found bool) {
	// Force the use of the shebang.
	if shebangLang, ok := overrideViaShebang(path, contents); ok {
		return shebangLang, true
	}

	// Lastly, fall back to whatever enry decides is a useful algorithm for calculating.

	c := contents
	// classifier is faster on small files without losing much accuracy
	if len(c) > 2048 {
		c = c[:2048]
	}

	lang, err := firstLanguage(enry.GetLanguages(path, []byte(c)))
	if err == nil {
		return NormalizeLanguage(lang), true
	}

	return NormalizeLanguage(lang), false
}

func firstLanguage(languages []string) (string, error) {
	for _, l := range languages {
		if l != "" {
			return l, nil
		}
	}
	return "", errors.New("UnrecognizedLanguage")
}

// overrideViaShebang handles explicitly using the shebang whenever possible.
//
// It also covers some edge cases when enry eagerly returns more languages
// than necessary, which ends up overriding the shebang completely (which,
// IMO is the highest priority match we can have).
//
// For example, enry will return "Perl" and "Pod" for a shebang of `#!/usr/bin/env perl`.
// This is actually unhelpful, because then enry will *not* select "Perl" as the
// language (which is our desired behavior).
func overrideViaShebang(path, content string) (lang string, ok bool) {
	shebangs := enry.GetLanguagesByShebang(path, []byte(content), []string{})
	if len(shebangs) == 0 {
		return "", false
	}

	if len(shebangs) == 1 {
		return shebangs[0], true
	}

	// There are some shebangs that enry returns that are not really
	// useful for our syntax highlighters to distinguish between.
	if slices.Equal(shebangs, []string{"Perl", "Pod"}) {
		return "Perl", true
	}

	return "", false
}
