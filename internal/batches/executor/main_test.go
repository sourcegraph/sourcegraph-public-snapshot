package executor

import (
	"github.com/sourcegraph/batch-change-utils/overridable"
	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
)

var testRepo1 = &graphql.Repository{
	ID:            "src-cli",
	Name:          "github.com/sourcegraph/src-cli",
	DefaultBranch: &graphql.Branch{Name: "main", Target: graphql.Target{OID: "d34db33f"}},
	FileMatches: map[string]bool{
		"README.md": true,
		"main.go":   true,
	},
}

var testRepo2 = &graphql.Repository{
	ID:   "sourcegraph",
	Name: "github.com/sourcegraph/sourcegraph",
	DefaultBranch: &graphql.Branch{
		Name:   "main",
		Target: graphql.Target{OID: "f00b4r3r"},
	},
}

var testChangesetTemplate = &batches.ChangesetTemplate{
	Title:  "commit title",
	Body:   "commit body",
	Branch: "commit-branch",
	Commit: batches.ExpandedGitCommitDescription{
		Message: "commit msg",
		Author: &batches.GitCommitAuthor{
			Name:  "Tester",
			Email: "tester@example.com",
		},
	},
	Published: overridable.FromBoolOrString(false),
}
