package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type SideEffect struct {
	Title_  string
	Detail_ string
}

func (v SideEffect) Title() string { return v.Title_ }

func (v SideEffect) Detail() *string {
	if v.Detail_ == "" {
		return nil
	}
	return &v.Detail_
}

type SideEffectInput struct {
	Title  string  `json:"title"`
	Detail *string `json:"detail,omitempty"`
}

func FromSideEffectInput(input ...SideEffectInput) []SideEffect {
	out := make([]SideEffect, len(input))
	for i, in := range input {
		out[i] = SideEffect{Title_: in.Title}
		if in.Detail != nil {
			out[i].Detail_ = *in.Detail
		}
	}
	return out
}

type SideEffectConnectionArgs struct {
	graphqlutil.ConnectionArgs
	Other bool
}

type hasSideEffects interface {
	SideEffects(context.Context, *SideEffectConnectionArgs) (SideEffectConnection, error)
}

// SideEffectConnection implements the SideEffectConnection GraphQL type.
type SideEffectConnection []SideEffect

func (c SideEffectConnection) Nodes(context.Context) ([]SideEffect, error) {
	return []SideEffect(c), nil
}

func (c SideEffectConnection) TotalCount(context.Context) (int32, error) { return int32(len(c)), nil }

func (c SideEffectConnection) PageInfo(context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}
