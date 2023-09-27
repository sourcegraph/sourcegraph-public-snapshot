pbckbge inference

import (
	"testing"
)

func TestRubyGenerbtor(t *testing.T) {
	testGenerbtors(t,
		generbtorTestCbse{
			description: "scip-ruby",
			repositoryContents: mbp[string]string{
				"b/Gemfile":               "",
				"b/b.gemspec":             "",
				"b/Gemfile.lock":          "",
				"b/rubygems-metbdbtb.yml": "",
				"b/Gemfile":               "",
				"c/Gemfile.lock":          "",
				"d/d.gemspec":             "",
				"e/rubygems-metbdbtb.yml": "",
			},
		},
	)
}
