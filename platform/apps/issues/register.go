package issues

import (
	issuesnext "src.sourcegraph.com/apps/issues/sgapp"
	"src.sourcegraph.com/sourcegraph/conf/feature"
	"src.sourcegraph.com/sourcegraph/platform"
)

func init() {
	switch feature.Features.IssuesNext {
	case true:
		issuesnext.Init()
	case false:
		platform.RegisterFrame(platform.RepoFrame{
			ID:      "issues",
			Title:   "Issues",
			Icon:    "issue-opened",
			Handler: Handler{},
		})
	}
}
