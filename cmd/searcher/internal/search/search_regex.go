package search

import (
	"context"
	"io"
	//nolint:depguard // using the grafana fork of regexp clashes with zoekt, which uses the std regexp/syntax.
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func regexSearchBatch(ctx context.Context, m matcher, pm *pathMatcher, zf *zipFile, limit int, patternMatchesContent, patternMatchesPaths bool, isPatternNegated bool) ([]protocol.FileMatch, bool, error) {
	ctx, cancel, sender := newLimitedStreamCollector(ctx, limit)
	defer cancel()
	err := regexSearch(ctx, m, pm, zf, patternMatchesContent, patternMatchesPaths, isPatternNegated, sender)
	return sender.collected, sender.LimitHit(), err
}

// regexSearch concurrently searches files in zr looking for matches using m.
func regexSearch(ctx context.Context, m matcher, pm *pathMatcher, zf *zipFile, patternMatchesContent, patternMatchesPaths bool, isPatternNegated bool, sender matchSender) (err error) {
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
			if match := pm.Matches(f.Name) && m.MatchesString(f.Name); match == !isPatternNegated {
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
		m := m.Copy()
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

				// process
				fm, err := m.MatchesZip(zf, f, sender.Remaining())
				if err != nil {
					return err
				}
				match := len(fm.ChunkMatches) > 0
				if !match && patternMatchesPaths {
					// Try matching against the file path.
					match = m.MatchesString(f.Name)
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
