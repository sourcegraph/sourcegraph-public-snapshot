package google

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

const maxPayloadSize = 10 * 1024 * 1024 // 10mb

var doneBytes = []byte("[DONE]")

// decoder decodes streaming events from a Server Sent Event stream.
type decoder struct {
	scanner *bufio.Scanner
	done    bool
	data    []byte
	err     error
}

func NewDecoder(r io.Reader) *decoder {
	// Custom split function to handle JSON objects separated by commas and newlines
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 262144), maxPayloadSize)
	// Custom split function to handle JSON objects separated by commas and newlines
	split := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		// Look for the end of a JSON object
		openBraces := 0
		start := 0
		for i := 0; i < len(data); i++ {
			switch data[i] {
			case '{':
				if openBraces == 0 {
					start = i
				}
				openBraces++
			case '}':
				openBraces--
				if openBraces == 0 {
					// Found a complete JSON object
					return i + 1, data[start : i+1], nil
				}
			case ',':
				// Ignore commas between top-level JSON objects
				if openBraces == 0 && start == i {
					start = i + 1
				}
			}
		}
		// If we're at EOF and we have a final, non-terminated JSON object
		if atEOF && openBraces == 0 {
			return len(data), data[start:], nil
		}
		// Request more data
		return 0, nil, nil
	}
	scanner.Split(split)
	return &decoder{
		scanner: scanner,
	}
}

// Scan advances the decoder to the next event in the stream. It returns
// false when it either hits the end of the stream or an error.
func (d *decoder) Scan() bool {
	if d.done {
		return false
	}
	for d.scanner.Scan() {
		line := d.scanner.Bytes()
		// Directly use the line as data without looking for "data:" prefix
		d.data = line
		fmt.Println("the line is", string(line))
		// Check for special sentinel value used by the Google API to
		// indicate that the stream is done.
		if bytes.Equal(line, doneBytes) {
			d.done = true
			return false
		}
		return true
	}

	d.err = d.scanner.Err()
	return false
}

// Event returns the event data of the last decoded event
func (d *decoder) Data() []byte {
	return d.data
}

// Err returns the last encountered error
func (d *decoder) Err() error {
	return d.err
}

func splitColon(data []byte) ([]byte, []byte) {
	i := bytes.Index(data, []byte(":"))
	if i < 0 {
		return bytes.TrimSpace(data), nil
	}
	return bytes.TrimSpace(data[:i]), bytes.TrimSpace(data[i+1:])
}
