package search

import (
	"bufio"
	"context"
	"errors"
	"io"
	"regexp"
	"regexp/syntax"
	"sync"
	"unicode"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/pathmatch"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/searcher/protocol"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/lazyzip"
)

const (
	// maxFileSize is the limit on file size in bytes. Only files smaller
	// than this are searched.
	maxFileSize = 1 << 19 // 512KB

	// maxLineSize is the maximum length of a line in bytes.
	// Lines larger than this are not scanned for results.
	// (e.g. minified javascript files that are all on one line).
	maxLineSize = 500

	// maxFileMatches is the limit on number of matching files we return.
	maxFileMatches = 1000

	// maxLineMatches is the limit on number of matches to return in a
	// file.
	maxLineMatches = 100

	// maxOffsets is the limit on number of matches to return on a line.
	maxOffsets = 10

	// numWorkers is how many concurrent readerGreps run per
	// concurrentFind
	numWorkers = 8
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
// files is done by the store.
//
// If there is no more low-hanging fruit and perf is not acceptable, we could
// consider an using ripgrep directly (modify it to search zip archives).
//
// TODO(keegan) return search statistics
type readerGrep struct {
	// re is the regexp to match.
	re *regexp.Regexp

	// ignoreCase if true means we need to do case insensitive matching.
	ignoreCase bool

	// buf is reused between file searches to avoid re-allocating. It
	// holds an entire file.
	buf []byte

	// transformBuf is reused between file searches to avoid
	// re-allocating. It is only used if we need to transform the input
	// before matching. For example we lower case the input in the case of
	// ignoreCase.
	transformBuf []byte

	// matchPath is compiled from the include/exclude path patterns and reports
	// whether a file path matches (and should be searched).
	matchPath pathmatch.PathMatcher
}

// compile returns a readerGrep for matching p.
func compile(p *protocol.PatternInfo) (*readerGrep, error) {
	var (
		expr       = p.Pattern
		ignoreCase bool
	)
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
	if !p.IsCaseSensitive {
		// We don't just use (?i) because regexp library doesn't seem
		// to contain good optimizations for case insensitive
		// search. Instead we lowercase the input and pattern.
		re, err := syntax.Parse(expr, syntax.Perl)
		if err != nil {
			return nil, err
		}
		lowerRegexpASCII(re)
		expr = re.String()
		ignoreCase = true
	}

	pathOptions := pathmatch.CompileOptions{
		RegExp:        p.PathPatternsAreRegExps,
		CaseSensitive: p.PathPatternsAreCaseSensitive,
	}
	matchPath, err := pathmatch.CompilePathPatterns(p.AllIncludePatterns(), p.ExcludePattern, pathOptions)
	if err != nil {
		return nil, err
	}

	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &readerGrep{
		re:         re,
		ignoreCase: ignoreCase,
		matchPath:  matchPath,
	}, nil
}

// Copy returns a copied version of rg that is safe to use from another
// goroutine.
func (rg *readerGrep) Copy() *readerGrep {
	return &readerGrep{
		re:         rg.re.Copy(),
		ignoreCase: rg.ignoreCase,
		matchPath:  rg.matchPath.Copy(),
	}
}

// Find returns a LineMatch for each line that matches rg in reader.
// LimitHit is true if some matches may not have been included in the result.
// NOTE: This is not safe to use concurrently.
func (rg *readerGrep) Find(reader io.Reader) (matches []protocol.LineMatch, limitHit bool, err error) {
	if rg.buf == nil {
		// reader should always have maxFileSize bytes or less.
		rg.buf = make([]byte, maxFileSize)
		if rg.ignoreCase {
			rg.transformBuf = make([]byte, maxFileSize)
		}
	}

	// Read the file into memory. We have a relatively low maxFileSize, so
	// we can simplify and avoid needing to stream the lines.
	n, err := readAll(reader, rg.buf)
	if err != nil {
		return nil, false, err
	}
	// fileMatchBuf is what we run match on, fileBuf is the original
	// data (for Preview).
	fileBuf := rg.buf[:n]
	fileMatchBuf := fileBuf

	// If we are ignoring case, we transform the input instead of
	// relying on the regular expression engine which can be
	// slow. compile has already lowercased the pattern. We also
	// trade some correctness for perf by using a non-utf8 aware
	// lowercase function.
	if rg.ignoreCase {
		fileMatchBuf = rg.transformBuf[:len(fileBuf)]
		bytesToLowerASCII(fileMatchBuf, fileBuf)
	}

	// Most files will not have a match and we bound the number of matched
	// files we return. So we can avoid the overhead of parsing out new
	// lines and repeatedly running the regex engine by running a single
	// match over the whole file. This does mean we duplicate work when
	// actually searching for results. We use the same approach when we
	// search per-line.
	if rg.re.Find(fileMatchBuf) == nil {
		return nil, false, nil
	}

	for i := 0; len(matches) < maxLineMatches; i++ {
		advance, lineBuf, err := bufio.ScanLines(fileBuf, true)
		if err != nil {
			// ScanLines should never return an err
			return nil, false, err
		}
		if advance == 0 { // EOF
			break
		}

		// matchBuf is what we actually match on. We have already done
		// the transform of fileBuf in fileMatchBuf. lineBuf is a
		// prefix of fileBuf, so matchBuf is the corresponding prefix.
		matchBuf := fileMatchBuf[:len(lineBuf)]

		// Advance file bufs in sync
		fileBuf = fileBuf[advance:]
		fileMatchBuf = fileMatchBuf[advance:]

		// Skip lines that are too long.
		if len(matchBuf) > maxLineSize {
			continue
		}

		// FindAllIndex allocates memory. We can avoid that by just
		// checking if we have a match first. We expect most lines to
		// not have a match, so we trade a bit of repeated computation
		// to avoid unnecessary allocations.
		if rg.re.Find(matchBuf) != nil {
			locs := rg.re.FindAllIndex(matchBuf, maxOffsets)
			lineLimitHit := len(locs) == maxOffsets
			offsetAndLengths := make([][]int, len(locs))
			for i, match := range locs {
				start, end := match[0], match[1]
				offsetAndLengths[i] = []int{start, end - start}
			}
			matches = append(matches, protocol.LineMatch{
				// making a copy of lineBuf is intentional, the underlying array of b can be overwritten by scanner.
				Preview:          string(lineBuf),
				LineNumber:       i,
				OffsetAndLengths: offsetAndLengths,
				LimitHit:         lineLimitHit,
			})
		}
	}
	limitHit = len(matches) == maxLineMatches
	return matches, limitHit, nil
}

// FindZip is a convenience function to run Find on f.
func (rg *readerGrep) FindZip(f *lazyzip.File) (protocol.FileMatch, error) {
	rc, err := f.Open()
	if err != nil {
		return protocol.FileMatch{}, err
	}
	lm, limitHit, err := rg.Find(rc)
	rc.Close()
	return protocol.FileMatch{
		Path:        f.Name,
		LineMatches: lm,
		LimitHit:    limitHit,
	}, err
}

// concurrentFind searches files in zr looking for matches using rg.
func concurrentFind(ctx context.Context, rg *readerGrep, zr *lazyzip.Reader, fileMatchLimit int) (fm []protocol.FileMatch, limitHit bool, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ConcurrentFind")
	ext.Component.Set(span, "matcher")
	span.SetTag("re", rg.re.String())
	span.SetTag("path", rg.matchPath.String())
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err.Error())
		}
		span.Finish()
	}()

	if fileMatchLimit > maxFileMatches || fileMatchLimit <= 0 {
		fileMatchLimit = maxFileMatches
	}

	// If we reach fileMatchLimit we use cancel to stop the search
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		files     = make(chan *lazyzip.File)
		matches   = make(chan protocol.FileMatch)
		wg        sync.WaitGroup
		wgErrOnce sync.Once
		wgErr     error
		filesErr  error
	)

	// goroutine responsible for writing to files. It also is the only
	// goroutine which listens for cancellation.
	go func() {
		var filesSkipped, filesSearched int
		defer func() {
			// We can write to span since nothing else will until files is
			// closed
			span.LogFields(
				otlog.Int("filesSkipped", filesSkipped),
				otlog.Int("filesSearched", filesSearched),
			)
			close(files)
		}()

		done := ctx.Done()
		for {
			f, err := zr.Next()
			if err != nil {
				if err != io.EOF {
					filesErr = err
				}
				return
			}
			if !rg.matchPath.MatchPath(f.Name) {
				filesSkipped++
				continue
			}
			select {
			case files <- f:
				filesSearched++
			case <-done:
				return
			}
		}
	}()

	// Start workers. They read from files and write to matches.
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(rg *readerGrep) {
			defer wg.Done()
			for f := range files {
				fm, err := rg.FindZip(f)
				if err != nil {
					wgErrOnce.Do(func() {
						wgErr = err
						// Drain files
						for range files {
						}
					})
					return
				}
				if len(fm.LineMatches) > 0 {
					matches <- fm
				}
			}
		}(rg.Copy())
	}

	// Wait for workers to be done. Signal to collector there is no more
	// results coming by closing matches.
	go func() {
		wg.Wait()
		close(matches)
	}()

	// Collect all matches. Do not return a nil slice if we find nothing
	// so we can nicely serialize it.
	m := []protocol.FileMatch{}
	for fm := range matches {
		m = append(m, fm)
		if len(m) >= fileMatchLimit {
			limitHit = true
			cancel()
			// drain matches
			for range matches {
			}
		}
	}

	err = wgErr
	if err == nil {
		err = filesErr
	}
	return m, limitHit, err
}

// lowerRegexpASCII lowers rune literals and expands char classes to include
// lowercase. It does it inplace. We can't just use strings.ToLower since it
// will change the meaning of regex shorthands like \S or \B.
func lowerRegexpASCII(re *syntax.Regexp) {
	for _, c := range re.Sub {
		if c != nil {
			lowerRegexpASCII(c)
		}
	}
	switch re.Op {
	case syntax.OpLiteral:
		// For literal strings we can simplify lower each character.
		for i := range re.Rune {
			re.Rune[i] = unicode.ToLower(re.Rune[i])
		}
	case syntax.OpCharClass:
		l := len(re.Rune)
		for i := 0; i < l; i += 2 {
			// We found a char class that includes a-z. No need to
			// modify.
			if re.Rune[i] <= 'a' && re.Rune[i+1] >= 'z' {
				return
			}
		}
		for i := 0; i < l; i += 2 {
			a, b := re.Rune[i], re.Rune[i+1]
			// This range doesn't include A-Z, so skip
			if a > 'Z' || b < 'A' {
				continue
			}
			simple := true
			if a < 'A' {
				simple = false
				a = 'A'
			}
			if b > 'Z' {
				simple = false
				b = 'Z'
			}
			a, b = unicode.ToLower(a), unicode.ToLower(b)
			if simple {
				// The char range is within A-Z, so we can
				// just modify it to be the equivalent in a-z.
				re.Rune[i], re.Rune[i+1] = a, b
			} else {
				// The char range includes characters outside
				// of A-Z. To be safe we just append a new
				// lowered range which is the intersection
				// with A-Z.
				re.Rune = append(re.Rune, a, b)
			}
		}
	default:
		return
	}
	// Copy to small storage if necessary
	for i := 0; i < 2 && i < len(re.Rune); i++ {
		re.Rune0[i] = re.Rune[i]
	}
}

// python to generate ', '.join(hex(ord(chr(i).lower())) for i in range(256))
var lowerTable = [256]uint8{
	0x0, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa, 0xb, 0xc, 0xd, 0xe, 0xf,
	0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
	0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
	0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
	0x40, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f,
	0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x78, 0x79, 0x7a, 0x5b, 0x5c, 0x5d, 0x5e, 0x5f,
	0x60, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f,
	0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x78, 0x79, 0x7a, 0x7b, 0x7c, 0x7d, 0x7e, 0x7f,
	0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8a, 0x8b, 0x8c, 0x8d, 0x8e, 0x8f,
	0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9a, 0x9b, 0x9c, 0x9d, 0x9e, 0x9f,
	0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8, 0xa9, 0xaa, 0xab, 0xac, 0xad, 0xae, 0xaf,
	0xb0, 0xb1, 0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7, 0xb8, 0xb9, 0xba, 0xbb, 0xbc, 0xbd, 0xbe, 0xbf,
	0xc0, 0xc1, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6, 0xc7, 0xc8, 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf,
	0xd0, 0xd1, 0xd2, 0xd3, 0xd4, 0xd5, 0xd6, 0xd7, 0xd8, 0xd9, 0xda, 0xdb, 0xdc, 0xdd, 0xde, 0xdf,
	0xe0, 0xe1, 0xe2, 0xe3, 0xe4, 0xe5, 0xe6, 0xe7, 0xe8, 0xe9, 0xea, 0xeb, 0xec, 0xed, 0xee, 0xef,
	0xf0, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7, 0xf8, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff,
}

func bytesToLowerASCII(dst, src []byte) {
	// we assume len(dst) >= len(src). We do the below to hint to the
	// compiler to eliminate bounds check on dst.
	dst = dst[:len(src)]
	for i := range src {
		dst[i] = lowerTable[src[i]]
	}
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
