package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
)

func init() {
	usage := `
Examples:

  Edit configuration property for the current user (authenticated by the src CLI's access token, if any):

    	$ src config edit -property motd -value '["Hello!"]'

  Overwrite all configuration settings for the current user:

    	$ src config edit -overwrite -value '{"motd":["Hello!"]}'

  Overwrite all configuration settings for the current user with the file contents:

    	$ src config edit -overwrite -value-file myconfig.json

  Edit a configuration property for the user with username alice:

    	$ src config edit -subject=$(src users get -f '{{.ID}}' -username=alice) -property motd -value '["Hello!"]'

  Overwrite all configuration settings for the organization named abc-org:

    	$ src config edit -subject=$(src orgs get -f '{{.ID}}' -name=abc-org) -overwrite -value '{"motd":["Hello!"]}'

`

	flagSet := flag.NewFlagSet("edit", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src config %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		subjectFlag   = flagSet.String("subject", "", "The ID of the configuration subject whose configuration to edit. (default: authenticated user)")
		propertyFlag  = flagSet.String("property", "", "The name of the configuration property to set.")
		valueFlag     = flagSet.String("value", "", "The value for the configuration property (when used with -property).")
		valueFileFlag = flagSet.String("value-file", "", "Read the value from this file instead of from the -value command-line option.")
		overwriteFlag = flagSet.Bool("overwrite", false, "Overwrite the entire settings with the value given in -value (not just a single property).")
		apiFlags      = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		keyPath := []KeyPath{}
		if *propertyFlag != "" {
			keyPath = []KeyPath{{Property: *propertyFlag}}
		} else if !*overwriteFlag {
			return &usageError{errors.New("either -property or -overwrite must be used")}
		}

		var value string
		if *valueFlag != "" {
			value = *valueFlag
		} else if *valueFileFlag != "" {
			data, err := ioutil.ReadFile(*valueFileFlag)
			if err != nil {
				return err
			}
			value = string(data)
		} else {
			return &usageError{errors.New("either -value or -value-file must be used")}
		}

		var subjectID string
		if *subjectFlag == "" {
			userID, err := getViewerUserID()
			if err != nil {
				return err
			}
			subjectID = userID
		} else {
			subjectID = *subjectFlag
		}

		lastID, err := getConfigurationSubjectLatestSettingsID(subjectID)
		if err != nil {
			return err
		}

		query := `
mutation EditConfiguration($input: ConfigurationMutationGroupInput!, $edit: ConfigurationEdit!) {
  configurationMutation(input: $input) {
    editConfiguration(edit: $edit) {
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
			ViewerConfiguration  *ConfigurationCascade
			ConfigurationSubject *ConfigurationSubject
		}
		return (&apiRequest{
			query:  query,
			vars:   queryVars,
			result: &result,
			flags:  apiFlags,
		}).do()
	}

	// Register the command.
	configCommands = append(configCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
