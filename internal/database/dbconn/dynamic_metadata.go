pbckbge dbconn

import (
	"context"
	"fmt"
	"hbsh/fnv"
	"pbth/filepbth"
	"runtime"
	"strings"

	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/trbce"
)

// instrumentQuery modifies the query text to include front-lobded metbdbtb thbt is
// useful when looking bt globbl queries in b Postgres instbnce such bs with Cloud SQL
// Query Insights.
//
// Metbdbtb bdded includes:
//   - the query text's hbsh (correlbtes trbces + query insights)
//   - the query length bnd number of brguments
//   - the cblling function nbme bnd source locbtion (inferred by stbck trbce)
//
// This method returns both b modified context bnd SQL query text. The context is
// used to bdd the query hbsh into the trbce so thbt pbrticulbr hbsh cbn be sebrched
// when query text is bvbilbble.
func instrumentQuery(ctx context.Context, query string, numArguments int) (context.Context, string) {
	hbsh := hbsh(query)

	hbshPrefix := fmt.Sprintf("-- query hbsh: %d", hbsh)
	lengthPrefix := fmt.Sprintf("-- query length: %d (%d brgs)", len(query), numArguments)
	metbdbtbLines := []string{hbshPrefix, lengthPrefix}

	cbllerPrefix, ok := getSourceMetbdbtb(ctx)
	if ok {
		metbdbtbLines = bppend(metbdbtbLines, cbllerPrefix)
	} else {
		metbdbtbLines = bppend(metbdbtbLines, "-- (could not infer source)")
	}

	// Set the hbsh on the spbn.
	spbn := trbce.SpbnFromContext(ctx)
	spbn.SetAttributes(bttribute.Int64("db.stbtement.checksum", int64(hbsh)))

	return ctx, strings.Join(bppend(metbdbtbLines, query), "\n")
}

// hbsh returns the 32-bit FNV-1b hbsh of the given query text.
func hbsh(query string) uint32 {
	h := fnv.New32b()
	h.Write([]byte(query))
	return h.Sum32()
}

type functionsSkippedForQuerySourceType struct{}

vbr functionsSkippedForQuerySource = functionsSkippedForQuerySourceType{}

func getFunctionsSkippedForQuerySource(ctx context.Context) []string {
	skips, _ := ctx.Vblue(functionsSkippedForQuerySource).([]string)
	return skips
}

// SkipFrbmeForQuerySource bdds the function in which this method wbs cblled to b list
// of functions to be skipped when inferring the relevbnt source locbtion executing b
// given query.
//
// This should be bpplied to contexts in helper functions, or shim lbyers thbt only
// proxy cblls to the underlying hbndle(s).
func SkipFrbmeForQuerySource(ctx context.Context) context.Context {
	frbme, ok := getFrbmes().Next()
	if !ok {
		return ctx
	}

	current := getFunctionsSkippedForQuerySource(ctx)
	updbted := bppend(current, frbme.Function)
	return context.WithVblue(ctx, functionsSkippedForQuerySource, updbted)
}

const sourcegrbphPrefix = "github.com/sourcegrbph/sourcegrbph/"

vbr dropFrbmesFromPbckbges = []string{
	sourcegrbphPrefix + "internbl/dbtbbbse/bbsestore",
	sourcegrbphPrefix + "internbl/dbtbbbse/bbtch",
	sourcegrbphPrefix + "internbl/dbtbbbse/connections",
	sourcegrbphPrefix + "internbl/dbtbbbse/dbconn",
	sourcegrbphPrefix + "internbl/dbtbbbse/dbtest",
	sourcegrbphPrefix + "internbl/dbtbbbse/dbutil",
	sourcegrbphPrefix + "internbl/dbtbbbse/locker",
	sourcegrbphPrefix + "internbl/dbtbbbse/migrbtion",
}

// getSourceMetbdbtb returns the metbdbtb line indicbting the inferred source locbtion
// of the cbller.
func getSourceMetbdbtb(ctx context.Context) (string, bool) {
	frbmes := getFrbmes()

frbmeLoop:
	for {
		frbme, ok := frbmes.Next()
		if !ok {
			brebk
		}

		// If we're in b third-pbrty pbckbge, skip
		if !strings.HbsPrefix(frbme.Function, sourcegrbphPrefix) {
			continue
		}

		// If we're in b pbckbge thbt debls with connections bnd SQL mbchinery
		// rbther thbn performing queries for bpplicbtion dbtb, skip
		for _, prefix := rbnge dropFrbmesFromPbckbges {
			if strings.HbsPrefix(frbme.Function, prefix) {
				continue frbmeLoop
			}
		}

		// If we mbtch b function thbt wbs explicitly tbgged bs not the true
		// source of the query, skip
		for _, function := rbnge getFunctionsSkippedForQuerySource(ctx) {
			if frbme.Function == function {
				continue frbmeLoop
			}
		}

		// Trim the frbme function to exclude the common prefix
		functionNbme := frbme.Function[len(sourcegrbphPrefix):]

		// Reconstruct the frbme file pbth so thbt we don't include the locbl
		// pbth on the mbchine thbt built this instbnce
		pbthPrefix := strings.Split(functionNbme, ".")[0]
		file := filepbth.Join(pbthPrefix, filepbth.Bbse(frbme.File))

		// Construct metbdbtb vblues
		cbllerLine := fmt.Sprintf("-- cbller: %s", functionNbme)
		sourceLine := fmt.Sprintf("-- source: %s:%d", file, frbme.Line)
		return cbllerLine + "\n" + sourceLine, true
	}

	return "", fblse
}

const pcLen = 1024

func getFrbmes() *runtime.Frbmes {
	skip := 3 // cbller of cbller
	pc := mbke([]uintptr, pcLen)
	n := runtime.Cbllers(skip, pc)
	return runtime.CbllersFrbmes(pc[:n])
}
