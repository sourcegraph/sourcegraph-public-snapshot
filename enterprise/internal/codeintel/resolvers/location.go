package resolvers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver/client"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lsif"
)

//
// Connection Resolver

type LocationsQueryOptions struct {
	Operation  string
	Repository string
	Commit     string
	Path       string
	Line       int32
	Character  int32
	Limit      *int32
	NextURL    *string
}

type locationConnectionResolver struct {
	opt LocationsQueryOptions

	once      sync.Once
	locations []*lsif.LSIFLocation
	nextURL   string
	err       error
}

var _ graphqlbackend.LocationConnectionResolver = &locationConnectionResolver{}

func (r *locationConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.LocationResolver, error) {
	locations, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	var l []graphqlbackend.LocationResolver
	for _, location := range locations {
		resolver, err := rangeToLocationResolver(ctx, location)
		if err != nil {
			return nil, err
		}

		l = append(l, resolver)
	}
	return l, nil
}

func (r *locationConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, nextURL, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if nextURL != "" {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(nextURL))), nil
	}

	return graphqlutil.HasNextPage(false), nil
}

func (r *locationConnectionResolver) compute(ctx context.Context) ([]*lsif.LSIFLocation, string, error) {
	r.once.Do(func() {
		var path string
		if r.opt.NextURL == nil {
			// first page of results
			path = fmt.Sprintf("/%s", r.opt.Operation)
		} else {
			// subsequent page of results
			path = *r.opt.NextURL
		}

		values := url.Values{}
		values.Set("repository", r.opt.Repository)
		values.Set("commit", r.opt.Commit)
		values.Set("path", r.opt.Path)
		values.Set("line", strconv.FormatInt(int64(r.opt.Line), 10))
		values.Set("character", strconv.FormatInt(int64(r.opt.Character), 10))
		if r.opt.Limit != nil {
			values.Set("limit", strconv.FormatInt(int64(*r.opt.Limit), 10))
		}

		resp, err := client.DefaultClient.BuildAndTraceRequest(ctx, "GET", path, values, nil)
		if err != nil {
			r.err = err
			return
		}

		payload := struct {
			Locations []*lsif.LSIFLocation
		}{
			Locations: []*lsif.LSIFLocation{},
		}

		if err := client.UnmarshalPayload(resp, &payload); err != nil {
			r.err = err
			return
		}

		r.locations = payload.Locations
		r.nextURL = client.ExtractNextURL(resp)
	})

	return r.locations, r.nextURL, r.err
}

//
// Helpers

func rangeToLocationResolver(ctx context.Context, location *lsif.LSIFLocation) (graphqlbackend.LocationResolver, error) {
	repo, err := backend.Repos.GetByName(ctx, api.RepoName(location.Repository))
	if err != nil {
		return nil, err
	}

	repoResolver, err := graphqlbackend.RepositoryByIDInt32(ctx, repo.ID)
	if err != nil {
		return nil, err
	}

	commitResolver, err := repoResolver.Commit(ctx, &graphqlbackend.RepositoryCommitArgs{Rev: location.Commit})
	if err != nil {
		return nil, err
	}

	gitTreeResolver, err := graphqlbackend.NewGitTreeEntryResolver(commitResolver, graphqlbackend.CreateFileInfo(location.Path, true)), nil
	if err != nil {
		return nil, err
	}

	return graphqlbackend.NewLocationResolver(gitTreeResolver, location.Range), nil
}
