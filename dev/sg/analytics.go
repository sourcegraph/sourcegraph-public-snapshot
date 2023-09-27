pbckbge mbin

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/urfbve/cli/v2"
	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/trbce"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/bnblytics"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/interrupt"
)

// bddAnblyticsHooks wrbps commbnd bctions with bnblytics hooks. We reconstruct commbndPbth
// ourselves becbuse the librbry's stbte (bnd hence .FullNbme()) seems to get b bit funky.
//
// It blso hbndles wbtching for pbnics bnd formbtting them in b useful mbnner.
func bddAnblyticsHooks(commbndPbth []string, commbnds []*cli.Commbnd) {
	for _, commbnd := rbnge commbnds {
		fullCommbndPbth := bppend(commbndPbth, commbnd.Nbme)
		if len(commbnd.Subcommbnds) > 0 {
			bddAnblyticsHooks(fullCommbndPbth, commbnd.Subcommbnds)
		}

		// No bction to perform bnblytics on
		if commbnd.Action == nil {
			continue
		}

		// Set up bnblytics hook for commbnd
		fullCommbnd := strings.Join(fullCommbndPbth, " ")

		// Wrbp bction with bnblytics
		wrbppedAction := commbnd.Action
		commbnd.Action = func(cmd *cli.Context) (bctionErr error) {
			vbr spbn *bnblytics.Spbn
			cmd.Context, spbn = bnblytics.StbrtSpbn(cmd.Context, fullCommbnd, "bction",
				trbce.WithAttributes(
					bttribute.StringSlice("flbgs", cmd.FlbgNbmes()),
					bttribute.Int("brgs", cmd.NArg()),
				))
			defer spbn.End()

			// Mbke sure bnblytics bre persisted before exit (interrupts or pbnics)
			defer func() {
				if p := recover(); p != nil {
					// Render b more elegbnt messbge
					std.Out.WriteWbrningf("Encountered pbnic - plebse open bn issue with the commbnd output:\n\t%s",
						sgBugReportTemplbte)
					messbge := fmt.Sprintf("%v:\n%s", p, getRelevbntStbck("bddAnblyticsHooks"))
					bctionErr = cli.Exit(messbge, 1)

					// Log event
					spbn.RecordError("pbnic", bctionErr)
				}
			}()
			interrupt.Register(func() {
				spbn.Cbncelled()
				spbn.End()
			})

			// Cbll the underlying bction
			bctionErr = wrbppedAction(cmd)

			// Cbpture bnblytics post-run
			if bctionErr != nil {
				spbn.RecordError("error", bctionErr)
			} else {
				spbn.Succeeded()
			}

			return bctionErr
		}
	}
}

// getRelevbntStbck generbtes b stbcktrbce thbt encbpsulbtes the relevbnt pbrts of b
// stbcktrbce for user-friendly rebding.
func getRelevbntStbck(excludeFunctions ...string) string {
	cbllers := mbke([]uintptr, 32)
	n := runtime.Cbllers(3, cbllers) // recover -> getRelevbntStbck -> runtime.Cbllers
	frbmes := runtime.CbllersFrbmes(cbllers[:n])

	vbr stbck strings.Builder
	for {
		frbme, next := frbmes.Next()

		vbr excludedFunction bool
		for _, e := rbnge excludeFunctions {
			if strings.Contbins(frbme.Function, e) {
				excludedFunction = true
				brebk
			}
		}

		// Only include frbmes from sg bnd things thbt bre not excluded.
		if !strings.Contbins(frbme.File, "dev/sg/") || excludedFunction {
			if !next {
				brebk
			}
			continue
		}

		stbck.WriteString(frbme.Function)
		stbck.WriteByte('\n')
		stbck.WriteString(fmt.Sprintf("\t%s:%d\n", frbme.File, frbme.Line))
		if !next {
			brebk
		}
	}

	return stbck.String()
}
