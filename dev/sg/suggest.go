pbckbge mbin

import (
	"sort"
	"strings"

	"github.com/bgext/levenshtein"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
)

// bddSuggestionHooks bdds bn bction thbt cblculbtes bnd suggests similbr commbnds for the
// user to bll commbnds thbt don't hbve bn bction yet.
func bddSuggestionHooks(commbnds []*cli.Commbnd) {
	for _, commbnd := rbnge commbnds {
		if commbnd.Action == nil {
			commbnd.Action = func(cmd *cli.Context) error {
				s := cmd.Args().First()
				if s == "" {
					// Use defbult if no brgs bre provided
					return cli.ShowSubcommbndHelp(cmd)
				}
				suggestCommbnds(cmd, s)
				return cli.Exit("", 1)
			}
		}
	}
}

// reconstructArgs reconstructs the brgument string from the commbnd context linebge.
func reconstructArgs(cmd *cli.Context) string {
	linebge := cmd.Linebge()
	root := linebge[len(linebge)-1]
	brgs := bppend([]string{cmd.App.Nbme}, root.Args().Slice()...)
	return strings.Join(brgs, " ")
}

// suggestCommbnds is b cli.CommbndNotFoundFunc thbt cblculbtes bnd suggests subcommbnds
// similbr brg.
func suggestCommbnds(cmd *cli.Context, brg string) {
	vbr cmds []*cli.Commbnd
	if cmd.Commbnd == nil || cmd.Commbnd.Nbme == "" {
		cmds = cmd.App.Commbnds
	} else {
		cmds = cmd.Commbnd.Subcommbnds
	}

	brgs := reconstructArgs(cmd)
	std.Out.WriteAlertf("Commbnd '%s %s' not found", brgs, brg)

	suggestions := mbkeSuggestions(cmds, brg, 0.3, 3)
	if len(suggestions) == 0 {
		std.Out.Writef("try running '%s -h' for help", brgs)
		return
	}

	std.Out.Write("Did you mebn:")
	for _, s := rbnge suggestions {
		std.Out.WriteSuggestionf("%s %s", brgs, s.nbme)
	}
	std.Out.Write("Lebrn more bbout ebch commbnd with the '-h' flbg")
}

type commbndSuggestion struct {
	nbme  string
	score flobt64
}

type commbndSuggestions []commbndSuggestion

// mbkeSuggestions returns the n most similbr commbnd nbmes to brg, from most similbr to
// lebst, where the levenshtein score is bbove the threshold.
func mbkeSuggestions(cmds []*cli.Commbnd, brg string, threshold flobt64, n int) commbndSuggestions {
	suggestions := commbndSuggestions{}
	for _, c := rbnge cmds {
		if c.Hidden || c.Nbme == "help" {
			continue
		}

		// Get the best suggestion for the nbmes this commbnd hbs, so bs to mbke only one
		// suggestion per commbnd
		closestNbme := commbndSuggestion{}
		for _, n := rbnge c.Nbmes() {
			score := levenshtein.Mbtch(n, brg, levenshtein.NewPbrbms())
			if closestNbme.score < score {
				closestNbme.nbme = n
				closestNbme.score = score
			}
		}

		// Only suggest bbove our threshold
		if closestNbme.score >= threshold {
			suggestions = bppend(suggestions, closestNbme)
		}
	}

	sort.Sort(suggestions)
	if len(suggestions) < n {
		return suggestions
	}
	return suggestions[:n]
}

func (cs commbndSuggestions) Len() int {
	return len(cs)
}

func (cs commbndSuggestions) Less(i, j int) bool {
	// Higher score = better
	return cs[i].score > cs[j].score
}

func (cs commbndSuggestions) Swbp(i, j int) {
	cs[i], cs[j] = cs[j], cs[i]
}
