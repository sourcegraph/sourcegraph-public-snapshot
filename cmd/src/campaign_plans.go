package main

import (
	"flag"
	"fmt"
)

var campaignPlansCommands commander

func init() {
	usage := `'src campaigns plans' creates and previews campaign plans (which can be used to create campaigns and changesets).

EXPERIMENTAL: Campaigns are experimental functionality on Sourcegraph and in the 'src' tool.

Usage:

	src campaigns plans command [command options]

The commands are:

	create-from-patches  creates plan from patches to repository branches

Use "src campaigns plans [command] -h" for more information about a command.
`

	flagSet := flag.NewFlagSet("plans", flag.ExitOnError)
	handler := func(args []string) error {
		campaignPlansCommands.run(flagSet, "src campaigns plans", usage, args)
		return nil
	}

	// Register the command.
	campaignsCommands = append(campaignsCommands, &command{
		flagSet: flagSet,
		aliases: []string{"plan"},
		handler: handler,
		usageFunc: func() {
			fmt.Println(usage)
		},
	})
}

func campaignPlanFragment(first int) string {
	return fmt.Sprintf(`
fragment CampaignPlanFields on CampaignPlan {
    id
    status {
        completedCount
        pendingCount
        state
        errors
    }
    changesets(first: %d) {
        nodes {
            repository {
                id
                name
                url
            }
            diff {
                fileDiffs {
                    rawDiff
                    diffStat {
                        added
                        deleted
                        changed
                    }
                    nodes {
                        oldPath
                        newPath
                        hunks {
                            body
                            section
                            newRange { startLine, lines }
                            oldRange { startLine, lines }
                            oldNoNewlineAt
                        }
                        stat {
                            added
                            deleted
                            changed
                        }
                        oldFile {
                            name
                            externalURLs {
                                serviceType
                                url
                            }
                        }
                    }
                }
            }
        }
    }
    previewURL
}
`, first)
}

type DiffRange struct {
	StartLine int `json:"startLine"`
	Lines     int `json:"lines"`
}

type FileDiffHunk struct {
	Body           string    `json:"body"`
	Section        string    `json:"section"`
	OldNoNewlineAt bool      `json:"oldNoNewlineAt"`
	OldRange       DiffRange `json:"oldRange"`
	NewRange       DiffRange `json:"newRange"`
}

type DiffStat struct {
	Added   int `json:"added"`
	Deleted int `json:"deleted"`
	Changed int `json:"changed"`
}

type File struct {
	Name string `json:"name"`
}

type FileDiff struct {
	OldPath string         `json:"oldPath"`
	NewPath string         `json:"newPath"`
	Hunks   []FileDiffHunk `json:"hunks"`
	Stat    DiffStat       `json:"stat"`
	OldFile File           `json:"oldFile"`
}

type FileDiffs struct {
	RawDiff  string     `json:"rawDiff"`
	DiffStat DiffStat   `json:"diffStat"`
	Nodes    []FileDiff `json:"nodes"`
}

type ChangesetPlan struct {
	Repository struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url,omitempty"`
	} `json:"repository"`
	Diff struct {
		FileDiffs FileDiffs `json:"fileDiffs"`
	} `json:"diff"`
}

type Status struct {
	CompletedCount int      `json:"completedCount"`
	PendingCount   int      `json:"pendingCount"`
	State          string   `json:"state"`
	Errors         []string `json:"errors"`
}

type CampaignPlan struct {
	ID         string `json:"id"`
	Status     Status `json:"status"`
	Changesets struct {
		Nodes      []ChangesetPlan `json:"nodes"`
		TotalCount int             `json:"totalCount"`
	} `json:"changesets"`
	PreviewURL string `json:"previewURL"`
}
