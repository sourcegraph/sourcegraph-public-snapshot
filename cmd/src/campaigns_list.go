package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/sourcegraph/src-cli/internal/api"
)

func init() {
	usage := `
Examples:

  List campaigns (default limit is 1000):

    	$ src campaigns list

  List only the first 5 campaigns:

    	$ src campaigns list -first=5

  List campaigns and only print their IDs (default is to print ID and Name):

    	$ src campaigns list -first=5 -f '{{.ID}}'

`

	flagSet := flag.NewFlagSet("list", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src repos %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		firstFlag      = flagSet.Int("first", 1000, "Returns the first n campaigns.")
		changesetsFlag = flagSet.Int("changesets", 1000, "Returns the first n changesets per campaign.")
		formatFlag     = flagSet.String("f", "{{.ID}}: {{.Name}}", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.ID}}: {{.Name}}") or "{{.|json}}")`)
		apiFlags       = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		err := flagSet.Parse(args)
		if err != nil {
			return err
		}

		tmpl, err := parseTemplate(*formatFlag)
		if err != nil {
			return err
		}

		query := campaignFragment + `
query Campaigns($first: Int, $changesetsFirst: Int) {
  campaigns(first: $first) {
    nodes {
	  ... campaign
    }
  }
}
`

		client := cfg.apiClient(apiFlags, flagSet.Output())

		var result struct {
			Campaigns struct {
				Nodes []Campaign
			}
		}

		if ok, err := client.NewRequest(query, map[string]interface{}{
			"first":           api.NullInt(*firstFlag),
			"changesetsFirst": api.NullInt(*changesetsFlag),
		}).Do(context.Background(), &result); err != nil || !ok {
			return err
		}

		for _, c := range result.Campaigns.Nodes {
			if err := execTemplate(tmpl, c); err != nil {
				return err
			}
		}
		return nil
	}

	// Register the command.
	campaignsCommands = append(campaignsCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}

const campaignFragment = `
fragment campaign on Campaign {
  id
  name
  description
  url
  publishedAt
  createdAt
  updatedAt

  changesets(first: $changesetsFirst) {
    nodes {
      id
      state
      reviewState
      repository {
        id
        name
      }
      externalURL {
        url
        serviceType
      }
      createdAt
      updatedAt
    }

    totalCount
    pageInfo { hasNextPage }
  }
}
`

type Campaign struct {
	ID          string
	Name        string
	Description string
	URL         string
	PublishedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Changesets  struct {
		Nodes []struct {
			ID          string
			State       string
			ReviewState string
			Repository  struct {
				ID   string
				Name string
			}
			ExternalURL struct {
				URL         string
				ServiceType string
			}
			CreatedAt time.Time
			UpdatedAt time.Time
		}
		TotalCount int
		PageInfo   struct{ HasNextPage bool }
	}
}
