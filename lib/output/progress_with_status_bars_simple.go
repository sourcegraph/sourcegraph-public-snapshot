package output

import (
	"time"
)

type progressWithStatusBarsSimple struct {
	*progressSimple

	statusBars []*StatusBar
}

func (p *progressWithStatusBarsSimple) Complete() {
	p.stop()
	writeBars(p.Output, p.bars)
	writeStatusBars(p.Output, p.statusBars)
}

func (p *progressWithStatusBarsSimple) StatusBarUpdatef(i int, format string, args ...any) {
	if p.statusBars[i] != nil {
		p.statusBars[i].Updatef(format, args...)
	}
}

func (p *progressWithStatusBarsSimple) StatusBarCompletef(i int, format string, args ...any) {
	if p.statusBars[i] != nil {
		wasComplete := p.statusBars[i].completed
		p.statusBars[i].Completef(format, args...)
		if !wasComplete {
			writeStatusBar(p.Output, p.statusBars[i])
		}
	}
}

func (p *progressWithStatusBarsSimple) StatusBarFailf(i int, format string, args ...any) {
	if p.statusBars[i] != nil {
		wasCompleted := p.statusBars[i].completed
		p.statusBars[i].Failf(format, args...)
		if !wasCompleted {
			writeStatusBar(p.Output, p.statusBars[i])
		}
	}
}

func (p *progressWithStatusBarsSimple) StatusBarResetf(i int, label, format string, args ...any) {
	if p.statusBars[i] != nil {
		p.statusBars[i].Resetf(label, format, args...)
	}
}

func newProgressWithStatusBarsSimple(bars []*ProgressBar, statusBars []*StatusBar, o *Output, opts *ProgressOpts) *progressWithStatusBarsSimple {
	p := &progressWithStatusBarsSimple{
		progressSimple: &progressSimple{
			Output: o,
			bars:   bars,
			done:   make(chan chan struct{}),
		},
		statusBars: statusBars,
	}

	if opts != nil && opts.NoSpinner {
		if p.Output.verbose {
			writeBars(p.Output, p.bars)
			writeStatusBars(p.Output, p.statusBars)
		}
		return p
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if p.Output.verbose {
					writeBars(p.Output, p.bars)
					writeStatusBars(p.Output, p.statusBars)
				}

			case c := <-p.done:
				c <- struct{}{}
				return
			}
		}
	}()

	return p
}

func writeStatusBar(w Writer, bar *StatusBar) {
	w.Writef("%s: "+bar.format, append([]any{bar.label}, bar.args...)...)
}

func writeStatusBars(o *Output, bars []*StatusBar) {
	if len(bars) > 1 {
		block := o.Block(Line("", StyleReset, "Status:"))
		defer block.Close()

		for _, bar := range bars {
			writeStatusBar(block, bar)
		}
	} else if len(bars) == 1 {
		writeStatusBar(o, bars[0])
	}
}
