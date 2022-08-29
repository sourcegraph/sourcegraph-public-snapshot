package search

import (
	"bytes"
	"context"
	"io"
	"regexp/syntax"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/grafana/regexp"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/pathmatch"
	"github.com/sourcegraph/sourcegraph/internal/search/casetransform"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/zoekt/query"
)

// readerGrep is responsible for finding LineMatches. It is not concurrency
// safe (it reuses buffers for performance).
//
// This code is base on reading the techniques detailed in
// http://blog.burntsushi.net/ripgrep/
//
// The stdlib regexp is pretty powerful and in fact implements many of the
// features in ripgrep. Our implementation gives high performance via pruning
// aggressively which files to consider (non-binary under a limit) and
// optimizing for assuming most lines will not contain a match. The pruning of
// files is done by the
//
// If there is no more low-hanging fruit and perf is not acceptable, we could
// consider using ripgrep directly (modify it to search zip archives).
//
// TODO(keegan) return search statistics
type readerGrep struct {
	// re is the regexp to match, or nil if empty ("match all files' content").
	re *regexp.Regexp

	// ignoreCase if true means we need to do case insensitive matching.
	ignoreCase bool

	// transformBuf is reused between file searches to avoid
	// re-allocating. It is only used if we need to transform the input
	// before matching. For example we lower case the input in the case of
	// ignoreCase.
	transformBuf []byte

	// matchPath is compiled from the include/exclude path patterns and reports
	// whether a file path matches (and should be searched).
	matchPath pathmatch.PathMatcher

	// literalSubstring is used to test if a file is worth considering for
	// matches. literalSubstring is guaranteed to appear in any match found by
	// re. It is the output of the longestLiteral function. It is only set if
	// the regex has an empty LiteralPrefix.
	literalSubstring []byte
}

// compile returns a readerGrep for matching p.
func compile(p *protocol.PatternInfo) (*readerGrep, error) {
	var (
		re               *regexp.Regexp
		literalSubstring []byte
	)
	if p.Pattern != "" {
		expr := p.Pattern
		if !p.IsRegExp {
			expr = regexp.QuoteMeta(expr)
		}
		if p.IsWordMatch {
			expr = `\b` + expr + `\b`
		}
		if p.IsRegExp {
			// We don't do the search line by line, therefore we want the
			// regex engine to consider newlines for anchors (^$).
			expr = "(?m:" + expr + ")"
		}

		// Transforms on the parsed regex
		{
			re, err := syntax.Parse(expr, syntax.Perl)
			if err != nil {
				return nil, err
			}

			if !p.IsCaseSensitive {
				// We don't just use (?i) because regexp library doesn't seem
				// to contain good optimizations for case insensitive
				// search. Instead we lowercase the input and pattern.
				casetransform.LowerRegexpASCII(re)
			}

			// OptimizeRegexp currently only converts capture groups into
			// non-capture groups (faster for stdlib regexp to execute).
			re = query.OptimizeRegexp(re, syntax.Perl)

			expr = re.String()
		}

		var err error
		re, err = regexp.Compile(expr)
		if err != nil {
			return nil, err
		}

		// Only use literalSubstring optimization if the regex engine doesn't
		// have a prefix to use.
		if pre, _ := re.LiteralPrefix(); pre == "" {
			ast, err := syntax.Parse(expr, syntax.Perl)
			if err != nil {
				return nil, err
			}
			ast = ast.Simplify()
			literalSubstring = []byte(longestLiteral(ast))
		}
	}

	pathOptions := pathmatch.CompileOptions{
		RegExp:        p.PathPatternsAreRegExps,
		CaseSensitive: p.PathPatternsAreCaseSensitive,
	}
	matchPath, err := pathmatch.CompilePathPatterns(p.IncludePatterns, p.ExcludePattern, pathOptions)
	if err != nil {
		return nil, err
	}

	return &readerGrep{
		re:               re,
		ignoreCase:       !p.IsCaseSensitive,
		matchPath:        matchPath,
		literalSubstring: literalSubstring,
	}, nil
}

// Copy returns a copied version of rg that is safe to use from another
// goroutine.
func (rg *readerGrep) Copy() *readerGrep {
	return &readerGrep{
		re:               rg.re,
		ignoreCase:       rg.ignoreCase,
		matchPath:        rg.matchPath,
		literalSubstring: rg.literalSubstring,
	}
}

// matchString returns whether rg's regexp pattern matches s. It is intended to be
// used to match file paths.
func (rg *readerGrep) matchString(s string) bool {
	if rg.re == nil {
		return true
	}
	if rg.ignoreCase {
		s = strings.ToLower(s)
	}
	return rg.re.MatchString(s)
}

// Find returns a LineMatch for each line that matches rg in reader.
// LimitHit is true if some matches may not have been included in the result.
// NOTE: This is not safe to use concurrently.
func (rg *readerGrep) Find(zf *zipFile, f *srcFile, limit int) (matches []protocol.ChunkMatch, err error) {
	// fileMatchBuf is what we run match on, fileBuf is the original
	// data (for Preview).
	fileBuf := zf.DataFor(f)
	fileMatchBuf := fileBuf

	// If we are ignoring case, we transform the input instead of
	// relying on the regular expression engine which can be
	// slow. compile has already lowercased the pattern. We also
	// trade some correctness for perf by using a non-utf8 aware
	// lowercase function.
	if rg.ignoreCase {
		if rg.transformBuf == nil {
			rg.transformBuf = make([]byte, zf.MaxLen)
		}
		fileMatchBuf = rg.transformBuf[:len(fileBuf)]
		casetransform.BytesToLowerASCII(fileMatchBuf, fileBuf)
	}

	// Most files will not have a match and we bound the number of matched
	// files we return. So we can avoid the overhead of parsing out new lines
	// and repeatedly running the regex engine by running a single match over
	// the whole file. This does mean we duplicate work when actually
	// searching for results. We use the same approach when we search
	// per-line. Additionally if we have a non-empty literalSubstring, we use
	// that to prune out files since doing bytes.Index is very fast.
	if !bytes.Contains(fileMatchBuf, rg.literalSubstring) {
		return nil, nil
	}

	// find limit+1 matches so we know whether we hit the limit
	locs := rg.re.FindAllIndex(fileMatchBuf, limit+1)
	if len(locs) == 0 {
		return nil, nil // short-circuit if we have no matches
	}
	ranges := locsToRanges(fileBuf, locs)
	chunks := chunkRanges(ranges, 0)
	return chunksToMatches(fileBuf, chunks), nil
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

// FindZip is a convenience function to run Find on f.
func (rg *readerGrep) FindZip(zf *zipFile, f *srcFile, limit int) (protocol.FileMatch, error) {
	cms, err := rg.Find(zf, f, limit)
	return protocol.FileMatch{
		Path:         f.Name,
		ChunkMatches: cms,
		LimitHit:     false,
	}, err
}

func regexSearchBatch(ctx context.Context, rg *readerGrep, zf *zipFile, limit int, patternMatchesContent, patternMatchesPaths bool, isPatternNegated bool) ([]protocol.FileMatch, bool, error) {
	ctx, cancel, sender := newLimitedStreamCollector(ctx, limit)
	defer cancel()
	err := regexSearch(ctx, rg, zf, patternMatchesContent, patternMatchesPaths, isPatternNegated, sender)
	return sender.collected, sender.LimitHit(), err
}

// regexSearch concurrently searches files in zr looking for matches using rg.
func regexSearch(ctx context.Context, rg *readerGrep, zf *zipFile, patternMatchesContent, patternMatchesPaths bool, isPatternNegated bool, sender matchSender) error {
	var err error
	span, ctx := ot.StartSpanFromContext(ctx, "RegexSearch")
	ext.Component.Set(span, "regex_search")
	if rg.re != nil {
		span.SetTag("re", rg.re.String())
	}
	span.SetTag("path", rg.matchPath.String())
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	if !patternMatchesContent && !patternMatchesPaths {
		patternMatchesContent = true
	}

	// If we reach limit we use cancel to stop the search
	var cancel context.CancelFunc
	if deadline, ok := ctx.Deadline(); ok {
		// If a deadline is set, try to finish before the deadline expires.
		timeout := time.Duration(0.9 * float64(time.Until(deadline)))
		span.LogFields(otlog.Int64("RegexSearchTimeout", int64(timeout)))
		ctx, cancel = context.WithTimeout(ctx, timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	var (
		files = zf.Files
	)

	if rg.re == nil || (patternMatchesPaths && !patternMatchesContent) {
		// Fast path for only matching file paths (or with a nil pattern, which matches all files,
		// so is effectively matching only on file paths).
		for _, f := range files {
			if match := rg.matchPath.MatchPath(f.Name) && rg.matchString(f.Name); match == !isPatternNegated {
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
		rg := rg.Copy()
		g.Go(func() error {
			for !contextCanceled.Load() {
				idx := int(lastFileIdx.Inc())
				if idx >= len(files) {
					return nil
				}

				f := &files[idx]

				// decide whether to process, record that decision
				if !rg.matchPath.MatchPath(f.Name) {
					filesSkipped.Inc()
					continue
				}
				filesSearched.Inc()

				// process
				fm, err := rg.FindZip(zf, f, sender.Remaining())
				if err != nil {
					return err
				}
				match := len(fm.ChunkMatches) > 0
				if !match && patternMatchesPaths {
					// Try matching against the file path.
					match = rg.matchString(f.Name)
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

	span.LogFields(
		otlog.Int("filesSkipped", int(filesSkipped.Load())),
		otlog.Int("filesSearched", int(filesSearched.Load())),
	)

	return err
}

// longestLiteral finds the longest substring that is guaranteed to appear in
// a match of re.
//
// Note: There may be a longer substring that is guaranteed to appear. For
// example we do not find the longest common substring in alternating
// group. Nor do we handle concatting simple capturing groups.
func longestLiteral(re *syntax.Regexp) string {
	switch re.Op {
	case syntax.OpLiteral:
		return string(re.Rune)
	case syntax.OpCapture, syntax.OpPlus:
		return longestLiteral(re.Sub[0])
	case syntax.OpRepeat:
		if re.Min >= 1 {
			return longestLiteral(re.Sub[0])
		}
	case syntax.OpConcat:
		longest := ""
		for _, sub := range re.Sub {
			l := longestLiteral(sub)
			if len(l) > len(longest) {
				longest = l
			}
		}
		return longest
	}
	return ""
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
