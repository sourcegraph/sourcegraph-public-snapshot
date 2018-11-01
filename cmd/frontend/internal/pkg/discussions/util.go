package discussions

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/discussions/ratelimit"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

// InsecureAddCommentToThread handles adding a new comment to an existing
// thread. It handles:
//
// 1. Rate limiting (NOT general permission handling).
// 2. Creating the actual database entry.
// 3. Notifying other users of the new comment.
// 4. Fetching and returning the updated thread.
//
// It does NOT verify that the user has permission to create this comment. That
// is the responsibility of the caller.
func InsecureAddCommentToThread(ctx context.Context, newComment *types.DiscussionComment) (*types.DiscussionThread, error) {
	if dc := conf.Get().Discussions; dc != nil && dc.AbuseProtection {
		if mustWait := ratelimit.TimeUntilUserCanAddCommentToThread(ctx, newComment.AuthorUserID, newComment.Contents); mustWait != 0 {
			return nil, fmt.Errorf("You are creating comments too quickly. You may create a new one after %v", mustWait.Round(time.Second))
		}
	}

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
