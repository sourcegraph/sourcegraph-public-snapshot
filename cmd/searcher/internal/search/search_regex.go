package search

import (
	"bytes"
	"context"
	"io"
	"time"
	"unicode/utf8"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/search/casetransform"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func regexSearchBatch(
	ctx context.Context,
	m matcher,
	pm *pathMatcher,
	zf *zipFile,
	limit int,
	patternMatchesContent, patternMatchesPaths bool,
	isPatternNegated bool,
	contextLines int32,
) ([]protocol.FileMatch, bool, error) {
	ctx, cancel, sender := newLimitedStreamCollector(ctx, limit)
	defer cancel()
	err := regexSearch(ctx, m, pm, zf, patternMatchesContent, patternMatchesPaths, isPatternNegated, sender, contextLines)
	return sender.collected, sender.LimitHit(), err
}

// regexSearch concurrently searches files in zr looking for matches using m.
//
// This code is base on reading the techniques detailed in
// http://blog.burntsushi.net/ripgrep/
//
// The stdlib regexp is pretty powerful and in fact implements many of the
// features in ripgrep. Our implementation gives high performance via pruning
// aggressively which files to consider (non-binary under a limit) and
// optimizing for assuming most lines will not contain a match.
//
// If there is no more low-hanging fruit and perf is not acceptable, we could
// consider using ripgrep directly (modify it to search zip archives).
//
// TODO(keegan) return search statistics
func regexSearch(
	ctx context.Context,
	m matcher,
	pm *pathMatcher,
	zf *zipFile,
	patternMatchesContent, patternMatchesPaths bool,
	isPatternNegated bool,
	sender matchSender,
	contextLines int32,
) (err error) {
	tr, ctx := trace.New(ctx, "regexSearch")
	defer tr.EndWithErr(&err)

	m.AddAttributes(tr)
	tr.SetAttributes(attribute.Stringer("path", pm))

	if !patternMatchesContent && !patternMatchesPaths {
		patternMatchesContent = true
	}

	// If we reach limit we use cancel to stop the search
	var cancel context.CancelFunc
	if deadline, ok := ctx.Deadline(); ok {
		// If a deadline is set, try to finish before the deadline expires.
		timeout := time.Duration(0.9 * float64(time.Until(deadline)))
		tr.AddEvent("set timeout", attribute.Stringer("duration", timeout))
		ctx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	var (
		files = zf.Files
	)

	if m.MatchesAllContent() || (patternMatchesPaths && !patternMatchesContent) {
		// Fast path for only matching file paths (or with a nil pattern, which matches all files,
		// so is effectively matching only on file paths).
		for _, f := range files {
			if match := pm.Matches(f.Name) && m.MatchesPath(f.Name); match == !isPatternNegated {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				fm := protocol.FileMatch{Path: f.Name}
				sender.Send(fm)
			}
		}
		return nil
	}

	var (
		lastFileIdx   = atomic.NewInt32(-1)
		filesSkipped  atomic.Uint32
		filesSearched atomic.Uint32
	)

	g, ctx := errgroup.WithContext(ctx)

	contextCanceled := atomic.NewBool(false)
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		contextCanceled.Store(true)
		close(done)
	}()
	defer func() { cancel(); <-done }()

	// Start workers. They read from files and write to matches.
	for i := 0; i < numWorkers; i++ {
		// transformBuf is reused between file searches to avoid
		// re-allocating. It is only used if we need to transform the input
		// before matching. For example we lower case the input in the case of
		// ignoreCase.
		var transformBuf []byte
		g.Go(func() error {
			for !contextCanceled.Load() {
				idx := int(lastFileIdx.Inc())
				if idx >= len(files) {
					return nil
				}

				f := &files[idx]

				// decide whether to process, record that decision
				if !pm.Matches(f.Name) {
					filesSkipped.Inc()
					continue
				}
				filesSearched.Inc()

				// fileMatchBuf is what we run match on, fileBuf is the original
				// data (for Preview).
				fileBuf := zf.DataFor(f)
				fileMatchBuf := fileBuf
				if m.IgnoreCase() {
					// If we are ignoring case, we transform the input instead of
					// relying on the regular expression engine which can be
					// slow. compilePattern has already lowercased the pattern. We also
					// trade some correctness for perf by using a non-utf8 aware
					// lowercase function.
					if transformBuf == nil {
						transformBuf = make([]byte, zf.MaxLen)
					}
					fileMatchBuf = transformBuf[:len(fileBuf)]
					casetransform.BytesToLowerASCII(fileMatchBuf, fileBuf)
				}

				locs := m.MatchesFile(fileMatchBuf, sender.Remaining())
				fm := locsToFileMatch(fileBuf, f.Name, locs, contextLines)
				match := len(fm.ChunkMatches) > 0
				if !match && patternMatchesPaths {
					// Try matching against the file path.
					match = m.MatchesPath(f.Name)
					if match {
						fm.Path = f.Name
					}
				}
				if match == !isPatternNegated {
					sender.Send(fm)
				}
			}
			return nil
		})
	}

	err = g.Wait()
	if err == nil && ctx.Err() == context.DeadlineExceeded {
		// We stopped early because we were about to hit the deadline.
		err = ctx.Err()
	}

	tr.AddEvent(
		"done",
		attribute.Int("filesSkipped", int(filesSkipped.Load())),
		attribute.Int("filesSearched", int(filesSearched.Load())),
	)

	return err
}

// readAll will read r until EOF into b. It returns the number of bytes
// read. If we do not reach EOF, an error is returned.
func readAll(r io.Reader, b []byte) (int, error) {
	n := 0
	for {
		if len(b) == 0 {
			// We may be at EOF, but it hasn't returned that
			// yet. Technically r.Read is allowed to return 0,
			// nil, but it is strongly discouraged. If they do, we
			// will just return an err.
			scratch := []byte{'1'}
			_, err := r.Read(scratch)
			if err == io.EOF {
				return n, nil
			}
			return n, errors.New("reader is too large")
		}

		m, err := r.Read(b)
		n += m
		b = b[m:]
		if err != nil {
			if err == io.EOF { // done
				return n, nil
			}
			return n, err
		}
	}
}

func locsToFileMatch(fileBuf []byte, name string, locs [][]int, contextLines int32) protocol.FileMatch {
	if len(locs) == 0 {
		return protocol.FileMatch{
			Path:     name,
			LimitHit: false,
		}
	}
	ranges := locsToRanges(fileBuf, locs)
	chunks := chunkRanges(ranges, contextLines*2)
	cms := chunksToMatches(fileBuf, chunks, contextLines)
	return protocol.FileMatch{
		Path:         name,
		ChunkMatches: cms,
		LimitHit:     false,
	}
}

// locs must be sorted, non-overlapping, and must be valid slices of buf.
func locsToRanges(buf []byte, locs [][]int) []protocol.Range {
	ranges := make([]protocol.Range, 0, len(locs))

	prevEnd := 0
	prevEndLine := 0

	for _, loc := range locs {
		start, end := loc[0], loc[1]

		startLine := prevEndLine + bytes.Count(buf[prevEnd:start], []byte{'\n'})
		endLine := startLine + bytes.Count(buf[start:end], []byte{'\n'})

		firstLineStart := 0
		if off := bytes.LastIndexByte(buf[:start], '\n'); off >= 0 {
			firstLineStart = off + 1
		}

		lastLineStart := firstLineStart
		if off := bytes.LastIndexByte(buf[:end], '\n'); off >= 0 {
			lastLineStart = off + 1
		}

		ranges = append(ranges, protocol.Range{
			Start: protocol.Location{
				Offset: int32(start),
				Line:   int32(startLine),
				Column: int32(utf8.RuneCount(buf[firstLineStart:start])),
			},
			End: protocol.Location{
				Offset: int32(end),
				Line:   int32(endLine),
				Column: int32(utf8.RuneCount(buf[lastLineStart:end])),
			},
		})

		prevEnd = end
		prevEndLine = endLine
	}

	return ranges
}
