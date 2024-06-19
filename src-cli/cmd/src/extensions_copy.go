package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sourcegraph/src-cli/internal/api"

	"github.com/sourcegraph/sourcegraph/lib/errors"
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
		apiFlags        = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		err := flagSet.Parse(args)
		if err != nil {
			return err
		}

		extensionID := *extensionIDFlag
		if extensionID == "" {
			return errors.Newf("must provide -extension-id")
		}

		currentUser := *currentUserFlag
		if currentUser == "" {
			return errors.Newf("must provide -current-user")
		}

		extensionIDParts := strings.Split(extensionID, "/")
		if len(extensionIDParts) != 2 {
			return errors.Newf("-extension-id must have the form <publisher>/<name>")
		}
		extensionName := extensionIDParts[1]

		ctx := context.Background()
		client := cfg.apiClient(apiFlags, flagSet.Output())
		ok := false

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

		withCfg(&config{Endpoint: "https://sourcegraph.com"}, func() {
			dotComClient := cfg.apiClient(apiFlags, flagSet.Output())
			query := `query GetExtension(
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
}`

			ok, err = dotComClient.NewRequest(query, map[string]interface{}{
				"extensionID": extensionID,
			}).Do(ctx, &extensionResult)
		})
		if err != nil || !ok {
			return err
		}

		rawManifest := []byte(extensionResult.ExtensionRegistry.Extension.Manifest.Raw)
		manifest, err := updatePropertyInManifest(rawManifest, "extensionID", extensionID)
		if err != nil {
			return err
		}
		// Remove sourcegraph.com bundle URL.
		manifest, err = updatePropertyInManifest(manifest, "url", "")
		if err != nil {
			return err
		}

		response, err := http.Get(extensionResult.ExtensionRegistry.Extension.Manifest.BundleURL)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		bundle, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		fmt.Printf("bundle: %s\n", string(bundle[0:100]))
		fmt.Printf("manifest: %s\n", string(manifest[0:]))

		query := `mutation PublishExtension(
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
}`

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
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"extensionID": currentUser + "/" + extensionName,
			"manifest":    string(manifest),
			"bundle":      bundle,
		}).Do(ctx, &publishResult); err != nil || !ok {
			return err
		}

		fmt.Println("Extension published!")
		fmt.Println()
		fmt.Printf("\tExtension ID: %s\n\n", publishResult.ExtensionRegistry.PublishExtension.Extension.ExtensionID)
		fmt.Printf("View, enable, and configure it at: %s\n", cfg.Endpoint+publishResult.ExtensionRegistry.PublishExtension.Extension.URL)
		return nil
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
