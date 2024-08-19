package ansi

import (
	"bytes"

	"github.com/charmbracelet/x/ansi/parser"
	"github.com/rivo/uniseg"
)

// Strip removes ANSI escape codes from a string.
func Strip(s string) string {
	var (
		buf    bytes.Buffer         // buffer for collecting printable characters
		ri     int                  // rune index
		rw     int                  // rune width
		pstate = parser.GroundState // initial state
	)

	// This implements a subset of the Parser to only collect runes and
	// printable characters.
	for i := 0; i < len(s); i++ {
		var state, action byte
		if pstate != parser.Utf8State {
			state, action = parser.Table.Transition(pstate, s[i])
		}

		switch {
		case pstate == parser.Utf8State:
			// During this state, collect rw bytes to form a valid rune in the
			// buffer. After getting all the rune bytes into the buffer,
			// transition to GroundState and reset the counters.
			buf.WriteByte(s[i])
			ri++
			if ri < rw {
				continue
			}
			pstate = parser.GroundState
			ri = 0
			rw = 0
		case action == parser.PrintAction:
			// This action happens when we transition to the Utf8State.
			if w := utf8ByteLen(s[i]); w > 1 {
				rw = w
				buf.WriteByte(s[i])
				ri++
				break
			}
			fallthrough
		case action == parser.ExecuteAction:
			// collects printable ASCII and non-printable characters
			buf.WriteByte(s[i])
		}

		// Transition to the next state.
		// The Utf8State is managed separately above.
		if pstate != parser.Utf8State {
			pstate = state
		}
	}

	return buf.String()
}

// StringWidth returns the width of a string in cells. This is the number of
// cells that the string will occupy when printed in a terminal. ANSI escape
// codes are ignored and wide characters (such as East Asians and emojis) are
// accounted for.
func StringWidth(s string) int {
	if s == "" {
		return 0
	}

	var (
		gstate  = -1
		pstate  = parser.GroundState // initial state
		cluster string
		width   int
	)

	for i := 0; i < len(s); i++ {
		state, action := parser.Table.Transition(pstate, s[i])
		switch action {
		case parser.PrintAction:
			if utf8ByteLen(s[i]) > 1 {
				var w int
				cluster, _, w, gstate = uniseg.FirstGraphemeClusterInString(s[i:], gstate)
				width += w
				i += len(cluster) - 1
				pstate = parser.GroundState
				continue
			}
			width++
			fallthrough
		default:
			// Reset uniseg state when we're not in a printable state.
			gstate = -1
		}

		pstate = state
	}

	return width
}
