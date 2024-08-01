package languages

import (
	"slices"
	"strings"
	"testing"

	"github.com/go-enry/go-enry/v2"               //nolint:depguard - This package is allowed to use enry
	enrydata "github.com/go-enry/go-enry/v2/data" //nolint:depguard - This package is allowed to use enry
	"github.com/stretchr/testify/require"
)

// Languages/extensions that we don't want to regress
var nonAmbiguousExtensionsCheck = map[string]string{
	".apex":    "Apex",
	".apxt":    "Apex",
	".apxc":    "Apex",
	".cls":     "Apex",
	".trigger": "Apex",
	".js":      "JavaScript",
	// Linguist removed JSX (but not TSX) as a separate language:
	// https://github.com/github-linguist/linguist/pull/5133
	".jsx":   "JavaScript",
	".ts":    "TypeScript",
	".tsx":   "TSX",
	".py":    "Python",
	".rb":    "Ruby",
	".go":    "Go",
	".java":  "Java",
	".kt":    "Kotlin",
	".magik": "Magik",
	".scala": "Scala",
	".cs":    "C#",
	".fs":    "F#",
	".rs":    "Rust",
	".c":     "C",
	".cpp":   "C++",
	".cxx":   "C++",
	".hpp":   "C++",
	".hxx":   "C++",
	".lua":   "Lua",
	".dart":  "Dart",
	".swift": "Swift",
	".css":   "CSS",
	".json":  "JSON",
	".yml":   "YAML",
	".xml":   "XML",
	".pkl":   "Pkl",
}

func TestGetLanguageByAlias_UnsupportedLanguages(t *testing.T) {
	for alias, name := range unsupportedByEnryAliasMap {
		resName, _ := GetLanguageByNameOrAlias(alias)
		require.Equal(t, name, resName,
			"maybe a typo in `unsupportedByEnryAliasMap`?")
	}
}

func TestGetLanguageByAlias_NonAmbiguousLanguages(t *testing.T) {
	for _, language := range nonAmbiguousExtensionsCheck {
		_, ok := GetLanguageByNameOrAlias(language)
		require.True(t, ok,
			"unable to find language %s in go-enry", language)
	}
}

func TestGetLanguageExtensions_UnsupportedExtensions(t *testing.T) {
	for language, exts := range unsupportedByEnryNameToExtensionMap {
		extensions := GetLanguageExtensions(language)
		for _, ext := range exts {
			require.Contains(t, extensions, ext,
				"maybe a typo in `unsupportedByEnryNameToExtensionMap`?")
		}
	}
}

func TestGetLanguageExtensions_NonAmbiguousExtensions(t *testing.T) {
	langMap := reverseMap(nonAmbiguousExtensionsCheck)
	for language, exts := range langMap {
		extensions := GetLanguageExtensions(language)
		for _, ext := range exts {
			require.Contains(t, extensions, ext,
				"If this test fails when updating enry, maybe `overrideAmbiguousExtensionsMap` needs updating")
		}
	}
}

func TestGetLanguagesByExtension_UnsupportedExtensions(t *testing.T) {
	for ext, language := range unsupportedByEnryExtensionToNameMap {
		filename := "foo" + ext
		languages, _ := getLanguagesByExtension(filename)
		require.Contains(t, languages, language,
			"maybe a typo in `unsupportedByEnryExtensionToNameMap`?")
	}
}

func TestGetLanguagesByExtension_OverrideExtensions(t *testing.T) {
	for ext, language := range overrideAmbiguousExtensionsMap {
		filename := "foo" + ext
		enryLangs := enry.GetLanguagesByExtension(filename, nil, nil)
		require.Contains(t, enryLangs, language,
			"maybe a typo in `overrideAmbiguousExtensionsMap`?")
		require.Greaterf(t, len(enryLangs), 1,
			"extension %v is not ambiguous according to enry, remove it from `overrideAmbiguousExtensionsMap`",
			ext)
	}
}

func TestGetLanguagesByExtension_NonAmbiguousExtensions(t *testing.T) {
	for ext, language := range nonAmbiguousExtensionsCheck {
		filename := "foo" + ext
		languages, isLikelyBinaryFile := getLanguagesByExtension(filename)
		require.False(t, isLikelyBinaryFile)
		require.Equal(t, []string{language}, languages,
			"If this test fails when updating enry, maybe `overrideAmbiguousExtensionsMap` needs updating")
	}
}

func TestGetLanguagesByExtension_BinaryExtensions(t *testing.T) {
	for _, ext := range []string{".png", ".jpg", ".gif"} {
		filename := "foo" + ext
		_, isLikelyBinary := getLanguagesByExtension(filename)
		require.Truef(t, isLikelyBinary, "filename: %v was not guessed to be binary;"+
			"bug in extension matching logic in getLanguagesByExtension maybe?",
			filename)
	}
}

func TestExtensionsConsistency(t *testing.T) {
	for ext, overrideLang := range overrideAmbiguousExtensionsMap {
		filepath := "foo" + ext
		enryLangsForExt := enry.GetLanguagesByExtension(filepath, nil, nil)
		require.Containsf(t, enryLangsForExt, overrideLang, "overrideAmbiguousExtensionsMap maps extension %q to language %q but "+
			"that mapping is not present in enry's list %v", ext, overrideLang, enryLangsForExt)
		require.Greaterf(t, len(enryLangsForExt), 1, "overrideAmbiguousExtensionsMap states that"+
			"%q extension is ambiguous, but only found langs: %v", ext, enryLangsForExt)

		candidates, isLikelyBinary := getLanguagesByExtension(filepath)
		require.False(t, isLikelyBinary, "ambiguous files are all source code")
		require.True(t, len(candidates) == 1, "getLanguagesByExtension should respect overrideAmbiguousExtensionsMap")

		shouldBeIgnoredLangsForExt := slices.DeleteFunc(enryLangsForExt, func(s string) bool {
			return s == overrideLang
		})
		for _, shouldBeIgnoredLang := range shouldBeIgnoredLangsForExt {
			ignoredExts, found := nicheExtensionUsages[shouldBeIgnoredLang]
			require.Truef(t, found, "expected lang: %q to have an entry in nicheExtensionUsages for consistency with GetLanguagesByExtension", shouldBeIgnoredLang)
			require.Truef(t, len(ignoredExts) >= 1, "sets in nicheExtensionUsages must be non-empty")

			nonNicheExts := GetLanguageExtensions(shouldBeIgnoredLang)
			for ignoredExt, _ := range ignoredExts {
				require.Falsef(t, slices.Contains(nonNicheExts, ignoredExt),
					"GetLanguageExtensions should not return %q for lang %q for consistency with GetLanguagesByExtension",
					ignoredExt, shouldBeIgnoredLang)
			}
		}
	}
}

func TestExtensionsConsistency2(t *testing.T) {
	for lang, _ := range enrydata.ExtensionsByLanguage {
		for _, ext := range GetLanguageExtensions(lang) {
			if strings.Count(ext, ".") > 1 {
				// Ignore unusual edge cases like .coffee.md for Literate CoffeeScript
				continue
			}
			langsByExt, isLikelyBinary := getLanguagesByExtension("foo" + ext)
			if !isLikelyBinary {
				require.Truef(t, slices.Contains(langsByExt, lang),
					"expected getLanguagesByExtension result %v to contain %q (extension: %q)", langsByExt, lang, ext)
			}
		}
	}
}

// TODO(id: remove-pkl-special-case) Linguist v7.30.0 adds support for Pkl,
// so when we upgrade to a matching go-enry version, we can remove special
// cases for Pkl.
func TestUnsupportedByEnry(t *testing.T) {
	for lang := range unsupportedByEnryNameToExtensionMap {
		enry_extensions, found := enrydata.ExtensionsByLanguage[lang]
		if found {
			validateLanguageAgainstGoEnry(t, "unsupportedByEnryNameToExtensionMap", enry_extensions, lang)
		}
	}
	for _, lang := range unsupportedByEnryAliasMap {
		enry_extensions, found := enrydata.ExtensionsByLanguage[lang]
		if found {
			validateLanguageAgainstGoEnry(t, "unsupportedByEnryAliasMap", enry_extensions, lang)
		}
	}
	for _, lang := range unsupportedByEnryExtensionToNameMap {
		enry_extensions, found := enrydata.ExtensionsByLanguage[lang]
		if found {
			validateLanguageAgainstGoEnry(t, "unsupportedByEnryExtensionToNameMap", enry_extensions, lang)
		}
	}
}

func validateLanguageAgainstGoEnry(t *testing.T, name string, enryExtensions []string, lang string) {
	enryExtensions = slices.Clone(enryExtensions)
	slices.Sort(enryExtensions)
	sgExtensions := slices.Clone(unsupportedByEnryNameToExtensionMap[lang])
	slices.Sort(sgExtensions)

	require.NotEqualf(t, enryExtensions, sgExtensions, "looks like language %q is supported by enry with the same extensions; remove it from %q", lang, name)
}
