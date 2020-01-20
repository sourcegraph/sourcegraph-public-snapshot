package resolvers

import (
	"context"
	"encoding/base64"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver/client"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lsif"
)

type lsifQueryResolver struct {
	repoID   api.RepoID
	repoName api.RepoName
	commit   graphqlbackend.GitObjectID
	path     string
	upload   *lsif.LSIFUpload
}

var _ graphqlbackend.LSIFQueryResolver = &lsifQueryResolver{}

func (r *lsifQueryResolver) Commit(ctx context.Context) (*graphqlbackend.GitCommitResolver, error) {
	return resolveCommit(ctx, r.repoID, r.upload.Commit)
}

func (r *lsifQueryResolver) Definitions(ctx context.Context, args *graphqlbackend.LSIFQueryPositionArgs) (graphqlbackend.LocationConnectionResolver, error) {
	opts := &struct {
		RepoID    api.RepoID
		RepoName  api.RepoName
		Commit    graphqlbackend.GitObjectID
		Path      string
		Line      int32
		Character int32
		UploadID  int64
	}{
		RepoID:    r.repoID,
		RepoName:  r.repoName,
		Commit:    r.commit,
		Path:      r.path,
		Line:      args.Line,
		Character: args.Character,
		UploadID:  r.upload.ID,
	}

	locations, nextURL, err := client.DefaultClient.Definitions(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &locationConnectionResolver{
		locations: locations,
		nextURL:   nextURL,
	}, nil
}

func (r *lsifQueryResolver) References(ctx context.Context, args *graphqlbackend.LSIFPagedQueryPositionArgs) (graphqlbackend.LocationConnectionResolver, error) {
	opts := &struct {
		RepoID    api.RepoID
		RepoName  api.RepoName
		Commit    graphqlbackend.GitObjectID
		Path      string
		Line      int32
		Character int32
		UploadID  int64
		Limit     *int32
		Cursor    *string
	}{
		RepoID:    r.repoID,
		RepoName:  r.repoName,
		Commit:    r.commit,
		Path:      r.path,
		Line:      args.Line,
		Character: args.Character,
		UploadID:  r.upload.ID,
	}
	if args.First != nil {
		opts.Limit = args.First
	}
	if args.After != nil {
		decoded, err := base64.StdEncoding.DecodeString(*args.After)
		if err != nil {
			return nil, err
		}
		nextURL := string(decoded)
		opts.Cursor = &nextURL
	}

	locations, nextURL, err := client.DefaultClient.References(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &locationConnectionResolver{
		locations: locations,
		nextURL:   nextURL,
	}, nil
}

func (r *lsifQueryResolver) Hover(ctx context.Context, args *graphqlbackend.LSIFQueryPositionArgs) (graphqlbackend.HoverResolver, error) {
	text, lspRange, err := client.DefaultClient.Hover(ctx, &struct {
		RepoID    api.RepoID
		RepoName  api.RepoName
		Commit    graphqlbackend.GitObjectID
		Path      string
		Line      int32
		Character int32
		UploadID  int64
	}{
		RepoID:    r.repoID,
		RepoName:  r.repoName,
		Commit:    r.commit,
		Path:      r.path,
		Line:      args.Line,
		Character: args.Character,
		UploadID:  r.upload.ID,
	})
	if err != nil {
		return nil, err
	}

	return &hoverResolver{
		text:     text,
		lspRange: lspRange,
	}, nil
}
