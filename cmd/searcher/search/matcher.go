package search

import (
	"archive/zip"
	"bufio"
	"context"
	"io"
	"regexp"
	"sync"
)

// This code is base on reading the techniques detailed in
// http://blog.burntsushi.net/ripgrep/
//
// If all low-hanging fruit has been reached, it likely makes sense to write
// some glue in rust to make ripgrep search on an archive and shell out to
// ripgrep.
//
// Note: There is still plenty of low-hanging fruit here! Please remove this
// comment once there isn't! TODO:
// - [x] Parallel search
// - [ ] Specialize searching a fixed string
// - [ ] Use grep optimized regex engine at github.com/google/codesearch/regexp
// - [ ] Skip bad files
// - [x] Re-use memory buffers as much as possible
// - [ ] Possibly use cgo + c++ re2 library
// - [ ] Parse regex to find fixed string to search for
// - [ ] Limit results returned
// - [ ] Avoid parsing new lines (most important optimization apparently)
// - [ ] Only support UTF-8/ASCII
// - [ ] Avoid parsing files out of tar/zip (like new line trick)

const (
	// maxFileSize is the limit on file size in bytes. Only files smaller
	// than this are searched.
	maxFileSize = 1 << 19 // 512KB
)

// readerGrep is responsible for finding LineMatches. It is not concurrency
// safe (it reuses buffers for performance).
type readerGrep struct {
	// re is the regexp to match
	re *regexp.Regexp

	// reader is reused between file searches to avoid re-allocating the
	// underlying buffer.
	reader *bufio.Reader
}

// compile returns a readerGrep for matching p.
func compile(p *Params) (*readerGrep, error) {
	expr := p.Pattern
	if !p.IsRegExp {
		expr = regexp.QuoteMeta(expr)
	}
	if p.IsWordMatch {
		expr = `\b` + expr + `\b`
	}
	if !p.IsCaseSensitive {
		expr = `(?i)(` + expr + `)`
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &readerGrep{re: re}, nil
}

// Copy returns a copied version of rg that is safe to use from another
// goroutine.
func (rg *readerGrep) Copy() *readerGrep {
	return &readerGrep{re: rg.re.Copy()}
}

// Find returns a LineMatch for each line that matches rg in reader.
//
// NOTE: This is not safe to use concurrently.
func (rg *readerGrep) Find(reader io.Reader) ([]LineMatch, error) {
	r := rg.reader
	if r == nil {
		r = bufio.NewReader(reader)
		rg.reader = r
	} else {
		r.Reset(reader)
	}

	var matches []LineMatch
	for i := 1; ; i++ {
		b, isPrefix, err := r.ReadLine()
		if isPrefix || err != nil {
			// We have either found a long line, encountered an
			// error or reached EOF. We skip files with long lines
			// since the user is unlikely interested in the
			// result, so the only case we want to return matches
			// is if we have reached the end of file.
			if err == io.EOF {
				return matches, nil
			}
			return nil, err
		}
		// FindAllIndex allocates memory. We can avoid that by just
		// checking if we have a match first. We expect most lines to
		// not have a match, so we trade a bit of repeated computation
		// to avoid unnecessary allocations.
		if rg.re.Find(b) != nil {
			locs := rg.re.FindAllIndex(b, -1)
			offsetAndLengths := make([][]int, len(locs))
			for i, match := range locs {
				start, end := match[0], match[1]
				offsetAndLengths[i] = []int{start, end - start}
			}
			matches = append(matches, LineMatch{
				// making a copy of b is intentional, the underlying array of b can be overwritten by scanner.
				Preview:          string(b),
				LineNumber:       i,
				OffsetAndLengths: offsetAndLengths,
			})
		}
	}
}

// FindZip is a convenience function to run Find on f.
func (rg *readerGrep) FindZip(f *zip.File) (FileMatch, error) {
	rc, err := f.Open()
	if err != nil {
		return FileMatch{}, err
	}
	lm, err := rg.Find(rc)
	rc.Close()
	return FileMatch{
		Path:        f.Name,
		LineMatches: lm,
	}, err
}

// concurrentFind searches files in zr looking for matches using rg.
func concurrentFind(ctx context.Context, rg *readerGrep, zr *zip.Reader) ([]FileMatch, error) {
	var (
		files      = make(chan *zip.File)
		matches    = make(chan FileMatch)
		numWorkers = 8 // TODO: Needs tuning!
		wg         sync.WaitGroup
		wgErrOnce  sync.Once
		wgErr      error
	)

	// goroutine responsible for writing to files. It also is the only
	// goroutine which listens for cancellation.
	go func() {
		done := ctx.Done()
		for _, f := range zr.File {
			if f.FileHeader.UncompressedSize64 > maxFileSize {
				continue
			}
			select {
			case files <- f:
			case <-done:
				close(files)
				return
			}
		}
		close(files)
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
	m := []FileMatch{}
	for fm := range matches {
		m = append(m, fm)
	}
	return m, wgErr
}
