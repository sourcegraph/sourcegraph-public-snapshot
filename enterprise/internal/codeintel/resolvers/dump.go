package resolvers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver/client"
	"github.com/sourcegraph/sourcegraph/internal/lsif"
)

type lsifDumpResolver struct {
	repo     *types.Repo
	lsifDump *lsif.LSIFDump
}

var _ graphqlbackend.LSIFDumpResolver = &lsifDumpResolver{}

func (r *lsifDumpResolver) ID() graphql.ID {
	return marshalLSIFDumpGQLID(r.lsifDump.Repository, r.lsifDump.ID)
}

func (r *lsifDumpResolver) ProjectRoot(ctx context.Context) (*graphqlbackend.GitTreeEntryResolver, error) {
	repoResolver := graphqlbackend.NewRepositoryResolver(r.repo)
	commitResolver, err := repoResolver.Commit(ctx, &graphqlbackend.RepositoryCommitArgs{Rev: r.lsifDump.Commit})
	if err != nil {
		return nil, err
	}

	return graphqlbackend.NewGitTreeEntryResolver(commitResolver, graphqlbackend.CreateFileInfo(r.lsifDump.Root, true)), nil
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
	RepositoryID    graphql.ID
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
	repo       *graphqlbackend.RepositoryResolver
	totalCount int
	nextURL    string
	err        error
}

var _ graphqlbackend.LSIFDumpConnectionResolver = &lsifDumpConnectionResolver{}

func (r *lsifDumpConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.LSIFDumpResolver, error) {
	dumps, repo, _, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	var l []graphqlbackend.LSIFDumpResolver
	for _, lsifDump := range dumps {
		l = append(l, &lsifDumpResolver{
			repo:     repo.Type(),
			lsifDump: lsifDump,
		})
	}
	return l, nil
}

func (r *lsifDumpConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	_, _, count, _, err := r.compute(ctx)
	return int32(count), err
}

func (r *lsifDumpConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, _, _, nextURL, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if nextURL != "" {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(nextURL))), nil
	}

	return graphqlutil.HasNextPage(false), nil
}

func (r *lsifDumpConnectionResolver) compute(ctx context.Context) ([]*lsif.LSIFDump, *graphqlbackend.RepositoryResolver, int, string, error) {
	r.once.Do(func() {
		repo, err := graphqlbackend.RepositoryByID(ctx, r.opt.RepositoryID)
		if err != nil {
			r.err = err
			return
		}

		var path string
		if r.opt.NextURL == nil {
			// first page of results
			path = fmt.Sprintf("/dumps/%s", url.PathEscape(repo.Name()))
		} else {
			// subsequent page of results
			path = *r.opt.NextURL
		}

		query := url.Values{}
		if r.opt.Query != nil {
			query.Set("query", *r.opt.Query)
		}
		if r.opt.IsLatestForRepo != nil && *r.opt.IsLatestForRepo {
			query.Set("visibleAtTip", "true")
		}
		if r.opt.Limit != nil {
			query.Set("limit", strconv.FormatInt(int64(*r.opt.Limit), 10))
		}

		resp, err := client.DefaultClient.BuildAndTraceRequest(ctx, "GET", path, query, nil)
		if err != nil {
			r.err = err
			return
		}

		payload := struct {
			Dumps      []*lsif.LSIFDump `json:"dumps"`
			TotalCount int              `json:"totalCount"`
		}{
			Dumps: []*lsif.LSIFDump{},
		}

		if err := client.UnmarshalPayload(resp, &payload); err != nil {
			r.err = err
			return
		}

		r.dumps = payload.Dumps
		r.repo = repo
		r.totalCount = payload.TotalCount
		r.nextURL = client.ExtractNextURL(resp)
	})

	return r.dumps, r.repo, r.totalCount, r.nextURL, r.err
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
