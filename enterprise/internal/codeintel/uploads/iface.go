package uploads

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/policies"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/background"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
)

type UploadService interface {
	GetDirtyRepositories(ctx context.Context) (_ map[int]int, err error)
	GetRepositoriesMaxStaleAge(ctx context.Context) (_ time.Duration, err error)
}

type Locker = background.Locker

type GitserverClient interface {
	policies.GitserverClient
	background.GitserverClient

	ListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) (_ []*gitdomain.Tag, err error)
}

type RepoStore = background.RepoStore

type PolicyService = background.PolicyService

type PolicyMatcher = background.PolicyMatcher
