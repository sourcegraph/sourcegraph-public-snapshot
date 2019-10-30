package graphqlbackend

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

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

func marshalLSIFDumpGQLID(repoName string, lsifDumpID int32) graphql.ID {
	// Encode both repository and ID, as we need both to make the backend request
	return relay.MarshalID("LSIFDump", fmt.Sprintf(
		"%s:%s",
		base64.StdEncoding.EncodeToString([]byte(repoName)),
		strconv.FormatInt(int64(lsifDumpID), 36),
	))
}

func unmarshalLSIFDumpGQLID(id graphql.ID) (string, int32, error) {
	var raw string
	if err := relay.UnmarshalSpec(id, &raw); err != nil {
		return "", 0, err
	}

	parts := strings.Split(raw, ":")
	if len(parts) != 2 {
		return "", 0, errors.New("malformed LSIF dump id")
	}

	repoName, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return "", 0, err
	}

	lsifDumpID, err = strconv.ParseInt(parts[1], 36, 32)
	if err != nil {
		return "", 0, err
	}

	return string(repoName), lsifDumpID, nil
}
