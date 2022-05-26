package result

import (
	"net/url"
	"path"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// File represents all the information we need to identify a file in a repository
type File struct {
	// InputRev is the Git revspec that the user originally requested to search. It is used to
	// preserve the original revision specifier from the user instead of navigating them to the
	// absolute commit ID when they select a result.
	InputRev *string           `json:"-"`
	Repo     types.MinimalRepo `json:"-"`
	CommitID api.CommitID      `json:"-"`
	Path     string
}

func (f *File) URL() *url.URL {
	var path strings.Builder
	path.Grow(len("/@/-/blob/") + len(f.Repo.Name) + len(f.Path) + 20)
	path.WriteRune('/')
	path.WriteString(string(f.Repo.Name))
	if f.InputRev != nil && len(*f.InputRev) > 0 {
		path.WriteRune('@')
		path.WriteString(*f.InputRev)
	}
	path.WriteString("/-/blob/")
	path.WriteString(f.Path)
	return &url.URL{Path: path.String()}
}

// FileMatch represents either:
// - A collection of symbol results (len(Symbols) > 0)
// - A collection of text content results (len(LineMatches) > 0)
// - A result representing the whole file (len(Symbols) == 0 && len(LineMatches) == 0)
type FileMatch struct {
	File

	MultilineMatches []MultilineMatch
	Symbols          []*SymbolMatch `json:"-"`

	LimitHit bool
}

func (fm *FileMatch) RepoName() types.MinimalRepo {
	return fm.File.Repo
}

func (fm *FileMatch) searchResultMarker() {}

func (fm *FileMatch) ResultCount() int {
	rc := len(fm.Symbols) + len(fm.MultilineMatches)
	if rc == 0 {
		return 1 // 1 to count "empty" results like type:path results
	}
	return rc
}

// IsPathMatch returns true if a `FileMatch` has no line or symbol matches. In
// the absence of a true `PathMatch` type, we use this function as a proxy
// signal to drive `select:file` logic that deduplicates path results.
func (fm *FileMatch) IsPathMatch() bool {
	return len(fm.MultilineMatches) == 0 && len(fm.Symbols) == 0
}

func (fm *FileMatch) Select(selectPath filter.SelectPath) Match {
	switch selectPath.Root() {
	case filter.Repository:
		return &RepoMatch{
			Name: fm.Repo.Name,
			ID:   fm.Repo.ID,
		}
	case filter.File:
		fm.MultilineMatches = nil
		fm.Symbols = nil
		if len(selectPath) > 1 && selectPath[1] == "directory" {
			fm.Path = path.Clean(path.Dir(fm.Path)) + "/" // Add trailing slash for clarity.
		}
		return fm
	case filter.Symbol:
		if len(fm.Symbols) > 0 {
			fm.MultilineMatches = nil // Only return symbol match if symbols exist
			if len(selectPath) > 1 {
				filteredSymbols := SelectSymbolKind(fm.Symbols, selectPath[1])
				if len(filteredSymbols) == 0 {
					return nil // Remove file match if there are no symbol results after filtering
				}
				fm.Symbols = filteredSymbols
			}
			return fm
		}
		return nil
	case filter.Content:
		// Only return file match if line matches exist
		if len(fm.MultilineMatches) > 0 {
			fm.Symbols = nil
			return fm
		}
		return nil
	case filter.Commit:
		return nil
	}
	return nil
}

// AppendMatches appends the line matches from src as well as updating match
// counts and limit.
func (fm *FileMatch) AppendMatches(src *FileMatch) {
	fm.MultilineMatches = append(fm.MultilineMatches, src.MultilineMatches...)
	fm.Symbols = append(fm.Symbols, src.Symbols...)
	fm.LimitHit = fm.LimitHit || src.LimitHit
}

// Limit will mutate fm such that it only has limit results. limit is a number
// greater than 0.
//
//   if limit >= ResultCount then nothing is done and we return limit - ResultCount.
//   if limit < ResultCount then ResultCount becomes limit and we return 0.
func (fm *FileMatch) Limit(limit int) int {
	matchCount := len(fm.MultilineMatches)
	symbolCount := len(fm.Symbols)

	// An empty FileMatch should still count against the limit -- see *FileMatch.ResultCount()
	if matchCount == 0 && symbolCount == 0 {
		return limit - 1
	}

	if limit < matchCount {
		fm.MultilineMatches = fm.MultilineMatches[:limit]
		limit = 0
		fm.LimitHit = true
	} else {
		limit -= matchCount
	}

	if limit < symbolCount {
		fm.Symbols = fm.Symbols[:limit]
		limit = 0
		fm.LimitHit = true
	} else {
		limit -= symbolCount
	}
	return limit
}

func (fm *FileMatch) Key() Key {
	k := Key{
		TypeRank: rankFileMatch,
		Repo:     fm.Repo.Name,
		Commit:   fm.CommitID,
		Path:     fm.Path,
	}

	if fm.InputRev != nil {
		k.Rev = *fm.InputRev
	}

	return k
}

type MultilineMatch struct {
	// Preview is a possibly-multiline string that contains all the
	// lines that the match overlaps.
	// The number of lines in Preview should be End.Line - Start.Line + 1
	Preview string
	Range   Range
}

func MultilineSliceAsLineMatchSlice(matches []MultilineMatch) []*LineMatch {
	lineMatches := make([]*LineMatch, 0, len(matches))
	for _, m := range matches {
		lineMatches = append(lineMatches, m.AsLineMatches()...)
	}
	sort.Slice(lineMatches, func(i, j int) bool {
		return lineMatches[i].LineNumber < lineMatches[j].LineNumber
	})
	res := lineMatches[:0]
	for i, lm := range lineMatches {
		if i == 0 {
			res = append(res, lm)
			continue
		}
		last := len(res) - 1
		if lm.LineNumber == res[last].LineNumber {
			res[last].OffsetAndLengths = append(res[last].OffsetAndLengths, lm.OffsetAndLengths...)
		} else {
			res = append(res, lm)
		}
	}
	return res
}

func (m MultilineMatch) AsLineMatches() []*LineMatch {
	lines := strings.Split(m.Preview, "\n")
	lineMatches := make([]*LineMatch, 0, len(lines))
	for i, line := range lines {
		offset := int32(0)
		if i == 0 {
			offset = int32(m.Range.Start.Column)
		}
		length := int32(utf8.RuneCountInString(line)) - offset
		if i == len(lines)-1 {
			length = int32(m.Range.End.Column) - offset
		}
		lineMatches = append(lineMatches, &LineMatch{
			Preview:          line,
			LineNumber:       int32(m.Range.Start.Line) + int32(i),
			OffsetAndLengths: [][2]int32{{offset, length}},
		})
	}
	return lineMatches
}

type LineMatch struct {
	// Preview is the full single line these offsets belong to
	Preview          string
	OffsetAndLengths [][2]int32
	LineNumber       int32
}
