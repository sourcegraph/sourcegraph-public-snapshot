package inference

import (
	"testing"
)

func TestRubyGenerator(t *testing.T) {
	testGenerators(t,
		generatorTestCase{
			description: "scip-ruby",
			repositoryContents: map[string]string{
				"a/Gemfile":               "",
				"a/a.gemspec":             "",
				"a/Gemfile.lock":          "",
				"a/rubygems-metadata.yml": "",
				"b/Gemfile":               "",
				"c/Gemfile.lock":          "",
				"d/d.gemspec":             "",
				"e/rubygems-metadata.yml": "",
			},
		},
	)
}
