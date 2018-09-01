package main

import (
	"flag"
	"fmt"
)

func init() {
	usage := `
Examples:

  Get extension with extension ID "alice/myextension":

    	$ src extensions get alice/myextension
    	$ src extensions get -extension-id=alice/myextension

`

	flagSet := flag.NewFlagSet("get", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src extensions %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		extensionIDFlag = flagSet.String("extension-id", "", `Look up extension by extension ID. (e.g. "alice/myextension")`)
		formatFlag      = flagSet.String("f", "{{.|json}}", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.ExtensionID}}: {{.Manifest.Title}} ({{.RemoteURL}})" or "{{.|json}}")`)
		apiFlags        = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		tmpl, err := parseTemplate(*formatFlag)
		if err != nil {
			return err
		}

		query := `query RegistryExtension(
  $extensionID: String!,
) {
  extensionRegistry {
    extension(
      extensionID: $extensionID
    ) {
      ...RegistryExtensionFields
    }
  }
}` + registryExtensionFragment

		extensionID := *extensionIDFlag
		if extensionID == "" && flagSet.NArg() == 1 {
			extensionID = flagSet.Arg(0)
		}

		var result struct {
			ExtensionRegistry struct {
				Extension *Extension
			}
		}
		return (&apiRequest{
			query: query,
			vars: map[string]interface{}{
				"extensionID": extensionID,
			},
			result: &result,
			done: func() error {
				return execTemplate(tmpl, result.ExtensionRegistry.Extension)
			},
			flags: apiFlags,
		}).do()
	}

	// Register the command.
	extensionsCommands = append(extensionsCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
