package githubapi

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/apps/tracker/issues"
)

// markRead marks the specified issue as read for current user.
func (s service) markRead(ctx context.Context, repo issues.RepoSpec, id uint64) error {
	if s.notifications == nil {
		return nil
	}
	return s.notifications.MarkRead(ctx, "Issue", repo, id)
}
