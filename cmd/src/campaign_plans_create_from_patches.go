package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
)

func init() {
	usage := `
Create a campaign plan from a set of patches (in unified diff format) to repository branches.

Standard input is expected to be a JSON array of {repository: string, baseRevision: string, patch: string}. The repository value is the repository's GraphQL ID (which you can look up given a repository name using 'src repos get -name=...').

Examples:

  Create a campaign plan from my.patch applied to a repository's master branch:

		$ patch=$(jq -M -R -s . < my.patch)
		$ repo=$(src repo get -name=github.com/sourcegraph/src-cli -f '{{.ID|json}}')
		$ echo '[{"repository": $repo, "baseRevision": "master", "patch": $patch}]' |\
		  src campaigns plans create-from-patches

  Create a campaign plan from patches.json produced by 'src actions exec':

		$ src actions exec -f action.json > patches.json
		$ src campaign plan create-from-patches < patches.json

  Create a campaign plan by piping output of 'src actions exec' into 'src campaign plan create-from-patches':

		$ src actions exec -f action.json | src campaign plan create-from-patches < patches.json

`

	flagSet := flag.NewFlagSet("create-from-patches", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src campaigns plans %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		changesetsFlag = flagSet.Int("changesets", 1000, "Returns the first n changesets in the plan.")
		formatFlag     = flagSet.String("f", "{{friendlyCampaignPlanCreatedMessage .}}", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.ID}}: {{len .Changesets}} changesets") or "{{.|json}}")`)
		apiFlags       = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		tmpl, err := parseTemplate(*formatFlag)
		if err != nil {
			return err
		}

		if isatty.IsTerminal(os.Stdin.Fd()) {
			log.Println("# Waiting for JSON patches input on stdin...")
		}

		var patches []CampaignPlanPatch
		if err := json.NewDecoder(os.Stdin).Decode(&patches); err != nil {
			return errors.Wrap(err, "invalid JSON patches input")
		}

		return createCampaignPlanFromPatches(apiFlags, patches, tmpl, *changesetsFlag)
	}

	// Register the command.
	campaignPlansCommands = append(campaignPlansCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}

const createCampaignPlanMutation = `
mutation CreateCampaignPlanFromPatches($patches: [CampaignPlanPatch!]!) {
  createCampaignPlanFromPatches(patches: $patches) {
    ...CampaignPlanFields
  }
}
`

func createCampaignPlanFromPatches(
	apiFlags *apiFlags,
	patches []CampaignPlanPatch,
	tmpl *template.Template,
	numChangesets int,
) error {
	query := createCampaignPlanMutation + campaignPlanFragment(numChangesets)

	var result struct {
		CreateCampaignPlanFromPatches CampaignPlan
	}
	return (&apiRequest{
		query:  query,
		vars:   map[string]interface{}{"patches": patches},
		result: &result,
		done: func() error {
			if err := execTemplate(tmpl, result.CreateCampaignPlanFromPatches); err != nil {
				return err
			}
			return nil
		},
		flags: apiFlags,
	}).do()
}
