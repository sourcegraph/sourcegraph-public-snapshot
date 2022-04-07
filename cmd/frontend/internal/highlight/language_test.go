package highlight

import (
	"testing"

	"github.com/grafana/regexp"
)

type languageTestCase struct {
	Config   syntaxHighlightConfig
	Path     string
	Expected string
	Found    bool
}

func TestGetLanguageFromConfig(t *testing.T) {
	cases := []languageTestCase{
		{
			Config: syntaxHighlightConfig{
				Extensions: map[string]string{
					"go": "not go",
				},
			},
			Path:     "example.go",
			Found:    true,
			Expected: "not go",
		},
		{
			Config: syntaxHighlightConfig{
				Extensions: map[string]string{},
			},
			Path:     "example.go",
			Found:    false,
			Expected: "",
		},

		{
			Config: syntaxHighlightConfig{
				Extensions: map[string]string{
					"strato": "scala",
				},
			},
			Path:     "test.strato",
			Found:    true,
			Expected: "scala",
		},

		{
			Config: syntaxHighlightConfig{
				Patterns: []languagePattern{
					{
						pattern:  regexp.MustCompile("asdf"),
						language: "not matching",
					},
					{
						pattern:  regexp.MustCompile("\\.bashrc"),
						language: "bash",
					},
				},
			},
			Path:     "/home/example/.bashrc",
			Found:    true,
			Expected: "bash",
		},
	}

	for _, testCase := range cases {
		language, found := getLanguageFromConfig(testCase.Config, testCase.Path)
		if found != testCase.Found {
			t.Fatalf("Got: %v, Expected: %v", testCase.Found, found)
		}

		if language != testCase.Expected {
			t.Fatalf("Got: %s, Expected: %s", testCase.Expected, language)
		}
	}
}
