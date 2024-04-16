package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/cmderrors"
)

func init() {
	usage := `
Examples:

  Edit settings property for the current user (authenticated by the src CLI's access token, if any):

    	$ src config edit -property motd -value '["Hello!"]'

  Overwrite all settings settings for the current user:

    	$ src config edit -overwrite -value '{"motd":["Hello!"]}'

  Overwrite all settings settings for the current user with the file contents:

    	$ src config edit -overwrite -value-file myconfig.json

  Edit a settings property for the user with username alice:

    	$ src config edit -subject=$(src users get -f '{{.ID}}' -username=alice) -property motd -value '["Hello!"]'

  Overwrite all settings settings for the organization named abc-org:

    	$ src config edit -subject=$(src orgs get -f '{{.ID}}' -name=abc-org) -overwrite -value '{"motd":["Hello!"]}'

  Change global settings:

    	$ src config edit -subject=$(echo 'query { site { id } }' | src api | jq .data.site.id --raw-output)
`

	flagSet := flag.NewFlagSet("edit", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src config %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		subjectFlag   = flagSet.String("subject", "", "The ID of the settings subject whose settings to edit. (default: authenticated user)")
		propertyFlag  = flagSet.String("property", "", "The name of the settings property to set.")
		valueFlag     = flagSet.String("value", "", "The value for the settings property (when used with -property).")
		valueFileFlag = flagSet.String("value-file", "", "Read the value from this file instead of from the -value command-line option.")
		overwriteFlag = flagSet.Bool("overwrite", false, "Overwrite the entire settings with the value given in -value (not just a single property).")
		apiFlags      = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		err := flagSet.Parse(args)
		if err != nil {
			return err
		}

		keyPath := []KeyPath{}
		if *propertyFlag != "" {
			keyPath = []KeyPath{{Property: *propertyFlag}}
		} else if !*overwriteFlag {
			return cmderrors.Usage("either -property or -overwrite must be used")
		}

		var value string
		if *valueFlag != "" {
			value = *valueFlag
		} else if *valueFileFlag != "" {
			data, err := os.ReadFile(*valueFileFlag)
			if err != nil {
				return err
			}
			value = string(data)
		} else {
			return cmderrors.Usage("either -value or -value-file must be used")
		}

		ctx := context.Background()
		client := cfg.apiClient(apiFlags, flagSet.Output())

		var subjectID string
		if *subjectFlag == "" {
			userID, err := getViewerUserID(ctx, client)
			if err != nil {
				return err
			}
			subjectID = userID
		} else {
			subjectID = *subjectFlag
		}

		lastID, err := getSettingsSubjectLatestSettingsID(ctx, client, subjectID)
		if err != nil {
			return err
		}

		query := `
mutation EditSettings($input: SettingsMutationGroupInput!, $edit: SettingsEdit!) {
  settingsMutation(input: $input) {
    editSettings(edit: $edit) {
      empty {
        alwaysNil
      }
    }
  }
}`
		queryVars := map[string]interface{}{
			"input": map[string]interface{}{
				"subject": subjectID,
				"lastID":  lastID,
			},
			"edit": map[string]interface{}{
				"keyPath":                   keyPath,
				"value":                     value,
				"valueIsJSONCEncodedString": true,
			},
		}

		var result struct {
			ViewerSettings  *SettingsCascade
			SettingsSubject *SettingsSubject
		}
		_, err = client.NewRequest(query, queryVars).Do(ctx, &result)
		return err
	}

	// Register the command.
	configCommands = append(configCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
