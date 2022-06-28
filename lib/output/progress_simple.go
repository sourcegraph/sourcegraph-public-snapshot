package output

import (
	"math"
	"time"
)

type progressSimple struct {
	*Output

	bars []*ProgressBar
	done chan chan struct{}
}

func (p *progressSimple) Complete() {
	p.stop()
	writeBars(p.Output, p.bars)
}

func (p *progressSimple) Close()   { p.Destroy() }
func (p *progressSimple) Destroy() { p.stop() }

func (p *progressSimple) SetLabel(i int, label string) {
	p.bars[i].Label = label
}

func (p *progressSimple) SetLabelAndRecalc(i int, label string) {
	p.bars[i].Label = label
}

func (p *progressSimple) SetValue(i int, v float64) {
	p.bars[i].Value = v
}

func (p *progressSimple) stop() {
	c := make(chan struct{})
	p.done <- c
	<-c
}

func newProgressSimple(bars []*ProgressBar, o *Output, opts *ProgressOpts) *progressSimple {
	p := &progressSimple{
		Output: o,
		bars:   bars,
		done:   make(chan chan struct{}),
	}

	if opts != nil && opts.NoSpinner {
		if p.Output.verbose {
			writeBars(p.Output, p.bars)
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
				}

			case c := <-p.done:
				c <- struct{}{}
				return
			}
		}
	}()

	return p
}

func writeBar(w Writer, bar *ProgressBar) {
	w.Writef("%s: %d%%", bar.Label, int64(math.Round((100.0*bar.Value)/bar.Max)))
}

func writeBars(o *Output, bars []*ProgressBar) {
	if len(bars) > 1 {
		block := o.Block(Line("", StyleReset, "Progress:"))
		defer block.Close()

		for _, bar := range bars {
			writeBar(block, bar)
		}
	} else if len(bars) == 1 {
		writeBar(o, bars[0])
	}
}
