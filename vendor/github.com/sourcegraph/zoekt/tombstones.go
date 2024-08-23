package zoekt

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

var mockRepos []*Repository

// SetTombstone idempotently sets a tombstone for repoName in .meta.
func SetTombstone(shardPath string, repoID uint32) error {
	return setTombstone(shardPath, repoID, true)
}

// UnsetTombstone idempotently removes a tombstones for reopName in .meta.
func UnsetTombstone(shardPath string, repoID uint32) error {
	return setTombstone(shardPath, repoID, false)
}

func setTombstone(shardPath string, repoID uint32, tombstone bool) error {
	var repos []*Repository
	var err error

	if mockRepos != nil {
		repos = mockRepos
	} else {
		repos, _, err = ReadMetadataPath(shardPath)
		if err != nil {
			return err
		}
	}

	for _, repo := range repos {
		if repo.ID == repoID {
			repo.Tombstone = tombstone
		}
	}

	tempPath, finalPath, err := JsonMarshalRepoMetaTemp(shardPath, repos)
	if err != nil {
		return err
	}

	err = os.Rename(tempPath, finalPath)
	if err != nil {
		os.Remove(tempPath)
	}

	return nil
}

// JsonMarshalRepoMetaTemp writes the json encoding of the given repository metadata to a temporary file
// in the same directory as the given shard path. It returns both the path of the temporary file and the
// path of the final file that the caller should use.
//
// The caller is responsible for renaming the temporary file to the final file path, or removing
// the temporary file if it is no longer needed.
// TODO: Should we stick this in a util package?
func JsonMarshalRepoMetaTemp(shardPath string, repositoryMetadata interface{}) (tempPath, finalPath string, err error) {
	finalPath = shardPath + ".meta"

	b, err := json.Marshal(repositoryMetadata)
	if err != nil {
		return "", "", fmt.Errorf("marshalling json: %w", err)
	}

	f, err := os.CreateTemp(filepath.Dir(finalPath), filepath.Base(finalPath)+".*.tmp")
	if err != nil {
		return "", "", fmt.Errorf("writing temporary file: %s", err)
	}

	defer func() {
		f.Close()
		if err != nil {
			_ = os.Remove(f.Name())
		}
	}()

	err = f.Chmod(0o666 &^ umask)
	if err != nil {
		return "", "", fmt.Errorf("chmoding temporary file: %s", err)
	}

	_, err = f.Write(b)
	if err != nil {
		return "", "", fmt.Errorf("writing json to temporary file: %s", err)
	}

	return f.Name(), finalPath, nil
}

// umask holds the Umask of the current process
var umask os.FileMode
