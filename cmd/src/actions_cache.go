package main

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

type actionExecutionCacheKey struct {
	Repo ActionRepo
	Runs []*ActionStep
}

type actionExecutionCache interface {
	get(ctx context.Context, key actionExecutionCacheKey) (result PatchInput, ok bool, err error)
	set(ctx context.Context, key actionExecutionCacheKey, result PatchInput) error
}

type actionExecutionDiskCache struct {
	dir string
}

func (c actionExecutionDiskCache) cacheFilePath(key actionExecutionCacheKey) (string, error) {
	keyJSON, err := json.Marshal(key)
	if err != nil {
		return "", errors.Wrap(err, "Failed to marshal JSON when generating action cache key")
	}

	b := sha256.Sum256(keyJSON)
	keyString := base64.RawURLEncoding.EncodeToString(b[:16])

	return filepath.Join(c.dir, keyString+".json"), nil
}

func (c actionExecutionDiskCache) get(ctx context.Context, key actionExecutionCacheKey) (PatchInput, bool, error) {
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

func (c actionExecutionDiskCache) set(ctx context.Context, key actionExecutionCacheKey, result PatchInput) error {
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

// actionExecutionNoOpCache is an implementation of actionExecutionCache that does not store or
// retrieve cache entries.
type actionExecutionNoOpCache struct{}

func (actionExecutionNoOpCache) get(ctx context.Context, key actionExecutionCacheKey) (result PatchInput, ok bool, err error) {
	return PatchInput{}, false, nil
}

func (actionExecutionNoOpCache) set(ctx context.Context, key actionExecutionCacheKey, result PatchInput) error {
	return nil
}
