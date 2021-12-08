package fileskip

import (
	"fmt"
	"github.com/cockroachdb/errors"
	"os"
)

func ReadCache(cacheDir string) (r *RepoIndex, err error) {
	file, err := os.Open(cacheDir)
	if err != nil {
		return nil, err
	}
	result, err := DeserializeRepoIndex(file)
	if err != nil {
		return nil, err
	}
	if result.Blobs == nil {
		return nil, errors.Errorf("results.Blobs is nil")
	}
	return result, nil
}

func WriteCache(dir, cacheDir string) error {
	r, err := NewInMemoryRepoIndex(&GitFileSystem{dir})
	if err != nil {
		return err
	}
	fmt.Println("Writing index...")
	return r.SerializeToFile(cacheDir)
}
