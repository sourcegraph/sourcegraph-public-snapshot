package resolvers

import (
	"context"
	"encoding/base64"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver/client"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type Resolver struct{}

var _ graphqlbackend.CodeIntelResolver = &Resolver{}

func NewResolver() graphqlbackend.CodeIntelResolver {
	return &Resolver{}
}

func (r *Resolver) LSIFDumpByID(ctx context.Context, id graphql.ID) (graphqlbackend.LSIFDumpResolver, error) {
	repoName, dumpID, err := unmarshalLSIFDumpGQLID(id)
	if err != nil {
		return nil, err
	}

	repo, err := backend.Repos.GetByName(ctx, api.RepoName(repoName))
	if err != nil {
		return nil, err
	}

	lsifDump, err := client.DefaultClient.GetDump(ctx, &struct {
		RepoName string
		DumpID   int64
	}{
		RepoName: repoName,
		DumpID:   dumpID,
	})
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

	err = client.DefaultClient.DeleteDump(ctx, &struct {
		RepoName string
		DumpID   int64
	}{
		RepoName: repoName,
		DumpID:   dumpID,
	})
	if err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

// LSIFDumps resolves LSIF dumps for a given repository.
//
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

func (r *Resolver) LSIFUploadByID(ctx context.Context, id graphql.ID) (graphqlbackend.LSIFUploadResolver, error) {
	uploadID, err := unmarshalLSIFUploadGQLID(id)
	if err != nil {
		return nil, err
	}

	lsifUpload, err := client.DefaultClient.GetUpload(ctx, &struct {
		UploadID string
	}{
		UploadID: uploadID,
	})
	if err != nil {
		return nil, err
	}

	return &lsifUploadResolver{lsifUpload: lsifUpload}, nil
}

func (r *Resolver) DeleteLSIFUpload(ctx context.Context, id graphql.ID) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may delete LSIF data for now
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	uploadID, err := unmarshalLSIFUploadGQLID(id)
	if err != nil {
		return nil, err
	}

	err = client.DefaultClient.DeleteUpload(ctx, &struct {
		UploadID string
	}{
		UploadID: uploadID,
	})
	if err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

// LSIFUploads resolves the LSIF uploads in a given state.
//
// This method implements cursor-based forward pagination. The `after` parameter
// should be an `endCursor` value from a previous request. This value is the rel="next"
// URL in the Link header of the LSIF server response. This URL includes all of the
// query variables required to fetch the subsequent page of results. This state is not
// dependent on the limit, so we can overwrite this value if the user has changed its
// value since making the last request.
func (r *Resolver) LSIFUploads(ctx context.Context, args *graphqlbackend.LSIFUploadsQueryArgs) (graphqlbackend.LSIFUploadConnectionResolver, error) {
	opt := LSIFUploadsListOptions{
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

	return &lsifUploadConnectionResolver{opt: opt}, nil
}

const lsifUploadStatsGQLID = "lsifUploadStats"

func (r *Resolver) LSIFUploadStats(ctx context.Context) (graphqlbackend.LSIFUploadStatsResolver, error) {
	return r.LSIFUploadStatsByID(ctx, marshalLSIFUploadStatsGQLID(lsifUploadStatsGQLID))
}

func (r *Resolver) LSIFUploadStatsByID(ctx context.Context, id graphql.ID) (graphqlbackend.LSIFUploadStatsResolver, error) {
	lsifUploadStatsID, err := unmarshalLSIFUploadStatsGQLID(id)
	if err != nil {
		return nil, err
	}
	if lsifUploadStatsID != lsifUploadStatsGQLID {
		return nil, fmt.Errorf("lsif upload stats not found: %q", lsifUploadStatsID)
	}

	stats, err := client.DefaultClient.GetUploadStats(ctx)
	if err != nil {
		return nil, err
	}

	return &lsifUploadStatsResolver{stats: stats}, nil
}

func (r *Resolver) LSIF(ctx context.Context, args *graphqlbackend.LSIFQueryArgs) (graphqlbackend.LSIFQueryResolver, error) {
	dump, err := client.DefaultClient.Exists(ctx, &struct {
		RepoName string
		Commit   string
		Path     string
	}{
		RepoName: args.RepoName,
		Commit:   string(args.Commit),
		Path:     args.Path,
	})

	if err != nil {
		return nil, err
	}

	return &lsifQueryResolver{
		repoName: args.RepoName,
		commit:   args.Commit,
		path:     args.Path,
		dump:     dump,
	}, nil
}
