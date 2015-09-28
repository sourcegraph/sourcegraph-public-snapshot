package issues

import "sourcegraph.com/sourcegraph/sourcegraph/platform"

func init() {
	platform.RegisterFrame(platform.RepoFrame{
		ID:      "issues",
		Title:   "Issues",
		Icon:    "issue-opened",
		Handler: Handler{},
	})
}
