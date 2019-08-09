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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_384(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
