package git

import (
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/wrexec"
)

// SetRepositoryType sets the type of the repository.
func SetRepositoryType(rcf *wrexec.RecordingCommandFactory, reposDir string, dir common.GitDir, typ string) error {
	return ConfigSet(rcf, reposDir, dir, "sourcegraph.type", typ)
}

// GetRepositoryType returns the type of the repository.
func GetRepositoryType(rcf *wrexec.RecordingCommandFactory, reposDir string, dir common.GitDir) (string, error) {
	val, err := ConfigGet(rcf, reposDir, dir, "sourcegraph.type")
	if err != nil {
		return "", err
	}
	return val, nil
}
