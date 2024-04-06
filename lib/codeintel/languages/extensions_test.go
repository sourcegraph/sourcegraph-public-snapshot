package languages

import (
	"testing"

	"github.com/go-enry/go-enry/v2"
	"github.com/stretchr/testify/require"
)

func TestOverrideExtensions(t *testing.T) {
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

func TestNonAmbiguousExtensions(t *testing.T) {
	// Languages/extensions that we don't want to regress
	nonAmbiguousExtensionsCheck := map[string]string{
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

	for ext, language := range nonAmbiguousExtensionsCheck {
		filename := "foo" + ext
		languages, isLikelyBinaryFile := getLanguagesByExtension(filename)
		require.False(t, isLikelyBinaryFile)
		require.Equal(t, []string{language}, languages,
			"If this test fails when updating enry, maybe `overrideAmbiguousExtensionsMap` needs updating")
	}
}

func TestBinaryExtensions(t *testing.T) {
	for _, ext := range []string{".png", ".jpg", ".gif"} {
		filename := "foo" + ext
		_, isLikelyBinary := getLanguagesByExtension(filename)
		require.Truef(t, isLikelyBinary, "filename: %v was not guessed to be binary;"+
			"bug in extension matching logic in getLanguagesByExtension maybe?",
			filename)
	}
}
