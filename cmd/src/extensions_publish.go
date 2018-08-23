package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
)

func init() {
	usage := `
Publish an extension to Sourcegraph, creating it (if necessary).

Examples:

  Publish the "alice/myextension" extension described by package.json in the current directory:

    	$ cat package.json
        {
          "name":      "myextension",
          "publisher": "alice",
          "title":     "My Extension",
          "url":       "https://example.com/bundled-extension.js"
        }
    	$ src extensions publish

`

	flagSet := flag.NewFlagSet("publish", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src extensions %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		extensionIDFlag = flagSet.String("extension-id", "", `Override the extension ID in the manifest. (default: read from -manifest file)`)
		manifestFlag    = flagSet.String("manifest", "package.json", `The extension manifest file.`)
		forceFlag       = flagSet.Bool("force", false, `Force publish the extension, even if there are validation problems or other warnings.`)
		apiFlags        = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		manifest, err := ioutil.ReadFile(*manifestFlag)
		if err != nil {
			return fmt.Errorf("%s\n\nRun this command in a directory with a %s file for an extension.\n\nSee 'src extensions %s -h' for help.", err, *manifestFlag, flagSet.Name())
		}
		extensionID := *extensionIDFlag
		if extensionID == "" {
			extensionID, err = readExtensionIDFromManifest(manifest)
			if err != nil {
				return err
			}
		}
		manifest, err = updateExtensionIDInManifest(manifest, extensionID)
		if err != nil {
			return err
		}

		query := `mutation PublishExtension(
  $extensionID: String!,
  $manifest: String!,
  $force: Boolean!,
) {
  extensionRegistry {
    publishExtension(
      extensionID: $extensionID,
      manifest: $manifest,
      force: $force,
    ) {
      extension {
        extensionID
        url
      }
    }
  }
}`

		var result struct {
			ExtensionRegistry struct {
				PublishExtension struct {
					Extension struct {
						ExtensionID string
						URL         string
					}
				}
			}
		}
		return (&apiRequest{
			query: query,
			vars: map[string]interface{}{
				"extensionID": extensionID,
				"manifest":    string(manifest),
				"force":       *forceFlag,
			},
			result: &result,
			done: func() error {
				fmt.Println("Extension published!")
				fmt.Println()
				fmt.Printf("\tExtension ID: %s\n\n", result.ExtensionRegistry.PublishExtension.Extension.ExtensionID)
				fmt.Printf("View, enable, and configure it at: %s\n", cfg.Endpoint+result.ExtensionRegistry.PublishExtension.Extension.URL)
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

	// Catch the mistake of omitting the "extensions" subcommand.
	commands = append(commands, didYouMeanOtherCommand("publish", []string{"extensions publish", "ext publish        (alias)"}))
}

func readExtensionIDFromManifest(manifest []byte) (string, error) {
	var o map[string]interface{}
	if err := json.Unmarshal(manifest, &o); err != nil {
		return "", err
	}

	extensionID, _ := o["extensionID"].(string)
	if extensionID != "" {
		return extensionID, nil
	}

	name, _ := o["name"].(string)
	publisher, _ := o["publisher"].(string)
	if name == "" && publisher == "" {
		return "", errors.New(`extension manifest must contain "name" and "publisher" string properties (the extension ID is of the form "publisher/name" and uses these values)`)
	}
	if name == "" {
		return "", fmt.Errorf(`extension manifest must contain a "name" string property for the extension name (the extension ID will be %q)`, publisher+"/name")
	}
	if publisher == "" {
		return "", fmt.Errorf(`extension manifest must contain a "publisher" string property referring to a username or organization name on Sourcegraph (the extension ID will be %q)`, "publisher/"+name)
	}
	return publisher + "/" + name, nil
}

func updateExtensionIDInManifest(manifest []byte, extensionID string) (updatedManifest []byte, err error) {
	var o map[string]interface{}
	if err := json.Unmarshal(manifest, &o); err != nil {
		return nil, err
	}
	if o == nil {
		o = map[string]interface{}{}
	}
	o["extensionID"] = extensionID
	return json.MarshalIndent(o, "", "  ")
}
