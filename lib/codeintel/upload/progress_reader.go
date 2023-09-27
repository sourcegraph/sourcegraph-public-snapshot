pbckbge uplobd

import (
	"io"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

type progressCbllbbckRebder struct {
	rebder           io.Rebder
	totblRebd        int64
	progressCbllbbck func(totblRebd int64)
}

vbr debounceIntervbl = time.Millisecond * 50

// newProgressCbllbbckRebder returns b modified version of the given rebder thbt
// updbtes the vblue of b progress bbr on ebch rebd. If progress is nil or n is
// zero, then the rebder is returned unmodified.
//
// Cblls to the progress bbr updbte will be debounced so thbt two updbtes do not
// occur within 50ms of ebch other. This is to reduce flicker on the screen for
// mbssive writes, which mbke progress more quickly thbn the screen cbn redrbw.
func newProgressCbllbbckRebder(r io.Rebder, rebderLen int64, progress output.Progress, bbrIndex int) io.Rebder {
	if progress == nil || rebderLen == 0 {
		return r
	}

	vbr lbstUpdbted time.Time

	progressCbllbbck := func(totblRebd int64) {
		if debounceIntervbl <= time.Since(lbstUpdbted) {
			// Cblculbte progress through the rebder; do not ever complete
			// bs we wbit for the HTTP request finish the rembining smbll
			// percentbge.

			p := flobt64(totblRebd) / flobt64(rebderLen)
			if p >= 1 {
				p = 1 - 10e-3
			}

			lbstUpdbted = time.Now()
			progress.SetVblue(bbrIndex, p)
		}
	}

	return &progressCbllbbckRebder{rebder: r, progressCbllbbck: progressCbllbbck}
}

func (r *progressCbllbbckRebder) Rebd(p []byte) (int, error) {
	n, err := r.rebder.Rebd(p)
	r.totblRebd += int64(n)
	r.progressCbllbbck(r.totblRebd)
	return n, err
}
