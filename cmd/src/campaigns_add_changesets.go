package main

import (
	"errors"
	"flag"
	"fmt"
)

func init() {
	usage := `
Add changesets for a given repository to a campaign.

Usage:

	src campaigns add-changesets -campaign <id> -repo-name <name> [external changeset IDs...]

Examples:

  Add GitHub pull requests #5662 and #5366 for repository "github.com/sourcegraph/sourcegraph" to the Sourcegraph campaign with GraphQL ID "Q2FtcGFpZ246MQ==":

    	$ src campaigns add-changesets -campaign=Q2FtcGFpZ246MQ== -repo-name=github.com/sourcegraph/sourcegraph 5662 5366

  Then we can add pull requests from another repository, "github.com/sourcegraph/src-cli" to the same campaign:

    	$ src campaigns add-changesets -campaign=Q2FtcGFpZ246MQ== -repo-name=github.com/sourcegraph/src-cli 33 41

Notes:

  You can NOT add changesets to a campaign if the repository is not mirrored on the Sourcegraph instance.

  The repository names are the names you get when you run "src repos list".
`

	flagSet := flag.NewFlagSet("add-changesets", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src campaigns %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		campaignIDFlag = flagSet.String("campaign", "", "ID of campaign to which to add changesets. (required)")
		repoNameFlag   = flagSet.String("repo-name", "", "Name of repository to which the changesets belong. (required)")
		apiFlags       = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		if *campaignIDFlag == "" {
			return &usageError{errors.New("-campaign must be specified")}
		}

		if *repoNameFlag == "" {
			return &usageError{errors.New("-repo-name must be specified")}
		}

		if len(args) <= 4 {
			return &usageError{errors.New("no external changeset IDs specified")}
		}

		externalIDs := args[4:]

		repoID, err := getRepoID(apiFlags, *repoNameFlag)
		if err != nil {
			return err
		}

		changesetIDs, err := createChangesets(apiFlags, repoID, externalIDs)
		if err != nil {
			return err
		}

		err = addChangesets(apiFlags, *campaignIDFlag, changesetIDs)
		if err != nil {
			return err
		}

		return nil
	}

	// Register the command.
	campaignsCommands = append(campaignsCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})

	// Catch the mistake of omitting the "extensions" subcommand.
	commands = append(commands, didYouMeanOtherCommand("add-changesets", []string{"campaigns add-changesets"}))
}

const getRepoIDQuery = `query Repository($name: String) { repository(name: $name) { id } }`

func getRepoID(f *apiFlags, name string) (string, error) {
	var result struct{ Repository struct{ ID string } }

	req := &apiRequest{
		query:  getRepoIDQuery,
		vars:   map[string]interface{}{"name": name},
		result: &result,
		flags:  f,
	}

	err := req.do()
	if err != nil {
		return "", err
	}

	return result.Repository.ID, err
}

const createChangesetsQuery = `
mutation CreateChangesets($input: [CreateChangesetInput!]!) {
  createChangesets(input: $input) {
    id
  }
}`

func createChangesets(f *apiFlags, repoID string, externalIDs []string) ([]string, error) {
	var result struct {
		CreateChangesets []struct {
			ID string `json:"id"`
		} `json:"createChangesets"`
	}

	pairs := make([]map[string]interface{}, len(externalIDs))
	for i, id := range externalIDs {
		pairs[i] = map[string]interface{}{
			"repository": repoID,
			"externalID": id,
		}
	}

	var changesetIDs []string

	req := &apiRequest{
		query:  createChangesetsQuery,
		vars:   map[string]interface{}{"input": pairs},
		result: &result,
		flags:  f,
	}

	err := req.do()
	if err != nil {
		return changesetIDs, err
	}

	for _, c := range result.CreateChangesets {
		changesetIDs = append(changesetIDs, c.ID)
	}

	fmt.Printf("Created %d changesets.\n", len(changesetIDs))

	return changesetIDs, err
}

const addChangesetsQuery = `
mutation AddChangesetsToCampaign($campaign: ID!, $changesets: [ID!]!) {
  addChangesetsToCampaign(campaign: $campaign, changesets: $changesets) {
    id
    changesets {
      totalCount
    }
  }
}
`

func addChangesets(f *apiFlags, campaignID string, changesetIDs []string) error {
	var result struct {
		AddChangesetsToCampaign struct {
			ID         string `json:"id"`
			Changesets struct {
				TotalCount int `json:"totalCount"`
			} `json:"changesets"`
		} `json:"addChangesetsToCampaign"`
	}

	req := &apiRequest{
		query: addChangesetsQuery,
		vars: map[string]interface{}{
			"campaign":   campaignID,
			"changesets": changesetIDs,
		},
		result: &result,
		flags:  f,
	}

	err := req.do()
	if err != nil {
		return err
	}

	fmt.Printf("Added changeset to campaign. Changesets now in campaign: %d\n", result.AddChangesetsToCampaign.Changesets.TotalCount)

	return err
}
