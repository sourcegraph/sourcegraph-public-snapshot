package backend

import (
	"bytes"
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"

	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/proto"
)

// OwnService gives access to code ownership data.
// At this point only data from CODEOWNERS file is presented, if available.
type OwnService interface {
	// OwnersFile returns a CODEOWNERS file from a given repository at given commit ID.
	// In the case the file can not be found, `nil` `*codeownerspb.File` and `nil` `error` is returned.
	OwnersFile(context.Context, api.RepoName, api.CommitID) (*codeownerspb.File, error)
}

var _ OwnService = ownService{}

func NewOwnService(git gitserver.Client) OwnService {
	return ownService{git: git}
}

type ownService struct {
	git gitserver.Client
}

// codeownersLocations contains the locations where CODEOWNERS file
// is expected to be found relative to the repository root directory.
// These are in line with GitHub and GitLab documentation.
var codeownersLocations = []string{
	"CODEOWNERS",
	".github/CODEOWNERS",
	".gitlab/CODEOWNERS",
	"docs/CODEOWNERS",
}

func (s ownService) OwnersFile(ctx context.Context, repoName api.RepoName, commitID api.CommitID) (*codeownerspb.File, error) {
	var content []byte
	var err error
	for _, path := range codeownersLocations {
		content, err = s.git.ReadFile(
			ctx,
			authz.DefaultSubRepoPermsChecker,
			repoName,
			commitID,
			path,
		)
		if err == nil && content != nil {
			break
		}
	}
	if content == nil {
		return nil, nil
	}
	return codeowners.Parse(bytes.NewReader(content))
}
