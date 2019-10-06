package graphqlbackend

import "context"

// ThreadConnectionFilters implements the ThreadConnectionFilters GraphQL type.
type ThreadConnectionFilters interface {
	Repository(context.Context) ([]RepositoryFilter, error)
	Label(context.Context) ([]LabelFilter, error)
	OpenCount(context.Context) (int32, error)
	ClosedCount(context.Context) (int32, error)
}

// RepositoryFilter implements the RepositoryFilter GraphQL type.
type RepositoryFilter struct {
	Repository_ *RepositoryResolver
	Count_      int32
	IsApplied_  bool
}

func (v RepositoryFilter) Repository() *RepositoryResolver { return v.Repository_ }
func (v RepositoryFilter) Count() *int32                   { return &v.Count_ }
func (v RepositoryFilter) IsApplied() bool                 { return v.IsApplied_ }

// LabelFilter implements the LabelFilter GraphQL type.
type LabelFilter struct {
	Label_     Label
	LabelName_ string
	Count_     int32
	IsApplied_ bool
}

func (v LabelFilter) Label() Label      { return v.Label_ }
func (v LabelFilter) LabelName() string { return v.LabelName_ }
func (v LabelFilter) Count() *int32     { return &v.Count_ }
func (v LabelFilter) IsApplied() bool   { return v.IsApplied_ }
