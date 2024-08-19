package ansi

import "strconv"

// SaveCursor (DECSC) is an escape sequence that saves the current cursor
// position.
//
//	ESC 7
//
// See: https://vt100.net/docs/vt510-rm/DECSC.html
const SaveCursor = "\x1b7"

// RestoreCursor (DECRC) is an escape sequence that restores the cursor
// position.
//
//	ESC 8
//
// See: https://vt100.net/docs/vt510-rm/DECRC.html
const RestoreCursor = "\x1b8"

// RequestCursorPosition (CPR) is an escape sequence that requests the current
// cursor position.
//
//	CSI 6 n
//
// The terminal will report the cursor position as a CSI sequence in the
// following format:
//
//	CSI Pl ; Pc R
//
// Where Pl is the line number and Pc is the column number.
// See: https://vt100.net/docs/vt510-rm/CPR.html
const RequestCursorPosition = "\x1b[6n"

// RequestExtendedCursorPosition (DECXCPR) is a sequence for requesting the
// cursor position report including the current page number.
//
//	CSI ? 6 n
//
// The terminal will report the cursor position as a CSI sequence in the
// following format:
//
//	CSI ? Pl ; Pc ; Pp R
//
// Where Pl is the line number, Pc is the column number, and Pp is the page
// number.
// See: https://vt100.net/docs/vt510-rm/DECXCPR.html
const RequestExtendedCursorPosition = "\x1b[?6n"

// CursorUp (CUU) returns a sequence for moving the cursor up n cells.
//
//	CSI n A
//
// See: https://vt100.net/docs/vt510-rm/CUU.html
func CursorUp(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b[" + s + "A"
}

// CursorUp1 is a sequence for moving the cursor up one cell.
//
// This is equivalent to CursorUp(1).
const CursorUp1 = "\x1b[A"

// CursorDown (CUD) returns a sequence for moving the cursor down n cells.
//
//	CSI n B
//
// See: https://vt100.net/docs/vt510-rm/CUD.html
func CursorDown(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b[" + s + "B"
}

// CursorDown1 is a sequence for moving the cursor down one cell.
//
// This is equivalent to CursorDown(1).
const CursorDown1 = "\x1b[B"

// CursorRight (CUF) returns a sequence for moving the cursor right n cells.
//
//	CSI n C
//
// See: https://vt100.net/docs/vt510-rm/CUF.html
func CursorRight(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b[" + s + "C"
}

// CursorRight1 is a sequence for moving the cursor right one cell.
//
// This is equivalent to CursorRight(1).
const CursorRight1 = "\x1b[C"

// CursorLeft (CUB) returns a sequence for moving the cursor left n cells.
//
//	CSI n D
//
// See: https://vt100.net/docs/vt510-rm/CUB.html
func CursorLeft(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b[" + s + "D"
}

// CursorLeft1 is a sequence for moving the cursor left one cell.
//
// This is equivalent to CursorLeft(1).
const CursorLeft1 = "\x1b[D"

// CursorNextLine (CNL) returns a sequence for moving the cursor to the
// beginning of the next line n times.
//
//	CSI n E
//
// See: https://vt100.net/docs/vt510-rm/CNL.html
func CursorNextLine(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b[" + s + "E"
}

// CursorPreviousLine (CPL) returns a sequence for moving the cursor to the
// beginning of the previous line n times.
//
//	CSI n F
//
// See: https://vt100.net/docs/vt510-rm/CPL.html
func CursorPreviousLine(n int) string {
	var s string
	if n > 1 {
		s = strconv.Itoa(n)
	}
	return "\x1b[" + s + "F"
}

// MoveCursor (CUP) returns a sequence for moving the cursor to the given row
// and column.
//
//	CSI n ; m H
//
// See: https://vt100.net/docs/vt510-rm/CUP.html
func MoveCursor(row, col int) string {
	if row < 0 {
		row = 0
	}
	if col < 0 {
		col = 0
	}
	return "\x1b[" + strconv.Itoa(row) + ";" + strconv.Itoa(col) + "H"
}

// MoveCursorOrigin is a sequence for moving the cursor to the upper left
// corner of the screen. This is equivalent to MoveCursor(1, 1).
const MoveCursorOrigin = "\x1b[1;1H"

// SaveCursorPosition (SCP or SCOSC) is a sequence for saving the cursor
// position.
//
//	CSI s
//
// This acts like Save, except the page number where the cursor is located is
// not saved.
//
// See: https://vt100.net/docs/vt510-rm/SCOSC.html
const SaveCursorPosition = "\x1b[s"

// RestoreCursorPosition (RCP or SCORC) is a sequence for restoring the cursor
// position.
//
//	CSI u
//
// This acts like Restore, except the cursor stays on the same page where the
// cursor was saved.
//
// See: https://vt100.net/docs/vt510-rm/SCORC.html
const RestoreCursorPosition = "\x1b[u"
