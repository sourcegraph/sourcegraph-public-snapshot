package main

import (
	"flag"
	"fmt"
)

func init() {
	usage := `
Examples:

  Get configuration for the current user (authenticated by the src CLI's access token, if any):

    	$ src config get

  Get configuration for the user with username alice:

    	$ src config get -subject=$(src users get -f '{{.ID}}' -username=alice)

  Get configuration for the organization named abc-org:

    	$ src config get -subject=$(src orgs get -f '{{.ID}}' -name=abc-org)

`

	flagSet := flag.NewFlagSet("get", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src config %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		subjectFlag = flagSet.String("subject", "", "The ID of the configuration subject whose configuration to get. (default: authenticated user)")
		formatFlag  = flagSet.String("f", "{{.Contents|jsonIndent}}", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.|json}}")`)
		apiFlags    = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		tmpl, err := parseTemplate(*formatFlag)
		if err != nil {
			return err
		}

		var query string
		var queryVars map[string]interface{}
		if *subjectFlag == "" {
			query = viewerConfigurationQuery
		} else {
			query = configurationSubjectCascadeQuery
			queryVars = map[string]interface{}{
				"subject": nullString(*subjectFlag),
			}
		}

		var result struct {
			ViewerConfiguration  *ConfigurationCascade
			ConfigurationSubject *ConfigurationSubject
		}
		return (&apiRequest{
			query:  query,
			vars:   queryVars,
			result: &result,
			done: func() error {
				var merged Configuration
				if result.ViewerConfiguration != nil {
					merged = result.ViewerConfiguration.Merged
				} else if result.ConfigurationSubject != nil {
					merged = result.ConfigurationSubject.ConfigurationCascade.Merged
				}
				return execTemplate(tmpl, merged)
			},
			flags: apiFlags,
		}).do()
	}

	// Register the command.
	configCommands = append(configCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
