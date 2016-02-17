// Package pktline contains a SplitFunc to be used with a Scanner that scans a
// stream for git's pkt-lines. The pkt-line format as specified in git's docs:
//
// 	pkt-line Format
// 	---------------
//
// 	Much (but not all) of the payload is described around pkt-lines.
//
// 	A pkt-line is a variable length binary string.  The first four bytes
// 	of the line, the pkt-len, indicates the total length of the line,
// 	in hexadecimal.  The pkt-len includes the 4 bytes used to contain
// 	the length's hexadecimal representation.
//
// 	A pkt-line MAY contain binary data, so implementors MUST ensure
// 	pkt-line parsing/formatting routines are 8-bit clean.
//
// 	A non-binary line SHOULD BE terminated by an LF, which if present
// 	MUST be included in the total length. Receivers MUST treat pkt-lines
// 	with non-binary data the same whether or not they contain the trailing
// 	LF (stripping the LF if present, and not complaining when it is
// 	missing).
//
// 	The maximum length of a pkt-line's data component is 65520 bytes.
// 	Implementations MUST NOT send pkt-line whose length exceeds 65524
// 	(65520 bytes of payload + 4 bytes of length data).
//
// 	Implementations SHOULD NOT send an empty pkt-line ("0004").
//
// 	A pkt-line with a length field of 0 ("0000"), called a flush-pkt,
// 	is a special case and MUST be handled differently than an empty
// 	pkt-line ("0004").
//
// 	----
// 	  pkt-line     =  data-pkt / flush-pkt
//
// 	  data-pkt     =  pkt-len pkt-payload
// 	  pkt-len      =  4*(HEXDIG)
// 	  pkt-payload  =  (pkt-len - 4)*(OCTET)
//
// 	  flush-pkt    = "0000"
// 	----
//
// 	Examples (as C-style strings):
//
// 	----
// 	  pkt-line          actual value
// 	  ---------------------------------
// 	  "0006a\n"         "a\n"
// 	  "0005a"           "a"
// 	  "000bfoobar\n"    "foobar\n"
// 	  "0004"            ""
// 	----
package pktline

import (
	"fmt"
	"strconv"
	"strings"
)

// parseLength assumes the passed data starts with a pkt-line.
// Returns an error if less than 4 bytes or contains non-hex characters in the
// first 4 bytes.
func parseLength(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("pkt-line is %d bytes must be at least 4 bytes", len(data))
	}
	data = data[:4]
	length, err := strconv.ParseInt(string(data), 16, 16)
	if err != nil {
		return 0, fmt.Errorf("failed to scan pkt-line length from %q", string(data))
	}
	return int(length), nil
}

// SplitFunc is to be used in a scanner to scan one packet line at a time.
func SplitFunc(data []byte, atEOF bool) (int, []byte, error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if len(data) < 4 {
		if atEOF {
			return 0, nil, fmt.Errorf("reached eof without whole pkt-line")
		}
		return 0, nil, nil
	}
	n, err := parseLength(data[:4])
	if err != nil {
		return 0, nil, err
	}
	if n == 0 {
		return 4, data[:4], nil // flush pkt-line
	}
	if len(data) < n {
		if atEOF {
			return 0, nil, fmt.Errorf("reached eof without whole pkt-line")
		}
		return 0, nil, nil
	}
	return n, data[:n], nil
}

// IsComment returns whether the pkt-line's payload is a comment.
func IsComment(data []byte) bool {
	return len(data) > 4 && data[4] == []byte("#")[0]
}

// IsFlush returns whether the pkt-line's payload is a flush.
func IsFlush(data []byte) bool {
	return len(data) == 4 && string(data) == "0000"
}

// HasPrefix returns whether the pkt-line payload starts with the passed in data.
// Does not account for a possible control byte identifying the band.
func HasPrefix(data, prefix []byte) bool {
	return len(data) > 3+len(prefix) && strings.HasPrefix(string(data[4:]), string(prefix))
}

const zeroID = "0000000000000000000000000000000000000000"

// IsCreate returns whether the pkt-line contains a create command.
func IsCreate(data []byte) bool {
	fields := strings.Fields(string(data[4:]))
	if len(fields) < 3 {
		return false
	}
	return fields[0] == zeroID && len(fields[1]) == 40 && fields[1] != zeroID
}

// IsDelete returns whether the pkt-line contains a delete command.
func IsDelete(data []byte) bool {
	fields := strings.Fields(string(data[4:]))
	if len(fields) < 3 {
		return false
	}
	return fields[1] == zeroID && len(fields[0]) == 40 && fields[0] != zeroID
}

// IsUpdate returns whether the pkt-line contains an update command.
func IsUpdate(data []byte) bool {
	fields := strings.Fields(string(data[4:]))
	if len(fields) < 3 {
		return false
	}
	return len(fields[0]) == 40 && len(fields[1]) == 40 && fields[0] != zeroID && fields[1] != zeroID
}
