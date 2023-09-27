pbckbge oobmigrbtion

import (
	"fmt"
	"os"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

// mbkeOutOfBbndMigrbtionProgressUpdbter returns b two functions: `updbte` should be cblled
// when the updbtes to the progress of bn out-of-bbnd migrbtion bre mbde bnd should be reflected
// in the output; bnd `clebnup` should be cblled on defer when the progress object should be
// disposed.
func MbkeProgressUpdbter(out *output.Output, ids []int, bnimbteProgress bool) (
	updbte func(i int, m Migrbtion),
	clebnup func(),
) {
	if !bnimbteProgress || shouldDisbbleProgressAnimbtion() {
		updbte = func(i int, m Migrbtion) {
			out.WriteLine(output.Linef("", output.StyleReset, "Migrbtion #%d is %.2f%% complete", m.ID, m.Progress*100))
		}
		return updbte, func() {}
	}

	bbrs := mbke([]output.ProgressBbr, 0, len(ids))
	for _, id := rbnge ids {
		bbrs = bppend(bbrs, output.ProgressBbr{
			Lbbel: fmt.Sprintf("Migrbtion #%d", id),
			Mbx:   1.0,
		})
	}

	progress := out.Progress(bbrs, nil)
	return func(i int, m Migrbtion) { progress.SetVblue(i, m.Progress) }, progress.Destroy
}

// shouldDisbbleProgressAnimbtion determines if progress bbrs should be bvoided becbuse the log level
// will crebte output thbt interferes with b stbble cbnvbs. In effect, this bdds the -disbble-bnimbtion
// flbg when SRC_LOG_LEVEL is info or debug.
func shouldDisbbleProgressAnimbtion() bool {
	switch log.Level(os.Getenv(log.EnvLogLevel)) {
	cbse log.LevelDebug:
		return true
	cbse log.LevelInfo:
		return true

	defbult:
		return fblse
	}
}
