package graphql

import (
	"context"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type usageConnectionResolver struct {
}

var _ resolverstubs.UsageConnectionResolver = &usageConnectionResolver{}

func (u *usageConnectionResolver) Nodes(ctx context.Context) ([]resolverstubs.UsageResolver, error) {
	//TODO implement me
	panic("implement me")
}

func (u *usageConnectionResolver) PageInfo() resolverstubs.PageInfo {
	//TODO implement me
	panic("implement me")
}

type usageResolver struct {
}

var _ resolverstubs.UsageResolver = &usageResolver{}

func (u *usageResolver) Symbol(ctx context.Context) (resolverstubs.SymbolInformationResolver, error) {
	//TODO implement me
	panic("implement me")
}

func (u *usageResolver) UsageRange(ctx context.Context) (resolverstubs.UsageRangeResolver, error) {
	//TODO implement me
	panic("implement me")
}

func (u *usageResolver) SurroundingContent(_ context.Context, args *struct {
	*resolverstubs.SurroundingLines `json:"surroundingLines"`
}) (*string, error) {
	//TODO implement me
	panic("implement me")
}

type symbolInformationResolver struct {
}

var _ resolverstubs.SymbolInformationResolver = &symbolInformationResolver{}

func (s *symbolInformationResolver) Name() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *symbolInformationResolver) Documentation() (*[]string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *symbolInformationResolver) Provenance() (resolverstubs.CodeGraphDataProvenance, error) {
	//TODO implement me
	panic("implement me")
}

func (s *symbolInformationResolver) DataSource() *string {
	//TODO implement me
	panic("implement me")
}

type usageRangeResolver struct{}

var _ resolverstubs.UsageRangeResolver = &usageRangeResolver{}

func (u *usageRangeResolver) Repository() string {
	//TODO implement me
	panic("implement me")
}

func (u *usageRangeResolver) Revision() string {
	//TODO implement me
	panic("implement me")
}

func (u *usageRangeResolver) Path() string {
	//TODO implement me
	panic("implement me")
}

func (u *usageRangeResolver) Range() resolverstubs.RangeResolver {
	//TODO implement me
	panic("implement me")
}
