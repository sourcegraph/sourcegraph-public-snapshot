package inference

import (
	"testing"
)

func TestTypeScriptGenerator(t *testing.T) {
	testGenerators(t,
		generatorTestCase{
			description: "javascript project with no tsconfig 1",
			repositoryContents: map[string]string{
				"package.json": "",
			},
		},
		generatorTestCase{
			description: "javascript project with no tsconfig 2",
			repositoryContents: map[string]string{
				"package.json": "",
				"yarn.lock":    "",
			},
		},
		generatorTestCase{
			description: "simple tsconfig",
			repositoryContents: map[string]string{
				"tsconfig.json": "",
			},
		},
		generatorTestCase{
			description: "tsconfig in subdirectories",
			repositoryContents: map[string]string{
				"a/tsconfig.json": "",
				"b/tsconfig.json": "",
				"c/tsconfig.json": "",
			},
		},
		generatorTestCase{
			description: "typescript installation steps",
			repositoryContents: map[string]string{
				"tsconfig.json":              "",
				"package.json":               "",
				"foo/bar/package.json":       "",
				"foo/bar/yarn.lock":          "",
				"foo/bar/baz/tsconfig.json":  "",
				"foo/bar/bonk/tsconfig.json": "",
				"foo/bar/bonk/package.json":  "",
				"foo/baz/tsconfig.json":      "",
			},
		},
		generatorTestCase{
			description: "typescript with lerna configuration",
			repositoryContents: map[string]string{
				"package.json":  "",
				"lerna.json":    `{"npmClient": "yarn"}`,
				"tsconfig.json": "",
			},
		},
		generatorTestCase{
			description: "typescript with node version",
			repositoryContents: map[string]string{
				"package.json":  `{"engines": {"node": "42"}}`,
				"tsconfig.json": "",
				".nvmrc":        "",
			},
		},
	)
}
