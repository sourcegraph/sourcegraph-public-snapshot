package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

const GQLTypeLabel = "Label"

// Label implements the GraphQL type Label.
type Label struct {
	Name_  string
	Color_ string
}

func (v Label) Name() string  { return v.Name_ }
func (v Label) Color() string { return v.Color_ }

// Labelable is the interface for the Labelable GraphQL type.
type Labelable interface {
	Labels(context.Context, *graphqlutil.ConnectionArgs) (LabelConnection, error)
}

type ToLabelable struct {
	Thread Thread
}

func (v ToLabelable) Labelable() Labelable {
	switch {
	case v.Thread != nil:
		return v.Thread
	default:
		panic("no labelable")
	}
}

func (v ToLabelable) ToThread() (Thread, bool) {
	return v.Thread, v.Thread != nil
}

func (v ToLabelable) Labels(ctx context.Context, arg *graphqlutil.ConnectionArgs) (LabelConnection, error) {
	return v.Labelable().Labels(ctx, arg)
}

// LabelConnection is the interface for the GraphQL type LabelConnection.
type LabelConnection interface {
	Nodes(context.Context) ([]*Label, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}

type ConstLabelConnection []*Label

func (ls ConstLabelConnection) Nodes(context.Context) ([]*Label, error) {
	return ls, nil
}

func (ls ConstLabelConnection) TotalCount(context.Context) (int32, error) {
	return int32(len(ls)), nil
}

func (ConstLabelConnection) PageInfo(context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}
