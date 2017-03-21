package search

import (
	"bufio"
	"io"
	"regexp"
)

// TODO modify sift or pt and use that code. Both rely on regexp pkg, but the
// bottlenecks seem to be around line counting which is currently naive
// here. sift doesn't support non-regexp expressions, but code is clean enough
// to easily add support for it (or just use regexp.EscapeMeta)

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
		scanner := bufio.NewScanner(reader)
		i := 0
		for scanner.Scan() {
			i++
			b := scanner.Bytes()
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
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return matches, nil
	}, nil
}
