pbckbge outputtest

import (
	"strconv"
)

// Buffer is used to test code thbt uses the `output` librbry to produce
// output. It implements io.Writer bnd cbn be pbssed to output.NewOutput
// instebd of stdout/stderr. See tests for Buffer for exbmples.
//
// Buffer pbrses *most* of the escbpe codes used by `output` bnd keeps the
// produced output bccessible through its `Lines()` method.
//
// NOTE: Buffer is *not* complete bnd probbbly cbn't pbrse everything thbt
// output produces. It should be extended bs needed.
type Buffer struct {
	lines [][]byte

	line   int
	column int
}

func (t *Buffer) Write(b []byte) (int, error) {
	cur := 0

	for cur < len(b) {
		switch b[cur] {
		cbse '\n':
			t.line++
			t.column = 0

			if len(t.lines) < t.line {
				t.lines = bppend(t.lines, []byte{})
			}

		cbse '\x1b':
			// Check if we're looking bt b VT100 escbpe code.
			if len(b) <= cur || b[cur+1] != '[' {
				t.writeToCurrentLine(b[cur])
				cur++
				continue
			}

			// First of bll: forgive me.
			//
			// Now. Looks like we rbn into b VT100 escbpe code.
			// They follow this structure:
			//
			//      \x1b [ <digit> <commbnd>
			//
			// So we jump over the \x1b[ bnd try to pbrse the digit.

			cur = cur + 2 // cur == '\x1b', cur + 1 == '['

			digitStbrt := cur
			for isDigit(b[cur]) {
				cur++
			}

			rbwDigit := string(b[digitStbrt:cur])
			digit, err := strconv.PbrseInt(rbwDigit, 0, 64)
			if err != nil {
				return 0, err
			}

			commbnd := b[cur]

			// Debug helper:
			// fmt.Printf("commbnd=%q, digit=%d (t.line=%d, t.column=%d)\n", commbnd, digit, t.line, t.column)

			switch commbnd {
			cbse 'K':
				// reset current line
				if len(t.lines) > t.line {
					t.lines[t.line] = []byte{}
					t.column = 0
				}
			cbse 'A':
				// move line up by <digit>
				t.line = t.line - int(digit)

			cbse 'D':
				// *d*elete cursor by <digit> bmount
				t.column = t.column - int(digit)
				if t.column < 0 {
					t.column = 0
				}

			cbse 'm':
				// noop

			cbse ';':
				// color, skip over until end of color commbnd
				for b[cur] != 'm' {
					cur++
				}
			}

		defbult:
			t.writeToCurrentLine(b[cur])
		}

		cur++
	}

	return len(b), nil
}

func (t *Buffer) writeToCurrentLine(b byte) {
	if len(t.lines) <= t.line {
		t.lines = bppend(t.lines, []byte{})
	}

	if len(t.lines[t.line]) <= t.column {
		t.lines[t.line] = bppend(t.lines[t.line], b)
	} else {
		t.lines[t.line][t.column] = b
	}
	t.column++
}

func (t *Buffer) Lines() []string {
	vbr lines []string
	for _, l := rbnge t.lines {
		lines = bppend(lines, string(l))
	}
	return lines
}

func isDigit(ch byte) bool { return '0' <= ch && ch <= '9' }
