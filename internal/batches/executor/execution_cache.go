package executor

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution/cache"
)

func NewDiskCache(dir string) cache.Cache {
	if dir == "" {
		return &ExecutionNoOpCache{}
	}

	return &ExecutionDiskCache{dir}
}

type ExecutionDiskCache struct {
	Dir string
}

const cacheFileExt = ".json"

func (c ExecutionDiskCache) cacheFilePath(key cache.Keyer) (string, error) {
	keyString, err := key.Key()
	if err != nil {
		return "", errors.Wrap(err, "calculating execution cache key")
	}

	return filepath.Join(c.Dir, key.Slug(), keyString+cacheFileExt), nil
}

func readCacheFile(path string, result interface{}) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	if err := json.Unmarshal(data, result); err != nil {
		// Delete the invalid data to avoid causing an error for next time.
		if err := os.Remove(path); err != nil {
			return false, errors.Wrap(err, "while deleting cache file with invalid JSON")
		}
		return false, errors.Wrapf(err, "reading cache file %s", path)
	}

	return true, nil
}

func (c ExecutionDiskCache) writeCacheFile(path string, result interface{}) error {
	raw, err := json.Marshal(result)
	if err != nil {
		return errors.Wrap(err, "serializing cache content to JSON")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	return os.WriteFile(path, raw, 0600)
}

func (c ExecutionDiskCache) Clear(ctx context.Context, key cache.Keyer) error {
	path, err := c.cacheFilePath(key)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(path)
}

func (c ExecutionDiskCache) Get(ctx context.Context, key cache.Keyer) (execution.AfterStepResult, bool, error) {
	var result execution.AfterStepResult
	path, err := c.cacheFilePath(key)
	if err != nil {
		return result, false, err
	}

	found, err := readCacheFile(path, &result)
	if err != nil {
		return result, false, err
	}

	return result, found, nil
}

func (c ExecutionDiskCache) Set(ctx context.Context, key cache.Keyer, result execution.AfterStepResult) error {
	path, err := c.cacheFilePath(key)
	if err != nil {
		return err
	}

	return c.writeCacheFile(path, &result)
}

// ExecutionNoOpCache is an implementation of ExecutionCache that does not store or
// retrieve cache entries.
type ExecutionNoOpCache struct{}

func (ExecutionNoOpCache) Clear(ctx context.Context, key cache.Keyer) error {
	return nil
}

func (ExecutionNoOpCache) Set(ctx context.Context, key cache.Keyer, result execution.AfterStepResult) error {
	return nil
}

func (ExecutionNoOpCache) Get(ctx context.Context, key cache.Keyer) (execution.AfterStepResult, bool, error) {
	return execution.AfterStepResult{}, false, nil
}
