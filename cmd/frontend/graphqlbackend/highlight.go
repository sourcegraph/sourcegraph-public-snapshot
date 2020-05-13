package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type highlightedRange struct {
	line      int32
	character int32
	length    int32
}

func (h *highlightedRange) Line() int32      { return h.line }
func (h *highlightedRange) Character() int32 { return h.character }
func (h *highlightedRange) Length() int32    { return h.length }

type highlightedString struct {
	value      string
	highlights []*highlightedRange
}

func (s *highlightedString) Value() string                   { return s.value }
func (s *highlightedString) Highlights() []*highlightedRange { return s.highlights }

func fromVCSHighlights(vcsHighlights []git.Highlight) []*highlightedRange {
	highlights := make([]*highlightedRange, len(vcsHighlights))
	for i, vh := range vcsHighlights {
		highlights[i] = &highlightedRange{
			line:      int32(vh.Line),
			character: int32(vh.Character),
			length:    int32(vh.Length),
		}
	}
	return highlights
}

// A highlighting implementation for diff nodes. Returns the highlighted lines for both base and head, aborted status and an optional error.
type DiffHighlighter interface {
	Highlight(ctx context.Context, args *HighlightArgs) (highlightedBase []string, highlightedHead []string, aborted bool, err error)
}

type HighlightArgs struct {
	DisableTimeout     bool
	IsLightTheme       bool
	HighlightLongLines bool
}
