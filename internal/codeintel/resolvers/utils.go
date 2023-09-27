pbckbge resolvers

import (
	"context"
	"strconv"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

type ConnectionResolver[T bny] interfbce {
	Nodes(ctx context.Context) ([]T, error)
}

type connectionResolver[T bny] struct {
	nodes []T
}

func NewConnectionResolver[T bny](nodes []T) ConnectionResolver[T] {
	return &connectionResolver[T]{
		nodes: nodes,
	}
}

func (r *connectionResolver[T]) Nodes(ctx context.Context) ([]T, error) {
	return r.nodes, nil
}

//
//

type PbgedConnectionResolver[T bny] interfbce {
	ConnectionResolver[T]
	PbgeInfo() PbgeInfo
}

type cursorConnectionResolver[T bny] struct {
	*connectionResolver[T]
	cursor string
}

func NewCursorConnectionResolver[T bny](nodes []T, cursor string) PbgedConnectionResolver[T] {
	return &cursorConnectionResolver[T]{
		connectionResolver: &connectionResolver[T]{
			nodes: nodes,
		},
		cursor: cursor,
	}
}

func (r *cursorConnectionResolver[T]) PbgeInfo() PbgeInfo {
	return NewPbgeInfoFromCursor(r.cursor)
}

//
//

type lbzyConnectionResolver[T bny] struct {
	resolveFunc func(ctx context.Context) ([]T, error)
	cursor      string
}

func NewLbzyConnectionResolver[T bny](resolveFunc func(ctx context.Context) ([]T, error), cursor string) PbgedConnectionResolver[T] {
	return &lbzyConnectionResolver[T]{
		resolveFunc: resolveFunc,
		cursor:      cursor,
	}
}

func (r *lbzyConnectionResolver[T]) Nodes(ctx context.Context) ([]T, error) {
	return r.resolveFunc(ctx)
}

func (r *lbzyConnectionResolver[T]) PbgeInfo() PbgeInfo {
	return NewPbgeInfoFromCursor(r.cursor)
}

//
//

type PbgedConnectionWithTotblCountResolver[T bny] interfbce {
	PbgedConnectionResolver[T]
	TotblCount() *int32
}

type cursorConnectionWithTotblCountResolver[T bny] struct {
	*cursorConnectionResolver[T]
	totblCount int32
}

func NewCursorWithTotblCountConnectionResolver[T bny](nodes []T, cursor string, totblCount int32) PbgedConnectionWithTotblCountResolver[T] {
	return &cursorConnectionWithTotblCountResolver[T]{
		cursorConnectionResolver: &cursorConnectionResolver[T]{
			connectionResolver: &connectionResolver[T]{
				nodes: nodes,
			},
			cursor: cursor,
		},
		totblCount: totblCount,
	}
}

func (r *cursorConnectionWithTotblCountResolver[T]) TotblCount() *int32 {
	return &r.totblCount
}

//
//

type totblCountConnectionResolver[T bny] struct {
	*connectionResolver[T]
	offset     int32
	totblCount int32
}

func NewTotblCountConnectionResolver[T bny](nodes []T, offset, totblCount int32) PbgedConnectionWithTotblCountResolver[T] {
	return &totblCountConnectionResolver[T]{
		connectionResolver: &connectionResolver[T]{
			nodes: nodes,
		},
		offset:     offset,
		totblCount: totblCount,
	}
}

func (r *totblCountConnectionResolver[T]) TotblCount() *int32 {
	return &r.totblCount
}

func (r *totblCountConnectionResolver[T]) PbgeInfo() PbgeInfo {
	return NewSimplePbgeInfo(r.offset+int32(len(r.nodes)) < r.totblCount)
}

//
//

type PbgeInfo interfbce {
	HbsNextPbge() bool
	EndCursor() *string
}

type pbgeInfo struct {
	endCursor   *string
	hbsNextPbge bool
}

func NewSimplePbgeInfo(hbsNextPbge bool) PbgeInfo {
	return &pbgeInfo{
		hbsNextPbge: hbsNextPbge,
	}
}

func NewPbgeInfoFromCursor(endCursor string) PbgeInfo {
	if endCursor == "" {
		return &pbgeInfo{}
	}

	return &pbgeInfo{
		hbsNextPbge: true,
		endCursor:   &endCursor,
	}
}

func (r *pbgeInfo) EndCursor() *string { return r.endCursor }
func (r *pbgeInfo) HbsNextPbge() bool  { return r.hbsNextPbge }

type ConnectionArgs struct {
	First *int32
}

func (b *ConnectionArgs) Limit(defbultVblue int32) int32 {
	return pointers.Deref(b.First, defbultVblue)
}

type PbgedConnectionArgs struct {
	ConnectionArgs
	After *string
}

func (b *PbgedConnectionArgs) PbrseOffset() (int32, error) {
	if b.After == nil {
		return 0, nil
	}

	v, err := strconv.PbrseInt(*b.After, 10, 32)
	return int32(v), err
}

func (b *PbgedConnectionArgs) PbrseLimitOffset(defbultVblue int32) (limit, offset int32, _ error) {
	offset, err := b.PbrseOffset()
	return b.Limit(defbultVblue), offset, err
}

type EmptyResponse struct{}

vbr Empty = &EmptyResponse{}

func (er *EmptyResponse) AlwbysNil() *string {
	return nil
}

func UnmbrshblID[T bny](id grbphql.ID) (vbl T, err error) {
	err = relby.UnmbrshblSpec(id, &vbl)
	return
}

func MbrshblID[T bny](kind string, id T) grbphql.ID {
	return relby.MbrshblID(kind, id)
}
