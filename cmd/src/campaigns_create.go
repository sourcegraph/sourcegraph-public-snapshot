package main

import (
	"flag"
	"fmt"

	"github.com/pkg/errors"
)

func init() {
	usage := `
Examples:

  Create a campaign with the given name, description and campaign plan:

		$ src campaigns create -name="Format Go code" \
		   -desc="This campaign runs gofmt over all Go repositories" \
		   -plan=Q2FtcGFpZ25QbGFuOjM=

  Create a manual campaign with the given name and description and adds two GitHub pull requests to it:

		$ src campaigns create -name="Migrate to Python 3" \
		   -desc="This campaign tracks all Python 3 migration PRs"
		$ src campaigns add-changesets -campaign=<id-returned-by-previous-command> \
		   -repo-name=github.com/our-org/a-python-repo 5612 7321

`

	flagSet := flag.NewFlagSet("create", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src campaigns create %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		nameFlag        = flagSet.String("name", "", "Name of the campaign. (required)")
		descriptionFlag = flagSet.String("desc", "", "Description for the campaign. (required)")
		namespaceFlag   = flagSet.String("namespace", "", "ID of the namespace under which to create the campaign. The namespace can be the GraphQL ID of a Sourcegraph user or organisation. If not specified, the ID of the authenticated user is queried and used. (Required)")
		planIDFlag      = flagSet.String("plan", "", "ID of campaign plan the campaign should turn into changesets. If no plan is specified, a campaign is created to which changesets can be added manually.")
		draftFlag       = flagSet.Bool("draft", false, "Create the campaign as a draft (which won't create pull requests on code hosts)")

		changesetsFlag = flagSet.Int("changesets", 1000, "Returns the first n changesets per campaign.")

		formatFlag = flagSet.String("f", "{{friendlyCampaignCreatedMessage .}}", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.ID}}: {{.Name}}") or "{{.|json}}")`)
		apiFlags   = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		if *nameFlag == "" {
			return &usageError{errors.New("-name must be specified")}
		}

		if *descriptionFlag == "" {
			return &usageError{errors.New("-desc must be specified")}
		}

		var namespace string
		if *namespaceFlag != "" {
			namespace = *namespaceFlag
		} else {
			var currentUserResult struct {
				CurrentUser *User
			}

			req := &apiRequest{
				query:  currentUserIDQuery,
				result: &currentUserResult,
				flags:  apiFlags,
			}
			err := req.do()
			if err != nil {
				return err
			}
			if currentUserResult.CurrentUser.ID == "" {
				return errors.New("Failed to query authenticated user's ID")
			}
			namespace = currentUserResult.CurrentUser.ID
		}

		tmpl, err := parseTemplate(*formatFlag)
		if err != nil {
			return err
		}

		input := map[string]interface{}{
			"name":        *nameFlag,
			"description": *descriptionFlag,
			"namespace":   namespace,
			"plan":        nullString(*planIDFlag),
			"draft":       *draftFlag,
		}

		var result struct {
			CreateCampaign Campaign
		}

		return (&apiRequest{
			query: campaignFragment + createcampaignMutation,
			vars: map[string]interface{}{
				"input":           input,
				"changesetsFirst": nullInt(*changesetsFlag),
			},
			result: &result,
			done: func() error {
				return execTemplate(tmpl, result.CreateCampaign)
			},
			flags: apiFlags,
		}).do()
	}

	// Register the command.
	campaignsCommands = append(campaignsCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}

const currentUserIDQuery = `query CurrentUser { currentUser { id } }`

const createcampaignMutation = `mutation CreateCampaign($input: CreateCampaignInput!, $changesetsFirst: Int) {
  createCampaign(input: $input) {
	... campaign
  }
}
`
