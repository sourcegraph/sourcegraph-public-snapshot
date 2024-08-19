package parser

import "fmt"

// Cursor allows you to record your current position so you can return to it
// later. Keeps track of its own position in the buffer of the parser.
type Cursor struct {
	// Rune is the value that the cursor points at.
	Rune rune
	// The length of the current rune in bytes.
	size int

	// The start position of the current rune.
	position int
	// The row and column of the current rune, NOT in bytes!
	row, column int
}

// Position returns the row and column of the cursors location.
func (c *Cursor) Position() (int, int) {
	return c.row, c.column
}

func (c *Cursor) String() string {
	return fmt.Sprintf("%U: %c", c.Rune, c.Rune)
}

// state manages the state of a parser. It contains a pointer to the last
// successfully parsed rune.
type state struct {
	end *Cursor
	p   *Parser
}

// Ok updates the end of the state and jumps to the next rune.
func (s *state) Ok(last *Cursor) {
	if last == nil {
		// Optional values have no last mark.
		return
	}
	s.end = last
	// We jump to the given cursor (last parsed rune) because it is not
	// guaranteed that the already parser did not pass it.
	s.p.Jump(last).Next()
}

// End returns a mark to the last successfully parsed rune.
func (s *state) End() *Cursor {
	return s.end
}
