package campaigns

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func UserCacheDir() (string, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userCacheDir, "sourcegraph-src"), nil
}

type ExecutionCacheKey struct {
	Repo ActionRepo
	Runs []*ActionStep
}

type ExecutionCache interface {
	Get(ctx context.Context, key ExecutionCacheKey) (result PatchInput, ok bool, err error)
	Set(ctx context.Context, key ExecutionCacheKey, result PatchInput) error
	Clear(ctx context.Context, key ExecutionCacheKey) error
}

type ExecutionDiskCache struct {
	Dir string
}

func (c ExecutionDiskCache) cacheFilePath(key ExecutionCacheKey) (string, error) {
	keyJSON, err := json.Marshal(key)
	if err != nil {
		return "", errors.Wrap(err, "Failed to marshal JSON when generating action cache key")
	}

	b := sha256.Sum256(keyJSON)
	keyString := base64.RawURLEncoding.EncodeToString(b[:16])

	return filepath.Join(c.Dir, keyString+".json"), nil
}

func (c ExecutionDiskCache) Get(ctx context.Context, key ExecutionCacheKey) (PatchInput, bool, error) {
	path, err := c.cacheFilePath(key)
	if err != nil {
		return PatchInput{}, false, err
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil // treat as not-found
		}
		return PatchInput{}, false, err
	}

	var result PatchInput
	if err := json.Unmarshal(data, &result); err != nil {
		// Delete the invalid data to avoid causing an error for next time.
		if err := os.Remove(path); err != nil {
			return PatchInput{}, false, errors.Wrap(err, "while deleting cache file with invalid JSON")
		}
		return PatchInput{}, false, errors.Wrapf(err, "reading cache file %s", path)
	}

	return result, true, nil
}

func (c ExecutionDiskCache) Set(ctx context.Context, key ExecutionCacheKey, result PatchInput) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	path, err := c.cacheFilePath(key)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0600)
}

func (c ExecutionDiskCache) Clear(ctx context.Context, key ExecutionCacheKey) error {
	path, err := c.cacheFilePath(key)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(path)
}

// ExecutionNoOpCache is an implementation of actionExecutionCache that does not store or
// retrieve cache entries.
type ExecutionNoOpCache struct{}

func (ExecutionNoOpCache) Get(ctx context.Context, key ExecutionCacheKey) (result PatchInput, ok bool, err error) {
	return PatchInput{}, false, nil
}

func (ExecutionNoOpCache) Set(ctx context.Context, key ExecutionCacheKey, result PatchInput) error {
	return nil
}

func (ExecutionNoOpCache) Clear(ctx context.Context, key ExecutionCacheKey) error {
	return nil
}
