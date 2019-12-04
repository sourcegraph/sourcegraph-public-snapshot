package resolvers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver/client"
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

func (r *Resolver) LSIFDumpByID(ctx context.Context, id graphql.ID) (graphqlbackend.LSIFDumpResolver, error) {
	repoName, dumpID, err := unmarshalLSIFDumpGQLID(id)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/dumps/%s/%d", url.PathEscape(repoName), dumpID)

	var lsifDump *lsif.LSIFDump
	if err := client.DefaultClient.TraceRequestAndUnmarshalPayload(ctx, "GET", path, nil, nil, &lsifDump); err != nil {
		return nil, err
	}

	repo, err := backend.Repos.GetByName(ctx, api.RepoName(repoName))
	if err != nil {
		return nil, err
	}

	return &lsifDumpResolver{repo: repo, lsifDump: lsifDump}, nil
}

func (r *Resolver) DeleteLSIFDump(ctx context.Context, id graphql.ID) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may delete LSIF data for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	repoName, dumpID, err := unmarshalLSIFDumpGQLID(id)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/dumps/%s/%d", url.PathEscape(repoName), dumpID)
	if err := client.DefaultClient.TraceRequestAndUnmarshalPayload(ctx, "DELETE", path, nil, nil, nil); !client.IsNotFound(err) {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

//
// Dump Connection Resolvers

// This method implements cursor-based forward pagination. The `after` parameter
// should be an `endCursor` value from a previous request. This value is the rel="next"
// URL in the Link header of the LSIF server response. This URL includes all of the
// query variables required to fetch the subsequent page of results. This state is not
// dependent on the limit, so we can overwrite this value if the user has changed its
// value since making the last request.

func (r *Resolver) LSIFDumps(ctx context.Context, args *graphqlbackend.LSIFRepositoryDumpsQueryArgs) (graphqlbackend.LSIFDumpConnectionResolver, error) {
	opt := LSIFDumpsListOptions{
		RepositoryID:    args.RepositoryID,
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

func (r *Resolver) LSIFJobByID(ctx context.Context, id graphql.ID) (graphqlbackend.LSIFJobResolver, error) {
	jobID, err := unmarshalLSIFJobGQLID(id)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/jobs/%s", url.PathEscape(jobID))

	var lsifJob *lsif.LSIFJob
	if err := client.DefaultClient.TraceRequestAndUnmarshalPayload(ctx, "GET", path, nil, nil, &lsifJob); err != nil {
		return nil, err
	}

	return &lsifJobResolver{lsifJob: lsifJob}, nil
}

func (r *Resolver) DeleteLSIFJob(ctx context.Context, id graphql.ID) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may delete LSIF data for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	jobID, err := unmarshalLSIFJobGQLID(id)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/jobs/%s", url.PathEscape(jobID))
	if err := client.DefaultClient.TraceRequestAndUnmarshalPayload(ctx, "DELETE", path, nil, nil, nil); !client.IsNotFound(err) {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
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
	return r.LSIFJobStatsByID(ctx, marshalLSIFJobStatsGQLID(lsifJobStatsGQLID))
}

func (r *Resolver) LSIFJobStatsByID(ctx context.Context, id graphql.ID) (graphqlbackend.LSIFJobStatsResolver, error) {
	lsifJobStatsID, err := unmarshalLSIFJobStatsGQLID(id)
	if err != nil {
		return nil, err
	}
	if lsifJobStatsID != lsifJobStatsGQLID {
		return nil, fmt.Errorf("lsif job stats not found: %q", lsifJobStatsID)
	}

	var stats *lsif.LSIFJobStats
	if err := client.DefaultClient.TraceRequestAndUnmarshalPayload(ctx, "GET", "/jobs/stats", nil, nil, &stats); err != nil {
		return nil, err
	}

	return &lsifJobStatsResolver{stats: stats}, nil
}
