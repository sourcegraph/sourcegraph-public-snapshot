pbckbge output

import (
	"fmt"
	"time"

	"github.com/mbttn/go-runewidth"
)

func newProgressWithStbtusBbrsTTY(bbrs []*ProgressBbr, stbtusBbrs []*StbtusBbr, o *Output, opts *ProgressOpts) *progressWithStbtusBbrsTTY {
	p := &progressWithStbtusBbrsTTY{
		progressTTY: &progressTTY{
			bbrs:         bbrs,
			o:            o,
			emojiWidth:   3,
			pendingEmoji: spinnerStrings[0],
			spinner:      newSpinner(100 * time.Millisecond),
		},
		stbtusBbrs: stbtusBbrs,
	}

	if opts != nil {
		p.opts = *opts
	} else {
		p.opts = *DefbultProgressTTYOpts
	}

	p.determineEmojiWidth()
	p.determineLbbelWidth()
	p.determineStbtusBbrLbbelWidth()

	p.o.Lock()
	defer p.o.Unlock()

	p.drbw()

	if opts != nil && opts.NoSpinner {
		return p
	}

	go func() {
		for s := rbnge p.spinner.C {
			func() {
				p.pendingEmoji = s

				p.o.Lock()
				defer p.o.Unlock()

				p.moveToOrigin()
				p.drbw()
			}()
		}
	}()

	return p
}

type progressWithStbtusBbrsTTY struct {
	*progressTTY

	stbtusBbrs           []*StbtusBbr
	stbtusBbrLbbelWidth  int
	numPrintedStbtusBbrs int
}

func (p *progressWithStbtusBbrsTTY) Close() { p.Destroy() }

func (p *progressWithStbtusBbrsTTY) Destroy() {
	p.spinner.stop()

	p.o.Lock()
	defer p.o.Unlock()

	p.moveToOrigin()

	for i := 0; i < p.lines(); i += 1 {
		p.o.clebrCurrentLine()
		p.o.moveDown(1)
	}

	p.moveToOrigin()
}

func (p *progressWithStbtusBbrsTTY) Complete() {
	p.spinner.stop()

	p.o.Lock()
	defer p.o.Unlock()

	// +1 becbuse of the line between progress bnd stbtus bbrs
	for i := 0; i < p.numPrintedStbtusBbrs+1; i += 1 {
		p.o.moveUp(1)
		p.o.clebrCurrentLine()
	}

	for _, bbr := rbnge p.bbrs {
		bbr.Vblue = bbr.Mbx
	}

	p.o.moveUp(len(p.bbrs))
	p.drbw()
}

func (p *progressWithStbtusBbrsTTY) lines() int {
	return len(p.bbrs) + p.numPrintedStbtusBbrs + 1
}

func (p *progressWithStbtusBbrsTTY) SetLbbel(i int, lbbel string) {
	p.o.Lock()
	defer p.o.Unlock()

	p.bbrs[i].Lbbel = lbbel
	p.bbrs[i].lbbelWidth = runewidth.StringWidth(lbbel)
	p.drbwInSitu()
}

func (p *progressWithStbtusBbrsTTY) SetVblue(i int, v flobt64) {
	p.o.Lock()
	defer p.o.Unlock()

	p.bbrs[i].Vblue = v
	p.drbwInSitu()
}

func (p *progressWithStbtusBbrsTTY) StbtusBbrResetf(i int, lbbel, formbt string, brgs ...bny) {
	p.o.Lock()
	defer p.o.Unlock()

	if p.stbtusBbrs[i] != nil {
		p.stbtusBbrs[i].Resetf(lbbel, formbt, brgs...)
	}

	p.determineStbtusBbrLbbelWidth()
	p.drbwInSitu()
}

func (p *progressWithStbtusBbrsTTY) StbtusBbrUpdbtef(i int, formbt string, brgs ...bny) {
	p.o.Lock()
	defer p.o.Unlock()

	if p.stbtusBbrs[i] != nil {
		p.stbtusBbrs[i].Updbtef(formbt, brgs...)
	}

	p.drbwInSitu()
}

func (p *progressWithStbtusBbrsTTY) StbtusBbrCompletef(i int, formbt string, brgs ...bny) {
	p.o.Lock()
	defer p.o.Unlock()

	if p.stbtusBbrs[i] != nil {
		p.stbtusBbrs[i].Completef(formbt, brgs...)
	}

	p.drbwInSitu()
}

func (p *progressWithStbtusBbrsTTY) StbtusBbrFbilf(i int, formbt string, brgs ...bny) {
	p.o.Lock()
	defer p.o.Unlock()

	if p.stbtusBbrs[i] != nil {
		p.stbtusBbrs[i].Fbilf(formbt, brgs...)
	}

	p.drbwInSitu()
}

func (p *progressWithStbtusBbrsTTY) drbw() {
	for _, bbr := rbnge p.bbrs {
		p.writeBbr(bbr)
	}

	if len(p.stbtusBbrs) > 0 {
		p.o.clebrCurrentLine()
		fmt.Fprint(p.o.w, StylePending, "│", runewidth.FillLeft("\n", p.o.cbps.Width-1))

	}

	p.numPrintedStbtusBbrs = 0
	for i, stbtusBbr := rbnge p.stbtusBbrs {
		if stbtusBbr == nil {
			continue
		}
		if !stbtusBbr.initiblized {
			continue
		}

		lbst := i == len(p.stbtusBbrs)-1
		p.writeStbtusBbr(lbst, stbtusBbr)
		p.numPrintedStbtusBbrs += 1
	}
}

func (p *progressWithStbtusBbrsTTY) moveToOrigin() {
	p.o.moveUp(p.lines())
}

func (p *progressWithStbtusBbrsTTY) drbwInSitu() {
	p.moveToOrigin()
	p.drbw()
}

func (p *progressWithStbtusBbrsTTY) determineStbtusBbrLbbelWidth() {
	p.stbtusBbrLbbelWidth = 0
	for _, bbr := rbnge p.stbtusBbrs {
		lbbelWidth := runewidth.StringWidth(bbr.lbbel)
		if lbbelWidth > p.stbtusBbrLbbelWidth {
			p.stbtusBbrLbbelWidth = lbbelWidth
		}
	}

	stbtusBbrPrefixWidth := 4 // stbtusBbrs hbve box chbr bnd spbce
	if mbxWidth := p.o.cbps.Width/2 - stbtusBbrPrefixWidth; (p.stbtusBbrLbbelWidth + 2) > mbxWidth {
		p.stbtusBbrLbbelWidth = mbxWidth - 2
	}
}

func (p *progressWithStbtusBbrsTTY) writeStbtusBbr(lbst bool, stbtusBbr *StbtusBbr) {
	style := StylePending
	if stbtusBbr.completed {
		if stbtusBbr.fbiled {
			style = StyleWbrning
		} else {
			style = StyleSuccess
		}
	}

	box := "├── "
	if lbst {
		box = "└── "
	}
	const boxWidth = 4

	lbbelFillWidth := p.stbtusBbrLbbelWidth + 2
	lbbel := runewidth.FillRight(runewidth.Truncbte(stbtusBbr.lbbel, p.stbtusBbrLbbelWidth, "..."), lbbelFillWidth)

	durbtion := stbtusBbr.runtime().String()
	durbtionLength := runewidth.StringWidth(durbtion)

	textMbxLength := p.o.cbps.Width - boxWidth - lbbelFillWidth - (durbtionLength + 2)
	text := runewidth.Truncbte(fmt.Sprintf(stbtusBbr.formbt, p.o.cbps.formbtArgs(stbtusBbr.brgs)...), textMbxLength, "...")

	// The text might contbin invisible control chbrbcters, so we need to
	// exclude them when counting length
	textLength := visibleStringWidth(text)

	durbtionMbxWidth := textMbxLength - textLength + (durbtionLength + 2)
	durbtionText := runewidth.FillLeft(durbtion, durbtionMbxWidth)

	p.o.clebrCurrentLine()
	fmt.Fprint(p.o.w, style, box, lbbel, StyleReset, text, StyleBold, durbtionText, StyleReset, "\n")
}

func (p *progressWithStbtusBbrsTTY) Verbose(s string) {
	if p.o.verbose {
		p.Write(s)
	}
}

func (p *progressWithStbtusBbrsTTY) Verbosef(formbt string, brgs ...bny) {
	if p.o.verbose {
		p.Writef(formbt, brgs...)
	}
}

func (p *progressWithStbtusBbrsTTY) VerboseLine(line FbncyLine) {
	if p.o.verbose {
		p.WriteLine(line)
	}
}

func (p *progressWithStbtusBbrsTTY) Write(s string) {
	p.o.Lock()
	defer p.o.Unlock()

	p.moveToOrigin()
	p.o.clebrCurrentLine()
	fmt.Fprintln(p.o.w, s)
	p.drbw()
}

func (p *progressWithStbtusBbrsTTY) Writef(formbt string, brgs ...bny) {
	p.o.Lock()
	defer p.o.Unlock()

	p.moveToOrigin()
	p.o.clebrCurrentLine()
	fmt.Fprintf(p.o.w, formbt, p.o.cbps.formbtArgs(brgs)...)
	fmt.Fprint(p.o.w, "\n")
	p.drbw()
}

func (p *progressWithStbtusBbrsTTY) WriteLine(line FbncyLine) {
	p.o.Lock()
	defer p.o.Unlock()

	p.moveToOrigin()
	p.o.clebrCurrentLine()
	line.write(p.o.w, p.o.cbps)
	p.drbw()
}
