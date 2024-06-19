package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/api"
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
          "main":      "dist/myext.js",
          "scripts":   {"sourcegraph:prepublish": "parcel build --out-file dist/myext.js src/myext.ts"}
        }
    	$ src extensions publish

Notes:

  Source maps are supported (for easier debugging of extensions). If the main JavaScript bundle is "dist/myext.js",
  it looks for a source map in "dist/myext.map".

`

	flagSet := flag.NewFlagSet("publish", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src extensions %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		extensionIDFlag = flagSet.String("extension-id", "", `Override the extension ID in the manifest. (default: read from -manifest file)`)
		urlFlag         = flagSet.String("url", "", `Override the URL for the bundle. (example: set to http://localhost:1234/myext.js for local dev with parcel)`)
		gitHeadFlag     = flagSet.String("git-head", "", "Override the current git commit for the bundle. (default: uses `git rev-parse head`)")
		manifestFlag    = flagSet.String("manifest", "package.json", `The extension manifest file.`)
		forceFlag       = flagSet.Bool("force", false, `Force publish the extension, even if there are validation problems or other warnings.`)
		apiFlags        = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		manifestPath, err := filepath.Abs(*manifestFlag)
		if err != nil {
			return err
		}
		manifestDir := filepath.Dir(manifestPath)

		manifest, err := os.ReadFile(manifestPath)
		if err != nil {
			return errors.Newf("%s\n\nRun this command in a directory with a %s file for an extension.\n\nSee 'src extensions %s -h' for help", err, *manifestFlag, flagSet.Name())
		}
		extensionID := *extensionIDFlag
		if extensionID == "" {
			extensionID, err = readExtensionIDFromManifest(manifest)
			if err != nil {
				return err
			}
		}
		manifest, err = updatePropertyInManifest(manifest, "extensionID", extensionID)
		if err != nil {
			return err
		}
		manifest, err = addReadmeToManifest(manifest, manifestDir)
		if err != nil {
			return err
		}

		var bundle, sourceMap *string
		if *urlFlag != "" {
			manifest, err = updatePropertyInManifest(manifest, "url", *urlFlag)
			if err != nil {
				return err
			}
		} else {
			// Prepare and upload bundle.
			if err := runManifestPrepublishScript(manifest, manifestDir); err != nil {
				return err
			}

			var err error
			bundle, sourceMap, err = readExtensionArtifacts(manifest, manifestDir)
			if err != nil {
				return err
			}
		}

		if *gitHeadFlag != "" {
			manifest, err = updatePropertyInManifest(manifest, "gitHead", *gitHeadFlag)
			if err != nil {
				return err
			}
		} else {
			command := exec.Command("git", "rev-parse", "head")
			command.Dir = manifestDir

			out, err := command.CombinedOutput()
			if err != nil {
				fmt.Printf("failed to determine git head: %q\n", err)
			} else {
				manifest, err = updatePropertyInManifest(manifest, "gitHead", strings.TrimSpace(string(out)))
				if err != nil {
					return err
				}
			}
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		query := `mutation PublishExtension(
  $extensionID: String!,
  $manifest: String!,
  $bundle: String,
  $sourceMap: String,
  $force: Boolean!,
) {
  extensionRegistry {
    publishExtension(
      extensionID: $extensionID,
      manifest: $manifest,
      bundle: $bundle,
      sourceMap: $sourceMap,
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
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"extensionID": extensionID,
			"manifest":    string(manifest),
			"bundle":      bundle,
			"sourceMap":   sourceMap,
			"force":       *forceFlag,
		}).Do(context.Background(), &result); err != nil || !ok {
			return err
		}

		fmt.Println("Extension published!")
		fmt.Println()
		fmt.Printf("\tExtension ID: %s\n\n", result.ExtensionRegistry.PublishExtension.Extension.ExtensionID)
		fmt.Printf("View, enable, and configure it at: %s\n", cfg.Endpoint+result.ExtensionRegistry.PublishExtension.Extension.URL)
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

func runManifestPrepublishScript(manifest []byte, dir string) error {
	var o struct {
		Scripts struct {
			SourcegraphPrepublish string `json:"sourcegraph:prepublish"`
		} `json:"scripts"`
	}
	if err := json.Unmarshal(manifest, &o); err != nil {
		return err
	}

	if o.Scripts.SourcegraphPrepublish == "" {
		return nil
	}
	cmd := exec.Command("bash", "-c", o.Scripts.SourcegraphPrepublish)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PATH=%s:%s", filepath.Join(dir, "node_modules", ".bin"), os.Getenv("PATH")))
	cmd.Dir = dir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	fmt.Fprintf(os.Stderr, "# sourcegraph:prepublish: %s\n", o.Scripts.SourcegraphPrepublish)
	if err := cmd.Run(); err != nil {
		return errors.Newf("sourcegraph:prepublish script failed: %w (see output above)", err)
	}
	fmt.Fprintln(os.Stderr)
	return nil
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
		return "", errors.Newf(`extension manifest must contain a "name" string property for the extension name (the extension ID will be %q)`, publisher+"/name")
	}
	if publisher == "" {
		return "", errors.Newf(`extension manifest must contain a "publisher" string property referring to a username or organization name on Sourcegraph (the extension ID will be %q)`, "publisher/"+name)
	}
	return publisher + "/" + name, nil
}

func updatePropertyInManifest(manifest []byte, property, value string) (updatedManifest []byte, err error) {
	var o map[string]interface{}
	if err := json.Unmarshal(manifest, &o); err != nil {
		return nil, err
	}
	if o == nil {
		o = map[string]interface{}{}
	}
	if value == "" {
		delete(o, property)
	} else {
		o[property] = value
	}
	return json.MarshalIndent(o, "", "  ")
}

func addReadmeToManifest(manifest []byte, dir string) ([]byte, error) {
	var readme string
	filenames := []string{"README.md", "README.txt", "README", "readme.md", "readme.txt", "readme", "Readme.md", "Readme.txt", "Readme"}
	for _, f := range filenames {
		data, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			continue
		}
		readme = string(data)
		break
	}

	if readme == "" {
		return manifest, nil
	}

	var o map[string]interface{}
	if err := json.Unmarshal(manifest, &o); err != nil {
		return nil, err
	}
	if o == nil {
		o = map[string]interface{}{}
	}
	o["readme"] = readme
	return json.MarshalIndent(o, "", "  ")
}

func readExtensionArtifacts(manifest []byte, dir string) (bundle, sourceMap *string, err error) {
	var o struct {
		Main string `json:"main"`
	}
	if err := json.Unmarshal(manifest, &o); err != nil {
		return nil, nil, err
	}
	if o.Main == "" {
		return nil, nil, nil
	}

	mainPath := filepath.Join(dir, o.Main)

	data, err := os.ReadFile(mainPath)
	if err != nil {
		return nil, nil, errors.Newf(`extension manifest "main" bundle file: %s`, err)
	}
	{
		tmp := string(data)
		bundle = &tmp
	}

	// Guess that source map is the main file with a ".map" extension.
	sourceMapPath := strings.TrimSuffix(mainPath, filepath.Ext(mainPath)) + ".map"
	data, err = os.ReadFile(sourceMapPath)
	if err == nil {
		tmp := string(data)
		sourceMap = &tmp
	} else if !os.IsNotExist(err) {
		return nil, nil, err
	}

	return bundle, sourceMap, nil
}
