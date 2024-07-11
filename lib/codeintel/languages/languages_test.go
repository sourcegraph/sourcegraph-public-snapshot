package languages

import (
	"testing"

	"github.com/go-enry/go-enry/v2" //nolint:depguard - This package is allowed to use enry
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestGetLanguages(t *testing.T) {
	const matlabContent = "function [out] = square(x)\nout = x * x;\nend"
	const mathematicaContent = "f[x_] := x ^ 2\ng[y_] := f[y]"
	const cppContent = "namespace x { }"
	const cContent = "typedef struct { int x; } Int;"
	const emptyContent = ""

	testCases := []struct {
		path              string
		content           string
		expectedLanguages []string
		compareFirstOnly  bool
	}{
		{path: "perlscript", content: "#!/usr/bin/env perl\n$version = $ARGV[0];", expectedLanguages: []string{"Perl"}},
		{path: "rakuscript", content: "#!/usr/bin/env perl6\n$version = $ARGV[0];", expectedLanguages: []string{"Raku"}},
		{path: "ambiguous.h", content: emptyContent, expectedLanguages: []string{"C", "C++", "Objective-C"}},
		{path: "cpp.h", content: cppContent, expectedLanguages: []string{"C++"}},
		{path: "c.h", content: cContent, expectedLanguages: []string{"C"}},
		{path: "matlab.m", content: matlabContent, expectedLanguages: []string{"MATLAB"}, compareFirstOnly: true},
		{path: "mathematica.m", content: mathematicaContent, expectedLanguages: []string{"Mathematica"}, compareFirstOnly: true},
		{
			path: "mathematica2.m",
			content: `
s := StringRiffle[{"a", "b", "c", "d", "e"}, ", "]
Flatten[{{a, b}, {c, {d}, e}, {f, {g, h}}}]
square[x_] := x ^ 2
fourthpower[x_] := square[square[x]]
`,
			expectedLanguages: []string{"Mathematica"},
			compareFirstOnly:  true,
		},
	}

	for _, testCase := range testCases {
		var getContent func() ([]byte, error)
		if testCase.content != "" {
			getContent = func() ([]byte, error) { return []byte(testCase.content), nil }
		}
		gotLanguages, err := GetLanguages(testCase.path, getContent)
		require.NoError(t, err)
		if testCase.compareFirstOnly {
			require.Equal(t, testCase.expectedLanguages, gotLanguages[0:1])
			continue
		}
		require.Equal(t, testCase.expectedLanguages, gotLanguages)
	}

	rapid.Check(t, func(t *rapid.T) {
		path := rapid.String().Draw(t, "path")
		content := rapid.SliceOfN(rapid.Byte(), 0, 100).Draw(t, "contents")
		require.NotPanics(t, func() {
			langs, err := GetLanguages(path, func() ([]byte, error) { return content, nil })
			require.NoError(t, err)
			if len(langs) != 0 {
				for _, l := range langs {
					require.NotEqual(t, enry.OtherLanguage, l)
				}
			}
		})
	})

	rapid.Check(t, func(t *rapid.T) {
		baseName := "abcd"
		exts := []string{".h", ".m", ".unknown", ""}
		extGens := []*rapid.Generator[string]{}
		for _, ext := range exts {
			extGens = append(extGens, rapid.Just(ext))
		}
		extension := rapid.OneOf(extGens...).Draw(t, "extension")
		path := baseName + extension
		contentGens := []*rapid.Generator[string]{}
		for _, content := range []string{cContent, cppContent, mathematicaContent, matlabContent, emptyContent} {
			contentGens = append(contentGens, rapid.Just(content))
		}
		content := rapid.OneOf(contentGens...).Draw(t, "content")
		langs, err := GetLanguages(path, func() ([]byte, error) {
			return []byte(content), nil
		})
		require.NoError(t, err)
		for _, lang := range langs {
			require.NotEqual(t, enry.OtherLanguage, lang)
		}
	})
}
