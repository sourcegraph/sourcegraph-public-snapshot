package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	isatty "github.com/mattn/go-isatty"
	"github.com/sourcegraph/jsonx"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/cmderrors"
)

func init() {
	usage := `
Examples:

  Edit an external service configuration on the Sourcegraph instance:

    	$ cat new-config.json | src extsvc edit -id 'RXh0ZXJuYWxTZXJ2aWNlOjQ='
    	$ src extsvc edit -name 'My GitHub connection' new-config.json

  Edit an external service name on the Sourcegraph instance:

    	$ src extsvc edit -name 'My GitHub connection' -rename 'New name'

  Add some repositories to the exclusion list of the external service:

    	$ src extsvc edit -name 'My GitHub connection' -exclude-repos 'github.com/foo/one' 'github.com/foo/two'
`

	flagSet := flag.NewFlagSet("edit", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src extsvc %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		nameFlag                = flagSet.String("name", "", "exact name of the external service to edit")
		idFlag                  = flagSet.String("id", "", "ID of the external service to edit")
		renameFlag              = flagSet.String("rename", "", "when specified, renames the external service")
		excludeRepositoriesFlag = flagSet.String("exclude-repos", "", "when specified, add these repositories to the exclusion list")
		apiFlags                = api.NewFlags(flagSet)
	)

	handler := func(args []string) (err error) {
		if err := flagSet.Parse(args); err != nil {
			return err
		}
		ctx := context.Background()
		client := cfg.apiClient(apiFlags, flagSet.Output())

		// Determine ID of external service we will edit.
		if *nameFlag == "" && *idFlag == "" {
			return cmderrors.Usage("one of -name or -id flag must be specified")
		}
		id := *idFlag
		if id == "" {
			svc, err := lookupExternalService(ctx, client, "", *nameFlag)
			if err != nil {
				return err
			}
			id = svc.ID
		}

		// Determine if we are updating the JSON configuration or not.
		var updateJSON []byte
		if len(flagSet.Args()) == 1 {
			updateJSON, err = os.ReadFile(flagSet.Arg(0))
			if err != nil {
				return err
			}
		}
		if !isatty.IsTerminal(os.Stdin.Fd()) {
			// stdin is a pipe not a terminal
			updateJSON, err = io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
		}

		if *excludeRepositoriesFlag != "" {
			if len(updateJSON) == 0 {
				// We need to fetch the current JSON then.
				svc, err := lookupExternalService(ctx, client, id, "")
				if err != nil {
					return err
				}
				updateJSON = []byte(svc.Config)
			}
			updated, err := appendExcludeRepositories(string(updateJSON), strings.Fields(*excludeRepositoriesFlag))
			if err != nil {
				return err
			}
			updateJSON = []byte(updated)
		}

		updateExternalServiceInput := map[string]interface{}{
			"id": id,
		}
		if *renameFlag != "" {
			updateExternalServiceInput["displayName"] = *renameFlag
		}
		if len(updateJSON) > 0 {
			updateExternalServiceInput["config"] = string(updateJSON)
		}
		if len(updateExternalServiceInput) == 1 {
			return nil // nothing to update
		}

		queryVars := map[string]interface{}{
			"input": updateExternalServiceInput,
		}
		var result struct{} // TODO: future: allow formatting resulting external service
		if ok, err := client.NewRequest(externalServicesUpdateMutation, queryVars).Do(ctx, &result); err != nil {
			if strings.Contains(err.Error(), "Additional property exclude is not allowed") {
				return errors.New(`specified external service does not support repository "exclude" list`)
			}
			return err
		} else if ok {
			fmt.Println("External service updated:", id)
		}
		return nil
	}

	// Register the command.
	extsvcCommands = append(extsvcCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}

const externalServicesUpdateMutation = `
mutation ($input: UpdateExternalServiceInput!) {
	updateExternalService(input: $input) {
		id
	}
}
`

// appendExcludeRepositories appends to the ".exclude" field of the given jsonx
// the input list of repo names to exclude. It creates the exclude field if it
// doesn't exist.
func appendExcludeRepositories(input string, excludeRepoNames []string) (string, error) {
	// Known issue: Comments are not retained in the existing array value.
	var m interface{}
	if err := jsonxUnmarshal(input, &m); err != nil {
		return "", err
	}
	root, ok := m.(map[string]interface{})
	if !ok {
		return "", errors.New("existing JSONx external service configuration is invalid (not an object)")
	}
	var exclude []interface{}
	alreadyExcludedNames := map[string]struct{}{}
	if existing, ok := root["exclude"]; ok {
		exclude, ok = existing.([]interface{})
		if !ok {
			return "", errors.New("existing JSONx external service configuration is invalid (exclude is not an array)")
		}
		for i, exclude := range exclude {
			obj, ok := exclude.(map[string]interface{})
			if !ok {
				return "", errors.Newf("existing JSONx external service configuration is invalid (exclude.%d is not object)", i)
			}
			name, ok := obj["name"]
			if !ok {
				continue
			}
			nameStr, ok := name.(string)
			if !ok {
				continue
			}
			alreadyExcludedNames[nameStr] = struct{}{}
		}
	}
	for _, repoName := range excludeRepoNames {
		if strings.TrimSpace(repoName) == "" {
			continue
		}
		if _, already := alreadyExcludedNames[repoName]; already {
			continue
		}
		exclude = append(exclude, map[string]interface{}{
			"name": repoName,
		})
	}
	edits, _, err := jsonx.ComputePropertyEdit(
		input,
		jsonx.PropertyPath("exclude"),
		exclude,
		nil,
		jsonx.FormatOptions{InsertSpaces: true, TabSize: 2},
	)
	if err != nil {
		return "", err
	}
	return jsonx.ApplyEdits(input, edits...)
}

// jsonxToJSON converts jsonx to plain JSON.
func jsonxToJSON(text string) ([]byte, error) {
	data, errs := jsonx.Parse(text, jsonx.ParseOptions{Comments: true, TrailingCommas: true})
	if len(errs) > 0 {
		return data, errors.Newf("failed to parse JSON: %v", errs)
	}
	return data, nil
}

// jsonxUnmarshal unmarshals jsonx into Go data.
//
// This process loses comments, trailing commas, formatting, etc.
func jsonxUnmarshal(text string, v interface{}) error {
	data, err := jsonxToJSON(text)
	if err != nil {
		return err
	}
	if strings.TrimSpace(text) == "" {
		return nil
	}
	return json.Unmarshal(data, v)
}
