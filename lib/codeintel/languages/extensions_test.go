package languages

import (
	"slices"
	"testing"

	"github.com/go-enry/go-enry/v2"
	"github.com/stretchr/testify/require"
)

// Languages/extensions that we don't want to regress
var nonAmbiguousExtensionsCheck = map[string]string{
	".js": "JavaScript",
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
		resName, _ := GetLanguageByAlias(alias)
		require.Equal(t, name, resName,
			"maybe a typo in `unsupportedByEnryAliasMap`?")
	}
}

func TestGetLanguageByAlias_NonAmbiguousLanguages(t *testing.T) {
	for _, language := range nonAmbiguousExtensionsCheck {
		_, ok := GetLanguageByAlias(language)
		require.True(t, ok,
			"unable to find language %s in go-enry", language)
	}
}

func TestGetLanguageExtensions_UnsupportedExtensions(t *testing.T) {
	for language, ext := range unsupportedByEnryNameToExtensionMap {
		extensions := GetLanguageExtensions(language)
		require.Contains(t, extensions, ext,
			"maybe a typo in `unsupportedByEnryNameToExtensionMap`?")
	}
}

func TestGetLanguageExtensions_NonAmbiguousExtensions(t *testing.T) {
	langMap := reverseMap(nonAmbiguousExtensionsCheck)
	for language, ext := range langMap {
		extensions := GetLanguageExtensions(language)
		require.Contains(t, extensions, ext,
			"If this test fails when updating enry, maybe `overrideAmbiguousExtensionsMap` needs updating")
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

func TestGetLanguageExtensions_Consistency(t *testing.T) {
	for ext, overrideLang := range overrideAmbiguousExtensionsMap {
		filepath := "foo" + ext
		langs := enry.GetLanguagesByExtension(filepath, nil, nil)
		require.Containsf(t, langs, overrideLang, "overrideAmbiguousExtensionsMap maps extension %q to language %q but "+
			"that mapping is not present in enry's list %v", ext, overrideLang, langs)
		require.Greaterf(t, len(langs), 1, "overrideAmbiguousExtensionsMap states that"+
			"%q extension is ambiguous, but only found langs: %v", ext, langs)
		langs = slices.DeleteFunc(langs, func(s string) bool {
			return s == overrideLang
		})
		candidates, isLikelyBinary := getLanguagesByExtension(filepath)
		require.False(t, isLikelyBinary, "ambiguous files are all source code")
		require.True(t, len(candidates) == 1, "getLanguagesByExtension should respect overrideAmbiguousExtensionsMap")
		for _, otherLang := range langs {
			// This is currently wrong.
			require.NotEqualf(t, len(GetLanguageExtensions(otherLang)), 0,
				"GetLanguageExtensions returned %q for alias %q, which is inconsistent with getLanguagesByExtension"+
					"only returning language %q for %[1]q", ext, otherLang, overrideLang)
		}
	}
}
