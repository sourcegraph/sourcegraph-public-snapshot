package graphql

import (
	"fmt"
	"strconv"

	"github.com/sourcegraph/go-lsp"
)

type PositionResolver interface {
	Line() int32
	Character() int32
}

type positionResolver struct{ pos lsp.Position }

func NewPositionResolver(pos lsp.Position) PositionResolver {
	return &positionResolver{pos: pos}
}

func (r *positionResolver) Line() int32      { return int32(r.pos.Line) }
func (r *positionResolver) Character() int32 { return int32(r.pos.Character) }

func (r *positionResolver) urlFragment(forceIncludeCharacter bool) string {
	if !forceIncludeCharacter && r.pos.Character == 0 {
		return strconv.Itoa(r.pos.Line + 1)
	}
	return fmt.Sprintf("%d:%d", r.pos.Line+1, r.pos.Character+1)
}
