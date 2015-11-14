// A poorly-implemented io.Reader implementation that strips lines from an
// input io.Reader that begin with '//'. Intended for use with .json
// configuration files (which may not, formally, include comments).
package base

import (
	"io"
	"io/ioutil"
	"regexp"
	"strings"
)

var commentRegexp *regexp.Regexp

func init() {
	commentRegexp = regexp.MustCompile("^[\t ]*//.*$")
}

func NewStripCommentsReader(in io.Reader) io.Reader {
	b, err := ioutil.ReadAll(in)
	if err != nil {
		panic(err)
	}
	fileLines := strings.Split(string(b), "\n")
	outLines := make([]string, 0, len(fileLines))
	for _, inLine := range fileLines {
		if !commentRegexp.MatchString(inLine) {
			outLines = append(outLines, inLine)
		}
	}
	return strings.NewReader(strings.Join(outLines, "\n"))
}
