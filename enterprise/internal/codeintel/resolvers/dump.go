package resolvers

import (
	"context"
	"encoding/base64"
	"strconv"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver/client"
	"github.com/sourcegraph/sourcegraph/internal/lsif"
)

type lsifDumpResolver struct {
	projectRoot *graphqlbackend.GitTreeEntryResolver
	lsifDump    *lsif.LSIFDump
}

var _ graphqlbackend.LSIFDumpResolver = &lsifDumpResolver{}

func (r *lsifDumpResolver) ID() graphql.ID {
	return marshalLSIFDumpGQLID(r.lsifDump.Repository, r.lsifDump.ID)
}

func (r *lsifDumpResolver) ProjectRoot() *graphqlbackend.GitTreeEntryResolver {
	return r.projectRoot
}

func (r *lsifDumpResolver) InputRepoName() string {
	return r.lsifDump.Repository
}

func (r *lsifDumpResolver) InputCommit() string {
	return r.lsifDump.Commit
}

func (r *lsifDumpResolver) InputRoot() string {
	return r.lsifDump.Root
}

func (r *lsifDumpResolver) IsLatestForRepo() bool {
	return r.lsifDump.VisibleAtTip
}

func (r *lsifDumpResolver) UploadedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.lsifDump.UploadedAt}
}

func (r *lsifDumpResolver) ProcessedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.lsifDump.ProcessedAt}
}

type LSIFDumpsListOptions struct {
	RepositoryName  string
	Query           *string
	IsLatestForRepo *bool
	Limit           *int32
	NextURL         *string
}

type lsifDumpConnectionResolver struct {
	opt LSIFDumpsListOptions

	// cache results because they are used by multiple fields
	once       sync.Once
	dumps      []*lsif.LSIFDump
	totalCount int
	nextURL    string
	err        error
}

var _ graphqlbackend.LSIFDumpConnectionResolver = &lsifDumpConnectionResolver{}

func (r *lsifDumpConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.LSIFDumpResolver, error) {
	dumps, _, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	collectionResolver := newRepositoryCollectionResolver()

	var l []graphqlbackend.LSIFDumpResolver
	for _, lsifDump := range dumps {
		projectRoot, err := collectionResolver.resolve(ctx, newTreeResolver(lsifDump), lsifDump.Repository, lsifDump.Commit, lsifDump.Root)
		if err != nil {
			return nil, err
		}

		if projectRoot == nil {
			continue
		}

		l = append(l, &lsifDumpResolver{
			lsifDump:    lsifDump,
			projectRoot: projectRoot,
		})
	}
	return l, nil
}

func (r *lsifDumpConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	_, count, _, err := r.compute(ctx)
	return int32(count), err
}

func (r *lsifDumpConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, _, nextURL, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if nextURL != "" {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(nextURL))), nil
	}

	return graphqlutil.HasNextPage(false), nil
}

func (r *lsifDumpConnectionResolver) compute(ctx context.Context) ([]*lsif.LSIFDump, int, string, error) {
	r.once.Do(func() {
		r.dumps, r.nextURL, r.totalCount, r.err = client.DefaultClient.GetDumps(ctx, &struct {
			RepoName        string
			Query           *string
			IsLatestForRepo *bool
			Limit           *int32
			Cursor          *string
		}{
			RepoName:        r.opt.RepositoryName,
			Query:           r.opt.Query,
			IsLatestForRepo: r.opt.IsLatestForRepo,
			Limit:           r.opt.Limit,
			Cursor:          r.opt.NextURL,
		})
	})

	return r.dumps, r.totalCount, r.nextURL, r.err
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
