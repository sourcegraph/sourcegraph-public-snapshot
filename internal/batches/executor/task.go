package executor

import (
	"os"
	"path/filepath"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution/cache"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/repozip"
)

type Task struct {
	Repository *graphql.Repository

	// Path is the folder relative to the repository's root in which the steps
	// should be executed. "" means root.
	Path string
	// OnlyFetchWorkspace determines whether the repository archive contains
	// the complete repository or just the files in Path (and additional files,
	// see RepoFetcher).
	// If Path is "" then this setting has no effect.
	OnlyFetchWorkspace bool

	Steps []batcheslib.Step

	// TODO(mrnugget): this should just be a single BatchSpec field instead, if
	// we can make it work with caching
	BatchChangeAttributes *template.BatchChangeAttributes `json:"-"`

	Archive repozip.Archive `json:"-"`

	CachedResultFound bool                      `json:"-"`
	CachedResult      execution.AfterStepResult `json:"-"`
}

func (t *Task) ArchivePathToFetch() string {
	if t.OnlyFetchWorkspace {
		return t.Path
	}
	return ""
}

func (t *Task) cacheKey(globalEnv []string, isRemote bool) *cache.ExecutionKeyWithGlobalEnv {
	var metadataRetriever cache.MetadataRetriever
	// If the task is being run locally, set the metadata retrieve to use the filesystem based implementation.
	if !isRemote {
		metadataRetriever = fileMetadataRetriever{}
	}
	return &cache.ExecutionKeyWithGlobalEnv{
		GlobalEnv: globalEnv,
		ExecutionKey: &cache.ExecutionKey{
			Repository: batcheslib.Repository{
				ID:          t.Repository.ID,
				Name:        t.Repository.Name,
				BaseRef:     t.Repository.BaseRef(),
				BaseRev:     t.Repository.Rev(),
				FileMatches: t.Repository.SortedFileMatches(),
			},
			Path:                  t.Path,
			OnlyFetchWorkspace:    t.OnlyFetchWorkspace,
			Steps:                 t.Steps,
			BatchChangeAttributes: t.BatchChangeAttributes,
			MetadataRetriever:     metadataRetriever,
		},
	}
}

type fileMetadataRetriever struct {
}

func (f fileMetadataRetriever) Get(steps []batcheslib.Step) ([]cache.MountMetadata, error) {
	var mountsMetadata []cache.MountMetadata
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

func getMountMetadata(path string) ([]cache.MountMetadata, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errors.Newf("path %s does not exist", path)
	} else if err != nil {
		return nil, err
	}
	var metadata []cache.MountMetadata
	if info.IsDir() {
		dirMetadata, err := getDirectoryMountMetadata(path)
		if err != nil {
			return nil, err
		}
		metadata = append(metadata, dirMetadata...)
	} else {
		metadata = append(metadata, cache.MountMetadata{Path: path, Size: info.Size(), Modified: info.ModTime().UTC()})
	}
	return metadata, nil
}

func getDirectoryMountMetadata(path string) ([]cache.MountMetadata, error) {
	dir, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var metadata []cache.MountMetadata
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

func cacheKeyForStep(key *cache.ExecutionKeyWithGlobalEnv, stepIndex int) *cache.StepsCacheKeyWithGlobalEnv {
	return &cache.StepsCacheKeyWithGlobalEnv{
		StepsCacheKey: &cache.StepsCacheKey{
			ExecutionKey: key.ExecutionKey,
			StepIndex:    stepIndex,
		},
		GlobalEnv: key.GlobalEnv,
	}
}
