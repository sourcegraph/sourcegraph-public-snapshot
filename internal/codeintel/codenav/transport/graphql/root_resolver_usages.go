package graphql

import (
	"context"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
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
	symbol             *symbolInformationResolver
	provenance         codenav.CodeGraphDataProvenance
	kind               resolverstubs.SymbolUsageKind
	surroundingContent string
	usageRange         *usageRangeResolver
}

var _ resolverstubs.UsageResolver = &usageResolver{}

func NewSyntacticUsageResolver(usage codenav.SyntacticMatch, repoName api.RepoName, revision api.CommitID) resolverstubs.UsageResolver {
	var kind resolverstubs.SymbolUsageKind
	if usage.IsDefinition {
		kind = resolverstubs.UsageKindDefinition
	} else {
		kind = resolverstubs.UsageKindReference
	}
	return &usageResolver{
		symbol: &symbolInformationResolver{
			name: usage.Symbol,
		},
		provenance:         codenav.ProvenanceSyntactic,
		kind:               kind,
		surroundingContent: usage.SurroundingContent,
		usageRange: &usageRangeResolver{
			repoName: repoName,
			revision: revision,
			path:     usage.Path,
			range_:   usage.Range,
		},
	}
}

func NewSearchBasedUsageResolver(usage codenav.SearchBasedMatch, repoName api.RepoName, revision api.CommitID) resolverstubs.UsageResolver {
	var kind resolverstubs.SymbolUsageKind
	if usage.IsDefinition {
		kind = resolverstubs.UsageKindDefinition
	} else {
		kind = resolverstubs.UsageKindReference
	}
	return &usageResolver{
		symbol:             nil,
		provenance:         codenav.ProvenanceSearchBased,
		kind:               kind,
		surroundingContent: usage.SurroundingContent,
		usageRange: &usageRangeResolver{
			repoName: repoName,
			revision: revision,
			path:     usage.Path,
			range_:   usage.Range,
		},
	}
}

func (u *usageResolver) Symbol(ctx context.Context) (resolverstubs.SymbolInformationResolver, error) {
	if u.symbol == nil {
		// NOTE: if I try to directly return u.symbol, I get a panic in the resolver.
		return nil, nil
	}
	return u.symbol, nil
}

func (u *usageResolver) Provenance(ctx context.Context) (codenav.CodeGraphDataProvenance, error) {
	return u.provenance, nil
}

func (u *usageResolver) DataSource() *string {
	//TODO implement me
	// NOTE: For search-based usages it would be good to return if this usage was found via Zoekt or Searcher
	panic("implement me")
}

func (u *usageResolver) UsageRange(ctx context.Context) (resolverstubs.UsageRangeResolver, error) {
	return u.usageRange, nil
}

func (u *usageResolver) SurroundingContent(ctx context.Context) string {
	return u.surroundingContent
}

func (u *usageResolver) UsageKind() resolverstubs.SymbolUsageKind {
	return u.kind
}

type symbolInformationResolver struct {
	name string
}

var _ resolverstubs.SymbolInformationResolver = &symbolInformationResolver{}

func (s *symbolInformationResolver) Name() (string, error) {
	return s.name, nil
}

func (s *symbolInformationResolver) Documentation() (*[]string, error) {
	//TODO implement me
	panic("implement me")
}

type usageRangeResolver struct {
	repoName api.RepoName
	revision api.CommitID
	path     core.RepoRelPath
	range_   scip.Range
}

var _ resolverstubs.UsageRangeResolver = &usageRangeResolver{}

func (u *usageRangeResolver) Repository() string {
	return string(u.repoName)
}

func (u *usageRangeResolver) Revision() string {
	return string(u.revision)
}

func (u *usageRangeResolver) Path() string {
	return u.path.RawValue()
}

func (u *usageRangeResolver) Range() resolverstubs.RangeResolver {
	return &rangeResolver{
		lspRange: convertRange(shared.TranslateRange(u.range_)),
	}
}
