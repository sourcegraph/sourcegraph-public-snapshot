package main

import (
	"flag"
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

func init() {
	flagSet := flag.NewFlagSet("delete", flag.ExitOnError)
	apiFlags := newAPIFlags(flagSet)

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

	deleteRepository := func(repoName string) error {
		repoID, err := fetchRepositoryID(repoName)
		if err != nil {
			return err
		}

		query := `mutation DeleteRepository($repoID: ID!){
			deleteRepository(repository: $repoID) {
				alwaysNil
			}
		}`
		var result struct{}
		return (&apiRequest{
			query: query,
			vars: map[string]interface{}{
				"repoID": repoID,
			},
			result: &result,
			done: func() error {
				fmt.Fprintf(flag.CommandLine.Output(), "Repository %q deleted\n", repoName)
				return nil
			},
			flags: apiFlags,
		}).do()
	}

	deleteRepositories := func(args []string) error {
		flagSet.Parse(args)
		var errs *multierror.Error
		for _, repoName := range flagSet.Args() {
			err := deleteRepository(repoName)
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
