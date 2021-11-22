package gitserver

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

type gitserverClient struct{}

func NewClient() GitserverClient {
	return &gitserverClient{}
}

func (c *gitserverClient) FetchTar(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string) (io.ReadCloser, error) {
	return gitserver.DefaultClient.Archive(ctx, repo, gitserver.ArchiveOptions{Treeish: string(commit), Format: "tar", Paths: paths})
}

func (c *gitserverClient) GitDiff(ctx context.Context, repo api.RepoName, commitA, commitB api.CommitID) (*Changes, error) {
	command := gitserver.DefaultClient.Command("git", "diff", "-z", "--name-status", "--no-renames", string(commitA), string(commitB))
	command.Repo = repo

	output, err := command.Output(ctx)
	if err != nil {
		return nil, err
	}

	// The output is a a repeated sequence of:
	//
	//     <status> NUL <path> NUL
	//
	// where NUL is the 0 byte.
	//
	// Example:
	//
	//     M NUL cmd/symbols/internal/symbols/fetch.go NUL

	changes := Changes{}
	slices := bytes.Split(output, []byte{0})
	for i := 0; i < len(slices)-1; i += 2 {
		statusIdx := i
		fileIdx := i + 1

		if len(slices[statusIdx]) == 0 {
			return nil, fmt.Errorf("unrecognized git diff output (from repo %q, commitA %q, commitB %q): status was empty at index %d", repo, commitA, commitB, i)
		}

		status := slices[statusIdx][0]
		path := string(slices[fileIdx])

		switch status {
		case 'A':
			changes.Added = append(changes.Added, path)
		case 'M':
			changes.Modified = append(changes.Modified, path)
		case 'D':
			changes.Deleted = append(changes.Deleted, path)
		}
	}

	return &changes, nil
}
