package discussions

import (
	"context"
	"fmt"
	"net/url"
	"path"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
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
	switch {
	case t.TargetRepo != nil:
		// TODO(slimsag:discussions): future: Consider how to handle cases like:
		// - repo renames
		// - file paths not existing on the default branch (or at all).

		repo, err := db.Repos.Get(ctx, t.TargetRepo.RepoID)
		if err != nil {
			return nil, errors.Wrap(err, "db.Repos.Get")
		}
		if t.TargetRepo.Path == nil {
			return nil, nil // Can't generate a link to this yet, we don't have a UI for it yet.
		}
		u = &url.URL{Path: path.Join("/", string(repo.URI), "/-/blob/", *t.TargetRepo.Path)}

		// TODO(slimsag:discussions): frontend doesn't link to the comment directly
		// unless these are in this exact order. Why?
		//fragment := url.Values{}
		//fragment.Set("tab", "discussions")
		//fragment.Set("threadID", strconv.FormatInt(t.ID, 10))
		//fragment.Set("commentID", strconv.FormatInt(c.ID, 10))
		//u.Fragment = fragment.Encode()
		u.Fragment = fmt.Sprintf("tab=discussions&threadID=%v", t.ID)
		if c != nil {
			u.Fragment += fmt.Sprintf("&commentID=%v", c.ID)
		}
	default:
		return nil, nil // can't generate a link to this target type
	}
	return globals.AppURL.ResolveReference(u), nil
}
