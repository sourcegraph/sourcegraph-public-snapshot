package executor

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
	*Task
}

// Key converts the key into a string form that can be used to uniquely identify
// the cache key in a more concise form than the entire Task.
func (key ExecutionCacheKey) Key() (string, error) {
	// We have to resolve the step environments and include them in the cache
	// key to ensure that the cache is properly invalidated when an environment
	// variable changes.
	//
	// Note that we don't base the cache key on the entire global environment:
	// if an unrelated environment variable changes, that's fine. We're only
	// interested in the ones that actually make it into the step container.
	global := os.Environ()
	envs := make([]map[string]string, len(key.Task.Steps))
	for i, step := range key.Task.Steps {
		env, err := step.Env.Resolve(global)
		if err != nil {
			return "", errors.Wrapf(err, "resolving environment for step %d", i)
		}
		envs[i] = env
	}

	raw, err := json.Marshal(struct {
		*Task
		Environments []map[string]string
	}{
		Task:         key.Task,
		Environments: envs,
	})
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(raw)
	return base64.RawURLEncoding.EncodeToString(hash[:16]), nil
}

type ExecutionCache interface {
	Get(ctx context.Context, key ExecutionCacheKey) (result executionResult, found bool, err error)
	Set(ctx context.Context, key ExecutionCacheKey, result executionResult) error
	Clear(ctx context.Context, key ExecutionCacheKey) error
}

func NewCache(dir string) ExecutionCache {
	if dir == "" {
		return &ExecutionNoOpCache{}
	}

	return &ExecutionDiskCache{dir}
}

type ExecutionDiskCache struct {
	Dir string
}

const cacheFileExt = ".v3.json"

func (c ExecutionDiskCache) cacheFilePath(key ExecutionCacheKey) (string, error) {
	keyString, err := key.Key()
	if err != nil {
		return "", errors.Wrap(err, "calculating execution cache key")
	}

	return filepath.Join(c.Dir, keyString+cacheFileExt), nil
}

func (c ExecutionDiskCache) Get(ctx context.Context, key ExecutionCacheKey) (executionResult, bool, error) {
	var result executionResult

	path, err := c.cacheFilePath(key)
	if err != nil {
		return result, false, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return result, false, nil
	}

	if err := c.readCacheFile(path, &result); err != nil {
		return result, false, err
	}

	return result, true, nil
}

// sortCacheFiles sorts cache file paths by their "version", so that files
// ending in `cacheFileExt` are first.
func sortCacheFiles(paths []string) {
	sort.Slice(paths, func(i, j int) bool {
		return !isOldCacheFile(paths[i]) && isOldCacheFile(paths[j])
	})
}

func isOldCacheFile(path string) bool { return !strings.HasSuffix(path, cacheFileExt) }

func (c ExecutionDiskCache) readCacheFile(path string, result *executionResult) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, result); err != nil {
		// Delete the invalid data to avoid causing an error for next time.
		if err := os.Remove(path); err != nil {
			return errors.Wrap(err, "while deleting cache file with invalid JSON")
		}
		return errors.Wrapf(err, "reading cache file %s", path)
	}
	return nil
}

func (c ExecutionDiskCache) Set(ctx context.Context, key ExecutionCacheKey, result executionResult) error {
	path, err := c.cacheFilePath(key)
	if err != nil {
		return err
	}

	raw, err := json.Marshal(&result)
	if err != nil {
		return errors.Wrap(err, "serializing execution result to JSON")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	return ioutil.WriteFile(path, raw, 0600)
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

func (ExecutionNoOpCache) Get(ctx context.Context, key ExecutionCacheKey) (result executionResult, found bool, err error) {
	return executionResult{}, false, nil
}

func (ExecutionNoOpCache) Set(ctx context.Context, key ExecutionCacheKey, result executionResult) error {
	return nil
}

func (ExecutionNoOpCache) Clear(ctx context.Context, key ExecutionCacheKey) error {
	return nil
}
