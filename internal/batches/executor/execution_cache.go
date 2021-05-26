package executor

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/sourcegraph/src-cli/internal/batches"
)

func UserCacheDir() (string, error) {
	userCacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userCacheDir, "sourcegraph-src"), nil
}

type CacheKeyer interface {
	Key() (string, error)
	Slug() string
}

func resolveStepsEnvironment(steps []batches.Step) ([]map[string]string, error) {
	// We have to resolve the step environments and include them in the cache
	// key to ensure that the cache is properly invalidated when an environment
	// variable changes.
	//
	// Note that we don't base the cache key on the entire global environment:
	// if an unrelated environment variable changes, that's fine. We're only
	// interested in the ones that actually make it into the step container.
	global := os.Environ()
	envs := make([]map[string]string, len(steps))
	for i, step := range steps {
		env, err := step.Env.Resolve(global)
		if err != nil {
			return nil, errors.Wrapf(err, "resolving environment for step %d", i)
		}
		envs[i] = env
	}
	return envs, nil
}

func marshalHash(t *Task, envs []map[string]string) (string, error) {
	raw, err := json.Marshal(struct {
		*Task
		Environments []map[string]string
	}{
		Task:         t,
		Environments: envs,
	})
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(raw)
	return base64.RawURLEncoding.EncodeToString(hash[:16]), nil
}

// StepsCacheKey implements the CacheKeyer interface for a Task and a *subset*
// of its Steps, up to and including the step with index StepIndex in
// Task.Steps.
type StepsCacheKey struct {
	*Task
	StepIndex int
}

// Key converts the key into a string form that can be used to uniquely identify
// the cache key in a more concise form than the entire Task.
func (key StepsCacheKey) Key() (string, error) {
	// Setup a copy of the Task that only includes the Steps up to and
	// including key.StepIndex.
	taskCopy := &Task{
		Repository:            key.Task.Repository,
		Path:                  key.Task.Path,
		OnlyFetchWorkspace:    key.Task.OnlyFetchWorkspace,
		BatchChangeAttributes: key.Task.BatchChangeAttributes,
		Template:              key.Task.Template,
		TransformChanges:      key.Task.TransformChanges,
		Archive:               key.Task.Archive,
	}

	taskCopy.Steps = key.Task.Steps[0 : key.StepIndex+1]

	// Resolve environment only for the subset of Steps
	envs, err := resolveStepsEnvironment(taskCopy.Steps)
	if err != nil {
		return "", err
	}

	hash, err := marshalHash(taskCopy, envs)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-step-%d", hash, key.StepIndex), err
}

func (key StepsCacheKey) Slug() string { return key.Repository.Slug() }

// TaskCacheKey implements the CacheKeyer interface for a Task and all its
// Steps.
type TaskCacheKey struct {
	*Task
}

// Key converts the key into a string form that can be used to uniquely identify
// the cache key in a more concise form than the entire Task.
func (key TaskCacheKey) Key() (string, error) {
	envs, err := resolveStepsEnvironment(key.Task.Steps)
	if err != nil {
		return "", err
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

func (key TaskCacheKey) Slug() string { return key.Repository.Slug() }

type ExecutionCache interface {
	Get(ctx context.Context, key CacheKeyer) (result executionResult, found bool, err error)
	Set(ctx context.Context, key CacheKeyer, result executionResult) error

	GetStepResult(ctx context.Context, key CacheKeyer) (result stepExecutionResult, found bool, err error)
	SetStepResult(ctx context.Context, key CacheKeyer, result stepExecutionResult) error

	Clear(ctx context.Context, key CacheKeyer) error
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

const cacheFileExt = ".json"

func (c ExecutionDiskCache) cacheFilePath(key CacheKeyer) (string, error) {
	keyString, err := key.Key()
	if err != nil {
		return "", errors.Wrap(err, "calculating execution cache key")
	}

	return filepath.Join(c.Dir, key.Slug(), keyString+cacheFileExt), nil
}

func (c ExecutionDiskCache) Get(ctx context.Context, key CacheKeyer) (executionResult, bool, error) {
	var result executionResult

	path, err := c.cacheFilePath(key)
	if err != nil {
		return result, false, err
	}

	found, err := c.readCacheFile(path, &result)

	return result, found, err
}

func (c ExecutionDiskCache) readCacheFile(path string, result interface{}) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	}

	data, err := ioutil.ReadFile(path)
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

	return ioutil.WriteFile(path, raw, 0600)
}

func (c ExecutionDiskCache) Set(ctx context.Context, key CacheKeyer, result executionResult) error {
	path, err := c.cacheFilePath(key)
	if err != nil {
		return err
	}

	return c.writeCacheFile(path, &result)
}

func (c ExecutionDiskCache) Clear(ctx context.Context, key CacheKeyer) error {
	path, err := c.cacheFilePath(key)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(path)
}

func (c ExecutionDiskCache) GetStepResult(ctx context.Context, key CacheKeyer) (stepExecutionResult, bool, error) {
	var result stepExecutionResult
	path, err := c.cacheFilePath(key)
	if err != nil {
		return result, false, err
	}

	found, err := c.readCacheFile(path, &result)

	return result, found, nil
}

func (c ExecutionDiskCache) SetStepResult(ctx context.Context, key CacheKeyer, result stepExecutionResult) error {
	path, err := c.cacheFilePath(key)
	if err != nil {
		return err
	}

	return c.writeCacheFile(path, &result)
}

// ExecutionNoOpCache is an implementation of ExecutionCache that does not store or
// retrieve cache entries.
type ExecutionNoOpCache struct{}

func (ExecutionNoOpCache) Get(ctx context.Context, key CacheKeyer) (result executionResult, found bool, err error) {
	return executionResult{}, false, nil
}

func (ExecutionNoOpCache) Set(ctx context.Context, key CacheKeyer, result executionResult) error {
	return nil
}

func (ExecutionNoOpCache) Clear(ctx context.Context, key CacheKeyer) error {
	return nil
}

func (ExecutionNoOpCache) SetStepResult(ctx context.Context, key CacheKeyer, result stepExecutionResult) error {
	return nil
}

func (ExecutionNoOpCache) GetStepResult(ctx context.Context, key CacheKeyer) (stepExecutionResult, bool, error) {
	return stepExecutionResult{}, false, nil
}
