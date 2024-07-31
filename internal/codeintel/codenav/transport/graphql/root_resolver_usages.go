package graphql

import (
	"context"
	"fmt"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers/gitresolvers"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
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

func NewPreciseUsageResolver(ctx context.Context, usage shared.UploadUsage, locResolver *gitresolvers.CachedLocationResolver) (resolverstubs.UsageResolver, error) {
	kind := convertKind(usage.Kind)
	repoID := api.RepoID(usage.Upload.RepositoryID)
	res, err := locResolver.Path(ctx, repoID, usage.TargetCommit, usage.Path.RawValue(), false)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve path in usage")
	}
	start := usage.TargetRange.Start.ToSCIPPosition()
	// TODO(id: GRAPH-781): Optimize the code to:
	// - Avoid refetching/re-splitting the contents repeatedly for the same file
	// - Avoid line-splitting beyond the needed ranges.
	// We don't want to make this computation _lazy_, as that would require making
	// the GraphQL field optional (it is currently non-optional).
	content, err := res.Content(ctx, &resolverstubs.GitTreeContentPageArgs{StartLine: &start.Line, EndLine: pointers.Ptr(start.Line + 1)})
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch contents")
	}

	return &usageResolver{
		symbol:             &symbolInformationResolver{name: usage.Symbol},
		provenance:         codenav.ProvenancePrecise,
		kind:               kind,
		surroundingContent: content,
		usageRange: &usageRangeResolver{
			repoName: api.RepoName(usage.Upload.RepositoryName),
			revision: api.CommitID(usage.TargetCommit),
			path:     usage.Path,
			range_:   usage.TargetRange.ToSCIPRange(),
		},
		dataSource: &usage.Upload.Indexer,
	}, nil
}

func convertKind(kind shared.UsageKind) resolverstubs.SymbolUsageKind {
	switch kind {
	case shared.UsageKindDefinition:
		return resolverstubs.UsageKindDefinition
	case shared.UsageKindImplementation:
		return resolverstubs.UsageKindImplementation
	case shared.UsageKindSuper:
		return resolverstubs.UsageKindSuper
	case shared.UsageKindReference:
		return resolverstubs.UsageKindReference
	}
	panic(fmt.Sprintf("unhandled kind of shared.UsageKind: %q", kind.String()))
}

type usageResolver struct {
	symbol             *symbolInformationResolver
	provenance         codenav.CodeGraphDataProvenance
	kind               resolverstubs.SymbolUsageKind
	surroundingContent string
	usageRange         *usageRangeResolver
	dataSource         *string
}

var _ resolverstubs.UsageResolver = &usageResolver{} //nolint:exhaustruct

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
		dataSource: nil,
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
		// TODO: Record if we got the results from Searcher or Zoekt
		dataSource: nil,
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
