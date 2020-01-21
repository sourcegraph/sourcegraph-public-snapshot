package resolvers

import (
	"context"
	"encoding/base64"

	graphql "github.com/graph-gophers/graphql-go"
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

func (r *Resolver) LSIFUploadByID(ctx context.Context, id graphql.ID) (graphqlbackend.LSIFUploadResolver, error) {
	uploadID, err := unmarshalLSIFUploadGQLID(id)
	if err != nil {
		return nil, err
	}

	lsifUpload, err := client.DefaultClient.GetUpload(ctx, &struct {
		UploadID int64
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
		UploadID int64
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
func (r *Resolver) LSIFUploads(ctx context.Context, args *graphqlbackend.LSIFRepositoryUploadsQueryArgs) (graphqlbackend.LSIFUploadConnectionResolver, error) {
	opt := LSIFUploadsListOptions{
		RepositoryID:    args.RepositoryID,
		Query:           args.Query,
		State:           args.State,
		IsLatestForRepo: args.IsLatestForRepo,
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

func (r *Resolver) LSIF(ctx context.Context, args *graphqlbackend.LSIFQueryArgs) (graphqlbackend.LSIFQueryResolver, error) {
	upload, err := client.DefaultClient.Exists(ctx, &struct {
		RepoID api.RepoID
		Commit string
		Path   string
	}{
		RepoID: args.Repository.Type().ID,
		Commit: string(args.Commit),
		Path:   args.Path,
	})

	if err != nil {
		return nil, err
	}

	if upload == nil {
		return nil, nil
	}

	return &lsifQueryResolver{
		repoID: args.Repository.Type().ID,
		commit: args.Commit,
		path:   args.Path,
		upload: upload,
	}, nil
}
