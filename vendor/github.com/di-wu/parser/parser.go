package parser

import (
	"github.com/di-wu/parser/op"
	"unicode/utf8"
)

// EOD indicates the End Of (the) Data.
const EOD = 1<<31 - 1

// Parser represents a general purpose parser.
type Parser struct {
	buffer []byte
	cursor *Cursor
	decode func([]byte) (rune, int)

	converter func(interface{}) interface{}
	operator  func(interface{}) (*Cursor, error)
}

// New creates a new Parser.
func New(input []byte) (*Parser, error) {
	p := Parser{
		buffer: input,
		decode: utf8.DecodeRune,
	}

	current, size := p.decode(p.buffer)
	if size == 0 {
		// Nothing got decoded.
		return nil, &InitError{
			Message: "failed to scan the first rune",
		}
	}

	p.cursor = &Cursor{
		Rune: current,
		size: size,
	}
	return &p, nil
}

// DecodeRune allows you to redefine the way runes are decoded form the byte
// stream. By default utf8.DecodeRune is used.
func (p *Parser) DecodeRune(d func(p []byte) (rune, int)) {
	p.decode = d
}

// SetConverter allows you to add additional (prioritized) converters to the
// parser. e.g. convert aliases to other types or overwrite defaults.
func (p *Parser) SetConverter(c func(i interface{}) interface{}) {
	p.converter = c
}

// SetOperator allows you to support additional (prioritized) operators.
// Should return an UnsupportedType error if the given value is not supported.
func (p *Parser) SetOperator(o func(i interface{}) (*Cursor, error)) {
	p.operator = o
}

// Next advances the parser by one rune.
func (p *Parser) Next() *Parser {
	if p.Done() {
		return p
	}

	// Move position to the next rune.
	// |__|_| < next rune is position + size
	//  ^
	//  rune of size 2, position 0
	p.cursor.position += p.cursor.size

	current, size := p.decode(p.buffer[p.cursor.position:])
	if size == 0 {
		// Nothing got decoded.
		current = EOD
	}

	// Previous rune was an end of line, we are on a new line now.
	if p.cursor.Rune == '\n' || (p.cursor.Rune == '\r' && current != '\n') {
		p.cursor.row += 1
		p.cursor.column = 0
	} else {
		p.cursor.column += p.cursor.size
	}

	p.cursor.Rune = current
	p.cursor.size = size

	return p
}

// Current returns the value to which the cursor is pointing at.
func (p *Parser) Current() rune {
	return p.cursor.Rune
}

// Done checks whether the parser is done parsing.
func (p *Parser) Done() bool {
	return p.cursor.Rune == EOD
}

// Mark returns a copy of the current cursor.
func (p *Parser) Mark() *Cursor {
	mark := *p.cursor
	return &mark
}

// LookBack returns the previous cursor without decreasing the parser.
func (p *Parser) LookBack() *Cursor {
	if p.cursor.position == 0 || p.Done() {
		// Not possible to go back
		return p.Mark()
	}

	// We don't know the size of the previous rune... 1 or more?
	previous, size := p.decode(p.buffer[p.cursor.position-1:])
	for i := 2; previous == utf8.RuneError; i++ {
		previous, size = p.decode(p.buffer[p.cursor.position-i:])
	}

	var (
		row    = p.cursor.row
		column = p.cursor.column
	)
	if p.cursor.Rune == '\n' || (p.cursor.Rune == '\r' && p.Peek().Rune != '\n') {
		row -= 1
		column = 0
	} else {
		column -= size
	}

	return &Cursor{
		Rune:     previous,
		size:     size,
		position: p.cursor.position - size,
		row:      row,
		column:   column,
	}
}

// Peek returns the next cursor without advancing the parser.
func (p *Parser) Peek() *Cursor {
	start := p.Mark()
	defer p.Jump(start)
	return p.Next().Mark()
}

// Jump goes to the position of the given mark.
func (p *Parser) Jump(mark *Cursor) *Parser {
	cursor := *mark
	p.cursor = &cursor
	return p
}

// Slice returns the value in between the two given cursors [start:end]. The end
// value is inclusive!
func (p *Parser) Slice(start *Cursor, end *Cursor) string {
	if start.Rune == EOD {
		return ""
	}
	if end == nil { // Just to be sure...
		end = start
	}
	return string(p.buffer[start.position : end.position+end.size])
}

// Expect checks whether the buffer contains the given value. It consumes their
// corresponding runes and returns a mark to the last rune of the consumed
// value. It returns an error if can not find a match with the given value.
//
// It currently supports:
//	- rune & string
//	- func(p *Parser) (*Cursor, bool)
//	  (== AnonymousClass)
//	- []interface{}
//	  (== op.And)
//	- operators: op.Not, op.And, op.Or & op.XOr
func (p *Parser) Expect(i interface{}) (*Cursor, error) {
	state := state{p: p}

	i = ConvertAliases(i)
	if p.converter != nil {
		// Can undo previous conversions!
		i = p.converter(i)
	}

	if p.operator != nil {
		// Takes priority over default values. If an unsupported error is
		// returned we can check if one of the predefined types match.
		mark, err := p.operator(i)
		if _, ok := err.(*UnsupportedType); !ok {
			return mark, err
		}
	}
	switch start := p.Mark(); v := i.(type) {
	case rune:
		if p.cursor.Rune != v {
			return nil, p.ExpectedParseError(v, start, start)
		}
		state.Ok(p.Mark())
	case string:
		if v == "" {
			return nil, &ExpectError{
				Message: "can not parse empty string",
			}
		}
		for _, r := range []rune(v) {
			if p.cursor.Rune != r {
				return nil, p.ExpectedParseError(v, start, p.Mark())
			}
			state.Ok(p.Mark())
		}

	case AnonymousClass:
		last, passed := v(p)
		if !passed {
			if last == nil {
				last = start
			}
			return nil, p.ExpectedParseError(v, start, p.Jump(last).Peek())
		}
		state.Ok(last)

	case op.Not:
		defer p.Jump(start)
		if last, err := p.Expect(v.Value); err == nil {
			return nil, p.ExpectedParseError(v, start, last)
		}
	case op.Ensure:
		if last, err := p.Expect(v.Value); err != nil {
			return last, err
		}
		p.Jump(start)
	case op.And:
		var last *Cursor
		for _, i := range v {
			mark, err := p.Expect(i)
			if err != nil {
				if last == nil {
					last = start
				}
				return nil, p.ExpectedParseError(v, start, p.Jump(last).Peek())
			}
			last = mark
		}
		state.Ok(last)
	case op.Or:
		var last *Cursor
		for _, i := range v {
			mark, err := p.Expect(i)
			if err == nil {
				last = mark
				break
			}
		}
		if last == nil {
			return nil, p.ExpectedParseError(v, start, start)
		}
		state.Ok(last)
	case op.XOr:
		var last *Cursor
		for _, i := range v {
			mark, err := p.Expect(i)
			if err == nil {
				if last != nil {
					p.Jump(start)
					return nil, p.ExpectedParseError(v, start, mark)
				}
				last = mark
				p.Jump(start) // Go back to the start.
			}
		}
		if last == nil {
			return nil, p.ExpectedParseError(v, start, last)
		}
		state.Ok(last)

	case op.Range:
		var (
			count int
			last  *Cursor
		)
		for {
			mark, err := p.Expect(v.Value)
			if err != nil {
				break
			}
			last = mark
			count++

			if v.Max != -1 && count == v.Max {
				// Break if you have parsed the maximum amount of values.
				// This way count will never be larger than v.Max.
				break
			}
		}
		if count < v.Min {
			return nil, p.ExpectedParseError(v, start, last)
		}
		state.Ok(last)

	default:
		return nil, &UnsupportedType{
			Value: i,
		}
	}
	return state.End(), nil
}

// Check works the same as Parser.Expect, but instead it returns a bool instead
// of an error.
func (p *Parser) Check(i interface{}) (*Cursor, bool) {
	mark, err := p.Expect(i)
	if err != nil {
		return mark, false
	}
	return mark, true
}

// ConvertAliases converts various default primitive types to aliases for type
// matching.
//
// - (int, rune)
// - ([]interface{}, op.And)
func ConvertAliases(i interface{}) interface{} {
	switch v := i.(type) {
	case int:
		return rune(v)

	case func(p *Parser) (*Cursor, bool):
		return AnonymousClass(v)
	case Class:
		return AnonymousClass(v.Check)

	case []interface{}:
		return op.And(v)

	default:
		return i
	}
}
