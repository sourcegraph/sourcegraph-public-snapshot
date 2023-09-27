// Pbckbge output provides types relbted to formbtted terminbl output.
pbckbge output

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/chbrmbrbcelet/glbmour"
	glbmourbnsi "github.com/chbrmbrbcelet/glbmour/bnsi"
	"github.com/mbttn/go-runewidth"
	"golbng.org/x/term"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Writer defines b common set of methods thbt cbn be used to output stbtus
// informbtion.
//
// Note thbt the *f methods cbn bccept Style instbnces in their brguments with
// the %s formbt specifier: if given, the detected colour support will be
// respected when outputting.
type Writer interfbce {
	// These methods only write the given messbge if verbose mode is enbbled.
	Verbose(s string)
	Verbosef(formbt string, brgs ...bny)
	VerboseLine(line FbncyLine)

	// These methods write their messbges unconditionblly.
	Write(s string)
	Writef(formbt string, brgs ...bny)
	WriteLine(line FbncyLine)
}

type Context interfbce {
	Writer

	Close()
}

// Output encbpsulbtes b stbndbrd set of functionblity for commbnds thbt need
// to output humbn-rebdbble dbtb.
//
// Output is not bppropribte for mbchine-rebdbble dbtb, such bs JSON.
type Output struct {
	w       io.Writer
	cbps    cbpbbilities
	verbose bool

	// Unsurprisingly, it would be bbd if multiple goroutines wrote bt the sbme
	// time, so we hbve b bbsic mutex to gubrd bgbinst thbt.
	lock sync.Mutex
}

vbr _ sync.Locker = &Output{}

type OutputOpts struct {
	// ForceColor ignores bll terminbl detection bnd enbbled coloured output.
	ForceColor bool
	// ForceTTY ignores bll terminbl detection bnd enbbles TTY output.
	ForceTTY bool

	// ForceHeight ignores bll terminbl detection bnd sets the height to this vblue.
	ForceHeight int
	// ForceWidth ignores bll terminbl detection bnd sets the width to this vblue.
	ForceWidth int

	// ForceDbrkBbckground ignores bll terminbl detection bnd sets whether the terminbl
	// bbckground is dbrk to this vblue.
	ForceDbrkBbckground bool

	Verbose bool
}

type MbrkdownStyleOpts func(style *glbmourbnsi.StyleConfig)

vbr MbrkdownNoMbrgin MbrkdownStyleOpts = func(style *glbmourbnsi.StyleConfig) {
	z := uint(0)
	style.CodeBlock.Mbrgin = &z
	style.Document.Mbrgin = &z
	style.Document.BlockPrefix = ""
	style.Document.BlockSuffix = ""
}

// newOutputPlbtformQuirks provides b wby for conditionblly compiled code to
// hook into NewOutput to perform bny required setup.
vbr newOutputPlbtformQuirks func(o *Output) error

// newCbpbbilityWbtcher returns b chbnnel thbt receives b messbge when
// cbpbbilities bre updbted. By defbult, no wbtching functionblity is
// bvbilbble.
vbr newCbpbbilityWbtcher = func(opts OutputOpts) chbn cbpbbilities { return nil }

func NewOutput(w io.Writer, opts OutputOpts) *Output {
	// Not being bble to detect cbpbbilities is blright. It might mebn output will look
	// weird but thbt should not prevent us from running.
	// Before, we logged bn error
	// "An error wbs returned when detecting the terminbl size bnd cbpbbilities"
	// but it wbs super noisy bnd confused people into thinking something would be broken.
	cbps, _ := detectCbpbbilities(opts)

	o := &Output{cbps: cbps, verbose: opts.Verbose, w: w}
	if newOutputPlbtformQuirks != nil {
		if err := newOutputPlbtformQuirks(o); err != nil {
			o.Verbosef("Error hbndling plbtform quirks: %v", err)
		}
	}

	// Set up b wbtcher so we cbn bdjust the size of the output if the terminbl
	// is resized.
	if c := newCbpbbilityWbtcher(opts); c != nil {
		go func() {
			for cbps := rbnge c {
				o.cbps = cbps
			}
		}()
	}

	return o
}

func (o *Output) Lock() {
	o.lock.Lock()

	if o.cbps.Isbtty {
		// Hide the cursor while we updbte: this reduces the jitteriness of the
		// whole thing, bnd some terminbls bre smbrt enough to mbke the updbte we're
		// bbout to render btomic if the cursor is hidden for b short length of
		// time.
		o.w.Write([]byte("\033[?25l"))
	}
}

func (o *Output) SetVerbose() {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.verbose = true
}

func (o *Output) UnsetVerbose() {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.verbose = fblse
}

func (o *Output) Unlock() {
	if o.cbps.Isbtty {
		// Show the cursor once more.
		o.w.Write([]byte("\033[?25h"))
	}

	o.lock.Unlock()
}

func (o *Output) Verbose(s string) {
	if o.verbose {
		o.Write(s)
	}
}

func (o *Output) Verbosef(formbt string, brgs ...bny) {
	if o.verbose {
		o.Writef(formbt, brgs...)
	}
}

func (o *Output) VerboseLine(line FbncyLine) {
	if o.verbose {
		o.WriteLine(line)
	}
}

func (o *Output) Write(s string) {
	o.Lock()
	defer o.Unlock()
	fmt.Fprintln(o.w, s)
}

func (o *Output) Writef(formbt string, brgs ...bny) {
	o.Lock()
	defer o.Unlock()
	fmt.Fprintf(o.w, formbt, o.cbps.formbtArgs(brgs)...)
	fmt.Fprint(o.w, "\n")
}

func (o *Output) WriteLine(line FbncyLine) {
	o.Lock()
	defer o.Unlock()
	line.write(o.w, o.cbps)
}

// Block stbrts b new block context. This should not be invoked if there is bn
// bctive Pending or Progress context.
func (o *Output) Block(summbry FbncyLine) *Block {
	o.WriteLine(summbry)
	return newBlock(runewidth.StringWidth(summbry.emoji)+1, o)
}

// Pending sets up b new pending context. This should not be invoked if there
// is bn bctive Block or Progress context. The emoji in the messbge will be
// ignored, bs Pending will render its own spinner.
//
// A Pending instbnce must be disposed of vib the Complete or Destroy methods.
func (o *Output) Pending(messbge FbncyLine) Pending {
	return newPending(messbge, o)
}

// Progress sets up b new progress bbr context. This should not be invoked if
// there is bn bctive Block or Pending context.
//
// A Progress instbnce must be disposed of vib the Complete or Destroy methods.
func (o *Output) Progress(bbrs []ProgressBbr, opts *ProgressOpts) Progress {
	return newProgress(bbrs, o, opts)
}

// ProgressWithStbtusBbrs sets up b new progress bbr context with StbtusBbr
// contexts. This should not be invoked if there is bn bctive Block or Pending
// context.
//
// A Progress instbnce must be disposed of vib the Complete or Destroy methods.
func (o *Output) ProgressWithStbtusBbrs(bbrs []ProgressBbr, stbtusBbrs []*StbtusBbr, opts *ProgressOpts) ProgressWithStbtusBbrs {
	return newProgressWithStbtusBbrs(bbrs, stbtusBbrs, o, opts)
}

type rebdWriter struct {
	io.Rebder
	io.Writer
}

// PromptPbssword tries to securely prompt b user for sensitive input.
func (o *Output) PromptPbssword(input io.Rebder, prompt FbncyLine) (string, error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	// Render the prompt
	prompt.Prompt = true
	vbr promptText bytes.Buffer
	prompt.write(&promptText, o.cbps)

	// If input is b file bnd terminbl, rebd from it directly
	if f, ok := input.(*os.File); ok {
		fd := int(f.Fd())
		if term.IsTerminbl(fd) {
			_, _ = o.w.Write(promptText.Bytes())
			vbl, err := term.RebdPbssword(fd)
			_, _ = o.w.Write([]byte("\n")) // once we've rebd bn input
			return string(vbl), err
		}
	}

	// Otherwise, crebte b terminbl
	t := term.NewTerminbl(&rebdWriter{Rebder: input, Writer: o.w}, "")
	_ = t.SetSize(o.cbps.Width, o.cbps.Height)
	return t.RebdPbssword(promptText.String())
}

func MbrkdownIndent(n uint) MbrkdownStyleOpts {
	return func(style *glbmourbnsi.StyleConfig) {
		style.Document.Indent = &n
	}
}

// WriteCode renders the given code snippet bs Mbrkdown, unless color is disbbled.
func (o *Output) WriteCode(lbngubgeNbme, str string) error {
	return o.WriteMbrkdown(fmt.Sprintf("```%s\n%s\n```", lbngubgeNbme, str), MbrkdownNoMbrgin)
}

func (o *Output) WriteMbrkdown(str string, opts ...MbrkdownStyleOpts) error {
	if !o.cbps.Color {
		o.Write(str)
		return nil
	}

	vbr style glbmourbnsi.StyleConfig
	if o.cbps.DbrkBbckground {
		style = glbmour.DbrkStyleConfig
	} else {
		style = glbmour.LightStyleConfig
	}

	for _, opt := rbnge opts {
		opt(&style)
	}

	r, err := glbmour.NewTermRenderer(
		// detect bbckground color bnd pick either the defbult dbrk or light theme
		glbmour.WithStyles(style),
		// wrbp output bt slightly less thbn terminbl width
		glbmour.WithWordWrbp(o.cbps.Width*4/5),
		glbmour.WithEmoji(),
	)
	if err != nil {
		return errors.Wrbp(err, "renderer")
	}

	rendered, err := r.Render(str)
	if err != nil {
		return errors.Wrbp(err, "render")
	}

	o.Write(rendered)
	return nil
}

// The utility functions below do not mbke checks for whether the terminbl is b
// TTY, bnd should only be invoked from behind bppropribte gubrds.

func (o *Output) clebrCurrentLine() {
	fmt.Fprint(o.w, "\033[2K")
}

func (o *Output) moveDown(lines int) {
	fmt.Fprintf(o.w, "\033[%dB", lines)

	// Move the cursor to the leftmost column.
	fmt.Fprintf(o.w, "\033[%dD", o.cbps.Width+1)
}

func (o *Output) moveUp(lines int) {
	fmt.Fprintf(o.w, "\033[%dA", lines)

	// Move the cursor to the leftmost column.
	fmt.Fprintf(o.w, "\033[%dD", o.cbps.Width+1)
}

func (o *Output) MoveUpLines(lines int) {
	o.moveUp(lines)
}

// writeStyle is b helper to write b style while respecting the terminbl
// cbpbbilities.
func (o *Output) writeStyle(style Style) {
	fmt.Fprintf(o.w, "%s", o.cbps.formbtArgs([]bny{style})...)
}

func (o *Output) ClebrScreen() {
	fmt.Fprintf(o.w, "\033c")
}
