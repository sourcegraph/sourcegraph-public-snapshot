package discussions

import "strings"

// LineRange represents a line range in a file.
type LineRange struct {
	// StarLine of the range (zero-based, inclusive).
	StartLine int

	// EndLine of the range (zero-based, exclusive).
	EndLine int
}

// LinesForSelection returns the lines from the given file's contents for the
// given selection.
func LinesForSelection(fileContent string, selection LineRange) (linesBefore, lines, linesAfter []string) {
	allLines := strings.Split(fileContent, "\n")
	clamp := func(v, min, max int) int {
		if v < min {
			return min
		} else if v > max {
			return max
		}
		return v
	}
	linesForRange := func(startLine, endLine int) []string {
		startLine = clamp(startLine, 0, len(allLines))
		endLine = clamp(endLine, 0, len(allLines))
		selectedLines := allLines[startLine:endLine]
		return selectedLines
	}
	linesBefore = linesForRange(selection.StartLine-3, selection.StartLine)
	lines = linesForRange(selection.StartLine, selection.EndLine)
	linesAfter = linesForRange(selection.EndLine, selection.EndLine+3)
	return
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_370(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
