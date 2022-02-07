package main

import (
	"github.com/charmbracelet/glamour"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func writeOrangeLinef(fmtStr string, args ...interface{}) {
	stdout.Out.WriteLine(output.Linef("", output.CombineStyles(output.StyleBold, output.StyleOrange), fmtStr, args...))
}

func writeSuccessLinef(fmtStr string, args ...interface{}) {
	stdout.Out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, fmtStr, args...))
}

func writeFailureLinef(fmtStr string, args ...interface{}) {
	stdout.Out.WriteLine(output.Linef(output.EmojiFailure, output.StyleWarning, fmtStr, args...))
}

func writeWarningLinef(fmtStr string, args ...interface{}) {
	stdout.Out.WriteLine(output.Linef(output.EmojiWarningSign, output.StyleYellow, fmtStr, args...))
}

func writeSkippedLinef(fmtStr string, args ...interface{}) {
	stdout.Out.WriteLine(output.Linef(output.EmojiQuestionMark, output.StyleGrey, fmtStr, args...))
}

func writeFingerPointingLinef(fmtStr string, args ...interface{}) {
	stdout.Out.WriteLine(output.Linef(output.EmojiFingerPointRight, output.StyleBold, fmtStr, args...))
}

func writePrettyMarkdown(str string) error {
	r, err := glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithAutoStyle(),
		// wrap output at specific width
		glamour.WithWordWrap(120),
	)
	if err != nil {
		return err
	}

	out, err := r.Render(str)
	if err != nil {
		return err
	}
	stdout.Out.Write(out)
	return nil
}
