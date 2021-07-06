package graphql

import (
	"context"
	"sort"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func (r *symbolUsageResolver) Callers(ctx context.Context) ([]gql.SymbolCallerEdgeResolver, error) {
	locations, _, err := r.symbol.references(ctx)
	if err != nil {
		return nil, err
	}

	type callerInfo struct {
		sig       git.Signature
		locations []resolvers.AdjustedLocation
	}
	callers := map[string] /* email */ *callerInfo{}
	for _, loc := range locations {
		callerSigs, err := getLocationBlameAuthors(ctx, loc)
		if err != nil {
			return nil, err
		}

		for _, callerSig := range callerSigs {
			info := callers[callerSig.Email]
			if info == nil {
				info = &callerInfo{sig: callerSig}
			}
			info.locations = append(info.locations, loc)
			callers[callerSig.Email] = info
		}
	}

	// TODO(sqs): dedupe by user (eg one user might have many emails)
	edges := make([]*symbolCallerEdgeResolver, 0, len(callers))
	for _, callerInfo := range callers {
		edges = append(edges, &symbolCallerEdgeResolver{
			person:           gql.NewPersonResolver(dbconn.Global, callerInfo.sig.Name, callerInfo.sig.Email, true),
			locations:        callerInfo.locations,
			locationResolver: r.locationResolver,
		})

	}

	// Sort by #locations.
	sort.Slice(edges, func(i, j int) bool {
		return len(edges[i].locations) > len(edges[j].locations)
	})

	// Convert to slice of interface type.
	edgeIfaces := make([]gql.SymbolCallerEdgeResolver, len(edges))
	for i, edge := range edges {
		edgeIfaces[i] = edge
	}

	return edgeIfaces, nil
}

func getLocationBlameAuthors(ctx context.Context, loc resolvers.AdjustedLocation) ([]git.Signature, error) {
	hunks, err := git.BlameFile(ctx, api.RepoName(loc.Dump.RepositoryName), loc.Path, &git.BlameOptions{
		NewestCommit: api.CommitID(loc.Dump.Commit),
		StartLine:    loc.AdjustedRange.Start.Line + 1,
		EndLine:      loc.AdjustedRange.End.Line + 1,
	})
	if err != nil {
		return nil, err
	}

	authors := make([]git.Signature, len(hunks))
	for i, hunk := range hunks {
		authors[i] = hunk.Author
	}
	return authors, nil
}

type symbolCallerEdgeResolver struct {
	person    *gql.PersonResolver
	locations []resolvers.AdjustedLocation

	locationResolver *CachedLocationResolver
}

func (r *symbolCallerEdgeResolver) Person() *gql.PersonResolver { return r.person }

func (r *symbolCallerEdgeResolver) Locations() gql.LocationConnectionResolver {
	return NewLocationConnectionResolver(r.locations, nil, r.locationResolver)
}
