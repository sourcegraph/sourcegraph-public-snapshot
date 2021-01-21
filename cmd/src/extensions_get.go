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
		apiFlags        = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		tmpl, err := parseTemplate(*formatFlag)
		if err != nil {
			return err
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

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
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"extensionID": extensionID,
		}).Do(context.Background(), &result); err != nil || !ok {
			return err
		}

		return execTemplate(tmpl, result.ExtensionRegistry.Extension)
	}

	// Register the command.
	extensionsCommands = append(extensionsCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
