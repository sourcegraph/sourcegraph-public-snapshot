package languages

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetLanguages(t *testing.T) {
	testCases := []struct {
		path              string
		content           string
		expectedLanguages []string
		compareFirstOnly  bool
	}{
		{path: "perlscript", content: "#!/usr/bin/env perl\n$version = $ARGV[0];", expectedLanguages: []string{"Perl"}},
		{path: "rakuscript", content: "#!/usr/bin/env perl6\n$version = $ARGV[0];", expectedLanguages: []string{"Raku"}},
		{path: "ambiguous.h", content: "", expectedLanguages: []string{"C", "C++", "Objective-C"}},
		{path: "cpp.h", content: "namespace x { }", expectedLanguages: []string{"C++"}},
		{path: "c.h", content: "typedef struct { int x; } Int;", expectedLanguages: []string{"C"}},
		{path: "matlab.m", content: "function [out] = square(x)\nout = x * x;\nend", expectedLanguages: []string{"MATLAB"}, compareFirstOnly: true},
		{path: "mathematica.m", content: "f[x_] := x ^ 2\ng[y_] := f[y]", expectedLanguages: []string{"Mathematica"}, compareFirstOnly: true},
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
}
