package azuredevops

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (c *Client) GetRepo(ctx context.Context, args OrgProjectRepoArgs) (Repository, error) {
	queryParams := make(url.Values)
	queryParams.Set("api-version", apiVersion)

	urlRepositoriesByProjects := url.URL{Path: fmt.Sprintf("%s/%s/_apis/git/repositories/%s", args.Org, args.Project, args.RepoNameOrID), RawQuery: queryParams.Encode()}

	req, err := http.NewRequest("GET", urlRepositoriesByProjects.String(), nil)
	if err != nil {
		return Repository{}, err
	}

	var repo Repository
	if _, err = c.do(ctx, req, "", &repo); err != nil {
		return Repository{}, err
	}

	return repo, nil
}

func (c *Client) ListRepositoriesByProjectOrOrg(ctx context.Context, args ListRepositoriesByProjectOrOrgArgs) ([]Repository, error) {
	queryParams := make(url.Values)
	queryParams.Set("api-version", apiVersion)

	urlRepositoriesByProjects := url.URL{Path: fmt.Sprintf("%s/_apis/git/repositories", args.ProjectOrOrgName), RawQuery: queryParams.Encode()}

	req, err := http.NewRequest("GET", urlRepositoriesByProjects.String(), nil)
	if err != nil {
		return nil, err
	}

	var repos ListRepositoriesResponse
	if _, err = c.do(ctx, req, "", &repos); err != nil {
		return nil, err
	}

	return repos.Value, nil
}

func (c *Client) ForkRepository(ctx context.Context, org string, input ForkRepositoryInput) (Repository, error) {
	queryParams := make(url.Values)
	queryParams.Set("api-version", apiVersion)

	data, err := json.Marshal(&input)
	if err != nil {
		return Repository{}, errors.Wrap(err, "marshalling request")
	}

	urlRepositoriesByProjects := url.URL{Path: fmt.Sprintf("%s/_apis/git/repositories", org), RawQuery: queryParams.Encode()}

	req, err := http.NewRequest("POST", urlRepositoriesByProjects.String(), bytes.NewBuffer(data))
	if err != nil {
		return Repository{}, err
	}

	var repo Repository
	if _, err = c.do(ctx, req, "", &repo); err != nil {
		return Repository{}, err
	}

	return repo, nil
}

func (c *Client) GetCommitForRepositoryBranchHead(ctx context.Context, args OrgProjectRepoArgs, branchName string) (Ref, error) {
	var allRefs []Ref
	continuationToken := ""

	for {
		queryParams := make(url.Values)
		queryParams.Set("api-version", apiVersion)
		// The filter here by branch name is only a substring match, so we aren't guaranteed to only get one result.
		queryParams.Set("filter", fmt.Sprintf("heads/%s", branchName))
		if continuationToken != "" {
			queryParams.Set("continuationToken", continuationToken)
		}
		urlRepositoriesByProjects := url.URL{Path: fmt.Sprintf("%s/%s/_apis/git/repositories/%s/refs", args.Org, args.Project, args.RepoNameOrID), RawQuery: queryParams.Encode()}

		req, err := http.NewRequest("GET", urlRepositoriesByProjects.String(), nil)
		if err != nil {
			return Ref{}, err
		}

		var refs ListRefsResponse
		continuationToken, err = c.do(ctx, req, "", &refs)
		if err != nil {
			return Ref{}, err
		}
		allRefs = append(allRefs, refs.Value...)

		if continuationToken == "" {
			break
		}
	}

	for _, ref := range allRefs {
		if ref.Name == fmt.Sprintf("refs/heads/%s", branchName) {
			return ref, nil
		}
	}

	return Ref{}, errors.New("branch not found")
}
