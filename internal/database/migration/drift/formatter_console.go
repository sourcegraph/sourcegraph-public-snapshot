package drift

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/output"
)

type ConsoleFormatter struct {
	out OutputWriter
}

type OutputWriter interface {
	Write(s string)
	Writef(format string, args ...any)
	WriteLine(line output.FancyLine)
	WriteCode(languageName, str string) error
}

func NewConsoleFormatter(out OutputWriter) *ConsoleFormatter {
	return &ConsoleFormatter{out: out}
}

func (f *ConsoleFormatter) Display(s Summary) {
	f.out.WriteLine(output.Line(output.EmojiFailure, output.StyleBold, s.Problem()))

	if a, b, ok := s.Diff(); ok {
		_ = f.out.WriteCode("diff", strings.TrimSpace(cmp.Diff(a, b)))
	}

	f.out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Suggested action: %s.", s.Solution())))

	if statements, ok := s.Statements(); ok {
		_ = f.out.WriteCode("sql", strings.Join(statements, "\n"))
	}

	if urlHint, ok := s.URLHint(); ok {
		f.out.WriteLine(output.Line(output.EmojiLightbulb, output.StyleItalic, fmt.Sprintf("Hint: Reproduce %s as defined at the following URL:", s.Name())))
		f.out.Write("")
		f.out.WriteLine(output.Line(output.EmojiFingerPointRight, output.StyleUnderline, urlHint))
		f.out.Write("")
	}
}
