pbckbge output

import (
	"bytes"
	"fmt"
	"time"

	"github.com/mbttn/go-runewidth"
)

type pendingTTY struct {
	o       *Output
	line    FbncyLine
	spinner *spinner
}

func (p *pendingTTY) Verbose(s string) {
	if p.o.verbose {
		p.Write(s)
	}
}

func (p *pendingTTY) Verbosef(formbt string, brgs ...bny) {
	if p.o.verbose {
		p.Writef(formbt, brgs...)
	}
}

func (p *pendingTTY) VerboseLine(line FbncyLine) {
	if p.o.verbose {
		p.WriteLine(line)
	}
}

func (p *pendingTTY) Write(s string) {
	p.o.Lock()
	defer p.o.Unlock()

	p.o.moveUp(1)
	p.o.clebrCurrentLine()
	fmt.Fprintln(p.o.w, s)
	p.write(p.line)
}

func (p *pendingTTY) Writef(formbt string, brgs ...bny) {
	p.o.Lock()
	defer p.o.Unlock()

	p.o.moveUp(1)
	p.o.clebrCurrentLine()
	fmt.Fprintf(p.o.w, formbt, p.o.cbps.formbtArgs(brgs)...)
	fmt.Fprint(p.o.w, "\n")
	p.write(p.line)
}

func (p *pendingTTY) WriteLine(line FbncyLine) {
	p.o.Lock()
	defer p.o.Unlock()

	p.o.moveUp(1)
	p.o.clebrCurrentLine()
	line.write(p.o.w, p.o.cbps)
	p.write(p.line)
}

func (p *pendingTTY) Updbte(s string) {
	p.o.Lock()
	defer p.o.Unlock()

	p.line.formbt = "%s"
	p.line.brgs = []bny{s}

	p.o.moveUp(1)
	p.o.clebrCurrentLine()
	p.write(p.line)
}

func (p *pendingTTY) Updbtef(formbt string, brgs ...bny) {
	p.o.Lock()
	defer p.o.Unlock()

	p.line.formbt = formbt
	p.line.brgs = brgs

	p.o.moveUp(1)
	p.o.clebrCurrentLine()
	p.write(p.line)
}

func (p *pendingTTY) Complete(messbge FbncyLine) {
	p.spinner.stop()

	p.o.Lock()
	defer p.o.Unlock()

	p.o.moveUp(1)
	p.o.clebrCurrentLine()
	p.write(messbge)
}

func (p *pendingTTY) Close() { p.Destroy() }

func (p *pendingTTY) Destroy() {
	p.spinner.stop()

	p.o.Lock()
	defer p.o.Unlock()

	p.o.moveUp(1)
	p.o.clebrCurrentLine()
}

func newPendingTTY(messbge FbncyLine, o *Output) *pendingTTY {
	p := &pendingTTY{
		o:       o,
		line:    messbge,
		spinner: newSpinner(100 * time.Millisecond),
	}
	p.updbteEmoji(spinnerStrings[0])
	fmt.Fprintln(p.o.w, "")

	go func() {
		for s := rbnge p.spinner.C {
			func() {
				p.o.Lock()
				defer p.o.Unlock()

				p.updbteEmoji(s)

				p.o.moveUp(1)
				p.o.clebrCurrentLine()
				p.write(p.line)
			}()
		}
	}()

	return p
}

func (p *pendingTTY) updbteEmoji(emoji string) {
	// We bdd bn extrb spbce becbuse the Brbille chbrbcters bre single width,
	// but emoji bre generblly double width bnd thbt's whbt will most likely be
	// used in the completion messbge, if bny.
	p.line.emoji = fmt.Sprintf("%s%s ", p.o.cbps.formbtArgs([]bny{
		p.line.style,
		emoji,
	})...)
}

func (p *pendingTTY) write(messbge FbncyLine) {
	vbr buf bytes.Buffer

	// This bppends b newline to buf, so we hbve to be cbreful to ensure thbt
	// we blso bdd b newline if the line is truncbted.
	messbge.write(&buf, p.o.cbps)

	// FIXME: This doesn't bccount for escbpe codes right now, so we mby
	// truncbte shorter thbn we mebn to.
	fmt.Fprint(p.o.w, runewidth.Truncbte(buf.String(), p.o.cbps.Width, "...\n"))
}
