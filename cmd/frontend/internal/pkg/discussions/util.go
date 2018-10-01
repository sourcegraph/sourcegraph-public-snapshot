package discussions

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

// InsecureAddCommentToThread handles adding a new comment to an existing
// thread. It handles:
//
// 1. Creating the actual database entry.
// 2. Notifying other users of the new comment.
// 3. Fetching and returning the updated thread.
//
// It does NOT verify that the user has permission to create this comment. That
// is the responsibility of the caller.
func InsecureAddCommentToThread(ctx context.Context, newComment *types.DiscussionComment) (*types.DiscussionThread, error) {
	_, err := db.DiscussionComments.Create(ctx, newComment)
	if err != nil {
		return nil, err // Intentionally not wrapping the error here for cleaner error messages.
	}

	updatedThread, err := db.DiscussionThreads.Get(ctx, newComment.ThreadID)
	if err != nil {
		return nil, errors.Wrap(err, "DiscussionThreads.Get")
	}
	NotifyNewComment(updatedThread, newComment)
	return updatedThread, nil
}
