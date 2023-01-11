package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/regexp"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/adr"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var adrCommand = &cli.Command{
	Name:  "adr",
	Usage: `List, search, view, and create Sourcegraph Architecture Decision Records (ADRs)`,
	Description: `We use Architecture Decision Records (ADRs) only for logging decisions that have notable
architectural impact on our codebase. Since we're a high-agency company, we encourage any
contributor to commit an ADR if they've made an architecturally significant decision.

ADRs are not meant to replace our current RFC process but to complement it by capturing
decisions made in RFCs. However, ADRs do not need to come out of RFCs only. GitHub issues
or pull requests, PoCs, team-wide discussions, and similar processes may result in an ADR
as well.

Learn more about ADRs here: https://docs.sourcegraph.com/dev/adr`,
	UsageText: `
# List all ADRs
sg adr list

# Search for an ADR
sg adr search "search terms"

# Open a specific index
sg adr view 420

# Create a new ADR!
sg adr create my ADR title
`,
	Category: CategoryCompany,
	Subcommands: []*cli.Command{
		{
			Name:  "list",
			Usage: "List all ADRs",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "asc",
					Usage: "List oldest ADRs first",
				},
			},
			Action: func(cmd *cli.Context) error {
				repoRoot, err := root.RepositoryRoot()
				if err != nil {
					return err
				}

				adrs, err := adr.List(filepath.Join(repoRoot, "doc", "dev", "adr"))
				if err != nil {
					return err
				}
				if !cmd.Bool("asc") {
					sort.Slice(adrs, func(i, j int) bool {
						return adrs[i].Date.After(adrs[j].Date)
					})
				}
				for _, r := range adrs {
					printADR(r)
				}
				return nil
			},
		},
		{
			Name:      "search",
			ArgsUsage: "[terms...]",
			Usage:     "Search ADR titles and content",
			Action: func(cmd *cli.Context) error {
				if cmd.NArg() == 0 {
					return errors.New("search arguments are required")
				}

				// Build a regexp out of terms
				var terms []string
				for _, arg := range cmd.Args().Slice() {
					terms = append(terms, fmt.Sprintf("(%s)", regexp.QuoteMeta(arg)))
				}
				// Case-insensitive, with implicit wildcard
				searchRegexp, err := regexp.Compile("(?i)" + strings.Join(terms, "((.|\n|\r)*)"))
				if err != nil {
					return errors.Wrap(err, "invalid search")
				}

				repoRoot, err := root.RepositoryRoot()
				if err != nil {
					return err
				}

				var found bool
				if err := adr.VisitAll(filepath.Join(repoRoot, "doc", "dev", "adr"), func(r adr.ArchitectureDecisionRecord) error {
					// Try to match the title
					if searchRegexp.MatchString(r.Title) {
						printADR(r)
						found = true
						return nil
					}

					// Otherwise, try to match the file contents
					content, err := os.ReadFile(r.Path)
					if err != nil {
						return err
					}
					if searchRegexp.Match(content) {
						printADR(r)
						found = true
						return nil
					}

					return nil
				}); err != nil {
					return err
				}

				if !found {
					return errors.New("no ADRs found matching the given terms")
				}
				return nil
			},
		},
		{
			Name:      "view",
			ArgsUsage: "[number]",
			Usage:     "View an ADR",
			Action: func(cmd *cli.Context) error {
				arg := cmd.Args().First()
				index, err := strconv.Atoi(arg)
				if err != nil {
					return errors.Wrap(err, "invalid ADR index")
				}

				repoRoot, err := root.RepositoryRoot()
				if err != nil {
					return err
				}

				var found bool
				if err := adr.VisitAll(filepath.Join(repoRoot, "doc", "dev", "adr"), func(r adr.ArchitectureDecisionRecord) error {
					if r.Number != index {
						return nil
					}

					found = true
					content, err := os.ReadFile(r.Path)
					if err != nil {
						return err
					}

					if err := std.Out.WriteMarkdown(string(content)); err != nil {
						return err
					}
					std.Out.WriteSuggestionf("If published, you can also see and share this ADR at %s%s",
						output.StyleUnderline, r.DocsiteURL())
					return nil
				}); err != nil {
					return err
				}

				if !found {
					return errors.New("ADR not found - use 'sg adr list' or 'sg adr search' to find an ADR")
				}
				return nil
			},
		},
		{
			Name:      "create",
			ArgsUsage: "<title>",
			Usage:     "Create an ADR!",
			Action: func(cmd *cli.Context) error {
				repoRoot, err := root.RepositoryRoot()
				if err != nil {
					return err
				}

				adrs, err := adr.List(filepath.Join(repoRoot, "doc", "dev", "adr"))
				if err != nil {
					return err
				}

				newADR := &adr.ArchitectureDecisionRecord{
					Number: len(adrs) + 1,
					Title:  strings.Join(cmd.Args().Slice(), " "),
					Date:   time.Now().UTC(),
				}
				if err := adr.Create(filepath.Join(repoRoot, "doc", "dev", "adr"), newADR); err != nil {
					return err
				}

				std.Out.WriteSuccessf("Created template for 'ADR %d %s' at %s",
					newADR.Number, newADR.Title, newADR.Path)
				return nil
			},
		},
	},
}

func printADR(r adr.ArchitectureDecisionRecord) {
	std.Out.Writef("ADR %d %s%s%s %s%s%s",
		r.Number, output.CombineStyles(output.StyleBold, output.StyleSuccess), r.Title, output.StyleReset, output.StyleSuggestion, r.Date.Format("2006-01-02"), output.StyleReset)
}
