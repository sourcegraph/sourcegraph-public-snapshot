package gitcli

import (
	"github.com/sourcegraph/log"

	"github.com/hashicorp/golang-lru/v2"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/git"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
)

func NewBackend(logger log.Logger, rcf *wrexec.RecordingCommandFactory, dir common.GitDir, repoName api.RepoName) git.GitBackend {
	return &gitCLIBackend{
		logger:         logger,
		rcf:            rcf,
		dir:            dir,
		repoName:       repoName,
		revAtTimeCache: globalRevAtTimeCache,
	}
}

type gitCLIBackend struct {
	logger         log.Logger
	rcf            *wrexec.RecordingCommandFactory
	dir            common.GitDir
	repoName       api.RepoName
	revAtTimeCache *lru.Cache[revAtTimeCacheKey, api.CommitID]
}
