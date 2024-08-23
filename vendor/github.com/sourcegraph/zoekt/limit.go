package zoekt

import "log"

// SortAndTruncateFiles is a convenience around SortFiles and
// DisplayTruncator. Given an aggregated files it will sort and then truncate
// based on the search options.
func SortAndTruncateFiles(files []FileMatch, opts *SearchOptions) []FileMatch {
	SortFiles(files)
	truncator, _ := NewDisplayTruncator(opts)
	files, _ = truncator(files)
	return files
}

// DisplayTruncator is a stateful function which enforces Document and Match
// display limits by truncating and mutating before. hasMore is true until the
// limits are exhausted. Once hasMore is false each subsequent call will
// return an empty after and hasMore false.
type DisplayTruncator func(before []FileMatch) (after []FileMatch, hasMore bool)

// NewDisplayTruncator will return a DisplayTruncator which enforces the limits in
// opts. If there are no limits to enforce, hasLimits is false and there is no
// need to call DisplayTruncator.
func NewDisplayTruncator(opts *SearchOptions) (_ DisplayTruncator, hasLimits bool) {
	docLimit := opts.MaxDocDisplayCount
	docLimited := docLimit > 0

	matchLimit := opts.MaxMatchDisplayCount
	matchLimited := matchLimit > 0

	done := false

	if !docLimited && !matchLimited {
		return func(fm []FileMatch) ([]FileMatch, bool) {
			return fm, true
		}, false
	}

	return func(fm []FileMatch) ([]FileMatch, bool) {
		if done {
			return nil, false
		}

		if docLimited {
			if len(fm) >= docLimit {
				done = true
				fm = fm[:docLimit]
			}
			docLimit -= len(fm)
		}

		if matchLimited {
			fm, matchLimit = limitMatches(fm, matchLimit, opts.ChunkMatches)
			if matchLimit <= 0 {
				done = true
			}
		}

		return fm, !done
	}, true
}

func limitMatches(files []FileMatch, limit int, chunkMatches bool) ([]FileMatch, int) {
	var limiter func(file *FileMatch, limit int) int
	if chunkMatches {
		limiter = limitChunkMatches
	} else {
		limiter = limitLineMatches
	}
	for i := range files {
		limit = limiter(&files[i], limit)
		if limit <= 0 {
			return files[:i+1], 0
		}
	}
	return files, limit
}

// Limit the number of ChunkMatches in the given FileMatch, returning the
// remaining limit, if any.
func limitChunkMatches(file *FileMatch, limit int) int {
	for i := range file.ChunkMatches {
		cm := &file.ChunkMatches[i]
		if len(cm.Ranges) > limit {
			// We potentially need to effect the limit upon 3 different fields:
			// Ranges, SymbolInfo, and Content.

			// Content is the most complicated: we need to remove the last N
			// lines from it, where N is the difference between the line number
			// of the end of the old last Range and that of the new last Range.
			// This calculation is correct in the presence of both context lines
			// and multiline Ranges, taking into account that Content never has
			// a trailing newline.
			n := cm.Ranges[len(cm.Ranges)-1].End.LineNumber - cm.Ranges[limit-1].End.LineNumber
			if n > 0 {
				for b := len(cm.Content) - 1; b >= 0; b-- {
					if cm.Content[b] == '\n' {
						n -= 1
					}
					if n == 0 {
						cm.Content = cm.Content[:b]
						break
					}
				}
				if n > 0 {
					// Should be impossible.
					log.Panicf("Failed to find enough newlines when truncating Content, %d left over, %d ranges", n, len(cm.Ranges))
				}
			}

			cm.Ranges = cm.Ranges[:limit]
			if cm.SymbolInfo != nil {
				// When non-nil, SymbolInfo is specified to have the same length
				// as Ranges.
				cm.SymbolInfo = cm.SymbolInfo[:limit]
			}
		}
		if len(cm.Ranges) == limit {
			file.ChunkMatches = file.ChunkMatches[:i+1]
			limit = 0
			break
		}
		limit -= len(cm.Ranges)
	}
	return limit
}

// Limit the number of LineMatches in the given FileMatch, returning the
// remaining limit, if any.
func limitLineMatches(file *FileMatch, limit int) int {
	for i := range file.LineMatches {
		lm := &file.LineMatches[i]
		if len(lm.LineFragments) > limit {
			lm.LineFragments = lm.LineFragments[:limit]
		}
		if len(lm.LineFragments) == limit {
			file.LineMatches = file.LineMatches[:i+1]
			limit = 0
			break
		}
		limit -= len(lm.LineFragments)
	}
	return limit
}
