package search

import (
	"testing"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func TestFormatDiff(t *testing.T) {
	diff := []*diff.FileDiff{{
		OrigName: "internal/actor/actor.go",
		NewName:  "internal/actor/actor.go",
		Hunks: []*diff.Hunk{{
			OrigStartLine: 49,
			OrigLines:     0,
			NewStartLine:  61,
			StartPosition: 4,
			Section:       "func (a *Actor) IsInternal() bool {",
			Body: []byte(
				"+// types.User using the fetcher, which is likely a *database.UserStore.\n" +
					"+func (a *Actor) User(ctx context.Context, fetcher userFetcher) (*types.User, error) {\n" +
					"+\ta.userOnce.Do(func() {\n" +
					"\t\ta.user, a.userErr = fetcher.GetByID(ctx, a.UID)\n",
			),
		}},
	}}

	highlights := map[int]protocol.FileDiffHighlight{
		0: {
			OldFile: protocol.Ranges{{
				Start: protocol.Location{Line: 0, Offset: 6, Column: 6},
				End:   protocol.Location{Line: 0, Offset: 11, Column: 11},
			}},
			HunkHighlights: map[int]protocol.HunkHighlight{
				0: protocol.HunkHighlight{
					LineHighlights: map[int]protocol.Ranges{
						1: protocol.Ranges{{
							Start: protocol.Location{Line: 0, Offset: 0, Column: 0},
							End:   protocol.Location{Line: 0, Offset: 4, Column: 4},
						}},
					},
				},
			},
		},
	}

	formatted, newHighlights := FormatDiff(diff, highlights)
	require.Equal(t, formatted, 
	"+// types.User using the fetcher, which is likely a *database.UserStore.\n"+
		"+func (a *Actor) User(ctx context.Context, fetcher userFetcher) (*types.User, error) {\n"+
		"+\ta.userOnce.Do(func() {\n"+
		"\t\ta.user, a.userErr = fetcher.GetByID(ctx, a.UID)\n"
	)

}
