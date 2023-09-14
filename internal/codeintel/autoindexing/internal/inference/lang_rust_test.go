package inference

import (
	"testing"
)

func TestRustGenerator(t *testing.T) {
	testGenerators(t,
		generatorTestCase{
			description: "rust-analyzer",
			repositoryContents: map[string]string{
				"foo/bar/Cargo.toml": "",
				"foo/baz/Cargo.toml": "",
			},
		},
	)
}
