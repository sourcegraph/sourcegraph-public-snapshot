package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sourcegraph/src-cli/internal/api"
)

func init() {
	usage := `
Examples:

  List users:

    	$ src users list

  List *all* users (may be slow!):

    	$ src users list -first='-1'

  List users whose names match the query:

    	$ src users list -query='myquery'

  List all users with the "foo" tag:

    	$ src users list -tag=foo

`

	flagSet := flag.NewFlagSet("list", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src users %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		firstFlag  = flagSet.Int("first", 1000, "Returns the first n users from the list. (use -1 for unlimited)")
		queryFlag  = flagSet.String("query", "", `Returns users whose names match the query. (e.g. "alice")`)
		tagFlag    = flagSet.String("tag", "", `Returns users with the given tag.`)
		formatFlag = flagSet.String("f", "{{.Username}}", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.ID}}: {{.Username}} ({{.DisplayName}})" or "{{.|json}}")`)
		apiFlags   = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		ctx := context.Background()
		client := cfg.apiClient(apiFlags, flagSet.Output())

		tmpl, err := parseTemplate(*formatFlag)
		if err != nil {
			return err
		}
		vars := map[string]interface{}{
			"first": api.NullInt(*firstFlag),
			"query": api.NullString(*queryFlag),
			"tag":   api.NullString(*tagFlag),
		}
		queryTagVar := ""
		queryTag := ""
		if maybeTagVar, ok := vars["tag"].(*string); ok && maybeTagVar != nil {
			queryTagVar = `$tag: String,`
			queryTag = `tag: $tag,`
		}
		query := `query Users(
  $first: Int,
  $query: String,
` + queryTagVar + `
) {
  users(
first: $first,
    query: $query,
` + queryTag + `
  ) {
    nodes {
      ...UserFields
    }
  }
}` + userFragment

		var result struct {
			Users struct {
				Nodes []User
			}
		}
		if ok, err := client.NewRequest(query, vars).Do(ctx, &result); err != nil || !ok {
			return err
		}

		for _, user := range result.Users.Nodes {
			if err := execTemplate(tmpl, user); err != nil {
				return err
			}
		}
		return nil
	}

	// Register the command.
	usersCommands = append(usersCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
