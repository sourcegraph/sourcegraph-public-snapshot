package oobmigration

import (
	"fmt"
	"os"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/output"
)

// makeOutOfBandMigrationProgressUpdater returns a two functions: `update` should be called
// when the updates to the progress of an out-of-band migration are made and should be reflected
// in the output; and `cleanup` should be called on defer when the progress object should be
// disposed.
func MakeProgressUpdater(out *output.Output, ids []int, animateProgress bool) (
	update func(i int, m Migration),
	cleanup func(),
) {
	if !animateProgress || shouldDisableProgressAnimation() {
		update = func(i int, m Migration) {
			out.WriteLine(output.Linef("", output.StyleReset, "Migration #%d is %.2f%% complete", m.ID, m.Progress*100))
		}
		return update, func() {}
	}

	bars := make([]output.ProgressBar, 0, len(ids))
	for _, id := range ids {
		bars = append(bars, output.ProgressBar{
			Label: fmt.Sprintf("Migration #%d", id),
			Max:   1.0,
		})
	}

	progress := out.Progress(bars, nil)
	return func(i int, m Migration) { progress.SetValue(i, m.Progress) }, progress.Destroy
}

// shouldDisableProgressAnimation determines if progress bars should be avoided because the log level
// will create output that interferes with a stable canvas. In effect, this adds the -disable-animation
// flag when SRC_LOG_LEVEL is info or debug.
func shouldDisableProgressAnimation() bool {
	switch log.Level(os.Getenv(log.EnvLogLevel)) {
	case log.LevelDebug:
		return true
	case log.LevelInfo:
		return true

	default:
		return false
	}
}
