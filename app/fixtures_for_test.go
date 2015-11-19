package app_test

import (
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sqs/pbtypes"
)

var aliceUser = sourcegraph.User{
	UID:          1,
	Login:        "alice",
	Name:         "Alice",
	RegisteredAt: &pbtypes.Timestamp{Seconds: 12345},
}

var aliceFooRepo = sourcegraph.Repo{
	URI:           "github.com/alice/foo",
	Name:          "foo",
	Description:   "this is the foo description",
	VCS:           sourcegraph.Git,
	HTTPCloneURL:  "https://github.com/alice/foo.git",
	DefaultBranch: "master",
	Language:      "Python",
}

var quxDef = sourcegraph.Def{
	Def: graph.Def{
		DefKey: graph.DefKey{
			Repo:     aliceFooRepo.URI,
			CommitID: "abc",
			UnitType: "GoPackage",
			Unit:     ".",
			Path:     "baz/Qux",
		},
		Kind: "func",
		File: "a.txt",
	},
	DocHTML: &pbtypes.HTML{"this is the Qux doc"},
}

const fakeCommitID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" // full 40-char commit ID
