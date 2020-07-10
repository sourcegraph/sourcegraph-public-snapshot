package testing

import (
	"fmt"

	"github.com/graph-gophers/graphql-go"
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

func NewRawChangesetSpecGitBranch(repo graphql.ID) string {
	tmpl := `{
		"baseRepository": %q,
		"baseRev":"d34db33f",
		"baseRef":"refs/heads/master",

		"headRepository": %q,
		"headRef":"refs/heads/my-branch",

		"title": "the title",
		"body": "the body of the PR",

		"published": false,

		"commits": [
		  {"message": "git commit message", "diff": "+/- diff"}
		]
	}`

	return fmt.Sprintf(tmpl, repo, repo)
}

func NewRawChangesetSpecExisting(repo graphql.ID, externalID string) string {
	tmpl := `{
		"baseRepository": %q,
		"externalID": %q
	}`

	return fmt.Sprintf(tmpl, repo, externalID)
}
