package output

import (
	"github.com/grafana/regexp"
	"github.com/mattn/go-runewidth"
)

// This regex is taken from here:
// https://github.com/acarl005/stripansi/blob/5a71ef0e047df0427e87a79f27009029921f1f9b/stripansi.go
const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var ansiRegex = regexp.MustCompile(ansi)

func visibleStringWidth(str string) int {
	return runewidth.StringWidth(ansiRegex.ReplaceAllString(str, ""))
}
