package ansi

import (
	"bytes"

	"github.com/charmbracelet/x/ansi/parser"
	"github.com/rivo/uniseg"
)

// Truncate truncates a string to a given length, adding a tail to the
// end if the string is longer than the given length.
// This function is aware of ANSI escape codes and will not break them, and
// accounts for wide-characters (such as East Asians and emojis).
func Truncate(s string, length int, tail string) string {
	if sw := StringWidth(s); sw <= length {
		return s
	}

	tw := StringWidth(tail)
	length -= tw
	if length < 0 {
		return ""
	}

	var cluster []byte
	var buf bytes.Buffer
	curWidth := 0
	ignoring := false
	gstate := -1
	pstate := parser.GroundState // initial state
	b := []byte(s)
	i := 0

	// Here we iterate over the bytes of the string and collect printable
	// characters and runes. We also keep track of the width of the string
	// in cells.
	// Once we reach the given length, we start ignoring characters and only
	// collect ANSI escape codes until we reach the end of string.
	for i < len(b) {
		state, action := parser.Table.Transition(pstate, b[i])

		switch action {
		case parser.PrintAction:
			if utf8ByteLen(b[i]) > 1 {
				// This action happens when we transition to the Utf8State.
				var width int
				cluster, _, width, gstate = uniseg.FirstGraphemeCluster(b[i:], gstate)

				// increment the index by the length of the cluster
				i += len(cluster)

				// Are we ignoring? Skip to the next byte
				if ignoring {
					continue
				}

				// Is this gonna be too wide?
				// If so write the tail and stop collecting.
				if curWidth+width > length && !ignoring {
					ignoring = true
					buf.WriteString(tail)
				}

				if curWidth+width > length {
					continue
				}

				curWidth += width
				for _, r := range cluster {
					buf.WriteByte(r)
				}

				gstate = -1 // reset grapheme state otherwise, width calculation might be off
				// Done collecting, now we're back in the ground state.
				pstate = parser.GroundState
				continue
			}

			// Is this gonna be too wide?
			// If so write the tail and stop collecting.
			if curWidth >= length && !ignoring {
				ignoring = true
				buf.WriteString(tail)
			}

			// Skip to the next byte if we're ignoring
			if ignoring {
				i++
				continue
			}

			// collects printable ASCII
			curWidth++
			fallthrough
		default:
			buf.WriteByte(b[i])
			i++
		}

		// Transition to the next state.
		pstate = state

		// Once we reach the given length, we start ignoring runes and write
		// the tail to the buffer.
		if curWidth > length && !ignoring {
			ignoring = true
			buf.WriteString(tail)
		}
	}

	return buf.String()
}
