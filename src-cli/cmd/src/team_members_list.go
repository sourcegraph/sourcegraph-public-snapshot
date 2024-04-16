package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/api"
)

func init() {
	usage := `
Examples:

  List team members:

    	$ src team members list -name=<teamName>

  List team members whose names match the query:

    	$ src team members list -name=<teamName> -query='myquery'
`

	flagSet := flag.NewFlagSet("list", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src team members %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		nameFlag   = flagSet.String("name", "", "The team of which to return members")
		firstFlag  = flagSet.Int("first", 1000, "Returns the first n teams from the list")
		queryFlag  = flagSet.String("query", "", `Returns teams whose name or displayname match the query. (e.g. "engineering")`)
		formatFlag = flagSet.String("f", "{{.Username}}", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.Name}}: {{.DisplayName}}" or "{{.|json}}")`)
		jsonFlag   = flagSet.Bool("json", false, `Format for the output as json`)
		apiFlags   = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		if *nameFlag == "" {
			return errors.New("must provide -name")
		}

		tmpl, err := parseTemplate(*formatFlag)
		if err != nil {
			return err
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		query := `query TeamMembers(
	$name: String!,
	$first: Int,
	$search: String
) {
	team(name: $name) {
		members (
			first: $first,
			search: $search
		) {
			nodes {
				...TeamMemberFields
			}
		}
	}
}
` + teamMemberFragment

		var result struct {
			Team struct {
				Members struct {
					Nodes []TeamMember
				}
			}
		}
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"name":   *nameFlag,
			"first":  api.NullInt(*firstFlag),
			"search": api.NullString(*queryFlag),
		}).Do(context.Background(), &result); err != nil || !ok {
			return err
		}

		if jsonFlag != nil && *jsonFlag {
			json.NewEncoder(os.Stdout).Encode(result.Team.Members.Nodes)
			return nil
		}

		for _, t := range result.Team.Members.Nodes {
			if err := execTemplate(tmpl, t); err != nil {
				return err
			}
		}
		return nil
	}

	// Register the command.
	teamMembersCommands = append(teamMembersCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
