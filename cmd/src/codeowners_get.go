package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/cmderrors"
)

func init() {
	usage := `
Examples:

  Read the current codeowners file for the repository "github.com/sourcegraph/sourcegraph":

    	$ src codeowners get -repo='github.com/sourcegraph/sourcegraph'
`

	flagSet := flag.NewFlagSet("get", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src codeowners %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		repoFlag = flagSet.String("repo", "", "The repository to attach the data to")
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

		query := `query GetCodeownersFile(
	$repoName: String!
) {
	repository(name: $repoName) {
		ingestedCodeowners {
			...CodeownersFileFields			
		}
	}
}
` + codeownersFragment

		var result struct {
			Repository *struct {
				IngestedCodeowners *CodeownersIngestedFile
			}
		}
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"repoName": *repoFlag,
		}).Do(context.Background(), &result); err != nil || !ok {
			return err
		}

		if result.Repository == nil {
			return cmderrors.ExitCode(2, errors.Newf("repository %q not found", *repoFlag))
		}

		if result.Repository.IngestedCodeowners == nil {
			return cmderrors.ExitCode(2, errors.Newf("no codeowners data found for %q", *repoFlag))
		}

		fmt.Fprintf(os.Stdout, "%s", result.Repository.IngestedCodeowners.Contents)

		return nil
	}

	// Register the command.
	codeownersCommands = append(codeownersCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
