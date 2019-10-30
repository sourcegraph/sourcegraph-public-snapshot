package graphqlbackend

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/lsif"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

func (r *schemaResolver) LSIFDump(ctx context.Context, args *struct{ ID graphql.ID }) (*lsifDumpResolver, error) {
	return lsifDumpByGQLID(ctx, args.ID)
}

type lsifDumpResolver struct {
	repo     *types.Repo
	lsifDump *types.LSIFDump
}

func lsifDumpByGQLID(ctx context.Context, id graphql.ID) (*lsifDumpResolver, error) {
	repoName, dumpID, err := unmarshalLSIFDumpGQLID(id)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("/dumps/%s/%d", url.PathEscape(repoName), dumpID)

	var lsifDump *types.LSIFDump
	if err := lsif.TraceRequestAndUnmarshalPayload(ctx, path, nil, &lsifDump); err != nil {
		return nil, err
	}

	repo, err := backend.Repos.GetByName(ctx, api.RepoName(repoName))
	if err != nil {
		return nil, err
	}

	return &lsifDumpResolver{repo: repo, lsifDump: lsifDump}, nil
}

func (r *lsifDumpResolver) ID() graphql.ID {
	return marshalLSIFDumpGQLID(r.lsifDump.Repository, r.lsifDump.ID)
}

func (r *lsifDumpResolver) ProjectRoot() *gitTreeEntryResolver {
	commitResolver := &GitCommitResolver{
		repo: &RepositoryResolver{repo: r.repo},
		oid:  GitObjectID(r.lsifDump.Commit),
	}

	return &gitTreeEntryResolver{
		commit: commitResolver,
		stat:   createFileInfo(r.lsifDump.Root, true),
	}
}

func (r *lsifDumpResolver) IsLatestForRepo() bool {
	return r.lsifDump.VisibleAtTip
}

func (r *lsifDumpResolver) UploadedAt() DateTime {
	return DateTime{Time: r.lsifDump.UploadedAt}
}

type lsifDumpIDPayload struct {
	RepoName string
	ID       string
}

func marshalLSIFDumpGQLID(repoName string, lsifDumpID int64) graphql.ID {
	return relay.MarshalID("LSIFDump", lsifDumpIDPayload{
		RepoName: repoName,
		ID:       strconv.FormatInt(lsifDumpID, 36),
	})
}

func unmarshalLSIFDumpGQLID(id graphql.ID) (string, int64, error) {
	var raw lsifDumpIDPayload
	if err := relay.UnmarshalSpec(id, &raw); err != nil {
		return "", 0, err
	}

	dumpID, err := strconv.ParseInt(raw.ID, 36, 64)
	if err != nil {
		return "", 0, err
	}

	return raw.RepoName, dumpID, nil
}
