package jobselector

import (
	"context"
	"time"

	"github.com/grafana/regexp"

	codeinteltypes "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type GitserverClient interface {
	Head(ctx context.Context, repositoryID int) (string, bool, error)
	CommitExists(ctx context.Context, repositoryID int, commit string) (bool, error)
	ListFiles(ctx context.Context, repositoryID int, commit string, pattern *regexp.Regexp) ([]string, error)
	FileExists(ctx context.Context, repositoryID int, commit, file string) (bool, error)
	RawContents(ctx context.Context, repositoryID int, commit, file string) ([]byte, error)
	ResolveRevision(ctx context.Context, repositoryID int, versionString string) (api.CommitID, error)
	ListTags(ctx context.Context, repo api.RepoName, commitObjs ...string) (_ []*gitdomain.Tag, err error)

	CommitDate(ctx context.Context, repositoryID int, commit string) (string, time.Time, bool, error)
	RefDescriptions(ctx context.Context, repositoryID int, gitOjbs ...string) (map[string][]gitdomain.RefDescription, error)
	CommitsUniqueToBranch(ctx context.Context, repositoryID int, branchName string, isDefaultBranch bool, maxAge *time.Time) (map[string]time.Time, error)
}

type InferenceService interface {
	InferIndexJobs(ctx context.Context, repo api.RepoName, commit, overrideScript string) ([]config.IndexJob, error)
	InferIndexJobHints(ctx context.Context, repo api.RepoName, commit, overrideScript string) ([]config.IndexJobHint, error)
}

type UploadService interface {
	GetRepoName(ctx context.Context, repositoryID int) (_ string, err error)                // upload service
	GetDirtyRepositories(ctx context.Context) (_ map[int]int, err error)                    // upload service
	GetUploadsByIDs(ctx context.Context, ids ...int) (_ []codeinteltypes.Upload, err error) // upload service
	GetUploadByID(ctx context.Context, id int) (codeinteltypes.Upload, bool, error)
	ReferencesForUpload(ctx context.Context, uploadID int) (shared.PackageReferenceScanner, error)
	GetRepositoriesForIndexScan(ctx context.Context, table, column string, processDelay time.Duration, allowGlobalPolicies bool, repositoryMatchLimit *int, limit int, now time.Time) (_ []int, err error)
}
