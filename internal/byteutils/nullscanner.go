package byteutils

import "bytes"

// ScanNullLines is a split function for a [Scanner] that returns each null-terminated
// line of text, stripped of any trailing end-of-line marker. The returned line may
// be empty. The end-of-line marker is a null character.
// The last non-empty line of input will be returned even if it has no
// null character.
func ScanNullLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\x00'); i >= 0 {
		// We have a full null-terminated line.
		return i + 1, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
