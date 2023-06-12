package drift

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var errOutOfSync = errors.Newf("database schema is out of sync")

func DisplaySchemaSummaries(rawOut *output.Output, summaries []Summary) (err error) {
	out := &preambledOutput{rawOut, false}

	for _, summary := range summaries {
		displaySummary(out, summary)
		err = errOutOfSync
	}

	if err == nil {
		rawOut.WriteLine(output.Line(output.EmojiSuccess, output.StyleSuccess, "No drift detected"))
	}
	return err
}

func displaySummary(out *preambledOutput, summary Summary) {
	out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, summary.Problem()))

	if a, b, ok := summary.Diff(); ok {
		_ = out.WriteCode("diff", strings.TrimSpace(cmp.Diff(a, b)))
	}

	out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Suggested action: %s.", summary.Solution())))

	if statements, ok := summary.Statements(); ok {
		_ = out.WriteCode("sql", strings.Join(statements, "\n"))
	}

	if urlHint, ok := summary.URLHint(); ok {
		out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Hint: Reproduce %s as defined at the following URL:", summary.Name())))
		out.Write("")
		out.WriteLine(output.Line(output.EmojiFingerPointRight, output.StyleUnderline, urlHint))
		out.Write("")
	}
}

type preambledOutput struct {
	out     *output.Output
	emitted bool
}

func (o *preambledOutput) check() {
	if o.emitted {
		return
	}

	o.out.WriteLine(output.Line(output.EmojiFailure, output.StyleFailure, "Drift detected!"))
	o.out.Write("")
	o.emitted = true
}

func (o *preambledOutput) Write(s string) {
	o.check()
	o.out.Write(s)
}

func (o *preambledOutput) Writef(format string, args ...any) {
	o.check()
	o.out.Writef(format, args...)
}

func (o *preambledOutput) WriteLine(line output.FancyLine) {
	o.check()
	o.out.WriteLine(line)
}

func (o *preambledOutput) WriteCode(languageName, str string) error {
	o.check()
	return o.out.WriteCode(languageName, str)
}
