package graphqlbackend

import (
	"context"
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/schema"
)

type issueSearchResultResolver struct {
	title string
	url   string
	body  string
}

func (r *issueSearchResultResolver) Title() string {
	return r.title
}

func (r *issueSearchResultResolver) URL() string {
	return r.url
}

func (r *issueSearchResultResolver) Body() string {
	return r.body
}

func searchIssues(ctx context.Context, args *search.Args) ([]*searchResultResolver, error) {
	gh := conf.Get().Github
	cl, err := createGitHubClients(gh)
	if err != nil {
		return nil, err
	}
	var results []*searchResultResolver

	for _, c := range cl {
		if c != nil {
			var reposToSearch []string
			for _, repo := range args.Repos {
				reposToSearch = append(reposToSearch, string(repo.Repo.Name))
			}

			issues := c.SearchIssues(ctx, reposToSearch, args.Pattern.Pattern)

			r := make([]*searchResultResolver, len(issues))
			for i, item := range issues {
				res := &issueSearchResultResolver{
					title: item.Title,
					url:   item.HTMLURL,
					body:  item.Body,
				}
				r[i] = &searchResultResolver{issue: res}
			}
			results = append(results, r...)
		}

	}

	return results, nil
}

func createGitHubClients(gh []*schema.GitHubConnection) ([]*github.Client, error) {
	cl := make([]*github.Client, len(gh))
	check := make(map[string]bool, len(gh))

	for i, g := range gh {
		apiURL, err := url.Parse(g.Url)
		if err != nil {
			return nil, err
		}
		url := extsvc.NormalizeBaseURL(apiURL)
		if check[url.Host] != true {
			cl[i] = github.NewClient(url, g.Token, nil, nil)
			check[url.Host] = true
		}

	}

	return cl, nil
}
