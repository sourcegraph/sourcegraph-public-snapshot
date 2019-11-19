package resolvers

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/pkg/codeintel/lsifserver/client"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lsif"
)

type Resolver struct{}

var _ graphqlbackend.CodeIntelResolver = &Resolver{}

func NewResolver() graphqlbackend.CodeIntelResolver {
	return &Resolver{}
}

//
// Dump Node Resolvers

func (r *Resolver) LSIFDump(ctx context.Context, args *struct{ ID graphql.ID }) (graphqlbackend.LSIFDumpResolver, error) {
	return r.LSIFDumpByGQLID(ctx, args.ID)
}

func (r *Resolver) LSIFDumpByGQLID(ctx context.Context, id graphql.ID) (graphqlbackend.LSIFDumpResolver, error) {
	repoName, dumpID, err := unmarshalLSIFDumpGQLID(id)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/dumps/%s/%d", url.PathEscape(repoName), dumpID)

	var lsifDump *lsif.LSIFDump
	if err := client.TraceRequestAndUnmarshalPayload(ctx, "GET", path, nil, nil, &lsifDump); err != nil {
		return nil, err
	}

	repo, err := backend.Repos.GetByName(ctx, api.RepoName(repoName))
	if err != nil {
		return nil, err
	}

	return &lsifDumpResolver{repo: repo, lsifDump: lsifDump}, nil
}

//
// Dump Connection Resolvers

// This method implements cursor-based forward pagination. The `after` parameter
// should be an `endCursor` value from a previous request. This value is the rel="next"
// URL in the Link header of the LSIF server response. This URL includes all of the
// query variables required to fetch the subsequent page of results. This state is not
// dependent on the limit, so we can overwrite this value if the user has changed its
// value since making the last request.

func (r *Resolver) LSIFDumps(ctx context.Context, args *graphqlbackend.LSIFDumpsQueryArgs) (graphqlbackend.LSIFDumpConnectionResolver, error) {
	opt := LSIFDumpsListOptions{
		Repository:      args.Repository,
		Query:           args.Query,
		IsLatestForRepo: args.IsLatestForRepo,
	}
	if args.First != nil {
		if *args.First < 0 || *args.First > 5000 {
			return nil, errors.New("lsifDumps: requested pagination 'first' value outside allowed range (0 - 5000)")
		}

		opt.Limit = args.First
	}
	if args.After != nil {
		decoded, err := base64.StdEncoding.DecodeString(*args.After)
		if err != nil {
			return nil, err
		}
		nextURL := string(decoded)
		opt.NextURL = &nextURL
	}

	return &lsifDumpConnectionResolver{opt: opt}, nil
}

//
// Job Node Resolvers

func (r *Resolver) LSIFJob(ctx context.Context, args *struct{ ID graphql.ID }) (graphqlbackend.LSIFJobResolver, error) {
	return r.LSIFJobByGQLID(ctx, args.ID)
}

func (r *Resolver) LSIFJobByGQLID(ctx context.Context, id graphql.ID) (graphqlbackend.LSIFJobResolver, error) {
	jobID, err := unmarshalLSIFJobGQLID(id)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/jobs/%s", url.PathEscape(jobID))

	var lsifJob *lsif.LSIFJob
	if err := client.TraceRequestAndUnmarshalPayload(ctx, "GET", path, nil, nil, &lsifJob); err != nil {
		return nil, err
	}

	return &lsifJobResolver{lsifJob: lsifJob}, nil
}

//
// Job Connection Resolvers

// This method implements cursor-based forward pagination. The `after` parameter
// should be an `endCursor` value from a previous request. This value is the rel="next"
// URL in the Link header of the LSIF server response. This URL includes all of the
// query variables required to fetch the subsequent page of results. This state is not
// dependent on the limit, so we can overwrite this value if the user has changed its
// value since making the last request.

func (r *Resolver) LSIFJobs(ctx context.Context, args *graphqlbackend.LSIFJobsQueryArgs) (graphqlbackend.LSIFJobConnectionResolver, error) {
	opt := LSIFJobsListOptions{
		State: args.State,
		Query: args.Query,
	}
	if args.First != nil {
		opt.Limit = args.First
	}
	if args.After != nil {
		decoded, err := base64.StdEncoding.DecodeString(*args.After)
		if err != nil {
			return nil, err
		}
		nextURL := string(decoded)
		opt.NextURL = &nextURL
	}

	return &lsifJobConnectionResolver{opt: opt}, nil
}

//
// Job Stats Resolvers

const lsifJobStatsGQLID = "lsifJobStats"

func (r *Resolver) LSIFJobStats(ctx context.Context) (graphqlbackend.LSIFJobStatsResolver, error) {
	return r.LSIFJobStatsByGQLID(ctx, marshalLSIFJobStatsGQLID(lsifJobStatsGQLID))
}

func (r *Resolver) LSIFJobStatsByGQLID(ctx context.Context, id graphql.ID) (graphqlbackend.LSIFJobStatsResolver, error) {
	lsifJobStatsID, err := unmarshalLSIFJobStatsGQLID(id)
	if err != nil {
		return nil, err
	}
	if lsifJobStatsID != lsifJobStatsGQLID {
		return nil, fmt.Errorf("lsif job stats not found: %q", lsifJobStatsID)
	}

	var stats *lsif.LSIFJobStats
	if err := client.TraceRequestAndUnmarshalPayload(ctx, "GET", "/jobs/stats", nil, nil, &stats); err != nil {
		return nil, err
	}

	return &lsifJobStatsResolver{stats: stats}, nil
}
