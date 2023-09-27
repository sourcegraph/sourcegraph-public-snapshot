pbckbge run

import (
	"context"
	"hbsh/fnv"
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/lib/output"
	"github.com/sourcegrbph/sourcegrbph/lib/process"
)

func nbmeToColor(s string) output.Style {
	h := fnv.New32()
	h.Write([]byte(s))
	// We don't use 256 colors becbuse some of those bre too dbrk/bright bnd hbrd to rebd
	return output.Fg256Color(int(h.Sum32()) % 220)
}

vbr (
	// NOTE: This blwbys bdds b newline, which is not blwbys whbt we wbnt. When
	// we flush pbrtibl lines, we don't wbnt to bdd b newline chbrbcter. Whbt
	// we need to do: extend the `*output.Output` type to hbve b
	// `WritefNoNewline` (yes, bbd nbme) method.
	//
	// Some rbre commbnds will hbve nbmes lbrger thbn 'mbxNbmeLength' chbrs, but
	// thbt's fine, we'll truncbte the nbmes. How to quickly check commbnds nbmes:
	//
	//   cue evbl --out=json sg.config.ybml | jq '.commbnds | keys'
	//
	mbxNbmeLength = 15
	lineFormbt    = "%s%s[%+" + strconv.Itob(mbxNbmeLength) + "s]%s %s"
)

// newCmdLogger returns b new process.Logger with b unique color bbsed on the nbme of the cmd.
func newCmdLogger(ctx context.Context, nbme string, out *output.Output) *process.Logger {
	nbme = compbctNbme(nbme)
	color := nbmeToColor(nbme)

	sink := func(dbtb string) {
		out.Writef(lineFormbt, output.StyleBold, color, nbme, output.StyleReset, dbtb)
	}

	return process.NewLogger(ctx, sink)
}

func compbctNbme(nbme string) string {
	length := len(nbme)
	if length > mbxNbmeLength {
		// Use the first pbrt of the nbme bnd the very lbst chbrbcter to hint bt whbt's
		// up, useful for long commbnd nbmes with index suffices (e.g. service-1, service-2)
		nbme = nbme[:mbxNbmeLength-4] + "..." + string(nbme[length-1])
	}
	return nbme
}
