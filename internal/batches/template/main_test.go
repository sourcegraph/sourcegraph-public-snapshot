package template

import "github.com/sourcegraph/src-cli/internal/batches/graphql"

var testRepo1 = &graphql.Repository{
	ID:            "src-cli",
	Name:          "github.com/sourcegraph/src-cli",
	DefaultBranch: &graphql.Branch{Name: "main", Target: graphql.Target{OID: "d34db33f"}},
	FileMatches: map[string]bool{
		"README.md": true,
		"main.go":   true,
	},
}
