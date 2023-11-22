package testing

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/go-diff/diff"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

const TestRawBatchSpec = `{
  "name": "my-unique-name",
  "description": "My description",
  "on": [
    {"repositoriesMatchingQuery": "lang:go func main"},
    {"repository": "github.com/sourcegraph/src-cli"}
  ],
  "steps": [
    {
      "run": "echo 'foobar'",
      "container": "alpine",
      "env": [
		{ "PATH": "/work/foobar:$PATH" },
		"FOO"
	  ]
    }
  ],
  "changesetTemplate": {
    "title": "Hello World",
    "body": "My first batch change!",
    "branch": "hello-world",
    "commit": {
      "message": "Append Hello World to all README.md files"
    },
    "published": false
  }
}`

const TestRawBatchSpecYAML = `
name: my-unique-name
description: My description
'on':
- repositoriesMatchingQuery: lang:go func main
- repository: github.com/sourcegraph/src-cli
steps:
- run: echo 'foobar'
  container: alpine
  env:
    - PATH: "/work/foobar:$PATH"
    - FOO
changesetTemplate:
  title: Hello World
  body: My first batch change!
  branch: hello-world
  commit:
    message: Append Hello World to all README.md files
  published: false
`

func BuildRawBatchSpecWithImportChangesets(t *testing.T, imports []batcheslib.ImportChangeset) string {
	t.Helper()

	spec := batcheslib.BatchSpec{
		Name:             "test-batch-change",
		Description:      "only importing",
		ImportChangesets: imports,
	}

	marshaledRawSpec, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("failed to marshal BatchSpec: %s", err)
	}

	return string(marshaledRawSpec)
}

var ChangesetSpecDiffStat = &diff.Stat{Added: 3, Deleted: 3}

const ChangesetSpecAuthorEmail = "mary@example.com"

var ChangesetSpecDiff = []byte(`diff --git INSTALL.md INSTALL.md
index e5af166..d44c3fc 100644
--- INSTALL.md
+++ INSTALL.md
@@ -3,10 +3,10 @@
 Line 1
 Line 2
 Line 3
-Line 4
+This is cool: Line 4
 Line 5
 Line 6
-Line 7
-Line 8
+Another Line 7
+Foobar Line 8
 Line 9
 Line 10
`)

var baseChangesetSpecGitBranch = batcheslib.ChangesetSpec{
	BaseRef: "refs/heads/master",

	HeadRef: "refs/heads/my-branch",
	Title:   "the title",
	Body:    "the body of the PR",

	Published: batcheslib.PublishedValue{Val: false},

	Commits: []batcheslib.GitCommitDescription{
		{
			Version:     2,
			Message:     "git commit message\n\nand some more content in a second paragraph.",
			Diff:        ChangesetSpecDiff,
			AuthorName:  "Mary McButtons",
			AuthorEmail: ChangesetSpecAuthorEmail,
		},
	},
}

func NewRawChangesetSpecGitBranch(repo graphql.ID, baseRev string) string {
	spec := baseChangesetSpecGitBranch
	spec.BaseRepository = string(repo)
	spec.BaseRev = baseRev
	spec.HeadRepository = string(repo)

	rawSpec, err := json.Marshal(spec)
	if err != nil {
		panic(err)
	}
	return string(rawSpec)
}

func NewPublishedRawChangesetSpecGitBranch(repo graphql.ID, baseRev string, published batcheslib.PublishedValue) string {
	spec := baseChangesetSpecGitBranch
	spec.BaseRepository = string(repo)
	spec.BaseRev = baseRev
	spec.HeadRepository = string(repo)

	spec.Published = published

	rawSpec, err := json.Marshal(spec)
	if err != nil {
		panic(err)
	}
	return string(rawSpec)
}

func NewRawChangesetSpecExisting(repo graphql.ID, externalID string) string {
	tmpl := `{
		"baseRepository": %q,
		"externalID": %q
	}`

	return fmt.Sprintf(tmpl, repo, externalID)
}

func MarshalJSON(t testing.TB, v any) string {
	t.Helper()

	bs, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}

	return string(bs)
}
