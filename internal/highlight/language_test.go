package highlight

import (
	"fmt"
	"testing"

	"github.com/grafana/regexp"
	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/schema"
)

type languageTestCase struct {
	Config   syntaxHighlightConfig[*regexp.Regexp]
	Path     string
	Expected string
	Found    bool
}

func TestInitializeConfig(t *testing.T) {
	type TestCase struct {
		siteConfig          *schema.SyntaxHighlighting
		baseEngineConfig    syntaxEngineConfig
		baseHighlightConfig syntaxHighlightConfig[*regexp.Regexp]
		wantEngineConfig    autogold.Value
		wantHighlightConfig autogold.Value
	}

	// Update golden values with go test ./internal/highlight/... -update
	testCases := []TestCase{
		{
			siteConfig:          nil,
			wantEngineConfig:    autogold.Expect(syntaxEngineConfig{Overrides: map[string]EngineType{}}),
			wantHighlightConfig: autogold.Expect(syntaxHighlightConfig[string]{Extensions: map[string]string{}}),
		},
		{
			siteConfig: &schema.SyntaxHighlighting{
				Engine: &schema.SyntaxHighlightingEngine{},
			},
			wantEngineConfig:    autogold.Expect(syntaxEngineConfig{Overrides: map[string]EngineType{}}),
			wantHighlightConfig: autogold.Expect(syntaxHighlightConfig[string]{Extensions: map[string]string{}}),
		},
		{
			siteConfig: &schema.SyntaxHighlighting{
				Engine: &schema.SyntaxHighlightingEngine{
					Overrides: map[string]string{"go": "syntect"},
				},
			},
			wantEngineConfig: autogold.Expect(syntaxEngineConfig{Overrides: map[string]EngineType{
				"go": EngineSyntect,
			}}),
			wantHighlightConfig: autogold.Expect(syntaxHighlightConfig[string]{Extensions: map[string]string{}}),
		},
		{
			siteConfig: &schema.SyntaxHighlighting{
				Engine: &schema.SyntaxHighlightingEngine{
					Default: "syntect",
				},
			},
			wantEngineConfig:    autogold.Expect(syntaxEngineConfig{Default: EngineSyntect, Overrides: map[string]EngineType{}}),
			wantHighlightConfig: autogold.Expect(syntaxHighlightConfig[string]{Extensions: map[string]string{}}),
		},
		{
			baseEngineConfig:    baseEngineConfig,
			baseHighlightConfig: baseHighlightConfig,
			wantEngineConfig: autogold.Expect(syntaxEngineConfig{
				Default: EngineTreeSitter,
				Overrides: map[string]EngineType{
					"go":     EngineScipSyntax,
					"java":   EngineScipSyntax,
					"matlab": EngineScipSyntax,
					"perl":   EngineScipSyntax,
				},
			}),
			wantHighlightConfig: autogold.Expect(syntaxHighlightConfig[string]{Extensions: map[string]string{
				"ncl":  "nickel",
				"sbt":  "scala",
				"sc":   "scala",
				"tsx":  "tsx",
				"xlsg": "xlsg",
			}}),
		},
		{
			siteConfig: &schema.SyntaxHighlighting{
				Engine: &schema.SyntaxHighlightingEngine{
					Default: "syntect",
				},
			},
			baseEngineConfig:    baseEngineConfig,
			baseHighlightConfig: baseHighlightConfig,
			wantEngineConfig: autogold.Expect(syntaxEngineConfig{Default: EngineSyntect, Overrides: map[string]EngineType{
				"go":     EngineScipSyntax,
				"java":   EngineScipSyntax,
				"matlab": EngineScipSyntax,
				"perl":   EngineScipSyntax,
			}}),
			wantHighlightConfig: autogold.Expect(syntaxHighlightConfig[string]{Extensions: map[string]string{
				"ncl":  "nickel",
				"sbt":  "scala",
				"sc":   "scala",
				"tsx":  "tsx",
				"xlsg": "xlsg",
			}}),
		},
		{
			siteConfig: &schema.SyntaxHighlighting{
				Languages: &schema.SyntaxHighlightingLanguage{
					Extensions: map[string]string{
						"hack": "php",
					},
				},
			},
			wantEngineConfig: autogold.Expect(syntaxEngineConfig{Overrides: map[string]EngineType{}}),
			wantHighlightConfig: autogold.Expect(syntaxHighlightConfig[string]{Extensions: map[string]string{
				"hack": "php",
			}}),
		},
		{
			siteConfig: &schema.SyntaxHighlighting{
				Languages: &schema.SyntaxHighlightingLanguage{
					Patterns: []*schema.SyntaxHighlightingLanguagePatterns{
						{
							Pattern:  "BUILD(\\.bzl|\\.bazel)?",
							Language: "starlark",
						},
						{
							Pattern:  "Makefile",
							Language: "Make",
						},
					},
				},
			},
			wantEngineConfig: autogold.Expect(syntaxEngineConfig{Overrides: map[string]EngineType{}}),
			wantHighlightConfig: autogold.Expect(syntaxHighlightConfig[string]{
				Extensions: map[string]string{},
				Patterns: []languagePattern[string]{
					{
						pattern:  "BUILD(\\.bzl|\\.bazel)?",
						language: "starlark",
					},
					{
						pattern:  "Makefile",
						language: "Make",
					},
				},
			}),
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			gotEngineConfig, gotHighlightConfig := initializeConfig(testCase.siteConfig, testCase.baseEngineConfig, testCase.baseHighlightConfig)
			testCase.wantEngineConfig.Equal(t, gotEngineConfig)
			var patterns []languagePattern[string]
			for _, p := range gotHighlightConfig.Patterns {
				patterns = append(patterns, languagePattern[string]{
					pattern:  p.pattern.String(),
					language: p.language,
				})
			}
			testCase.wantHighlightConfig.Equal(t, syntaxHighlightConfig[string]{Extensions: gotHighlightConfig.Extensions, Patterns: patterns})
		})
	}
}

func TestGetLanguageFromConfig(t *testing.T) {
	cases := []languageTestCase{
		{
			Config: syntaxHighlightConfig[*regexp.Regexp]{
				Extensions: map[string]string{
					"go": "not go",
				},
			},
			Path:     "example.go",
			Found:    true,
			Expected: "not go",
		},
		{
			Config: syntaxHighlightConfig[*regexp.Regexp]{
				Extensions: map[string]string{},
			},
			Path:     "example.go",
			Found:    false,
			Expected: "",
		},

		{
			Config: syntaxHighlightConfig[*regexp.Regexp]{
				Extensions: map[string]string{
					"strato": "scala",
				},
			},
			Path:     "test.strato",
			Found:    true,
			Expected: "scala",
		},

		{
			Config: syntaxHighlightConfig[*regexp.Regexp]{
				Patterns: []languagePattern[*regexp.Regexp]{
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

func TestShebang(t *testing.T) {
	type testCase struct {
		Contents string
		Expected string
	}

	cases := []testCase{
		{
			Contents: "#!/usr/bin/env python",
			Expected: "python",
		},
		{
			Contents: "#!/usr/bin/env node",
			Expected: "javascript",
		},
		{
			Contents: "#!/usr/bin/env ruby",
			Expected: "ruby",
		},
		{
			Contents: "#!/usr/bin/env perl",
			Expected: "perl",
		},
		{
			Contents: "#!/usr/bin/env php",
			Expected: "php",
		},
		{
			Contents: "#!/usr/bin/env lua",
			Expected: "lua",
		},
		{
			Contents: "#!/usr/bin/env tclsh",
			Expected: "tcl",
		},
		{
			Contents: "#!/usr/bin/env fish",
			Expected: "fish",
		},
	}

	for _, testCase := range cases {
		language, _ := getLanguage("", testCase.Contents)
		if language != testCase.Expected {
			t.Fatalf("%s\nGot: %s, Expected: %s", testCase.Contents, language, testCase.Expected)
		}
	}
}

func TestGetLanguageFromContent(t *testing.T) {
	type testCase struct {
		Filename string
		Contents string
		Expected string
	}

	cases := []testCase{
		{
			Filename: "bruh.m",
			Contents: `#import "Import.h"
@interface Interface ()
@end`,
			Expected: "objective-c",
		},
		{
			Filename: "slay.m",
			Contents: `function setupPythonIfNeeded()
%setupPythonIfNeeded Check if python is installed and configured.  If it's`,
			Expected: "matlab",
		},
		{
			Filename: "hack_hh.hh",
			Contents: `<?hh`,
			Expected: "hack",
		},
		{
			Filename: "cpp_hh.hh",
			// empty file counts as cpp since we only do an explicit check to see if it is hack
			Contents: ``,
			Expected: "c++",
		},
		{
			Filename: "cpp_hh.hh",
			Contents: `#pragma once`,
			Expected: "c++",
		},
		{
			Filename: "cpp_hh.hh",
			Contents: `#help_index "Bit"`,
			// Linguist only considers .hc to be HolyC so this is marked as C++
			Expected: "c++",
		}, {
			Filename: "cpp_hh.hh",
			Contents: `/*
* I am a comment
*/
#ifndef __CPU_O3_CPU_HH__
#define __CPU_O3_CPU_HH__`,
			Expected: "c++",
		},
	}

	for _, testCase := range cases {
		language, _ := getLanguage(testCase.Filename, testCase.Contents)
		if language != testCase.Expected {
			t.Fatalf("%s\nGot: %s, Expected: %s", testCase.Contents, language, testCase.Expected)
		}
	}
}

func TestGetLanguageFromPath(t *testing.T) {
	type testCase struct {
		Filename string
		Expected string
	}

	// JSX is mapped to javascript but TSX is mapped to 'tsx' language
	// because Linguist contains a separate language for TSX but not JSX
	cases := []testCase{
		{
			Filename: "file.js",
			Expected: "javascript",
		},
		{
			Filename: "react.jsx",
			Expected: "javascript",
		},
		{
			Filename: "file.ts",
			Expected: "typescript",
		},
		{
			Filename: "react.tsx",
			Expected: "tsx",
		},
		{
			Filename: "hack.hack",
			Expected: "hack",
		},
		{
			Filename: "cpp.cpp",
			Expected: "c++",
		},

		// This resolves to C++ and not hack since
		// it does not find the <?hh to indicate hack
		{
			Filename: "file.hh",
			Expected: "c++",
		},
	}

	for _, testCase := range cases {
		language, _ := getLanguage(testCase.Filename, "")
		if language != testCase.Expected {
			t.Fatalf("%s\nGot: %s, Expected: %s", testCase.Filename, language, testCase.Expected)
		}
	}
}
