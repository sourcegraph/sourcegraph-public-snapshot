package resolvers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
// Node Resolver

type locationWithConfidenceResolver struct {
	*graphqlbackend.LocationResolver
}

var _ graphqlbackend.LocationWithConfidenceResolver = &locationWithConfidenceResolver{}

func (r *locationWithConfidenceResolver) Placeholder() *string { return nil }

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

type locationWithConfidenceConnectionResolver struct {
	opt LocationsQueryOptions

	once      sync.Once
	locations []*lsif.LSIFLocation
	nextURL   string
	err       error
}

func (r *locationWithConfidenceConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.LocationWithConfidenceResolver, error) {
	locations, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	var l []graphqlbackend.LocationWithConfidenceResolver
	for _, location := range locations {
		resolver, err := rangeToLocationWithConfidenceResolver(ctx, location)
		if err != nil {
			return nil, err
		}

		l = append(l, resolver)
	}
	return l, nil
}

func (r *locationWithConfidenceConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, nextURL, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if nextURL != "" {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(nextURL))), nil
	}

	return graphqlutil.HasNextPage(false), nil
}

func (r *locationWithConfidenceConnectionResolver) compute(ctx context.Context) ([]*lsif.LSIFLocation, string, error) {
	r.once.Do(func() {
		var path string
		if r.opt.NextURL == nil {
			// first page of results
			path = fmt.Sprintf("/request")
		} else {
			fmt.Printf("paths: %#v\n", *r.opt.NextURL)

			// subsequent page of results
			path = *r.opt.NextURL
		}

		values := url.Values{}
		values.Set("repository", r.opt.Repository)
		values.Set("commit", r.opt.Commit)
		if r.opt.Limit != nil {
			values.Set("limit", strconv.FormatInt(int64(*r.opt.Limit), 10))
		}

		body, err := json.Marshal(map[string]interface{}{
			"method": r.opt.Operation,
			"path":   r.opt.Path,
			"position": map[string]interface{}{
				"line":      r.opt.Line,
				"character": r.opt.Character,
			},
		})
		if err != nil {
			r.err = err
			return
		}

		resp, err := client.DefaultClient.BuildAndTraceRequest(ctx, "POST", path, values, ioutil.NopCloser(bytes.NewReader(body)))
		if err != nil {
			r.err = err
			return
		}

		var payload []*lsif.LSIFLocation
		if err := client.UnmarshalPayload(resp, &payload); err != nil {
			r.err = err
			return
		}

		r.locations = payload
		r.nextURL = client.ExtractNextURL(resp)
	})

	return r.locations, r.nextURL, r.err
}

//
// Helpers

func rangeToLocationWithConfidenceResolver(ctx context.Context, location *lsif.LSIFLocation) (*locationWithConfidenceResolver, error) {
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

	return &locationWithConfidenceResolver{
		LocationResolver: graphqlbackend.NewLocationResolver(gitTreeResolver, location.Range),
	}, nil
}
