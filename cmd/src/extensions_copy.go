package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func withCfg(new *config, f func()) {
	old := cfg
	cfg = new
	f()
	cfg = old
}

func init() {
	usage := `
Copy an extension from Sourcegraph.com to your private registry.
`

	flagSet := flag.NewFlagSet("copy", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src extensions %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		extensionIDFlag = flagSet.String("extension-id", "", `The <extID> in https://sourcegraph.com/extensions/<extID> (e.g. sourcegraph/java)`)
		currentUserFlag = flagSet.String("current-user", "", `The current user`)
		apiFlags        = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		extensionID := *extensionIDFlag
		if extensionID == "" {
			return fmt.Errorf("must provide -extension-id")
		}

		currentUser := *currentUserFlag
		if currentUser == "" {
			return fmt.Errorf("must provide -current-user")
		}

		extensionIDParts := strings.Split(extensionID, "/")
		if len(extensionIDParts) != 2 {
			return fmt.Errorf("-extension-id must have the form <publisher>/<name>")
		}
		extensionName := extensionIDParts[1]

		var extensionResult struct {
			ExtensionRegistry struct {
				Extension struct {
					Manifest struct {
						Raw       string
						BundleURL string
					}
				}
			}
		}

		var err error
		withCfg(&config{Endpoint: "https://sourcegraph.com"}, func() {
			err = (&apiRequest{
				query: `query GetExtension(
	$extensionID: String!
){
  extensionRegistry{
    extension(extensionID: $extensionID) {
      manifest{
        raw
        bundleURL
      }
    }
  }
}`,
				vars: map[string]interface{}{
					"extensionID": extensionID,
				},
				result: &extensionResult,
				flags:  apiFlags,
			}).do()
		})
		if err != nil {
			return err
		}

		rawManifest := []byte(extensionResult.ExtensionRegistry.Extension.Manifest.Raw)
		manifest, err := updatePropertyInManifest(rawManifest, "extensionID", extensionID)
		if err != nil {
			return err
		}

		response, err := http.Get(extensionResult.ExtensionRegistry.Extension.Manifest.BundleURL)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		bundle, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}
		fmt.Printf("bundle: %s\n", string(bundle[0:100]))
		fmt.Printf("manifest: %s\n", string(manifest[0:]))

		var publishResult struct {
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
			query: `mutation PublishExtension(
	$extensionID: String!,
	$manifest: String!,
	$bundle: String,
) {
	extensionRegistry {
		publishExtension(
			extensionID: $extensionID,
			manifest: $manifest,
			bundle: $bundle
		) {
			extension {
				extensionID
				url
			}
		}
	}
}`,
			vars: map[string]interface{}{
				"extensionID": currentUser + "/" + extensionName,
				"manifest":    string(manifest),
				"bundle":      bundle,
			},
			result: &publishResult,
			done: func() error {
				fmt.Println("Extension published!")
				fmt.Println()
				fmt.Printf("\tExtension ID: %s\n\n", publishResult.ExtensionRegistry.PublishExtension.Extension.ExtensionID)
				fmt.Printf("View, enable, and configure it at: %s\n", cfg.Endpoint+publishResult.ExtensionRegistry.PublishExtension.Extension.URL)
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
