package ansihtml

import (
	"errors"
	"io"
)

// Parser parses ANSI-encoded console output from an io.Reader.
type Parser struct {
	in  io.Reader
	out io.Writer

	parserState
}

// NewParser creates a Parser which reads from rd and writes output with escape
// sequences removed to w.
func NewParser(rd io.Reader, w io.Writer) *Parser {
	p := &Parser{in: rd, out: w}
	p.utf8Escapes = true
	return p
}

// Parse reads from the io.Reader, calling escapeHandler with any parsed
// ANSI escape sequences and writing normal output to the io.Writer, until
// either EOF is reached or an error occurs.
//
// Writes to w and calls to escapeHandler are done in the same order as data is
// read from the io.Reader, meaning escapeHandler can write to the io.Writer to
// insert text formatting data as necessary.
//
// escapeHandler takes the finalByte and intermediateBytes from the escape
// sequence and any parameterBytes from after the escape sequence as
// parameters.  For example, the escape sequence '\x1b[0;33m' will result in
// escapeHandler being called with finalByte '[', intermediateBytes '', and
// parameterBytes '0;33m'.
//
// intermediateBytes is rarely present in ANSI escape sequences, with one
// example being the switching between JIS encodings done by ISO-2022-JP.
func (p *Parser) Parse(escapeHandler func(finalByte byte, intermediateBytes, parameterBytes []byte) error) error {
	buf := make([]byte, 4096)
	return p.ParseBuffer(buf, escapeHandler)
}

// ParseBuffer performs the same action as Parse, but with a caller-supplied
// buffer for copying data from the reader to the writer.
func (p *Parser) ParseBuffer(buf []byte, escapeHandler func(finalByte byte, intermediateBytes, parameterBytes []byte) error) error {
	if len(buf) == 0 {
		return errors.New("buffer must not be empty")
	}

	var start, ofs, i int
	var werr error
	p.escapeHandler = func(finalByte byte, intermediateBytes, parameterBytes []byte) error {
		if werr == nil {
			_, werr = p.out.Write(buf[start : i-ofs])
		}
		start = i - ofs
		if escapeHandler == nil {
			return nil
		}
		return escapeHandler(finalByte, intermediateBytes, parameterBytes)
	}

	for {
		n, err := p.in.Read(buf)
		start = 0
		ofs = 0
		werr = nil
		for i = 0; i < n; i++ {
			output, herr := p.handle(buf[i])
			if herr != nil {
				return herr
			}
			if !output {
				ofs++
			} else {
				if p.extraByte != 0 {
					if ofs == 0 {
						// there's not room in the buffer to put the last byte,
						// so write the single byte.
						var wbuf [1]byte
						wbuf[0] = p.extraByte
						if werr == nil {
							_, werr = p.out.Write(wbuf[:])
						}
					} else {
						buf[i-ofs] = p.extraByte
						ofs--
					}
					p.extraByte = 0
				}
				if ofs > 0 {
					buf[i-ofs] = buf[i]
				}
			}
		}
		if start <= n-ofs && werr == nil {
			_, werr = p.out.Write(buf[start : n-ofs])
		}
		if werr != nil {
			return werr
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

const escape byte = '\x1b'

type readEscapeState uint8

const (
	// parser is not inside an escape sequence
	readNormal readEscapeState = iota

	// parser has read \x1b; next byte is intermediate or final byte
	readEscape

	// parser has read final byte; next byte is parameter data
	readEscapeParams
)

// maximum number of bytes in an escape sequence, not including the initial \x1b
const escapeSequenceMaxLength = 255

// maximum number of parameter bytes in an escape sequence
const escapeSequenceParamsMaxLength = 255

type parserState struct {
	// bytes in the input stream after \x1b
	seqBuf  [escapeSequenceMaxLength]byte
	seqBufI uint8

	state readEscapeState

	// parameter bytes for escape sequences which have params
	paramsBuf  [escapeSequenceParamsMaxLength]byte
	paramsBufI uint8

	previousByte byte

	// if true, C1 control codes may be encoded as utf-8 codepoints
	utf8Escapes bool

	// utf8escapes can lead to the start of an escape sequence being removed
	// from output incorrectly; extraByte will be set to this value if that
	// happens
	extraByte byte

	// will be called when an escape sequence is parsed
	escapeHandler func(finalByte byte, intermediateBytes, parameterBytes []byte) error
}

// handle a byte and return whether the byte should go to output
func (s *parserState) handle(b byte) (bool, error) {
	previousByte := s.previousByte
	s.previousByte = b
	var handlerError error
	switch s.state {
	default: // readNormal
		if b == escape || (s.utf8Escapes && b == 0xc2) {
			s.state = readEscape
			return false, handlerError
		}
		return true, handlerError
	case readEscape:
		if s.utf8Escapes && previousByte == 0xc2 {
			if b >= 0x80 && b <= 0x9f {
				s.seqBuf[0] = b - 0x40
				s.seqBufI = 1
				s.paramsBufI = 0
				if s.hasParams(b - 0x40) {
					s.state = readEscapeParams
				} else {
					handlerError = s.handleEscape()
					s.state = readNormal
				}
				return false, handlerError
			}

			s.extraByte = previousByte
			s.state = readNormal
			return true, handlerError
		}
		// intermediate or final byte
		if b >= 0x20 && b <= 0x7e {
			if s.seqBufI < escapeSequenceMaxLength {
				s.seqBuf[s.seqBufI] = b
				s.seqBufI++
			}
			// final byte
			if b >= 0x30 {
				s.paramsBufI = 0
				if s.hasParams(b) {
					s.state = readEscapeParams
				} else {
					handlerError = s.handleEscape()
					s.state = readNormal
				}
			}
		} else {
			// unknown sequence; swallow the sequence but resume normal output
			s.state = readNormal
			s.seqBufI = 0
		}
		return false, handlerError
	case readEscapeParams:
		var finalByte byte
		if s.seqBufI > 0 {
			finalByte = s.seqBuf[s.seqBufI-1]
		}
		if s.paramsBufI < escapeSequenceParamsMaxLength {
			s.paramsBuf[s.paramsBufI] = b
			s.paramsBufI++
		}
		switch finalByte {
		case '[': // CSI
			if s.paramsBufI == 1 || (previousByte >= 0x30 && previousByte <= 0x3f) {
				if !(b >= 0x20 && b <= 0x7e) {
					// invalid parameters
					s.state = readNormal
					s.seqBufI = 0
					return false, handlerError
				}
			} else if previousByte >= 0x20 && previousByte <= 0x2f {
				if !((b >= 0x20 && b <= 0x2f) || (b >= 0x40 && b <= 0x7e)) {
					// invalid parameters
					s.state = readNormal
					s.seqBufI = 0
					return false, handlerError
				}
			}
			if b >= 0x40 && b <= 0x7e {
				handlerError = s.handleEscape()
				s.state = readNormal
			}
		default: // ST-terminated
			if (previousByte == escape && b == '\\') ||
				(s.utf8Escapes && previousByte == 0xc2 && b == 0x9c) ||
				// if an ST-terminated sequence is too long, just truncate it
				s.paramsBufI == escapeSequenceParamsMaxLength ||
				// allow xterm BEL-terminated OSC
				(finalByte == ']' && b == '\x07') {

				handlerError = s.handleEscape()
				s.state = readNormal
				s.seqBufI = 0
			}
		}
		return false, handlerError
	}
}

func (s *parserState) hasParams(b byte) bool {
	switch b {
	case '[', 'P', 'X', '^', '_', ']':
		return true
	default:
		return false
	}
}

func (s *parserState) handleEscape() error {
	// seqBuf contains the escape sequence with \x1b, any intermediate bytes,
	// and the final byte
	// paramsBuf contains parameter bytes for e.g. CSI
	seqBufI := s.seqBufI
	s.seqBufI = 0
	paramsBufI := s.paramsBufI
	s.paramsBufI = 0
	if s.escapeHandler == nil {
		return nil
	}
	var finalByte byte
	if seqBufI > 0 {
		finalByte = s.seqBuf[seqBufI-1]
	}
	var intermediateBytes []byte
	if seqBufI > 1 {
		intermediateBytes = s.seqBuf[:seqBufI-1]
	}
	var parameterBytes []byte
	if paramsBufI > 0 {
		parameterBytes = s.paramsBuf[:paramsBufI]
	}
	return s.escapeHandler(finalByte, intermediateBytes, parameterBytes)
}
