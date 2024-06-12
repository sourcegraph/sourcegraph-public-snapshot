package languages

import (
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
