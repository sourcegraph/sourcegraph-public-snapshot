package graphql

import (
	"context"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

type usageConnectionResolver struct {
	nodes    []resolverstubs.UsageResolver
	pageInfo resolverstubs.PageInfo
}

var _ resolverstubs.UsageConnectionResolver = &usageConnectionResolver{}

func (u *usageConnectionResolver) Nodes(ctx context.Context) ([]resolverstubs.UsageResolver, error) {
	return u.nodes, nil
}

func (u *usageConnectionResolver) PageInfo() resolverstubs.PageInfo {
	return u.pageInfo
}

type usageResolver struct {
	symbol     resolverstubs.SymbolInformationResolver
	usageRange resolverstubs.UsageRangeResolver
}

var _ resolverstubs.UsageResolver = &usageResolver{}

func (u *usageResolver) Symbol(ctx context.Context) (resolverstubs.SymbolInformationResolver, error) {
	return u.symbol, nil
}

func (u *usageResolver) UsageRange(ctx context.Context) (resolverstubs.UsageRangeResolver, error) {
	return u.usageRange, nil
}

func (u *usageResolver) SurroundingContent(_ context.Context, args *struct {
	*resolverstubs.SurroundingLines `json:"surroundingLines"`
}) (*string, error) {
	//TODO implement me
	panic("implement me")
}

type symbolInformationResolver struct {
	name       string
	provenance resolverstubs.CodeGraphDataProvenance
}

var _ resolverstubs.SymbolInformationResolver = &symbolInformationResolver{}

func (s *symbolInformationResolver) Name() (string, error) {
	return s.name, nil
}

func (s *symbolInformationResolver) Documentation() (*[]string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *symbolInformationResolver) Provenance() (resolverstubs.CodeGraphDataProvenance, error) {
	return s.provenance, nil
}

func (s *symbolInformationResolver) DataSource() *string {
	//TODO implement me
	panic("implement me")
}

type usageRangeResolver struct {
	repository string
	revision   string
	path       string
	rx         *rangeResolver
}

var _ resolverstubs.UsageRangeResolver = &usageRangeResolver{}

func (u *usageRangeResolver) Repository() string {
	return u.repository
}

func (u *usageRangeResolver) Revision() string {
	return u.revision
}

func (u *usageRangeResolver) Path() string {
	return u.path
}

func (u *usageRangeResolver) Range() resolverstubs.RangeResolver {
	return u.rx
}
