package discussions

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

// URLToInlineThread returns a URL to the discussion thread's 'inline' view
// (i.e. the filepath/blob view).
//
// Returns nil, nil if the thread does not have an inline thread view. e.g.,
// for threads created not on a file but on something else.
func URLToInlineThread(ctx context.Context, thread *types.DiscussionThread) (*url.URL, error) {
	return urlToInline(ctx, thread, nil)
}

// URLToInlineComment returns a URL to the discussion thread comment's 'inline'
// view (i.e. the filepath/blob view).
//
// Returns nil, nil if the thread does not have an inline thread view. e.g.,
// for threads created not on a file but on something else.
func URLToInlineComment(ctx context.Context, thread *types.DiscussionThread, comment *types.DiscussionComment) (*url.URL, error) {
	return urlToInline(ctx, thread, comment)
}

func urlToInline(ctx context.Context, t *types.DiscussionThread, c *types.DiscussionComment) (*url.URL, error) {
	targets, err := db.DiscussionThreads.ListTargets(ctx, db.DiscussionThreadsListTargetsOptions{ThreadID: t.ID})
	if err != nil {
		return nil, errors.Wrap(err, "DiscussionThreads.ListTargets")
	}
	// TODO(sqs): This only takes the 1st target. Support multiple targets.
	if len(targets) > 0 {
		var commentID *int64
		if c != nil {
			commentID = &c.ID
		}
		return URLToInlineTarget(ctx, targets[0], &t.ID, commentID)
	}
	return nil, nil // can't generate a link to this target type
}

// URLToInlineTarget returns a URL to the discussion thread target's inline view (i.e., in a file).
func URLToInlineTarget(ctx context.Context, target *types.DiscussionThreadTargetRepo, threadID, commentID *int64) (*url.URL, error) {
	// TODO(slimsag:discussions): future: Consider how to handle cases like:
	// - repo renames
	// - file paths not existing on the default branch (or at all).
	repo, err := db.Repos.Get(ctx, target.RepoID)
	if err != nil {
		return nil, errors.Wrap(err, "db.Repos.Get")
	}
	if target.Path == nil {
		return nil, nil // Can't generate a link to this yet, we don't have a UI for it yet.
	}
	u := &url.URL{Path: path.Join("/", string(repo.Name), "/-/blob/", *target.Path)}

	fragment := url.Values{}
	fragment.Set("tab", "discussions")
	if threadID != nil {
		fragment.Set("threadID", strconv.FormatInt(*threadID, 10))
	}
	if commentID != nil {
		fragment.Set("commentID", strconv.FormatInt(*commentID, 10))
	}
	encFragment := fragment.Encode()
	if target.StartLine != nil {
		encFragment = fmt.Sprintf("L%d&%s", *target.StartLine+1, encFragment)
	}
	u.Fragment = encFragment
	return u, nil
}
