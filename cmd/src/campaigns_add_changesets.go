package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/sourcegraph/src-cli/internal/api"
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
		apiFlags       = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		err := flagSet.Parse(args)
		if err != nil {
			return err
		}

		if *campaignIDFlag == "" {
			return &usageError{errors.New("-campaign must be specified")}
		}

		if *repoNameFlag == "" {
			return &usageError{errors.New("-repo-name must be specified")}
		}

		if len(args) <= 2 {
			return &usageError{errors.New("no external changeset IDs specified")}
		}

		externalIDs := args[2:]

		ctx := context.Background()
		client := cfg.apiClient(apiFlags, flagSet.Output())

		repoID, err := getRepoID(ctx, client, *repoNameFlag)
		if err != nil {
			return err
		}

		changesetIDs, err := createChangesets(ctx, client, repoID, externalIDs)
		if err != nil {
			return err
		}

		err = addChangesets(ctx, client, *campaignIDFlag, changesetIDs)
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

func getRepoID(ctx context.Context, client api.Client, name string) (string, error) {
	var result struct{ Repository struct{ ID string } }

	_, err := client.NewRequest(getRepoIDQuery, map[string]interface{}{
		"name": name,
	}).Do(ctx, &result)

	return result.Repository.ID, err
}

const createChangesetsQuery = `
mutation CreateChangesets($input: [CreateChangesetInput!]!) {
  createChangesets(input: $input) {
    id
  }
}`

func createChangesets(ctx context.Context, client api.Client, repoID string, externalIDs []string) ([]string, error) {
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

	if ok, err := client.NewRequest(createChangesetsQuery, map[string]interface{}{
		"input": pairs,
	}).Do(ctx, &result); err != nil || !ok {
		return changesetIDs, err
	}

	for _, c := range result.CreateChangesets {
		changesetIDs = append(changesetIDs, c.ID)
	}

	fmt.Printf("Created %d changesets.\n", len(changesetIDs))

	return changesetIDs, nil
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

func addChangesets(ctx context.Context, client api.Client, campaignID string, changesetIDs []string) error {
	var result struct {
		AddChangesetsToCampaign struct {
			ID         string `json:"id"`
			Changesets struct {
				TotalCount int `json:"totalCount"`
			} `json:"changesets"`
		} `json:"addChangesetsToCampaign"`
	}

	if ok, err := client.NewRequest(addChangesetsQuery, map[string]interface{}{
		"campaign":   campaignID,
		"changesets": changesetIDs,
	}).Do(ctx, &result); err != nil || !ok {
		return err
	}

	fmt.Printf("Added changeset to campaign. Changesets now in campaign: %d\n", result.AddChangesetsToCampaign.Changesets.TotalCount)
	return nil
}
