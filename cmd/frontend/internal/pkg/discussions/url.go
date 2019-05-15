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
	var u *url.URL

	// TODO(sqs): This only takes the 1st target. Support multiple targets.
	targets, err := db.DiscussionThreads.ListTargets(ctx, t.ID)
	if err != nil {
		return nil, errors.Wrap(err, "DiscussionThreads.ListTargets")
	}
	switch {
	case len(targets) > 0:
		// TODO(slimsag:discussions): future: Consider how to handle cases like:
		// - repo renames
		// - file paths not existing on the default branch (or at all).

		target := targets[0]
		repo, err := db.Repos.Get(ctx, target.RepoID)
		if err != nil {
			return nil, errors.Wrap(err, "db.Repos.Get")
		}
		if target.Path == nil {
			return nil, nil // Can't generate a link to this yet, we don't have a UI for it yet.
		}
		u = &url.URL{Path: path.Join("/", string(repo.Name), "/-/blob/", *target.Path)}

		fragment := url.Values{}
		fragment.Set("tab", "discussions")
		fragment.Set("threadID", strconv.FormatInt(t.ID, 10))
		if c != nil {
			fragment.Set("commentID", strconv.FormatInt(c.ID, 10))
		}
		encFragment := fragment.Encode()
		if target.StartLine != nil {
			encFragment = fmt.Sprintf("L%d&%s", *target.StartLine+1, encFragment)
		}
		u.Fragment = encFragment
	default:
		return nil, nil // can't generate a link to this target type
	}
	return u, nil
}
