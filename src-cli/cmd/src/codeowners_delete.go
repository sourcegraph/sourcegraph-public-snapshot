package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/cmderrors"
)

func init() {
	usage := `
Examples:

  Delete a codeowners file for the repository "github.com/sourcegraph/sourcegraph":

    	$ src codeowners delete -repo='github.com/sourcegraph/sourcegraph'
`

	flagSet := flag.NewFlagSet("delete", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src codeowners %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		repoFlag = flagSet.String("repo", "", "The repository to delete the data for")
		apiFlags = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		if *repoFlag == "" {
			return errors.New("provide a repo name using -repo")
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		query := `mutation DeleteCodeownersFile(
	$repoName: String!,
) {
	deleteCodeownersFiles(repositories: [{
		repoName: $repoName,
	}]) {
		alwaysNil
	}
}
`

		var result struct {
			DeleteCodeownersFile CodeownersIngestedFile
		}
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"repoName": *repoFlag,
		}).Do(context.Background(), &result); err != nil || !ok {
			var gqlErr api.GraphQlErrors
			if errors.As(err, &gqlErr) {
				for _, e := range gqlErr {
					if strings.Contains(e.Error(), "repo not found:") {
						return cmderrors.ExitCode(2, errors.Newf("repository %q not found", *repoFlag))
					}
					if strings.Contains(e.Error(), "codeowners file not found:") {
						return cmderrors.ExitCode(2, errors.Newf("no data found for repository %q", *repoFlag))
					}
				}
			}
			return err
		}

		return nil
	}

	// Register the command.
	codeownersCommands = append(codeownersCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
