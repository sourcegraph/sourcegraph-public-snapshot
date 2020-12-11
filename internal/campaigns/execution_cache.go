package campaigns

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
	Get(ctx context.Context, key ExecutionCacheKey) (diff string, found bool, err error)
	Set(ctx context.Context, key ExecutionCacheKey, diff string) error
	Clear(ctx context.Context, key ExecutionCacheKey) error
}

type ExecutionDiskCache struct {
	Dir string
}

func (c ExecutionDiskCache) cacheFilePath(key ExecutionCacheKey) (string, error) {
	keyString, err := key.Key()
	if err != nil {
		return "", errors.Wrap(err, "calculating execution cache key")
	}

	return filepath.Join(c.Dir, keyString+".diff"), nil
}

func (c ExecutionDiskCache) Get(ctx context.Context, key ExecutionCacheKey) (string, bool, error) {
	path, err := c.cacheFilePath(key)
	if err != nil {
		return "", false, err
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil // treat as not-found
		}
		return "", false, err
	}

	// We previously cached complete ChangesetSpecs instead of just the diffs.
	// To be backwards compatible, we keep reading these:
	if strings.HasSuffix(path, ".json") {
		var result ChangesetSpec
		if err := json.Unmarshal(data, &result); err != nil {
			// Delete the invalid data to avoid causing an error for next time.
			if err := os.Remove(path); err != nil {
				return "", false, errors.Wrap(err, "while deleting cache file with invalid JSON")
			}
			return "", false, errors.Wrapf(err, "reading cache file %s", path)
		}
		if len(result.Commits) != 1 {
			return "", false, errors.New("cached result has no commits")
		}
		return result.Commits[0].Diff, true, nil
	}

	if strings.HasSuffix(path, ".diff") {
		return string(data), true, nil
	}

	return "", false, fmt.Errorf("unknown file format for cache file %q", path)
}

func (c ExecutionDiskCache) Set(ctx context.Context, key ExecutionCacheKey, diff string) error {
	path, err := c.cacheFilePath(key)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	return ioutil.WriteFile(path, []byte(diff), 0600)
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

func (ExecutionNoOpCache) Get(ctx context.Context, key ExecutionCacheKey) (diff string, found bool, err error) {
	return "", false, nil
}

func (ExecutionNoOpCache) Set(ctx context.Context, key ExecutionCacheKey, diff string) error {
	return nil
}

func (ExecutionNoOpCache) Clear(ctx context.Context, key ExecutionCacheKey) error {
	return nil
}
