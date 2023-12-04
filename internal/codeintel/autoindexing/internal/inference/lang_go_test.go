package inference

import (
	"testing"
)

func TestGoGenerator(t *testing.T) {
	testGenerators(t,
		generatorTestCase{
			description: "go modules",
			repositoryContents: map[string]string{
				"foo/bar/go.mod": "",
				"foo/baz/go.mod": "",
			},
		},
		generatorTestCase{
			description: "go files in root",
			repositoryContents: map[string]string{
				"main.go":       "",
				"internal/a.go": "",
				"internal/b.go": "",
			},
		},
		generatorTestCase{
			description: "go files in non-root (no match)",
			repositoryContents: map[string]string{
				"cmd/src/main.go": "",
			},
		},
	)
}
