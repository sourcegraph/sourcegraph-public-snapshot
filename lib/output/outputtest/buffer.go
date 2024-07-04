package outputtest

import (
	"strconv"
	"sync"
)

// Buffer is used to test code that uses the `output` library to produce
// output. It implements io.Writer and can be passed to output.NewOutput
// instead of stdout/stderr. See tests for Buffer for examples.
//
// Buffer parses *most* of the escape codes used by `output` and keeps the
// produced output accessible through its `Lines()` method.
//
// NOTE: Buffer is *not* complete and probably can't parse everything that
// output produces. It should be extended as needed.
type Buffer struct {
	sync.Mutex
	lines [][]byte

	line   int
	column int
}

func (t *Buffer) Write(b []byte) (int, error) {
	t.Lock()
	defer t.Unlock()

	cur := 0

	// Debug helper:
	// fmt.Printf("b: %q\n", string(b))

	for cur < len(b) {
		switch b[cur] {
		case '\n':
			t.line++
			t.column = 0

			if len(t.lines) < t.line {
				t.lines = append(t.lines, []byte{})
			}

		case '\x1b':
			// Check if we're looking at a VT100 escape code.
			if len(b) <= cur || b[cur+1] != '[' {
				t.writeToCurrentLine(b[cur])
				cur++
				continue
			}

			// First of all: forgive me.
			//
			// Now. Looks like we ran into a VT100 escape code.
			// They follow this structure:
			//
			//      \x1b [ <digit> <command>
			//
			// So we jump over the \x1b[ and try to parse the digit.

			cur = cur + 2 // cur == '\x1b', cur + 1 == '['

			digitStart := cur
			for isDigit(b[cur]) {
				cur++
			}

			rawDigit := string(b[digitStart:cur])
			digit, err := strconv.ParseInt(rawDigit, 0, 64)
			if err != nil {
				return 0, err
			}

			command := b[cur]

			// Debug helper:
			// fmt.Printf("command=%q, digit=%d (t.line=%d, t.column=%d)\n", command, digit, t.line, t.column)

			switch command {
			case 'K':
				// reset current line
				if len(t.lines) > t.line {
					t.lines[t.line] = []byte{}
					t.column = 0
				}
			case 'A':
				// move line up by <digit>
				t.line = t.line - int(digit)

			case 'D':
				// *d*elete cursor by <digit> amount
				t.column = t.column - int(digit)
				if t.column < 0 {
					t.column = 0
				}

			case 'm':
				// noop

			case ';':
				// color, skip over until end of color command
				for b[cur] != 'm' {
					cur++
				}
			}

		default:
			t.writeToCurrentLine(b[cur])
		}

		cur++
	}

	return len(b), nil
}

func (t *Buffer) writeToCurrentLine(b byte) {
	if len(t.lines) <= t.line {
		t.lines = append(t.lines, []byte{})
	}

	if len(t.lines[t.line]) <= t.column {
		t.lines[t.line] = append(t.lines[t.line], b)
	} else {
		t.lines[t.line][t.column] = b
	}
	t.column++
}

func (t *Buffer) Lines() []string {
	t.Lock()
	defer t.Unlock()

	var lines []string
	for _, l := range t.lines {
		lines = append(lines, string(l))
	}
	return lines
}

func isDigit(ch byte) bool { return '0' <= ch && ch <= '9' }
