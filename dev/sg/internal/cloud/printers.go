package cloud

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type Printer interface {
	Print([]*Instance) error
}

type terminalInstancePrinter struct {
	headingFmt string
	headings   []any
	valueFunc  func(i *Instance) []any
}

type jsonInstancePrinter struct {
	w io.Writer
}

func newTerminalInstancePrinter(valueFunc func(i *Instance) []any, headingFmt string, headings ...string) *terminalInstancePrinter {
	anyHeadings := make([]any, 0, len(headings))
	for _, h := range headings {
		anyHeadings = append(anyHeadings, h)
	}

	return &terminalInstancePrinter{
		headingFmt: headingFmt,
		headings:   anyHeadings,
		valueFunc:  valueFunc,
	}
}

func (p *terminalInstancePrinter) Print(items []*Instance) error {
	heading := fmt.Sprintf(p.headingFmt, p.headings...)
	std.Out.WriteLine(output.Line("", output.StyleBold, heading))
	for _, inst := range items {
		values := p.valueFunc(inst)
		line := fmt.Sprintf("%-20s %-11s %s", values...)
		std.Out.WriteLine(output.Line("", output.StyleGrey, line))
	}
	return nil
}

func newJSONInstancePrinter(w io.Writer) *jsonInstancePrinter {
	return &jsonInstancePrinter{w: w}
}

func (p *jsonInstancePrinter) Print(items []*Instance) error {
	return json.NewEncoder(p.w).Encode(items)
}
