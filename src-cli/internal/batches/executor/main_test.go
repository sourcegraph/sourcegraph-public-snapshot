package executor

import (
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/overridable"

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

var testPublished = overridable.FromBoolOrString(false)

var testChangesetTemplate = &batcheslib.ChangesetTemplate{
	Title:  "commit title",
	Body:   "commit body",
	Branch: "commit-branch",
	Commit: batcheslib.ExpandedGitCommitDescription{
		Message: "commit msg",
		Author: &batcheslib.GitCommitAuthor{
			Name:  "Tester",
			Email: "tester@example.com",
		},
	},
	Published: &testPublished,
}
