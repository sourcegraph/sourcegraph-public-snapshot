package testing

import (
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/go-diff/diff"
)

const TestRawCampaignSpec = `{
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
      "env": {
        "PATH": "/work/foobar:$PATH"
      }
    }
  ],
  "changesetTemplate": {
    "title": "Hello World",
    "body": "My first campaign!",
    "branch": "hello-world",
    "commit": {
      "message": "Append Hello World to all README.md files"
    },
    "published": false
  }
}`

const TestRawCampaignSpecYAML = `
name: my-unique-name
description: My description
'on':
- repositoriesMatchingQuery: lang:go func main
- repository: github.com/sourcegraph/src-cli
steps:
- run: echo 'foobar'
  container: alpine
  env:
    PATH: "/work/foobar:$PATH"
changesetTemplate:
  title: Hello World
  body: My first campaign!
  branch: hello-world
  commit:
    message: Append Hello World to all README.md files
  published: false
`

var ChangesetSpecDiffStat = &diff.Stat{Added: 1, Changed: 2, Deleted: 1}

func NewRawChangesetSpecGitBranch(repo graphql.ID, baseRev string) string {
	diff := `diff --git INSTALL.md INSTALL.md
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
`
	tmpl := `{

		"baseRepository": %q,
		"baseRev": %q,
		"baseRef":"refs/heads/master",

		"headRepository": %q,
		"headRef":"refs/heads/my-branch",

		"title": "the title",
		"body": "the body of the PR",

		"published": false,

		"commits": [
		  {"message": "git commit message", "diff": %q, "authorName": "Mary McButtons", "authorEmail": "mary@example.com"}]
	}`

	return fmt.Sprintf(tmpl, repo, baseRev, repo, diff)
}

func NewRawChangesetSpecExisting(repo graphql.ID, externalID string) string {
	tmpl := `{
		"baseRepository": %q,
		"externalID": %q
	}`

	return fmt.Sprintf(tmpl, repo, externalID)
}
