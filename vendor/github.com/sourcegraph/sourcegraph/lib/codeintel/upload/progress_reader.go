package upload

import (
	"io"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/output"
)

type progressCallbackReader struct {
	reader           io.Reader
	totalRead        int64
	progressCallback func(totalRead int64)
}

var debounceInterval = time.Millisecond * 50

// newProgressCallbackReader returns a modified version of the given reader that
// updates the value of a progress bar on each read. If progress is nil or n is
// zero, then the reader is returned unmodified.
//
// Calls to the progress bar update will be debounced so that two updates do not
// occur within 50ms of each other. This is to reduce flicker on the screen for
// massive writes, which make progress more quickly than the screen can redraw.
func newProgressCallbackReader(r io.Reader, readerLen int64, progress output.Progress, barIndex int) io.Reader {
	if progress == nil || readerLen == 0 {
		return r
	}

	var lastUpdated time.Time

	progressCallback := func(totalRead int64) {
		if debounceInterval <= time.Since(lastUpdated) {
			// Calculate progress through the reader; do not ever complete
			// as we wait for the HTTP request finish the remaining small
			// percentage.

			p := float64(totalRead) / float64(readerLen)
			if p >= 1 {
				p = 1 - 10e-3
			}

			lastUpdated = time.Now()
			progress.SetValue(barIndex, p)
		}
	}

	return &progressCallbackReader{reader: r, progressCallback: progressCallback}
}

func (r *progressCallbackReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.totalRead += int64(n)
	r.progressCallback(r.totalRead)
	return n, err
}
