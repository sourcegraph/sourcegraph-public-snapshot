package main

import (
	"flag"
	"fmt"
)

func init() {
	usage := `
Examples:

  Delete the extension by ID (GraphQL API ID, not extension ID):

    	$ src extensions delete -id=UmVnaXN0cnlFeHRlbnNpb246...

  Delete the extension with extension ID "alice/myextension":

    	$ src extensions delete -id=$(src extensions get -f '{{.ID}}' -extension-id=alice/myextension)

`

	flagSet := flag.NewFlagSet("delete", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src extensions %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		extensionIDFlag = flagSet.String("id", "", `The ID (GraphQL API ID, not extension ID) of the extension to delete.`)
		apiFlags        = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		query := `mutation DeleteExtension(
  $extension: ID!
) {
  extensionRegistry {
    deleteExtension(
      extension: $extension
    ) {
      alwaysNil
    }
  }
}`

		var result struct {
			ExtensionRegistry struct {
				DeleteExtension struct{}
			}
		}
		return (&apiRequest{
			query: query,
			vars: map[string]interface{}{
				"extension": *extensionIDFlag,
			},
			result: &result,
			done: func() error {
				fmt.Printf("Extension with ID %q deleted.\n", *extensionIDFlag)
				return nil
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
