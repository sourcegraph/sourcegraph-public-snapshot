package testing

import (
	"fmt"

	"github.com/graph-gophers/graphql-go"
)

const TestRawCampaignSpec = `{
  "name": "The name",
  "description": "My description",
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

func NewRawChangesetSpec(repo graphql.ID) string {
	tmpl := `{
		"repoID": %q,
		"rev":"d34db33f",
		"baseRef":"refs/heads/master",
		"diff":"+-"
	}`

	return fmt.Sprintf(tmpl, repo)
}
