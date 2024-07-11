package inference

import (
	"testing"
)

func TestDotNetGenerator(t *testing.T) {
	testGenerators(t,
		generatorTestCase{
			description: "dotnet sln files exist",
			repositoryContents: map[string]string{
				"one.sln":            "",
				"one.csproj":         "",
				"foo/baz/two.sln":    "",
				"foo/baz/two.vbproj": "",
			},
		},
		generatorTestCase{
			description: "dotnet sln files do not exist",
			repositoryContents: map[string]string{
				"one.csproj":         "",
				"foo/baz/two.vbproj": "",
			},
		},
	)
}
