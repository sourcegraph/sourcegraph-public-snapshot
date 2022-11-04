package cache

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Cache interface {
	Get(ctx context.Context, key Keyer) (result execution.AfterStepResult, found bool, err error)
	Set(ctx context.Context, key Keyer, result execution.AfterStepResult) error

	Clear(ctx context.Context, key Keyer) error
}

type Keyer interface {
	Key() (string, error)
	Slug() string
}

// MetadataRetriever retrieves mount metadata.
type MetadataRetriever interface {
	// Get returns the mount metadata from the provided steps.
	Get([]batches.Step) ([]MountMetadata, error)
}

// MountMetadata is the metadata of a file that is mounted by a Step.
type MountMetadata struct {
	Path     string
	Size     int64
	Modified time.Time
}

func (key CacheKey) mountsMetadata() ([]MountMetadata, error) {
	if key.MetadataRetriever != nil {
		return key.MetadataRetriever.Get(key.Steps)
	}
	return nil, nil
}

// resolveStepsEnvironment returns a slice of environments for each of the steps,
// containing only the env vars that are actually used.
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

func marshalAndHash(key *CacheKey, envs []map[string]string, metadata []MountMetadata) (string, error) {
	raw, err := json.Marshal(struct {
		*CacheKey
		Environments []map[string]string
		// Omit if empty to be backwards compatible.
		MountsMetadata []MountMetadata `json:"MountsMetadata,omitempty"`
	}{
		CacheKey:       key,
		Environments:   envs,
		MountsMetadata: metadata,
	})
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(raw)
	return base64.RawURLEncoding.EncodeToString(hash[:16]), nil
}

// CacheKey implements the Keyer interface for a batch spec execution in a
// repository workspace and a *subset* of its Steps, up to and including the
// step with index StepIndex in Task.Steps.
type CacheKey struct {
	Repository            batches.Repository
	Path                  string
	OnlyFetchWorkspace    bool
	Steps                 []batches.Step
	BatchChangeAttributes *template.BatchChangeAttributes

	// Ignore from serialization.
	MetadataRetriever MetadataRetriever `json:"-"`
	// Ignore from serialization.
	GlobalEnv []string `json:"-"`

	StepIndex int
}

// Key converts the key into a string form that can be used to uniquely identify
// the cache key in a more concise form than the entire Task.
func (key CacheKey) Key() (string, error) {
	// Setup a copy of the cache key that only includes the Steps up to and
	// including key.StepIndex.
	clone := key
	clone.Steps = key.Steps[0 : key.StepIndex+1]

	// Resolve environment only for the subset of Steps.
	envs, err := resolveStepsEnvironment(key.GlobalEnv, clone.Steps)
	if err != nil {
		return "", err
	}
	metadata, err := key.mountsMetadata()
	if err != nil {
		return "", err
	}

	hash, err := marshalAndHash(&clone, envs, metadata)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-step-%d", hash, key.StepIndex), err
}

func (key CacheKey) Slug() string {
	return SlugForRepo(key.Repository.Name, key.Repository.BaseRev)
}

func KeyForWorkspace(batchChangeAttributes *template.BatchChangeAttributes, r batches.Repository, path string, onlyFetchWorkspace bool, steps []batches.Step, stepIndex int) Keyer {
	sort.Strings(r.FileMatches)

	return CacheKey{
		Repository:            r,
		Path:                  path,
		OnlyFetchWorkspace:    onlyFetchWorkspace,
		Steps:                 steps,
		BatchChangeAttributes: batchChangeAttributes,
		StepIndex:             stepIndex,
	}
}

// ChangesetSpecsFromCache takes the execution.Result and generates all changeset specs from it.
func ChangesetSpecsFromCache(spec *batches.BatchSpec, r batches.Repository, result execution.AfterStepResult, path string) ([]*batches.ChangesetSpec, error) {
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
		Path:             path,
	}

	return batches.BuildChangesetSpecs(input)
}
