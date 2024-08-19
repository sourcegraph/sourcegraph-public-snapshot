package input

import (
	"bytes"
	"io"
	"unicode/utf8"

	"github.com/erikgeiser/coninput"
	"github.com/muesli/cancelreader"
)

// Driver represents an ANSI terminal input Driver.
// It reads input events and parses ANSI sequences from the terminal input
// buffer.
type Driver struct {
	rd    cancelreader.CancelReader
	table map[string]Key // table is a lookup table for key sequences.

	term string // term is the terminal name $TERM.

	// paste is the bracketed paste mode buffer.
	// When nil, bracketed paste mode is disabled.
	paste []byte

	buf [256]byte // do we need a larger buffer?

	// prevMouseState keeps track of the previous mouse state to determine mouse
	// up button events.
	prevMouseState coninput.ButtonState // nolint: unused

	flags int // control the behavior of the driver.
}

// NewDriver returns a new ANSI input driver.
// This driver uses ANSI control codes compatible with VT100/VT200 terminals,
// and XTerm. It supports reading Terminfo databases to overwrite the default
// key sequences.
func NewDriver(r io.Reader, term string, flags int) (*Driver, error) {
	d := new(Driver)
	cr, err := newCancelreader(r)
	if err != nil {
		return nil, err
	}

	d.rd = cr
	d.table = buildKeysTable(flags, term)
	d.term = term
	d.flags = flags
	return d, nil
}

// Cancel cancels the underlying reader.
func (d *Driver) Cancel() bool {
	return d.rd.Cancel()
}

// Close closes the underlying reader.
func (d *Driver) Close() error {
	return d.rd.Close()
}

func (d *Driver) readEvents() (e []Event, err error) {
	nb, err := d.rd.Read(d.buf[:])
	if err != nil {
		return nil, err
	}

	buf := d.buf[:nb]

	// Lookup table first
	if bytes.HasPrefix(buf, []byte{'\x1b'}) {
		if k, ok := d.table[string(buf)]; ok {
			e = append(e, KeyDownEvent(k))
			return
		}
	}

	var i int
	for i < len(buf) {
		nb, ev := ParseSequence(buf[i:])

		// Handle bracketed-paste
		if d.paste != nil {
			if _, ok := ev.(PasteEndEvent); !ok {
				d.paste = append(d.paste, buf[i])
				i++
				continue
			}
		}

		switch ev.(type) {
		case UnknownCsiEvent, UnknownSs3Event, UnknownEvent:
			// If the sequence is not recognized by the parser, try looking it up.
			if k, ok := d.table[string(buf[i:i+nb])]; ok {
				ev = KeyDownEvent(k)
			}
		case PasteStartEvent:
			d.paste = []byte{}
		case PasteEndEvent:
			// Decode the captured data into runes.
			var paste []rune
			for len(d.paste) > 0 {
				r, w := utf8.DecodeRune(d.paste)
				if r != utf8.RuneError {
					paste = append(paste, r)
				}
				d.paste = d.paste[w:]
			}
			d.paste = nil // reset the buffer
			e = append(e, PasteEvent(paste))
		case nil:
			i++
			continue
		}

		if mevs, ok := ev.(MultiEvent); ok {
			e = append(e, []Event(mevs)...)
		} else {
			e = append(e, ev)
		}
		i += nb
	}

	return
}
