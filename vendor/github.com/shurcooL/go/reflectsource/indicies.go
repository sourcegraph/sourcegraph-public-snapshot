package reflectsource

import (
	"bytes"
)

// getLineStartEndIndicies gets the starting and ending caret indicies of line with specified lineIndex.
// Does not include newline character.
// First line has index 0.
// Returns (-1, -1) if line is not found.
func getLineStartEndIndicies(b []byte, lineIndex int) (startIndex, endIndex int) {
	index := 0
	for line := 0; ; line++ {
		lineLength := bytes.IndexByte(b[index:], '\n')
		if line == lineIndex {
			if lineLength == -1 {
				return index, len(b)
			} else {
				return index, index + lineLength
			}
		}
		if lineLength == -1 {
			break
		}
		index += lineLength + 1
	}

	return -1, -1
}
