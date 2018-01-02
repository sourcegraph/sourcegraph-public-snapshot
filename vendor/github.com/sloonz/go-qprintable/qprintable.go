// Copyright 2010 Simon Lipp.
// Distributed under a BSD-like license. See COPYING for more
// details

// Package qprintable implements quoted-printable encoding as specified
// by RFC 2045. It is strict on ouput, generous on input.
//
// Quoting RFC 2045:
//  The Quoted-Printable encoding is intended to represent data that
//  largely consists of octets that correspond to printable characters in
//  the US-ASCII character set.  It encodes the data in such a way that
//  the resulting octets are unlikely to be modified by mail transport.
//  If the data being encoded are mostly US-ASCII text, the encoded form
//  of the data remains largely recognizable by humans.  A body which is
//  entirely US-ASCII may also be encoded in Quoted-Printable to ensure
//  the integrity of the data should the message pass through a
//  character-translating, and/or line-wrapping gateway.
package qprintable

import (
	"bytes"
	"io"
	"strings"
)

const maxLineSize = 76
const hexTable = "0123456789ABCDEF"

/*
 * Encodings
 */

type Encoding struct {
	isText    bool
	nativeEol string
}

// A text encoding has to convert its input in the canonical form (as
// defined by RFC 2045) : native ends of line (CR for MacTextEncoding,
// LF for UnixTextEncoding, CRLF for WindowsTextEncoding) are converted
// into CRLF sequences. Non-native EOL sequences (for example, CR on
// UnixTextEncoding) are treated as control characters and escaped.
//
// In the decoding process, CRLF sequences are converted to native ends of
// line.
var MacTextEncoding = &Encoding{true, "\r"}
var UnixTextEncoding = &Encoding{true, "\n"}
var WindowsTextEncoding = &Encoding{true, "\r\n"}

// In binary encoding, CR and LF characters are treated like other control
// characters sequence and are escaped.
var BinaryEncoding = &Encoding{false, ""}

// Try to detect encoding of string:
// strings with no \r will be Unix,
// strings with \r and no \n will be Mac,
// strings with count(\r\n) == count(\r) == count(\n) will be Windows,
// other strings will be binary
func DetectEncoding(data string) *Encoding {
	if strings.Count(data, "\r") == 0 {
		return UnixTextEncoding
	} else if strings.Count(data, "\n") == 0 {
		return MacTextEncoding
	} else if strings.Count(data, "\r") == strings.Count(data, "\n") && strings.Count(data, "\r\n") == strings.Count(data, "\n") {
		return WindowsTextEncoding
	}
	return BinaryEncoding
}

/*
 * Encoder
 */

type encoder struct {
	eol      string
	enc      *Encoding
	w        io.Writer
	lineSize int
	wasCR    bool
}

// Return first character position where the character has to be escaped
func (e *encoder) nextSpecialChar(p []byte) (i int) {
	for i = 0; i < len(p); i++ {
		// ASCII 32-126 (printable) + '\t' - '=' can appear in the qprintable stream
		if !((p[i] >= 32 && p[i] <= 126 && p[i] != byte('=')) || p[i] == byte('\t')) {
			return i
		}
	}
	return i
}

func (e *encoder) writeAndWrap(p []byte, atomic bool) (err error) {
	// -1 is to keep enough size for the trailing =
	for e.lineSize+len(p) > (maxLineSize - 1) {
		if !atomic {
			wSize := (maxLineSize - 1) - e.lineSize
			if _, err := e.w.Write(p[:wSize]); err != nil {
				return err
			}
			p = p[wSize:]
		}
		if _, err = e.w.Write([]byte("=" + e.eol)); err != nil {
			return err
		}
		e.lineSize = 0
	}
	if _, err = e.w.Write(p); err == nil {
		e.lineSize += len(p)
	}
	return err
}

func (e *encoder) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	var pos int

	if e.wasCR {
		if p[0] == '\n' {
			// EOL: consume LF and then convert it to CRLF
			if _, err = e.w.Write([]byte(e.eol)); err == nil {
				n++
				p = p[1:]
				e.lineSize = 0
			}
		} else {
			// CR + anything on windows means that we must escape CR
			// and then emit b[0] normally in the loop below
			if err = e.writeAndWrap([]byte("=0D"), true); err != nil {
				return n, err
			}
		}
		e.wasCR = false
	}

	for pos = e.nextSpecialChar(p); pos < len(p); pos = e.nextSpecialChar(p) {
		// Write ascii chars
		if err = e.writeAndWrap(p[:pos], false); err != nil {
			return n, err
		}
		n += pos

		// Hande special char
		if e.enc.isText && p[pos] == byte('\r') && e.enc.nativeEol == "\r\n" {
			// CR on windows
			e.wasCR = true
			n++
			if pos+1 < len(p) {
				nn, err := e.Write(p[pos+1:])
				n += nn
				return n, err
			}
		} else if e.enc.isText && p[pos] == e.enc.nativeEol[0] {
			// Other EOL
			if _, err = e.w.Write([]byte(e.eol)); err != nil {
				return n, err
			}
			e.lineSize = 0
			n++
		} else {
			// Control char
			if err = e.writeAndWrap([]byte{'=', hexTable[p[pos]>>4], hexTable[p[pos]&0xf]}, true); err != nil {
				return n, err
			}
			n++
		}

		// Consume printed chars
		if pos+1 < len(p) {
			p = p[pos+1:]
		} else {
			p = nil
		}
	}

	// Non-consumed data can be directly written
	if p != nil {
		if err = e.writeAndWrap(p, false); err != nil {
			return n, err
		}
		n += len(p)
	}

	return n, err
}

func (e *encoder) Close() error {
	if e.wasCR {
		if err := e.writeAndWrap([]byte("=0D"), true); err != nil {
			return err
		}
		e.wasCR = false
	}
	return nil
}

// Returns a new encoder. Any data passed to Write will be encoded
// according to enc and then written to w.
//
// Data passed to Write must be in its canonical form. The canonical
// form depends on the encoding:
//
// for binary encoding, anything goes.
//
// for text encodings, there shouldn't be any CR or LF characters
// other than the one used for end-of-line representation, that is,
// LF on Unix, CR on old Mac, CR+LF on Windows.
//
// It is the responsibility of the caller to ensure that the input stream
// is in its canonical form. Any CR of LF character which is not part of
// an end-of-line representation will be quoted.
//
// This returns a WriteCloser, but Close has no effect for encoding other
// than WindowsEncoding.
//
// For WindowsEncoding, any trailing CR will not be written unless
// you call this function. However, note that for a text conforming
// to windows canonical form, this should never happen. So this
// function is useful only for invalid WindowsEncoding text streams,
// you can safely ignore it in all other cases.
func NewEncoder(enc *Encoding, w io.Writer) io.WriteCloser {
	return &encoder{eol: "\r\n", enc: enc, w: w}
}

// Returns an encoder where the line-endings of the resulting stream
// is not the standard value (CRLF)
//
// Standard requires CRLF line endings, but there are some variants
// out there (like Maildir) which requires LF line endings.
func NewEncoderWithEOL(eol string, enc *Encoding, w io.Writer) io.WriteCloser {
	return &encoder{eol: eol, enc: enc, w: w}
}

/*
 * Decoder
 */

type decoder struct {
	enc            *Encoding
	r              io.Reader
	buf, leftovers *bytes.Buffer
}

func isHex(b byte) bool {
	return (b >= byte('0') && b <= byte('9')) ||
		(b >= byte('a') && b <= byte('f')) ||
		(b >= byte('A') && b <= byte('F'))
}

func hexValue(b byte) int {
	if b >= byte('0') && b <= byte('9') {
		return int(b - byte('0'))
	}
	if b >= byte('A') && b <= byte('F') {
		return 10 + int(b-byte('A'))
	}
	return 10 + int(b-byte('a'))
}

func isSpace(b byte) bool {
	return b == byte(' ') || b == byte('\t')
}

// Return first character position where the character has to be decoded
func (d *decoder) nextSpecialChar(p []byte) (i int) {
	for i = 0; i < len(p); i++ {
		if p[i] == byte('=') || (d.enc.isText && p[i] == byte('\r')) {
			return i
		}
	}
	return i
}

func (d *decoder) handleLeftovers(rawData []byte) []byte {
	// We have 5 different possible situations for leftovers
	//  - "="
	//  - "\r"
	//  - "=(hex character)"
	//  - "=(spaces)" with spaces = [\t ]+
	//  - "=\r"

	consume := func(rawData []byte) []byte {
		if len(rawData) > 1 {
			return rawData[1:]
		}
		return nil
	}

	// First, handle the first situation
	if d.leftovers.Bytes()[0] == byte('=') && d.leftovers.Len() == 1 {
		if isHex(rawData[0]) || isSpace(rawData[0]) || rawData[0] == byte('\r') {
			// Fall back to one of the last three situations
			d.leftovers.WriteByte(rawData[0])
			return consume(rawData)
		} else if rawData[0] == '\n' {
			// (ill-formed) soft line break, just discard leftovers
			d.leftovers.Truncate(0)
			return consume(rawData)
		} else {
			// non-escaped "=" sign, just add leftover to buffer
			d.buf.WriteByte(byte('='))
			d.leftovers.Truncate(0)
			return rawData
		}
	}

	// Handle "\r"
	if d.leftovers.Bytes()[0] == byte('\r') {
		if rawData[0] == byte('\n') {
			d.buf.WriteString(d.enc.nativeEol)
			d.leftovers.Truncate(0)
			return consume(rawData)
		} else {
			d.buf.WriteByte(byte('\r'))
			d.leftovers.Truncate(0)
			return rawData
		}
	}

	// Handle "=(hex character)"
	if isHex(d.leftovers.Bytes()[1]) {
		if isHex(rawData[0]) {
			var hexVal byte = byte(hexValue(rawData[0]) + (hexValue(d.leftovers.Bytes()[1]) << 4))
			d.buf.WriteByte(hexVal)
			d.leftovers.Truncate(0)
			return consume(rawData)
		} else {
			// ill-formed stream, but hard choice here; should =1x be treated like =01x or =3D1x ?
			// the second one is easier to implement, use it
			d.buf.Write(d.leftovers.Bytes())
			d.leftovers.Truncate(0)
			return rawData
		}
	}

	// Handle a =(space) sequence. This can appear in two occasions:
	//  - well formed streams, when trailing spaces after a soft line break must be ignored
	//    (this is the ugliest part of RFC 2045)
	//  - ill-formed streams when the "=" sign has not been escaped
	// We have to consume all space characters to make a choice : if the first non-space character
	// is a \r or \n (to deal with ill-formed streams where EOL is not CRLF) it's the first case,
	// otherwise it's the second
	if isSpace(d.leftovers.Bytes()[1]) {
		var i int
		for i = 0; i < len(rawData) && isSpace(rawData[i]); i++ {
		}
		if i == len(rawData) {
			// Couldn't decide (long sequence of spaces), put buf in leftovers and return
			d.leftovers.Write(rawData)
			return nil
		} else if rawData[i] == byte('\r') {
			// First case; replace leftovers with =\r, it will be handled more thoroughly later
			d.leftovers.Truncate(0)
			d.leftovers.WriteString("=\r")
			return consume(rawData[i:])
		} else if byte(rawData[i]) == '\n' {
			// First case with ill-encoded stream, just discard the whole
			d.leftovers.Truncate(0)
			return consume(rawData[i:])
		} else {
			// Second case, just add leftovers to buffer
			d.buf.Write(d.leftovers.Bytes())
			d.leftovers.Truncate(0)
			return rawData
		}
	}

	// Handle "=\r". We know it's a soft line break, we just have to consume any following LF
	d.leftovers.Truncate(0)
	if rawData[0] == byte('\n') {
		return consume(rawData)
	}
	return rawData
}

func (d *decoder) parseBytes(rawData []byte) {
	for d.leftovers.Len() > 0 && len(rawData) > 0 {
		rawData = d.handleLeftovers(rawData)
	}

	if len(rawData) > 0 {
		pos := d.nextSpecialChar(rawData)
		d.buf.Write(rawData[:pos])
		if pos < len(rawData) {
			rawData = rawData[pos:]
			d.leftovers.WriteByte(rawData[0])
			if len(rawData) > 1 {
				d.parseBytes(rawData[1:])
			}
		}
	}
}

func (d *decoder) Read(p []byte) (n int, err error) {
	var read int
	canContinue := true
	for n < len(p) && canContinue {
		if (len(p) - n) > d.buf.Len() {
			rawData := make([]byte, 1024)
			read, err = d.r.Read(rawData)
			if read < 1024 || err != nil {
				canContinue = false
			}
			if err == io.EOF && read == 0 && d.leftovers.Len() > 0 {
				// Underlying Reader is exhausted and there's still data in leftovers
				// This can't happen in well-formed streams. For ill-formed streams :
				//  - if leftovers = "=\r", "=(spaces)", just discard it
				//  - if leftovers = "=", "=(hex)", "\r" and add it to buffer
				if d.leftovers.Len() == 2 && (isSpace(d.leftovers.Bytes()[1]) || d.leftovers.Bytes()[1] == byte('\r')) {
					d.leftovers.Truncate(0)
				} else {
					d.buf.Write(d.leftovers.Bytes())
					d.leftovers.Truncate(0)
				}
			} else {
				d.parseBytes(rawData[:read])
			}
		}
		copied, _ := d.buf.Read(p[n:])
		n += copied
	}
	return n, err
}

// Returns a new decoder. Data will be read from r, and decoded
// according to enc.
func NewDecoder(enc *Encoding, r io.Reader) io.Reader {
	return &decoder{enc: enc, r: r,
		buf:       bytes.NewBuffer(nil),
		leftovers: bytes.NewBuffer(nil)}
}
