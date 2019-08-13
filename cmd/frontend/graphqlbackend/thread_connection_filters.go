package graphqlbackend

import "context"

// ThreadConnectionFilters implements the ThreadConnectionFilters GraphQL type.
type ThreadConnectionFilters interface {
	Repository(context.Context) ([]RepositoryFilter, error)
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
