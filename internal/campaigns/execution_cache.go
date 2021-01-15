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
	Get(ctx context.Context, key ExecutionCacheKey) (result ExecutionResult, found bool, err error)
	Set(ctx context.Context, key ExecutionCacheKey, result ExecutionResult) error
	Clear(ctx context.Context, key ExecutionCacheKey) error
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

func (c ExecutionDiskCache) Get(ctx context.Context, key ExecutionCacheKey) (ExecutionResult, bool, error) {
	var result ExecutionResult

	path, err := c.cacheFilePath(key)
	if err != nil {
		return result, false, err
	}

	// We try to be backwards compatible and see if we also find older cache
	// files.
	//
	// There are three different cache versions out in the wild and to be
	// backwards compatible we read all of them.
	//
	// In Sourcegraph/src-cli 3.26 we can remove the code here and simply read
	// the cache from `path`, since all the old cache files should be deleted
	// until then.
	globPattern := strings.TrimSuffix(path, cacheFileExt) + ".*"
	matches, err := filepath.Glob(globPattern)
	if err != nil {
		return result, false, err
	}

	switch len(matches) {
	case 0:
		// Nothing found
		return result, false, nil
	case 1:
		// One cache file found
		if err := c.readCacheFile(matches[0], &result); err != nil {
			return result, false, err
		}

		// If it's an old cache file, we rewrite the cache and delete the old file
		if isOldCacheFile(matches[0]) {
			if err := c.Set(ctx, key, result); err != nil {
				return result, false, errors.Wrap(err, "failed to rewrite cache in new format")
			}
			if err := os.Remove(matches[0]); err != nil {
				return result, false, errors.Wrap(err, "failed to remove old cache file")
			}
		}

		return result, true, err

	default:
		// More than one cache file found.
		// Sort them so that we'll can possibly read from the one with the most
		// current version.
		sortCacheFiles(matches)

		newest := matches[0]
		toDelete := matches[1:]

		// Read from newest
		if err := c.readCacheFile(newest, &result); err != nil {
			return result, false, err
		}

		// If the newest was also an older version, we write a new version...
		if isOldCacheFile(newest) {
			if err := c.Set(ctx, key, result); err != nil {
				return result, false, errors.Wrap(err, "failed to rewrite cache in new format")
			}
			// ... and mark the file also as to-be-deleted
			toDelete = append(toDelete, newest)
		}

		// Now we clean up the old ones
		for _, path := range toDelete {
			if err := os.Remove(path); err != nil {
				return result, false, errors.Wrap(err, "failed to remove old cache file")
			}
		}

		return result, true, nil
	}
}

// sortCacheFiles sorts cache file paths by their "version", so that files
// ending in `cacheFileExt` are first.
func sortCacheFiles(paths []string) {
	sort.Slice(paths, func(i, j int) bool {
		return !isOldCacheFile(paths[i]) && isOldCacheFile(paths[j])
	})
}

func isOldCacheFile(path string) bool { return !strings.HasSuffix(path, cacheFileExt) }

func (c ExecutionDiskCache) readCacheFile(path string, result *ExecutionResult) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	switch {
	case strings.HasSuffix(path, ".v3.json"):
		// v3 of the cache: we cache the diff and the outputs produced by the step.
		if err := json.Unmarshal(data, result); err != nil {
			// Delete the invalid data to avoid causing an error for next time.
			if err := os.Remove(path); err != nil {
				return errors.Wrap(err, "while deleting cache file with invalid JSON")
			}
			return errors.Wrapf(err, "reading cache file %s", path)
		}
		return nil

	case strings.HasSuffix(path, ".diff"):
		// v2 of the cache: we only cached the diff, since that's the
		// only bit of data we were interested in.
		result.Diff = string(data)
		result.Outputs = map[string]interface{}{}
		// Conversion is lossy, though: we don't populate result.StepChanges.
		result.ChangedFiles = &StepChanges{}

		return nil

	case strings.HasSuffix(path, ".json"):
		// v1 of the cache: we cached the complete ChangesetSpec instead of just the diffs.
		var spec ChangesetSpec
		if err := json.Unmarshal(data, &spec); err != nil {
			// Delete the invalid data to avoid causing an error for next time.
			if err := os.Remove(path); err != nil {
				return errors.Wrap(err, "while deleting cache file with invalid JSON")
			}
			return errors.Wrapf(err, "reading cache file %s", path)
		}
		if len(spec.Commits) != 1 {
			return errors.New("cached result has no commits")
		}

		result.Diff = spec.Commits[0].Diff
		result.Outputs = map[string]interface{}{}
		result.ChangedFiles = &StepChanges{}

		return nil
	}

	return fmt.Errorf("unknown file format for cache file %q", path)
}

func (c ExecutionDiskCache) Set(ctx context.Context, key ExecutionCacheKey, result ExecutionResult) error {
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

func (ExecutionNoOpCache) Get(ctx context.Context, key ExecutionCacheKey) (result ExecutionResult, found bool, err error) {
	return ExecutionResult{}, false, nil
}

func (ExecutionNoOpCache) Set(ctx context.Context, key ExecutionCacheKey, result ExecutionResult) error {
	return nil
}

func (ExecutionNoOpCache) Clear(ctx context.Context, key ExecutionCacheKey) error {
	return nil
}
