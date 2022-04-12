package bk

import (
	"strings"

	"github.com/grafana/regexp"
)

// CleanANSI cleans up Buildkite log output markup.
func CleanANSI(s string) string {
	// https://github.com/acarl005/stripansi/blob/master/stripansi.go
	ansi := regexp.MustCompile("[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))")
	s = ansi.ReplaceAllString(s, "")
	// Other weird markers not caught be above regex
	s = strings.ReplaceAll(s, "\x1BE", "")
	s = strings.ReplaceAll(s, "\x1b", "")
	s = strings.ReplaceAll(s, "\a", "")
	return s
}
