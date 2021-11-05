package cache

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
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
}

// Key converts the key into a string form that can be used to uniquely identify
// the cache key in a more concise form than the entire Task.
func (key *ExecutionKey) Key() (string, error) {
	envs, err := resolveStepsEnvironment(key.Steps)
	if err != nil {
		return "", err
	}

	return marshalHash(key, envs)
}

func (key ExecutionKey) Slug() string {
	return SlugForRepo(key.Repository.Name, key.Repository.BaseRev)
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
		// TODO: This should also render templates inside env vars.
		env, err := step.Env.Resolve(global)
		if err != nil {
			return nil, errors.Wrapf(err, "resolving environment for step %d", i)
		}
		envs[i] = env
	}
	return envs, nil
}

func marshalHash(t *ExecutionKey, envs []map[string]string) (string, error) {
	raw, err := json.Marshal(struct {
		*ExecutionKey
		Environments []map[string]string
	}{
		ExecutionKey: t,
		Environments: envs,
	})
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(raw)
	return base64.RawURLEncoding.EncodeToString(hash[:16]), nil
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
	envs, err := resolveStepsEnvironment(clone.Steps)
	if err != nil {
		return "", err
	}

	hash, err := marshalHash(clone, envs)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-step-%d", hash, key.StepIndex), err
}

func (key StepsCacheKey) Slug() string {
	return SlugForRepo(key.Repository.Name, key.Repository.BaseRev)
}
