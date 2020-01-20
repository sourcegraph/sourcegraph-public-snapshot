package main

import (
	"flag"
	"fmt"
)

func init() {
	usage := `
Examples:

  Preview a comby campaign:

    	$ src campaigns plans preview -type=comby -args='{"scopeQuery":"repo:sourcegraph/go-diff", "matchTemplate": "fmt.Errorf", "rewriteTemplate": "errors.Wrapf"}' -f '{{.|json}}'

`

	flagSet := flag.NewFlagSet("preview", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src campaigns plans %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		typeFlag       = flagSet.String("type", "", `The campaign type. (required)`)
		argumentsFlag  = flagSet.String("args", "", `The arguments that define what changes to make.`)
		waitFlag       = flagSet.Bool("wait", true, `Wait for the preview to be generated.`)
		changesetsFlag = flagSet.Int("changesets", 1000, "Returns the first n changesets in the plan.")
		formatFlag     = flagSet.String("f", "{{.PreviewURL}}", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.ID}}: {{.Changesets|json}}") or "{{.|json}}")`)
		apiFlags       = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		tmpl, err := parseTemplate(*formatFlag)
		if err != nil {
			return err
		}

		query := `mutation PreviewCampaignPlan(
  $specification: CampaignPlanSpecification!,
  $wait: Boolean!,
) {
  previewCampaignPlan(
    specification: $specification,
    wait: $wait,
  ) {
    ...CampaignPlanFields
  }
}
` + campaignPlanFragment(*changesetsFlag)

		var result struct {
			PreviewCampaignPlan CampaignPlan
		}
		return (&apiRequest{
			query: query,
			vars: map[string]interface{}{
				"specification": map[string]interface{}{
					"type":      *typeFlag,
					"arguments": *argumentsFlag,
				},
				"wait": *waitFlag,
			},
			result: &result,
			done: func() error {
				if err := execTemplate(tmpl, result.PreviewCampaignPlan); err != nil {
					return err
				}
				return nil
			},
			flags: apiFlags,
		}).do()
	}

	// Register the command.
	campaignPlansCommands = append(campaignPlansCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
