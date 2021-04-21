package result

import (
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type FileMatch struct {
	Path        string
	LineMatches []*LineMatch
	LimitHit    bool

	Symbols  []*SymbolMatch `json:"-"`
	Repo     types.RepoName `json:"-"`
	CommitID api.CommitID   `json:"-"`

	// InputRev is the Git revspec that the user originally requested to search. It is used to
	// preserve the original revision specifier from the user instead of navigating them to the
	// absolute commit ID when they select a result.
	InputRev *string `json:"-"`
}

func (fm *FileMatch) searchResultMarker() {}

func (fm *FileMatch) URL() string {
	var b strings.Builder
	var ref string
	if fm.InputRev != nil {
		ref = url.QueryEscape(*fm.InputRev)
	}
	b.Grow(len(fm.Repo.Name) + len(ref) + len(fm.Path) + len("git://?#"))
	b.WriteString("git://")
	b.WriteString(string(fm.Repo.Name))
	if ref != "" {
		b.WriteByte('?')
		b.WriteString(ref)
	}
	b.WriteByte('#')
	b.WriteString(fm.Path)
	return b.String()
}

func (fm *FileMatch) ResultCount() int {
	rc := len(fm.Symbols)
	for _, m := range fm.LineMatches {
		rc += len(m.OffsetAndLengths)
	}
	if rc == 0 {
		return 1 // 1 to count "empty" results like type:path results
	}
	return rc
}

func (fm *FileMatch) Select(t filter.SelectPath) Match {
	switch t.Type {
	case filter.Repository:
		return &RepoMatch{
			Name: fm.Repo.Name,
			ID:   fm.Repo.ID,
		}
	case filter.File:
		fm.LineMatches = nil
		fm.Symbols = nil
		return fm
	case filter.Symbol:
		if len(fm.Symbols) > 0 {
			fm.LineMatches = nil // Only return symbol match if symbols exist
			if len(t.Fields) > 0 {
				filteredSymbols := SelectSymbolKind(fm.Symbols, t.Fields[0])
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
		if len(fm.LineMatches) > 0 {
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
	fm.LineMatches = append(fm.LineMatches, src.LineMatches...)
	fm.Symbols = append(fm.Symbols, src.Symbols...)
	fm.LimitHit = fm.LimitHit || src.LimitHit
}

// Limit will mutate fm such that it only has limit results. limit is a number
// greater than 0.
//
//   if limit >= ResultCount then nothing is done and we return limit - ResultCount.
//   if limit < ResultCount then ResultCount becomes limit and we return 0.
func (fm *FileMatch) Limit(limit int) int {
	// Check if we need to limit.
	if after := limit - fm.ResultCount(); after >= 0 {
		return after
	}

	// Invariant: limit > 0
	for i, m := range fm.LineMatches {
		after := limit - len(m.OffsetAndLengths)
		if after <= 0 {
			fm.Symbols = nil
			fm.LineMatches = fm.LineMatches[:i+1]
			m.OffsetAndLengths = m.OffsetAndLengths[:limit]
			return 0
		}
		limit = after
	}

	fm.Symbols = fm.Symbols[:limit]
	return 0
}

// LineMatch is the struct used by vscode to receive search results for a line
type LineMatch struct {
	Preview          string
	OffsetAndLengths [][2]int32
	LineNumber       int32
}
