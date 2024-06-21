package graphql

import (
	"context"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/types"
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
	kind       resolverstubs.SymbolUsageKind
	usageRange resolverstubs.UsageRangeResolver
}

var _ resolverstubs.UsageResolver = &usageResolver{}

func NewSyntacticUsageResolver(usage codenav.SyntacticMatch, repository types.Repo, revision api.CommitID) resolverstubs.UsageResolver {
	var kind resolverstubs.SymbolUsageKind
	if scip.SymbolRole_Definition.Matches(usage.Occurrence) {
		kind = resolverstubs.UsageKindDefinition
	} else {
		kind = resolverstubs.UsageKindReference
	}
	return &usageResolver{
		symbol: &symbolInformationResolver{
			name:       usage.Occurrence.Symbol,
			provenance: resolverstubs.ProvenanceSyntactic,
		},
		kind: kind,
		usageRange: &usageRangeResolver{
			repository: repository,
			revision:   revision,
			path:       usage.Path,
			range_:     usage.Range(),
		},
	}
}

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

func (u *usageResolver) UsageKind() resolverstubs.SymbolUsageKind {
	return u.kind
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
	repository types.Repo
	revision   api.CommitID
	path       string
	range_     scip.Range
}

var _ resolverstubs.UsageRangeResolver = &usageRangeResolver{}

func (u *usageRangeResolver) Repository() string {
	return string(u.repository.Name)
}

func (u *usageRangeResolver) Revision() string {
	return string(u.revision)
}

func (u *usageRangeResolver) Path() string {
	return u.path
}

func (u *usageRangeResolver) Range() resolverstubs.RangeResolver {
	return &rangeResolver{
		lspRange: convertRange(shared.TranslateRange(u.range_)),
	}
}
