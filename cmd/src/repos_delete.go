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
	flagSet := flag.NewFlagSet("delete", flag.ExitOnError)
	apiFlags := api.NewFlags(flagSet)

	printUsage := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src repos %s'\n", flagSet.Name())

		flagSet.PrintDefaults()

		examples := `
Examples:

   Delete one or more repositories:

    	$ src repos delete github.com/my/repo github.com/my/repo2
`
		fmt.Fprint(flag.CommandLine.Output(), examples)
	}

	deleteRepository := func(ctx context.Context, client api.Client, repoName string) error {
		repoID, err := fetchRepositoryID(ctx, client, repoName)
		if err != nil {
			return err
		}

		query := `mutation DeleteRepository($repoID: ID!){
			deleteRepository(repository: $repoID) {
				alwaysNil
			}
		}`
		var result struct{}
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"repoID": repoID,
		}).Do(ctx, &result); err != nil || !ok {
			return err
		}

		fmt.Fprintf(flag.CommandLine.Output(), "Repository %q deleted\n", repoName)
		return nil
	}

	deleteRepositories := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		ctx := context.Background()
		client := cfg.apiClient(apiFlags, flagSet.Output())

		var errs *multierror.Error
		for _, repoName := range flagSet.Args() {
			err := deleteRepository(ctx, client, repoName)
			if err != nil {
				err = errors.Wrapf(err, "Failed to delete repository %q", repoName)
				errs = multierror.Append(errs, err)
			}
		}
		return errs.ErrorOrNil()
	}

	// Register the command.
	reposCommands = append(reposCommands, &command{
		flagSet:   flagSet,
		handler:   deleteRepositories,
		usageFunc: printUsage,
	})
}
