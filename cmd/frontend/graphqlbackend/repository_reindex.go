package graphqlbackend

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ReindexRepository will trigger Zoekt indexserver to reindex the repository.
func (r *schemaResolver) ReindexRepository(ctx context.Context, args *struct {
	Repository graphql.ID
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: There is no reason why non-site-admins would need to run this operation.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repo, err := r.repositoryByID(ctx, args.Repository)
	if err != nil {
		return nil, err
	}

	// Find the Zoekt webserver hosting the index of "repo"
	ep, err := search.Indexers().Map.Get(repo.Name())
	if err != nil {
		return nil, err
	}

	// We add the scheme http:// on a best-effort basis.
	//
	// ep doesn't have to be a valid URL. For example, locally ep can just be
	// localhost:<port>, which would be parsed by url.Parse without error, with
	// localhost as scheme. The reason is that the Go 1.19 scheme parser parses
	// all valid characters [a-zA-Z][a-zA-Z0-9+-.]*) before the first ':' as
	// scheme.
	if !strings.HasPrefix(ep, "http://") {
		ep = "http://" + ep
	}
	u, err := url.Parse(ep)
	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Add("repo", strconv.Itoa(int(repo.IDInt32())))

	// <host:port>/indexerver/?headless
	u = u.ResolveReference(&url.URL{Path: "/indexserver/", RawQuery: "headless"})

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpcli.InternalClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(b))
	}

	return &EmptyResponse{}, nil
}
