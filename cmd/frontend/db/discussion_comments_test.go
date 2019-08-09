package db

import (
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

// TODO(slimsag:discussions): future: test that DiscussionCommentsListOptions.AuthorUserID works
// TODO(slimsag:discussions): future: test that DiscussionCommentsListOptions.ThreadID works

func TestDiscussionComments_Create(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	user, err := Users.Create(ctx, NewUser{
		Email:                 "a@a.com",
		Username:              "u",
		Password:              "p",
		EmailVerificationCode: "c",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create a repository to comply with the postgres repo constraint.
	if err := Repos.Upsert(ctx, api.InsertRepoOp{Name: "myrepo", Description: "", Fork: false, Enabled: true}); err != nil {
		t.Fatal(err)
	}
	repo, err := Repos.GetByName(ctx, "myrepo")
	if err != nil {
		t.Fatal(err)
	}

	// Create the thread.
	thread, err := DiscussionThreads.Create(ctx, &types.DiscussionThread{
		AuthorUserID: user.ID,
		Title:        "Hello world!",
		TargetRepo: &types.DiscussionThreadTargetRepo{
			RepoID:   repo.ID,
			Path:     strPtr("foo/bar/mux.go"),
			Branch:   strPtr("master"),
			Revision: strPtr("0c1a96370c1a96370c1a96370c1a96370c1a9637"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DiscussionThreads.Get(ctx, thread.ID); err != nil {
		t.Fatal("expected to get created thread", err)
	}

	// Create the comment.
	comment, err := DiscussionComments.Create(ctx, &types.DiscussionComment{
		ThreadID:     thread.ID,
		AuthorUserID: user.ID,
		Contents:     "What do you think of Hello World as a Service?",
	})
	if err != nil {
		t.Fatal(err)
	}
	if comment.CreatedAt == (time.Time{}) {
		t.Fatal("expected CreatedAt to be set, got zero value time")
	}
	if _, err := DiscussionComments.Get(ctx, comment.ID); err != nil {
		t.Fatal("expected to get created comment", err)
	}

	// Update the comment.
	const wantCommentContents = "x"
	if _, err := DiscussionComments.Update(ctx, comment.ID, &DiscussionCommentsUpdateOptions{
		Contents: strPtr(wantCommentContents),
	}); err != nil {
		t.Fatal(err)
	}
	if comment, err := DiscussionComments.Get(ctx, comment.ID); err != nil {
		t.Fatal("expected to get created comment", err)
	} else if comment.Contents != wantCommentContents {
		t.Errorf("got comment contents %q, want %q", comment.Contents, wantCommentContents)
	}

	// Test deleting the repo cascade deletes
	err = Repos.Delete(ctx, repo.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = DiscussionComments.Get(ctx, comment.ID)
	if _, ok := err.(*ErrCommentNotFound); !ok {
		t.Fatal("expected to not find deleted comment", err)
	}
	_, err = DiscussionThreads.Get(ctx, thread.ID)
	if _, ok := err.(*ErrThreadNotFound); !ok {
		t.Fatal("expected to not find deleted thread", err)
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_43(size int) error {
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
