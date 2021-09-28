package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/dev/codeintel-qa/internal"
)

type uploadMeta struct {
	id       string
	repoName string
	commit   string
}

func uploadAll(ctx context.Context, commitsByRepo map[string][]string, limiter *limiter, start time.Time) ([]uploadMeta, error) {
	n := 0
	for _, commits := range commitsByRepo {
		n += len(commits)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, n)
	uploadCh := make(chan uploadMeta, n)

	for repoName, commits := range commitsByRepo {
		for i, commit := range commits {
			wg.Add(1)

			go func(repoName, commit, file string) {
				defer wg.Done()

				if err := limiter.Acquire(ctx); err != nil {
					errCh <- err
					return
				}
				defer limiter.Release()

				fmt.Printf("[%5s] Uploading index for %s@%s\n", internal.TimeSince(start), repoName, commit[:7])

				graphqlID, err := upload(ctx, makeRepoName(repoName), commit, file)
				if err != nil {
					errCh <- err
					return
				}

				fmt.Printf("[%5s] Finished uploading index for %s@%s\n", internal.TimeSince(start), repoName, commit[:7])

				uploadCh <- uploadMeta{
					id:       graphqlID,
					repoName: repoName,
					commit:   commit,
				}
			}(repoName, commit, fmt.Sprintf("%s.%d.%s.dump", repoName, i, commit))
		}
	}

	go func() {
		wg.Wait()
		close(errCh)
		close(uploadCh)
	}()

	for err := range errCh {
		return nil, err
	}

	uploads := make([]uploadMeta, 0, n)
	for upload := range uploadCh {
		uploads = append(uploads, upload)
	}

	return uploads, nil
}

func upload(ctx context.Context, repoName, commit, file string) (string, error) {
	argMap := map[string]string{
		"root":   "/",
		"repo":   repoName,
		"commit": commit,
		"file":   file,
	}

	args := make([]string, 0, len(argMap))
	for k, v := range argMap {
		args = append(args, fmt.Sprintf("-%s=%s", k, v))
	}

	cmd := exec.CommandContext(ctx, "src", append([]string{"lsif", "upload", "-json"}, args...)...)
	cmd.Dir = indexDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("failed to upload index: %s", output))
	}

	resp := struct {
		UploadURL string `json:"uploadUrl"`
	}{}
	if err := json.Unmarshal(output, &resp); err != nil {
		return "", err
	}

	parts := strings.Split(resp.UploadURL, "/")
	return parts[len(parts)-1], nil
}
