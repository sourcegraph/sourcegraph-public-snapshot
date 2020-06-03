package main

import (
	"flag"
	"fmt"
)

var campaignPatchSetsCommands commander

func init() {
	usage := `'src campaigns patchsets' creates and previews patch sets (which can be used to create campaigns and changesets).

EXPERIMENTAL: Campaigns are experimental functionality on Sourcegraph and in the 'src' tool.

Usage:

	src campaigns patchsets command [command options]

The commands are:

	create-from-patches  creates patch set from patches to repository branches

Use "src campaigns patchsets [command] -h" for more information about a command.
`

	flagSet := flag.NewFlagSet("patchsets", flag.ExitOnError)
	handler := func(args []string) error {
		campaignPatchSetsCommands.run(flagSet, "src campaigns patchsets", usage, args)
		return nil
	}

	// Register the command.
	campaignsCommands = append(campaignsCommands, &command{
		flagSet: flagSet,
		aliases: []string{"patchset"},
		handler: handler,
		usageFunc: func() {
			fmt.Println(usage)
		},
	})
}

func patchSetFragment(first int) string {
	return fmt.Sprintf(`
fragment PatchSetFields on PatchSet {
    id
    patches(first: %d) {
        nodes {
            __typename
            ... on HiddenPatch {
                id
            }
            ... on Patch {
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

type Patch struct {
	Repository struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url,omitempty"`
	} `json:"repository"`
	Diff struct {
		FileDiffs FileDiffs `json:"fileDiffs"`
	} `json:"diff"`
}

type PatchSet struct {
	ID      string `json:"id"`
	Patches struct {
		Nodes      []Patch `json:"nodes"`
		TotalCount int     `json:"totalCount"`
	} `json:"patches"`
	PreviewURL string `json:"previewURL"`
}
