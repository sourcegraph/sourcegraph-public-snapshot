package main

import (
	"flag"
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
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
	apiFlags := newAPIFlags(flagSet)

	setRepositoryEnabled := func(repoName string, enabled bool) error {
		repoID, err := fetchRepositoryID(repoName)
		if err != nil {
			return err
		}

		query := `mutation SetRepositoryEnabled($repoID: ID!, $enabled: Boolean!){
  setRepositoryEnabled(repository: $repoID, enabled: $enabled) {
    alwaysNil
  }
}`

		var result struct{}
		return (&apiRequest{
			query: query,
			vars: map[string]interface{}{
				"repoID":  repoID,
				"enabled": enabled,
			},
			result: &result,
			done: func() error {
				fmt.Printf("repository %sd: %s\n", cmdName, repoName)
				return nil
			},
			flags: apiFlags,
		}).do()
	}

	handler := func(args []string) error {
		flagSet.Parse(args)

		var errs *multierror.Error
		for _, repoName := range flagSet.Args() {
			if err := setRepositoryEnabled(repoName, enable); err != nil {
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

func fetchRepositoryID(repoName string) (string, error) {
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
	err := (&apiRequest{
		query: query,
		vars: map[string]interface{}{
			"repoName": repoName,
		},
		result: &result,
	}).do()
	if err != nil {
		return "", err
	}
	if result.Repository.ID == "" {
		return "", fmt.Errorf("repository not found: %s", repoName)
	}
	return result.Repository.ID, nil
}
