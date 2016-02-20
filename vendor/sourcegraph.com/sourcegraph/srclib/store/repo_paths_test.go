package store

import (
	"encoding/hex"
	"fmt"

	"sourcegraph.com/sourcegraph/rwvfs"
)

type customRepoPaths struct{}

// RepoToPath implements RepoPaths.
func (e *customRepoPaths) RepoToPath(repo string) []string {
	h := make([]byte, hex.EncodedLen(len(repo)))
	hex.Encode(h, []byte(repo))
	return []string{string(h)}
}

// PathToRepo implements RepoPaths.
func (e *customRepoPaths) PathToRepo(path []string) string {
	c, err := hex.DecodeString(path[0])
	if err != nil {
		panic(fmt.Sprintf("hex-decoding path %q: %s", path, err))
	}
	return string(c)
}

// ListRepoPaths implements RepoPaths.
func (e *customRepoPaths) ListRepoPaths(vfs rwvfs.WalkableFileSystem, after string, max int) ([][]string, error) {
	entries, err := vfs.ReadDir(".")
	if err != nil {
		return nil, err
	}
	paths := make([][]string, len(entries))
	for i, e := range entries {
		paths[i] = []string{e.Name()}
	}
	return paths, nil
}
