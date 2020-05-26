package jsonlines

import (
	"bufio"
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/internal/correlation/lsif"
)

// Read reads the given content as line-separated JSON objects representing a single LSIF vertex or edge and
// returns a channel of lsif.Pair values for each non-empty line.
func Read(ctx context.Context, r io.Reader) <-chan lsif.Pair {
	return ReadLines(ctx, r, unmarshalElement)
}

// LineBufferSize is the maximum size of the buffer used to read each line of a raw LSIF index. Lines in
// LSIF can get very long as it include escaped hover text (package documentation), as well as large edges
// such as the contains edge of large documents.
//
// This corresponds a 10MB buffer that can accommodate 10 million characters.
const LineBufferSize = 1e7

// TODO - document
const ChannelBufferSize = 500

// TODO - document
const NumUnmarshalRoutines = 100

// TODO - document
func ReadLines(ctx context.Context, r io.Reader, unmarshal func(line []byte) (_ lsif.Element, err error)) <-chan lsif.Pair {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	scanner.Buffer(make([]byte, LineBufferSize), LineBufferSize)

	// TODO - document
	lineCh := make(chan []byte, ChannelBufferSize)
	go func() {
		defer close(lineCh)

		for scanner.Scan() {
			if line := scanner.Bytes(); len(line) != 0 {
				lineCh <- line
			}
		}
	}()

	pairCh := make(chan lsif.Pair, ChannelBufferSize)
	go func() {
		defer close(pairCh)

		// TODO - document
		work := make(chan int, NumUnmarshalRoutines)
		defer close(work)

		// TODO - document
		signal := make(chan struct{}, NumUnmarshalRoutines)
		defer close(signal)

		lines := make([][]byte, NumUnmarshalRoutines)
		pairs := make([]lsif.Pair, NumUnmarshalRoutines)

		// TODO - document
		for i := 0; i < NumUnmarshalRoutines; i++ {
			go func() {
				for idx := range work {
					element, err := unmarshal(lines[idx])
					pairs[idx].Element = element
					pairs[idx].Err = err
					signal <- struct{}{}
				}
			}()
		}

		done := false
		for !done {
			i := 0

			// TODO - document
			for i < NumUnmarshalRoutines {
				line, ok := <-lineCh
				if !ok {
					done = true
					break
				}

				lines[i] = line
				work <- i
				i++
			}

			// TODO - document
			for j := 0; j < i; j++ {
				<-signal
			}

			// TODO - document
			for j := 0; j < i; j++ {
				select {
				case pairCh <- pairs[j]:
				case <-ctx.Done():
					return
				}
			}
		}

		// TODO - document
		if err := scanner.Err(); err != nil {
			pairCh <- lsif.Pair{Err: err}
		}
	}()

	return pairCh
}
