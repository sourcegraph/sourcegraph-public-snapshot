package news

import (
	"context"
	"net/http"

	"github.com/shurcooL/githubv4"
)

type label struct {
	Name string
}

type discussion struct {
	Title  string
	Body   string
	Author struct {
		Login string
	}
	Labels struct {
		Nodes []label
	} `graphql:"labels(first:5)"`
}

func getDiscussions(ctx context.Context) ([]discussion, error) {
	client := githubv4.NewClient(http.DefaultClient)

	var q struct {
		Repository struct {
			Discussions struct {
				Nodes []discussion
			} `graphql:"discussions(categoryId:\"DIC_kwDOAnYEBM4B_WVJ\",first:10)"`
		} `graphql:"repository(owner:\"sourcegraph\",name:\"sourcegraph\")"`
	}

	err := client.Query(ctx, &q, map[string]any{})
	if err != nil {
		return nil, err
	}
	return q.Repository.Discussions.Nodes, nil
}
