package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/dev/codeintel-qa/internal"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type uploadMeta struct {
	id       string
	repoName string
	commit   string
}

// uploadAll uploads the dumps for the commits present in the given commitsByRepo map.
// Uploads are performed concurrently given the limiter instance as well as the set of
// flags supplied by the user. This function returns a slice of uploadMeta containing
// the graphql identifier of the uploaded resources.
func uploadAll(ctx context.Context, commitsByRepo map[string][]string, limiter *internal.Limiter) ([]uploadMeta, error) {
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

				fmt.Printf("[%5s] %s Uploading index for %s@%s\n", internal.TimeSince(start), internal.EmojiLightbulb, repoName, commit[:7])

				graphqlID, err := upload(ctx, internal.MakeTestRepoName(repoName), commit, file)
				if err != nil {
					errCh <- err
					return
				}

				fmt.Printf("[%5s] %s Finished uploading index for %s@%s\n", internal.TimeSince(start), internal.EmojiSuccess, repoName, commit[:7])

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

// upload invokes `src lsif upload` on the host and returns the graphql identifier of
// the uploaded resource.
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
	cmd.Env = []string{
		fmt.Sprintf("SRC_ENDPOINT=%s", internal.SourcegraphEndpoint),
		fmt.Sprintf("SRC_ACCESS_TOKEN=%s", internal.SourcegraphAccessToken),
	}

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
