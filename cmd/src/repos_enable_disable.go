package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
)

func init() {
	initEnableDisable("disable", false, `
Examples:

Disable one or more repositories:

	$ src repos disable github.com/my/repo github.com/my/repo2

`)

	initEnableDisable("enable", true, `
Examples:

Enable one or more repositories:

	$ src repos enable github.com/my/repo github.com/my/repo2

`)
}

func initEnableDisable(cmdName string, enabled bool, usage string) {
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

		errs := false
		for _, repoName := range flagSet.Args() {
			if err := setRepositoryEnabled(repoName, enabled); err != nil {
				errs = true
				log.Println(err)
			}
		}
		if errs {
			return errors.New("(errors occurred)")
		}
		return nil
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
