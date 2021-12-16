package main

import "github.com/sourcegraph/sourcegraph/lib/output"

func writeOrangeLinef(fmtStr string, args ...interface{}) {
	out.WriteLine(output.Linef("", output.CombineStyles(output.StyleBold, output.StyleOrange), fmtStr, args...))
}

func writeSuccessLinef(fmtStr string, args ...interface{}) {
	out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, fmtStr, args...))
}

func writeFailureLinef(fmtStr string, args ...interface{}) {
	out.WriteLine(output.Linef(output.EmojiFailure, output.StyleWarning, fmtStr, args...))
}

func writeWarningLinef(fmtStr string, args ...interface{}) {
	out.WriteLine(output.Linef(output.EmojiWarningSign, output.StyleYellow, fmtStr, args...))
}

func writeSkippedLinef(fmtStr string, args ...interface{}) {
	out.WriteLine(output.Linef(output.EmojiQuestionMark, output.StyleGrey, fmtStr, args...))
}

func writeFingerPointingLinef(fmtStr string, args ...interface{}) {
	out.WriteLine(output.Linef(output.EmojiFingerPointRight, output.StyleBold, fmtStr, args...))
}
