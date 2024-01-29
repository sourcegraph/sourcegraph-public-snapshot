package codyignore

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/paths"
)

// Service allows for operating .cody/ignore files.
type Service interface {

	// IsIgnored returns true if given file is ignored for Cody at given commit, and false otherwise.
	// Error indicates failure to check. In case of error false is always returned.
	// The client should handle the error properly and gracefully ensure Cody keeps operating.
	IsIgnored(ctx context.Context, repoName api.RepoName, commitID api.CommitID, filePath string) (bool, error)
}

// NewService creates a .cody/ignore service operating against a gitserver.
func NewService(g gitserver.Client) Service {
	return &service{gitserverClient: g}
}

type service struct {
	gitserverClient gitserver.Client
}

var codyignorePaths = []string{
	".cody/ignore",
}

func (s *service) IsIgnored(ctx context.Context, repoName api.RepoName, commitID api.CommitID, filePath string) (bool, error) {
	// read the first file gitserver finds, among possible paths for codyignore:
	var (
		r io.ReadCloser
		lastErr error
	)
	for _, path := range codyignorePaths {
		var err error
		r, err = s.gitserverClient.NewFileReader(
			ctx,
			repoName,
			commitID,
			path,
		)
		if err != nil {
			if !os.IsNotExist(err) {
				lastErr = err
			}
			continue
		}
	}
	if r == nil {
		return false, lastErr
	}
	defer r.Close()
	// read line by line, ignoring blank lines:
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		p, err := paths.Compile(line)
		if err != nil {
			// TODO: Log path compile errors
			continue
		}
		if p.Match(filePath) {
			return true, nil
		}
	}
	// no match found
	return false, nil
}
