package search

import (
	"bufio"
	"io"
	"regexp"
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
// - [ ] Parallel search
// - [ ] Specialize searching a fixed string
// - [ ] Use grep optimized regex engine at github.com/google/codesearch/regexp
// - [ ] Skip bad files
// - [ ] Re-use memory buffers as much as possible
// - [ ] Possibly use cgo + c++ re2 library
// - [ ] Parse regex to find fixed string to search for
// - [ ] Limit results returned
// - [ ] Avoid parsing new lines (most important optimization apparently)
// - [ ] Only support UTF-8/ASCII
// - [ ] Avoid parsing files out of tar/zip (like new line trick)

func compile(p *Params) (func(reader io.Reader) ([]LineMatch, error), error) {
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

	return func(reader io.Reader) ([]LineMatch, error) {
		var matches []LineMatch
		r := bufio.NewReader(reader)
		for i := 1; ; i++ {
			// This skips large lines, but this implementation
			// will be replaced for a more correct one.
			b, isPrefix, err := r.ReadLine()
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}
			locs := re.FindAllIndex(b, -1)
			if len(locs) > 0 {
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
			for isPrefix && err != nil {
				_, isPrefix, err = r.ReadLine()
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				return nil, err
			}
		}
		return matches, nil
	}, nil
}
