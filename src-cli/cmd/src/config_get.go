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

  Get settings for the current user (authenticated by the src CLI's access token, if any):

    	$ src config get

  Get settings for the user with username alice:

    	$ src config get -subject=$(src users get -f '{{.ID}}' -username=alice)

  Get settings for the organization named abc-org:

    	$ src config get -subject=$(src orgs get -f '{{.ID}}' -name=abc-org)

`

	flagSet := flag.NewFlagSet("get", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src config %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		subjectFlag = flagSet.String("subject", "", "The ID of the settings subject whose settings to get. (default: authenticated user)")
		formatFlag  = flagSet.String("f", "{{.|jsonIndent}}", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.|json}}")`)
		apiFlags    = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		err := flagSet.Parse(args)
		if err != nil {
			return err
		}

		tmpl, err := parseTemplate(*formatFlag)
		if err != nil {
			return err
		}

		var query string
		var queryVars map[string]interface{}
		if *subjectFlag == "" {
			query = viewerSettingsQuery
		} else {
			query = settingsSubjectCascadeQuery
			queryVars = map[string]interface{}{
				"subject": api.NullString(*subjectFlag),
			}
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		var result struct {
			ViewerSettings  *SettingsCascade
			SettingsSubject *SettingsSubject
		}

		ok, err := client.NewRequest(query, queryVars).Do(context.Background(), &result)
		if err != nil || !ok {
			return err
		}

		var final string
		if result.ViewerSettings != nil {
			final = result.ViewerSettings.Final
		} else if result.SettingsSubject != nil {
			final = result.SettingsSubject.SettingsCascade.Final
		}
		return execTemplate(tmpl, final)
	}

	// Register the command.
	configCommands = append(configCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
