package main

import "github.com/sourcegraph/sourcegraph/lib/output"

func writeOrangeLine(fmtStr string, args ...interface{}) {
	out.WriteLine(output.Linef("", output.CombineStyles(output.StyleBold, output.StyleOrange), fmtStr, args...))
}

func writeSuccessLine(fmtStr string, args ...interface{}) {
	out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, fmtStr, args...))
}

func writeFailureLine(fmtStr string, args ...interface{}) {
	out.WriteLine(output.Linef(output.EmojiFailure, output.StyleWarning, fmtStr, args...))
}

func writeSkippedLine(fmtStr string, args ...interface{}) {
	out.WriteLine(output.Linef(output.EmojiQuestionMark, output.StyleGrey, fmtStr, args...))
}

func writeFingerPointingLine(fmtStr string, args ...interface{}) {
	out.WriteLine(output.Linef(output.EmojiFingerPointRight, output.StyleBold, fmtStr, args...))
}
