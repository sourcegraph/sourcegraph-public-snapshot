package gitserverfs

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TempDirName is the name used for the temporary directory under ReposDir.
const TempDirName = ".tmp"

// P4HomeName is the name used for the directory that git p4 will use as $HOME
// and where it will store cache data.
const P4HomeName = ".p4home"

func MakeP4HomeDir(reposDir string) (p4home string, _ error) {
	p4Home := filepath.Join(reposDir, P4HomeName)
	// Ensure the directory exists
	if err := os.MkdirAll(p4Home, os.ModePerm); err != nil {
		return "", errors.Wrapf(err, "ensuring p4Home exists: %q", p4Home)
	}
	return p4home, nil
}

func RepoDirFromName(reposDir string, name api.RepoName) common.GitDir {
	p := string(protocol.NormalizeRepo(name))
	return common.GitDir(filepath.Join(reposDir, filepath.FromSlash(p), ".git"))
}

func RepoNameFromDir(reposDir string, dir common.GitDir) api.RepoName {
	// dir == ${s.ReposDir}/${name}/.git
	parent := filepath.Dir(string(dir))                   // remove suffix "/.git"
	name := strings.TrimPrefix(parent, reposDir)          // remove prefix "${s.ReposDir}"
	name = strings.Trim(name, string(filepath.Separator)) // remove /
	name = filepath.ToSlash(name)                         // filepath -> path
	return protocol.NormalizeRepo(api.RepoName(name))
}

// TempDir is a wrapper around os.MkdirTemp, but using the given reposDir
// temporary directory filepath.Join(s.ReposDir, tempDirName).
//
// This directory is cleaned up by gitserver and will be ignored by repository
// listing operations.
func TempDir(reposDir, prefix string) (name string, err error) {
	// TODO: At runtime, this directory always exists. We only need to ensure
	// the directory exists here because tests use this function without creating
	// the directory first. Ideally, we can remove this later.
	tmp := filepath.Join(reposDir, TempDirName)
	if err := os.MkdirAll(tmp, os.ModePerm); err != nil {
		return "", err
	}
	return os.MkdirTemp(tmp, prefix)
}

func IgnorePath(reposDir string, path string) bool {
	// We ignore any path which starts with .tmp or .p4home in ReposDir
	if filepath.Dir(path) != reposDir {
		return false
	}
	base := filepath.Base(path)
	return strings.HasPrefix(base, TempDirName) || strings.HasPrefix(base, P4HomeName)
}
