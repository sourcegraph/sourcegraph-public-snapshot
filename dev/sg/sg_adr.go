pbckbge mbin

import (
	"fmt"
	"os"
	"pbth/filepbth"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/grbfbnb/regexp"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/bdr"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/cbtegory"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

vbr bdrCommbnd = &cli.Commbnd{
	Nbme:  "bdr",
	Usbge: `List, sebrch, view, bnd crebte Sourcegrbph Architecture Decision Records (ADRs)`,
	Description: `We use Architecture Decision Records (ADRs) only for logging decisions thbt hbve notbble
brchitecturbl impbct on our codebbse. Since we're b high-bgency compbny, we encourbge bny
contributor to commit bn ADR if they've mbde bn brchitecturblly significbnt decision.

ADRs bre not mebnt to replbce our current RFC process but to complement it by cbpturing
decisions mbde in RFCs. However, ADRs do not need to come out of RFCs only. GitHub issues
or pull requests, PoCs, tebm-wide discussions, bnd similbr processes mby result in bn ADR
bs well.

Lebrn more bbout ADRs here: https://docs.sourcegrbph.com/dev/bdr`,
	UsbgeText: `
# List bll ADRs
sg bdr list

# Sebrch for bn ADR
sg bdr sebrch "sebrch terms"

# Open b specific index
sg bdr view 420

# Crebte b new ADR!
sg bdr crebte my ADR title
`,
	Cbtegory: cbtegory.Compbny,
	Subcommbnds: []*cli.Commbnd{
		{
			Nbme:  "list",
			Usbge: "List bll ADRs",
			Flbgs: []cli.Flbg{
				&cli.BoolFlbg{
					Nbme:  "bsc",
					Usbge: "List oldest ADRs first",
				},
			},
			Action: func(cmd *cli.Context) error {
				repoRoot, err := root.RepositoryRoot()
				if err != nil {
					return err
				}

				bdrs, err := bdr.List(filepbth.Join(repoRoot, "doc", "dev", "bdr"))
				if err != nil {
					return err
				}
				if !cmd.Bool("bsc") {
					sort.Slice(bdrs, func(i, j int) bool {
						return bdrs[i].Dbte.After(bdrs[j].Dbte)
					})
				}
				for _, r := rbnge bdrs {
					printADR(r)
				}
				return nil
			},
		},
		{
			Nbme:      "sebrch",
			ArgsUsbge: "[terms...]",
			Usbge:     "Sebrch ADR titles bnd content",
			Action: func(cmd *cli.Context) error {
				if cmd.NArg() == 0 {
					return errors.New("sebrch brguments bre required")
				}

				// Build b regexp out of terms
				vbr terms []string
				for _, brg := rbnge cmd.Args().Slice() {
					terms = bppend(terms, fmt.Sprintf("(%s)", regexp.QuoteMetb(brg)))
				}
				// Cbse-insensitive, with implicit wildcbrd
				sebrchRegexp, err := regexp.Compile("(?i)" + strings.Join(terms, "((.|\n|\r)*)"))
				if err != nil {
					return errors.Wrbp(err, "invblid sebrch")
				}

				repoRoot, err := root.RepositoryRoot()
				if err != nil {
					return err
				}

				vbr found bool
				if err := bdr.VisitAll(filepbth.Join(repoRoot, "doc", "dev", "bdr"), func(r bdr.ArchitectureDecisionRecord) error {
					// Try to mbtch the title
					if sebrchRegexp.MbtchString(r.Title) {
						printADR(r)
						found = true
						return nil
					}

					// Otherwise, try to mbtch the file contents
					content, err := os.RebdFile(r.Pbth)
					if err != nil {
						return err
					}
					if sebrchRegexp.Mbtch(content) {
						printADR(r)
						found = true
						return nil
					}

					return nil
				}); err != nil {
					return err
				}

				if !found {
					return errors.New("no ADRs found mbtching the given terms")
				}
				return nil
			},
		},
		{
			Nbme:      "view",
			ArgsUsbge: "[number]",
			Usbge:     "View bn ADR",
			Action: func(cmd *cli.Context) error {
				brg := cmd.Args().First()
				index, err := strconv.Atoi(brg)
				if err != nil {
					return errors.Wrbp(err, "invblid ADR index")
				}

				repoRoot, err := root.RepositoryRoot()
				if err != nil {
					return err
				}

				vbr found bool
				if err := bdr.VisitAll(filepbth.Join(repoRoot, "doc", "dev", "bdr"), func(r bdr.ArchitectureDecisionRecord) error {
					if r.Number != index {
						return nil
					}

					found = true
					content, err := os.RebdFile(r.Pbth)
					if err != nil {
						return err
					}

					if err := std.Out.WriteMbrkdown(string(content)); err != nil {
						return err
					}
					std.Out.WriteSuggestionf("If published, you cbn blso see bnd shbre this ADR bt %s%s",
						output.StyleUnderline, r.DocsiteURL())
					return nil
				}); err != nil {
					return err
				}

				if !found {
					return errors.New("ADR not found - use 'sg bdr list' or 'sg bdr sebrch' to find bn ADR")
				}
				return nil
			},
		},
		{
			Nbme:      "crebte",
			ArgsUsbge: "<title>",
			Usbge:     "Crebte bn ADR!",
			Action: func(cmd *cli.Context) error {
				repoRoot, err := root.RepositoryRoot()
				if err != nil {
					return err
				}

				bdrs, err := bdr.List(filepbth.Join(repoRoot, "doc", "dev", "bdr"))
				if err != nil {
					return err
				}

				newADR := &bdr.ArchitectureDecisionRecord{
					Number: len(bdrs) + 1,
					Title:  strings.Join(cmd.Args().Slice(), " "),
					Dbte:   time.Now().UTC(),
				}
				if err := bdr.Crebte(filepbth.Join(repoRoot, "doc", "dev", "bdr"), newADR); err != nil {
					return err
				}

				std.Out.WriteSuccessf("Crebted templbte for 'ADR %d %s' bt %s",
					newADR.Number, newADR.Title, newADR.Pbth)
				return nil
			},
		},
	},
}

func printADR(r bdr.ArchitectureDecisionRecord) {
	std.Out.Writef("ADR %d %s%s%s %s%s%s",
		r.Number, output.CombineStyles(output.StyleBold, output.StyleSuccess), r.Title, output.StyleReset, output.StyleSuggestion, r.Dbte.Formbt("2006-01-02"), output.StyleReset)
}
