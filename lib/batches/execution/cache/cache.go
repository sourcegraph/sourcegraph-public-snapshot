package cache

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Cache interface {
	Get(ctx context.Context, key Keyer) (result execution.Result, found bool, err error)
	Set(ctx context.Context, key Keyer, result execution.Result) error

	GetStepResult(ctx context.Context, key Keyer) (result execution.AfterStepResult, found bool, err error)
	SetStepResult(ctx context.Context, key Keyer, result execution.AfterStepResult) error

	Clear(ctx context.Context, key Keyer) error
}

type Keyer interface {
	Key() (string, error)
	Slug() string
}

// ExecutionKey implements the Keyer interface for the execution of a batch
// spec in a repository workspace and all its Steps.
type ExecutionKey struct {
	Repository batches.Repository

	Path               string
	OnlyFetchWorkspace bool
	Steps              []batches.Step

	BatchChangeAttributes *template.BatchChangeAttributes

	// Ignore from serialization.
	MetadataRetriever MetadataRetriever `json:"-"`
}

// MetadataRetriever retrieves mount metadata.
type MetadataRetriever interface {
	// Get returns the mount metadata from the provided steps.
	Get([]batches.Step) ([]MountMetadata, error)
}

// Key converts the key into a string form that can be used to uniquely identify
// the cache key in a more concise form than the entire Task.
func (key *ExecutionKey) Key() (string, error) {
	envs, err := resolveStepsEnvironment([]string{}, key.Steps)
	if err != nil {
		return "", err
	}
	metadata, err := key.mountsMetadata()
	if err != nil {
		return "", err
	}

	return marshalAndHash(key, envs, metadata)
}

func (key ExecutionKey) mountsMetadata() ([]MountMetadata, error) {
	if key.MetadataRetriever != nil {
		return key.MetadataRetriever.Get(key.Steps)
	}
	return nil, nil
}

func (key ExecutionKey) Slug() string {
	return SlugForRepo(key.Repository.Name, key.Repository.BaseRev)
}

func (key *ExecutionKey) WithGlobalEnv(global []string) *ExecutionKeyWithGlobalEnv {
	return &ExecutionKeyWithGlobalEnv{
		ExecutionKey: key,
		GlobalEnv:    global,
	}
}

// ExecutionKeyWithGlobalEnv implements the Keyer interface by embedding
// ExecutionKey but adding a global environment in which the steps could be
// resolved.
type ExecutionKeyWithGlobalEnv struct {
	*ExecutionKey
	GlobalEnv []string
}

func (key *ExecutionKeyWithGlobalEnv) Key() (string, error) {
	envs, err := resolveStepsEnvironment(key.GlobalEnv, key.Steps)
	if err != nil {
		return "", err
	}
	metadata, err := key.mountsMetadata()
	if err != nil {
		return "", err
	}

	return marshalAndHash(key.ExecutionKey, envs, metadata)
}

func resolveStepsEnvironment(globalEnv []string, steps []batches.Step) ([]map[string]string, error) {
	// We have to resolve the step environments and include them in the cache
	// key to ensure that the cache is properly invalidated when an environment
	// variable changes.
	//
	// Note that we don't base the cache key on the entire global environment:
	// if an unrelated environment variable changes, that's fine. We're only
	// interested in the ones that actually make it into the step container.
	envs := make([]map[string]string, len(steps))
	for i, step := range steps {
		// TODO: This should also render templates inside env vars.
		env, err := step.Env.Resolve(globalEnv)
		if err != nil {
			return nil, errors.Wrapf(err, "resolving environment for step %d", i)
		}
		envs[i] = env
	}
	return envs, nil
}

func marshalAndHash(key *ExecutionKey, envs []map[string]string, metadata []MountMetadata) (string, error) {
	raw, err := json.Marshal(struct {
		*ExecutionKey
		Environments []map[string]string
		// Omit if empty to be backwards compatible
		MountsMetadata []MountMetadata `json:"MountsMetadata,omitempty"`
	}{
		ExecutionKey:   key,
		Environments:   envs,
		MountsMetadata: metadata,
	})
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(raw)
	return base64.RawURLEncoding.EncodeToString(hash[:16]), nil
}

// MountMetadata is the metadata of a file that is mounted by a Step.
type MountMetadata struct {
	Path     string
	Size     int64
	Modified time.Time
}

// FileMetadataRetriever retrieves mount metadata from a filesystem.
type FileMetadataRetriever struct {
}

func (f FileMetadataRetriever) Get(steps []batches.Step) ([]MountMetadata, error) {
	var mountsMetadata []MountMetadata
	for _, step := range steps {
		// Build up the metadata for each mount for each step
		for _, mount := range step.Mount {
			metadata, err := getMountMetadata(mount.Path)
			if err != nil {
				return nil, err
			}
			// A mount could be a directory containing multiple files
			mountsMetadata = append(mountsMetadata, metadata...)
		}
	}
	return mountsMetadata, nil
}

func getMountMetadata(path string) ([]MountMetadata, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errors.Newf("path %s does not exist", path)
	} else if err != nil {
		return nil, err
	}
	var metadata []MountMetadata
	if info.IsDir() {
		dirMetadata, err := getDirectoryMountMetadata(path)
		if err != nil {
			return nil, err
		}
		metadata = append(metadata, dirMetadata...)
	} else {
		metadata = append(metadata, MountMetadata{Path: path, Size: info.Size(), Modified: info.ModTime().UTC()})
	}
	return metadata, nil
}

func getDirectoryMountMetadata(path string) ([]MountMetadata, error) {
	dir, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var metadata []MountMetadata
	for _, dirEntry := range dir {
		newPath := filepath.Join(path, dirEntry.Name())
		// Go back to the very start. Need to get the FileInfo again for the new path and figure out if it is a
		// directory or a file.
		fileMetadata, err := getMountMetadata(newPath)
		if err != nil {
			return nil, err
		}
		metadata = append(metadata, fileMetadata...)
	}
	return metadata, nil
}

// StepsCacheKey implements the Keyer interface for a batch spec execution in a
// repository workspace and a *subset* of its Steps, up to and including the
// step with index StepIndex in Task.Steps.
type StepsCacheKey struct {
	*ExecutionKey
	StepIndex int
}

// Key converts the key into a string form that can be used to uniquely identify
// the cache key in a more concise form than the entire Task.
func (key StepsCacheKey) Key() (string, error) {
	return marshalAndHashStepsCacheKey(key, []string{})
}

func (key StepsCacheKey) Slug() string {
	return SlugForRepo(key.Repository.Name, key.Repository.BaseRev)
}

// StepsCacheKeyWithGlobalEnv implements the Keyer interface by embedding
// StepsCacheKey but adding a global environment in which the steps could be
// resolved.
type StepsCacheKeyWithGlobalEnv struct {
	*StepsCacheKey
	GlobalEnv []string
}

func (key *StepsCacheKeyWithGlobalEnv) Key() (string, error) {
	return marshalAndHashStepsCacheKey(*key.StepsCacheKey, key.GlobalEnv)
}

func marshalAndHashStepsCacheKey(key StepsCacheKey, globalEnv []string) (string, error) {
	// Setup a copy of the Task that only includes the Steps up to and
	// including key.StepIndex.
	clone := &ExecutionKey{
		Repository:            key.ExecutionKey.Repository,
		Path:                  key.ExecutionKey.Path,
		OnlyFetchWorkspace:    key.ExecutionKey.OnlyFetchWorkspace,
		Steps:                 key.ExecutionKey.Steps[0 : key.StepIndex+1],
		BatchChangeAttributes: key.ExecutionKey.BatchChangeAttributes,
	}

	// Resolve environment only for the subset of Steps
	envs, err := resolveStepsEnvironment(globalEnv, clone.Steps)
	if err != nil {
		return "", err
	}
	metadata, err := key.mountsMetadata()
	if err != nil {
		return "", err
	}

	hash, err := marshalAndHash(clone, envs, metadata)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-step-%d", hash, key.StepIndex), err
}

func KeyForWorkspace(batchChangeAttributes *template.BatchChangeAttributes, r batches.Repository, path string, onlyFetchWorkspace bool, steps []batches.Step) ExecutionKey {
	sort.Strings(r.FileMatches)

	executionKey := ExecutionKey{
		Repository:            r,
		Path:                  path,
		OnlyFetchWorkspace:    onlyFetchWorkspace,
		Steps:                 steps,
		BatchChangeAttributes: batchChangeAttributes,
	}
	return executionKey
}

func ChangesetSpecsFromCache(spec *batches.BatchSpec, r batches.Repository, result execution.Result) ([]*batches.ChangesetSpec, error) {
	if result.Diff == "" {
		return []*batches.ChangesetSpec{}, nil
	}

	sort.Strings(r.FileMatches)

	input := &batches.ChangesetSpecInput{
		Repository: r,
		BatchChangeAttributes: &template.BatchChangeAttributes{
			Name:        spec.Name,
			Description: spec.Description,
		},
		Template:         spec.ChangesetTemplate,
		TransformChanges: spec.TransformChanges,
		Result:           result,
	}

	return batches.BuildChangesetSpecs(input, batches.ChangesetSpecFeatureFlags{
		IncludeAutoAuthorDetails: true,
		AllowOptionalPublished:   true,
	})
}
