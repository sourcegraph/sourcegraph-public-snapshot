package graphql

import (
	"context"
	"fmt"
	"strings"
	"sync"

	genslices "github.com/life4/genesis/slices"
	conciter "github.com/sourcegraph/conc/iter"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers/gitresolvers"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

func NewPreciseUsageResolvers(ctx context.Context, gitserverClient gitserver.Client, usages []shared.UploadUsage) ([]resolverstubs.UsageResolver, error) {
	uniquePaths := genslices.Uniq(genslices.Map(usages, shared.UploadUsage.RepoCommitPath))
	contentMap := map[types.RepoCommitPath][]string{}
	contentMapLock := sync.Mutex{}
	conciter.Iterator[types.RepoCommitPath]{MaxGoroutines: 4}.ForEach(uniquePaths, func(rcp *types.RepoCommitPath) {
		byteContents, err := gitresolvers.GetFullContents(ctx, gitserverClient,
			api.RepoName(rcp.Repo), api.CommitID(rcp.Commit), core.NewRepoRelPathUnchecked(rcp.Path))
		if err != nil {
			return
		}
		splitContents := strings.Split(string(byteContents), "\n")
		contentMapLock.Lock()
		contentMap[*rcp] = splitContents
		contentMapLock.Unlock()
	})
	out := make([]resolverstubs.UsageResolver, 0, len(usages))
	var err error
	for _, usage := range usages {
		contents, ok := contentMap[usage.RepoCommitPath()]
		if !ok {
			continue
		}
		start := usage.TargetRange.Start.Line
		if start < 0 {
			err = errors.CombineErrors(err, errors.Newf("malformed usage: %+v", usage))
			continue
		}
		if len(contents) <= start {
			err = errors.CombineErrors(err, errors.Newf("start position of usage (%+v) is out-of-bounds (file line count: %d)",
				usage, len(contents)))
			continue
		}
		kind := convertKind(usage.Kind)
		lineContent := contents[start]
		out = append(out, &usageResolver{
			symbol:             &symbolInformationResolver{name: usage.Symbol},
			provenance:         codenav.ProvenancePrecise,
			kind:               kind,
			surroundingContent: lineContent,
			usageRange: &usageRangeResolver{
				repoName: api.RepoName(usage.Upload.RepositoryName),
				revision: api.CommitID(usage.TargetCommit),
				path:     usage.Path,
				range_:   usage.TargetRange.ToSCIPRange(),
			},
			dataSource: &usage.Upload.Indexer,
		})
	}

	return out, err
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

func (u *usageResolver) UsageRange(ctx context.Context) resolverstubs.UsageRangeResolver {
	return u.usageRange
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
		range_: u.range_,
	}
}
