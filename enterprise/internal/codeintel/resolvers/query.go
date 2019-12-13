package resolvers

import (
	"context"
	"encoding/base64"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver/client"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lsif"
)

type lsifQueryResolver struct {
	repoName string
	commit   graphqlbackend.GitObjectID
	path     string
	dump     *lsif.LSIFDump
}

var _ graphqlbackend.LSIFQueryResolver = &lsifQueryResolver{}

func (r *lsifQueryResolver) Commit(ctx context.Context) (*graphqlbackend.GitCommitResolver, error) {
	repo, err := backend.Repos.GetByName(ctx, api.RepoName(r.repoName))
	if err != nil {
		return nil, err
	}

	return graphqlbackend.NewRepositoryResolver(repo).Commit(
		ctx,
		&graphqlbackend.RepositoryCommitArgs{Rev: string(r.dump.Commit)},
	)
}

func (r *lsifQueryResolver) Definitions(ctx context.Context, args *graphqlbackend.LSIFQueryPositionArgs) (graphqlbackend.LocationConnectionResolver, error) {
	opts := &struct {
		RepoName  string
		Commit    graphqlbackend.GitObjectID
		Path      string
		Line      int32
		Character int32
		DumpID    int64
	}{
		RepoName:  r.repoName,
		Commit:    r.commit,
		Path:      r.path,
		Line:      args.Line,
		Character: args.Character,
		DumpID:    r.dump.ID,
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
		RepoName  string
		Commit    graphqlbackend.GitObjectID
		Path      string
		Line      int32
		Character int32
		DumpID    int64
		Limit     *int32
		Cursor    *string
	}{
		RepoName:  r.repoName,
		Commit:    r.commit,
		Path:      r.path,
		Line:      args.Line,
		Character: args.Character,
		DumpID:    r.dump.ID,
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
		RepoName  string
		Commit    graphqlbackend.GitObjectID
		Path      string
		Line      int32
		Character int32
		DumpID    int64
	}{
		RepoName:  r.repoName,
		Commit:    r.commit,
		Path:      r.path,
		Line:      args.Line,
		Character: args.Character,
		DumpID:    r.dump.ID,
	})
	if err != nil {
		return nil, err
	}

	return &hoverResolver{
		text:     text,
		lspRange: lspRange,
	}, nil
}
