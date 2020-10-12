package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"text/template"

	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/campaigns"
)

func init() {
	usage := `
'src campaigns new' creates a new campaign spec YAML, prefilled with all
required fields.

Usage:

    src campaigns new [-f FILE]

Examples:


    $ src campaigns new -f campaign.spec.yaml

`

	flagSet := flag.NewFlagSet("new", flag.ExitOnError)

	var (
		fileFlag = flagSet.String("f", "campaign.yaml", "The name of campaign spec file to create.")
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

		tmpl, err := template.New("").Parse(campaignSpecTmpl)
		if err != nil {
			return err
		}

		author := campaigns.GitCommitAuthor{
			Name:  "Sourcegraph",
			Email: "campaigns@sourcegraph.com",
		}

		// Try to get better default values from git, ignore any errors.
		if err := checkExecutable("git", "version"); err == nil {
			var gitAuthorName, gitAuthorEmail string

			gitAuthorName, err = getGitConfig("user.name")
			gitAuthorEmail, err = getGitConfig("user.email")

			if err == nil && gitAuthorName != "" && gitAuthorEmail != "" {
				author.Name = gitAuthorName
				author.Email = gitAuthorEmail
			}
		}

		err = tmpl.Execute(f, map[string]interface{}{"Author": author})
		if err != nil {
			return errors.Wrap(err, "failed to write campaign spec to file")
		}

		fmt.Printf("%s created.\n", *fileFlag)
		return nil
	}

	campaignsCommands = append(campaignsCommands, &command{
		flagSet: flagSet,
		aliases: []string{},
		handler: handler,
		usageFunc: func() {
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src campaigns %s':\n", flagSet.Name())
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

const campaignSpecTmpl = `name: NAME-OF-YOUR-CAMPAIGN
description: DESCRIPTION-OF-YOUR-CAMPAIGN


# "on" specifies on which repositories to execute the "steps"
on:
  # Example: find all repositories that contain a README.md file.
  - repositoriesMatchingQuery: file:README.md

# steps are run in each repository. Each repository's resulting diff is captured.
steps:
  # Example: append "Hello World" to every README.md
  - run: echo "Hello World" | tee -a $(find -name README.md)
    container: alpine:3

# Describe the changeset (e.g., GitHub pull request) you want for each repository.
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
