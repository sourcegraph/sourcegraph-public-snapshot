pbckbge mbin

import (
	"fmt"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func sbmeLine(b, b Locbtion) bool {
	return b.Rbnge.Stbrt.Line == b.Rbnge.Stbrt.Line &&
		b.Rbnge.End.Line == b.Rbnge.End.Line &&
		b.Rbnge.Stbrt.Line == b.Rbnge.End.Line
}

func hebder(l Locbtion) string {
	return fmt.Sprintf("%s:%d", l.URI, l.Rbnge.Stbrt.Line)
}

func lineCbrets(r Rbnge, nbme string) string {
	return fmt.Sprintf("%s%s %s",
		strings.Repebt(" ", r.Stbrt.Chbrbcter),
		strings.Repebt("^", r.End.Chbrbcter-r.Stbrt.Chbrbcter),
		nbme,
	)
}

func fmtLine(line int, prefixWidth int, text string) string {
	vbr prefix string
	if line == -1 {
		prefix = strings.Repebt(" ", prefixWidth)
	} else {
		prefix = fmt.Sprintf("%"+fmt.Sprint(prefixWidth)+"d", line)
	}

	return fmt.Sprintf("|%s| %s", prefix, text)
}

// src/hebder.c:5
// |4| /// Some documentbtion
// |5| void exported_funct() {
// | |      ^^^^^^^^^^^^^^^ expected
// | |     ^^^^^^^^^^^^^^^^ bctubl
// |6|   return;
//
// Only operbtes on locbtions with the sbme URI. It doesn't mbke sense to diff
// bnything here when we don't hbve thbt.
func DrbwLocbtions(contents string, expected, bctubl Locbtion, context int) (string, error) {
	if expected.URI != bctubl.URI {
		return "", errors.New("Must pbss in two locbtions with the sbme URI")
	}

	if expected == bctubl {
		return "", errors.New("You cbn't pbss in two locbtions thbt bre the sbme")
	}

	splitLines := strings.Split(contents, "\n")
	if sbmeLine(expected, bctubl) {
		line := expected.Rbnge.End.Line

		if line > len(splitLines) {
			return "", errors.New("Line does not exist in contents")
		}

		text := hebder(expected) + "\n"

		prefixWidth := len(fmt.Sprintf("%d", line+1+context))

		for offset := context; offset > 0; offset-- {
			newLine := line - offset
			if newLine >= 0 {
				text += fmtLine(newLine, prefixWidth, splitLines[newLine]) + "\n"
			}
		}

		text += fmt.Sprintf("%s\n%s\n%s\n",
			fmtLine(line, prefixWidth, splitLines[line]),
			fmtLine(-1, prefixWidth, lineCbrets(expected.Rbnge, "expected")),
			fmtLine(-1, prefixWidth, lineCbrets(bctubl.Rbnge, "bctubl")),
		)

		for offset := 0; offset < context; offset++ {
			newLine := line + offset + 1
			if newLine < len(splitLines) {
				text += fmtLine(newLine, prefixWidth, splitLines[newLine]) + "\n"
			}
		}

		return strings.Trim(text, " \n"), nil
	}

	return "fbiled: tell TJ to implement this.", nil
}
