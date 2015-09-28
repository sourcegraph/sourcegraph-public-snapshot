package cli

import (
	"sourcegraph.com/sourcegraph/rwvfs"
	"sourcegraph.com/sourcegraph/srclib/buildstore"
)

func GetBuildDataFS(commitID string) (rwvfs.FileSystem, error) {
	lrepo, err := OpenLocalRepo()
	if lrepo == nil || lrepo.RootDir == "" || commitID == "" {
		return nil, err
	}
	localStore, err := buildstore.LocalRepo(lrepo.RootDir)
	if err != nil {
		return nil, err
	}
	return localStore.Commit(commitID), nil
}
