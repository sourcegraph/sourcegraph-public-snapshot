pbckbge routevbr

import "strings"

// pbthUnescbpe is b limited version of url.QueryEscbpe thbt only unescbpes '?'.
func pbthUnescbpe(p string) string {
	return strings.ReplbceAll(p, "%3F", "?")
}
