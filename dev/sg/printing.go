package main

import (
	"fmt"

	"github.com/charmbracelet/glamour"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func writeOrangeLinef(fmtStr string, args ...interface{}) {
	stdout.Out.WriteLine(output.Linef("", output.CombineStyles(output.StyleBold, output.StyleOrange), fmtStr, args...))
}

func printSgSetupWelcomeScreen() {
	genLine := func(style output.Style, content string) string {
		return fmt.Sprintf("%s%s%s", output.CombineStyles(output.StyleBold, style), content, output.StyleReset)
	}

	boxContent := func(content string) string { return genLine(output.StyleWhiteOnPurple, content) }
	shadow := func(content string) string { return genLine(output.StyleGreyBackground, content) }

	stdout.Out.Write(boxContent(`┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ sg ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓`))
	stdout.Out.Write(boxContent(`┃            _       __     __                             __                ┃`))
	stdout.Out.Write(boxContent(`┃           | |     / /__  / /________  ____ ___  ___     / /_____           ┃`) + shadow(`  `))
	stdout.Out.Write(boxContent(`┃           | | /| / / _ \/ / ___/ __ \/ __ '__ \/ _ \   / __/ __ \          ┃`) + shadow(`  `))
	stdout.Out.Write(boxContent(`┃           | |/ |/ /  __/ / /__/ /_/ / / / / / /  __/  / /_/ /_/ /          ┃`) + shadow(`  `))
	stdout.Out.Write(boxContent(`┃           |__/|__/\___/_/\___/\____/_/ /_/ /_/\___/   \__/\____/           ┃`) + shadow(`  `))
	stdout.Out.Write(boxContent(`┃                                           __              __               ┃`) + shadow(`  `))
	stdout.Out.Write(boxContent(`┃                  ___________   ________  / /___  ______  / /               ┃`) + shadow(`  `))
	stdout.Out.Write(boxContent(`┃                 / ___/ __  /  / ___/ _ \/ __/ / / / __ \/ /                ┃`) + shadow(`  `))
	stdout.Out.Write(boxContent(`┃                (__  ) /_/ /  (__  )  __/ /_/ /_/ / /_/ /_/                 ┃`) + shadow(`  `))
	stdout.Out.Write(boxContent(`┃               /____/\__, /  /____/\___/\__/\__,_/ .___(_)                  ┃`) + shadow(`  `))
	stdout.Out.Write(boxContent(`┃                    /____/                      /_/                         ┃`) + shadow(`  `))
	stdout.Out.Write(boxContent(`┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛`) + shadow(`  `))
	stdout.Out.Write(`  ` + shadow(`                                                                              `))
	stdout.Out.Write(`  ` + shadow(`                                                                              `))
}

func writeSuccessLinef(fmtStr string, args ...interface{}) {
	stdout.Out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, fmtStr, args...))
}

func writeFailureLinef(fmtStr string, args ...interface{}) {
	stdout.Out.WriteLine(output.Linef(output.EmojiFailure, output.StyleWarning, fmtStr, args...))
}

func newWarningLinef(fmtStr string, args ...interface{}) output.FancyLine {
	return output.Linef(output.EmojiWarningSign, output.StyleYellow, fmtStr, args...)
}

func writeWarningLinef(fmtStr string, args ...interface{}) {
	stdout.Out.WriteLine(newWarningLinef(fmtStr, args...))
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
		glamour.WithEmoji(),
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
