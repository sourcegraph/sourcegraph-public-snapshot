package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
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

type restIssue struct {
	Title   string
	HTMLURL string `json:"html_url"`
	Body    string
}

func searchIssues(ctx context.Context, args *search.Args) []*searchResultResolver {
	var constructedQuery strings.Builder
	for _, repo := range args.Repos {
		repoArg := fmt.Sprintf("repo:%s+", strings.Replace(string(repo.Repo.URI), "github.com/", "", 1))
		constructedQuery.Write([]byte(repoArg))
	}

	constructedQuery.Write([]byte(strings.Replace(args.Pattern.Pattern, " ", "+", -1)))
	fmt.Println(constructedQuery.String())
	resp, _ := http.Get(fmt.Sprintf("https://api.github.com/search/issues?q=%s", constructedQuery.String()))
	var t struct {
		Items []restIssue
	}
	json.NewDecoder(resp.Body).Decode(&t)
	results := make([]*searchResultResolver, len(t.Items))
	for i, item := range t.Items {
		res := &issueSearchResultResolver{
			title: item.Title,
			url:   item.HTMLURL,
			body:  item.Body,
		}
		results[i] = &searchResultResolver{issue: res}
	}
	return results
}
