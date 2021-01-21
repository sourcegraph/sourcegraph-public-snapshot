package main

import (
	"context"
	"flag"
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/api"
)

func init() {
	initReposEnableDisable("disable", false, `
Examples:

  Disable one or more repositories:

    	$ src repos disable github.com/my/repo github.com/my/repo2

`)

	initReposEnableDisable("enable", true, `
Examples:

  Enable one or more repositories:

    	$ src repos enable github.com/my/repo github.com/my/repo2

`)
}

func initReposEnableDisable(cmdName string, enable bool, usage string) {
	flagSet := flag.NewFlagSet(cmdName, flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src repos %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	apiFlags := api.NewFlags(flagSet)

	setRepositoryEnabled := func(ctx context.Context, client api.Client, repoName string, enabled bool) error {
		repoID, err := fetchRepositoryID(ctx, client, repoName)
		if err != nil {
			return err
		}

		query := `mutation SetRepositoryEnabled($repoID: ID!, $enabled: Boolean!){
  setRepositoryEnabled(repository: $repoID, enabled: $enabled) {
    alwaysNil
  }
}`

		var result struct{}
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"repoID":  repoID,
			"enabled": enabled,
		}).Do(ctx, &result); err != nil || !ok {
			return err
		}

		fmt.Printf("repository %sd: %s\n", cmdName, repoName)
		return nil
	}

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		ctx := context.Background()
		client := cfg.apiClient(apiFlags, flagSet.Output())

		var errs *multierror.Error
		for _, repoName := range flagSet.Args() {
			if err := setRepositoryEnabled(ctx, client, repoName, enable); err != nil {
				err = errors.Wrapf(err, "Failed to %s repository %q", cmdName, repoName)
				errs = multierror.Append(errs, err)
			}
		}
		return errs.ErrorOrNil()
	}

	// Register the command.
	reposCommands = append(reposCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}

func fetchRepositoryID(ctx context.Context, client api.Client, repoName string) (string, error) {
	query := `query RepositoryID($repoName: String!) {
  repository(name: $repoName) {
    id
  }
}`

	var result struct {
		Repository struct {
			ID string
		}
	}
	if ok, err := client.NewRequest(query, map[string]interface{}{
		"repoName": repoName,
	}).Do(ctx, &result); err != nil || !ok {
		return "", err
	}
	if result.Repository.ID == "" {
		return "", fmt.Errorf("repository not found: %s", repoName)
	}
	return result.Repository.ID, nil
}
