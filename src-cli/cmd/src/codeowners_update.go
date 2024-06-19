package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/cmderrors"
)

func init() {
	usage := `
Examples:

  Update a codeowners file for the repository "github.com/sourcegraph/sourcegraph":

    	$ src codeowners update -repo='github.com/sourcegraph/sourcegraph' -f CODEOWNERS

  Update a codeowners file for the repository "github.com/sourcegraph/sourcegraph" from stdin:

    	$ src codeowners update -repo='github.com/sourcegraph/sourcegraph' -f -
`

	flagSet := flag.NewFlagSet("update", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src codeowners %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		repoFlag      = flagSet.String("repo", "", "The repository to attach the data to")
		fileFlag      = flagSet.String("file", "", "File path to read ownership information from (- for stdin)")
		fileShortFlag = flagSet.String("f", "", "File path to read ownership information from (- for stdin). Alias for -file")
		apiFlags      = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		if *repoFlag == "" {
			return errors.New("provide a repo name using -repo")
		}

		if *fileFlag == "" && *fileShortFlag == "" {
			return errors.New("provide a file using -file")
		}
		if *fileFlag != "" && *fileShortFlag != "" {
			return errors.New("have to provide either -file or -f")
		}
		if *fileShortFlag != "" {
			*fileFlag = *fileShortFlag
		}

		file, err := readFile(*fileFlag)
		if err != nil {
			return err
		}

		content, err := io.ReadAll(file)
		if err != nil {
			return err
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		query := `mutation UpdateCodeownersFile(
	$repoName: String!,
	$content: String!
) {
	updateCodeownersFile(input: {
		repoName: $repoName,
		fileContents: $content,
	}) {
		...CodeownersFileFields
	}
}
` + codeownersFragment

		var result struct {
			UpdateCodeownersFile CodeownersIngestedFile
		}
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"repoName": *repoFlag,
			"content":  string(content),
		}).Do(context.Background(), &result); err != nil || !ok {
			var gqlErr api.GraphQlErrors
			if errors.As(err, &gqlErr) {
				for _, e := range gqlErr {
					if strings.Contains(e.Error(), "repo not found:") {
						return cmderrors.ExitCode(2, errors.Newf("repository %q not found", *repoFlag))
					}
					if strings.Contains(e.Error(), "could not update codeowners file: codeowners file not found:") {
						return cmderrors.ExitCode(2, errors.New("no codeowners data has been found for this repository"))
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
