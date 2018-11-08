package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

type RestIssue struct {
	Title   string
	HTMLURL string `json:"html_url"`
	Body    string
}

func (c *Client) SearchIssues(ctx context.Context, repoNames []string, searchPattern string) []RestIssue {
	var constructedQuery strings.Builder
	for _, repo := range repoNames {
		repoArg := fmt.Sprintf("repo:%s+", strings.Replace(repo, "github.com/", "", 1))
		constructedQuery.Write([]byte(repoArg))
	}
	constructedQuery.Write([]byte(strings.Replace(searchPattern, " ", "+", -1)))
	var t struct {
		Items []RestIssue
	}
	c.requestGet(ctx, fmt.Sprintf("/search/issues?q=%s", constructedQuery.String()), &t)

	return t.Items
}

func (c *Client) FetchIssues(ctx context.Context, repoName api.RepoName) []RestIssue {

	items := make([]RestIssue, 0)
	// repoName = strings.Replace(string(repoName), "github.com/", "", 1)
	// fmt.Println("REPONAME", fmt.Sprintf("repos/%s/issues", string(repoName)))
	c.requestGet(ctx, "/repos/sourcegraph/sourcegraph/issues", &items)
	fmt.Println("T ITEMS: %#v", items)
	return items
}
