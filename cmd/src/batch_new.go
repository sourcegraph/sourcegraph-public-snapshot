package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"text/template"

	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/batches"
)

func init() {
	usage := `
'src batch new' creates a new batch spec YAML, prefilled with all required
fields.

Usage:

    src batch new [-f FILE]

Examples:


    $ src batch new -f batch.spec.yaml

`

	flagSet := flag.NewFlagSet("new", flag.ExitOnError)

	var (
		fileFlag = flagSet.String("f", "batch.yaml", "The name of the batch spec file to create.")
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		f, err := os.OpenFile(*fileFlag, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			if os.IsExist(err) {
				return fmt.Errorf("file %s already exists", *fileFlag)
			}
			return errors.Wrapf(err, "failed to create file %s", *fileFlag)
		}
		defer f.Close()

		tmpl, err := template.New("").Parse(batchSpecTmpl)
		if err != nil {
			return err
		}

		author := batches.GitCommitAuthor{
			Name:  "Sourcegraph",
			Email: "batch-changes@sourcegraph.com",
		}

		// Try to get better default values from git, ignore any errors.
		if err := checkExecutable("git", "version"); err == nil {
			var gitAuthorName, gitAuthorEmail string
			var err1, err2 error
			gitAuthorName, err1 = getGitConfig("user.name")
			gitAuthorEmail, err2 = getGitConfig("user.email")

			if err1 == nil && err2 == nil && gitAuthorName != "" && gitAuthorEmail != "" {
				author.Name = gitAuthorName
				author.Email = gitAuthorEmail
			}
		}

		err = tmpl.Execute(f, map[string]interface{}{"Author": author})
		if err != nil {
			return errors.Wrap(err, "failed to write batch spec to file")
		}

		fmt.Printf("%s created.\n", *fileFlag)
		return nil
	}

	batchCommands = append(batchCommands, &command{
		flagSet: flagSet,
		aliases: []string{},
		handler: handler,
		usageFunc: func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src batch %s':\n", flagSet.Name())
			flagSet.PrintDefaults()
			fmt.Println(usage)
		},
	})
}

func getGitConfig(attribute string) (string, error) {
	cmd := exec.Command("git", "config", "--get", attribute)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

const batchSpecTmpl = `name: NAME-OF-YOUR-BATCH-CHANGE
description: DESCRIPTION-OF-YOUR-BATCH-CHANGE

# "on" specifies on which repositories to execute the "steps".
on:
  # Example: find all repositories that contain a README.md file.
  - repositoriesMatchingQuery: file:README.md

# "steps" are run in each repository. Each step is run in a Docker container
# with the repository as the working directory. Once complete, each
# repository's resulting diff is captured.
steps:
  # Example: append "Hello World" to every README.md
  - run: echo "Hello World" | tee -a $(find -name README.md)
    container: alpine:3

# "changesetTemplate" describes the changeset (e.g., GitHub pull request) that
# will be created for each repository.
changesetTemplate:
  title: Hello World
  body: This adds Hello World to the README

  branch: BRANCH-NAME-IN-EACH-REPOSITORY # Push the commit to this branch.

  commit:
    author:
      name: {{ .Author.Name }}
      email: {{ .Author.Email }}
    message: Append Hello World to all README.md files

  # Change published to true once you're ready to create changesets on the code host.
  published: false
`
