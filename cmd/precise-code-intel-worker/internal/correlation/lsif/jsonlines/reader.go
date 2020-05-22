package jsonlines

import (
	"bufio"
	"context"
	"fmt"
	"io"

	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/lsif"
)

// LineBufferSize is the maximum size of the buffer used to read each line of a raw LSIF index. Lines in
// LSIF can get very long as it include escaped hover text (package documentation), as well as large edges
// such as the contains edge of large documents.
//
// This corresponds a 10MB buffer that can accommodate 10 million characters.
const LineBufferSize = 1e7

// TODO(efritz) - document
func Read(ctx context.Context, r io.Reader) <-chan lsif.Pair {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	scanner.Buffer(make([]byte, LineBufferSize), LineBufferSize)

	// TODO(efritz) - configure a buffer
	ch := make(chan lsif.Pair)

	go func() {
		defer close(ch)

		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			element, err := unmarshalElement(line)
			if err != nil {
				fmt.Printf("NOPE: %v\n", line)
				ch <- lsif.Pair{Err: err}
			} else {
				select {
				case ch <- lsif.Pair{Element: element}:
				case <-ctx.Done():
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- lsif.Pair{Err: err}
		}
	}()

	return ch
}
