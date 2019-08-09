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
		u = &url.URL{Path: path.Join("/", string(repo.Name), "/-/blob/", *t.TargetRepo.Path)}

		fragment := url.Values{}
		fragment.Set("tab", "discussions")
		fragment.Set("threadID", strconv.FormatInt(t.ID, 10))
		if c != nil {
			fragment.Set("commentID", strconv.FormatInt(c.ID, 10))
		}
		encFragment := fragment.Encode()
		if t.TargetRepo.StartLine != nil {
			encFragment = fmt.Sprintf("L%d&%s", *t.TargetRepo.StartLine+1, encFragment)
		}
		u.Fragment = encFragment
	default:
		return nil, nil // can't generate a link to this target type
	}
	return u, nil
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_383(size int) error {
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
