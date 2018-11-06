package github

import (
	"context"
	"fmt"
	"strings"
)

type RestIssue struct {
	Title   string
	HTMLURL string `json:"html_url"`
	Body    string
}

func (c *Client) SearchIssues(ctx context.Context, repoNames []string, repoPattern string) []RestIssue {
	var constructedQuery strings.Builder
	for _, repo := range repoNames {
		repoArg := fmt.Sprintf("repo:%s+", strings.Replace(repo, "github.com/", "", 1))
		constructedQuery.Write([]byte(repoArg))
	}
	constructedQuery.Write([]byte(strings.Replace(repoPattern, " ", "+", -1)))
	var t struct {
		Items []RestIssue
	}
	c.requestGet(ctx, fmt.Sprintf("/search/issues?q=%s", constructedQuery.String()), &t)

	return t.Items
}
