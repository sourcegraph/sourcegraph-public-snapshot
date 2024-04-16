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
	OnlyFetchWorkspace    bool
	Steps                 []batcheslib.Step
	BatchChangeAttributes *template.BatchChangeAttributes
	// CachedStepResultFound is true when a partial execution result was found in the cache.
	// When this field is true, CachedStepResult is also populated.
	CachedStepResultFound bool
	CachedStepResult      execution.AfterStepResult
}

func (t *Task) ArchivePathToFetch() string {
	if t.OnlyFetchWorkspace {
		return t.Path
	}
	return ""
}

func (t *Task) CacheKey(globalEnv []string, workingDir string, stepIndex int) cache.Keyer {
	return &cache.CacheKey{
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
		MetadataRetriever:     fileMetadataRetriever{workingDirectory: workingDir},

		GlobalEnv: globalEnv,

		StepIndex: stepIndex,
	}
}

type fileMetadataRetriever struct {
	workingDirectory string
}

func (f fileMetadataRetriever) Get(steps []batcheslib.Step) ([]cache.MountMetadata, error) {
	var mountsMetadata []cache.MountMetadata
	for _, step := range steps {
		// Build up the metadata for each mount for each step
		for _, mount := range step.Mount {
			metadata, err := f.getMountMetadata(f.workingDirectory, mount.Path)
			if err != nil {
				return nil, err
			}
			// A mount could be a directory containing multiple files
			mountsMetadata = append(mountsMetadata, metadata...)
		}
	}
	return mountsMetadata, nil
}

func (f fileMetadataRetriever) getMountMetadata(baseDir string, path string) ([]cache.MountMetadata, error) {
	fullPath := path
	if !filepath.IsAbs(path) {
		fullPath = filepath.Join(baseDir, path)
	}
	info, err := os.Stat(fullPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, errors.Newf("path %s does not exist", path)
	} else if err != nil {
		return nil, err
	}
	var metadata []cache.MountMetadata
	if info.IsDir() {
		dirMetadata, err := f.getDirectoryMountMetadata(fullPath)
		if err != nil {
			return nil, err
		}
		metadata = append(metadata, dirMetadata...)
	} else {
		relativePath, err := filepath.Rel(f.workingDirectory, fullPath)
		if err != nil {
			return nil, err
		}
		metadata = append(metadata, cache.MountMetadata{Path: relativePath, Size: info.Size(), Modified: info.ModTime().UTC()})
	}
	return metadata, nil
}

// getDirectoryMountMetadata reads all the files in the directory with the given
// path and returns the cache.MountMetadata for all of them.
func (f fileMetadataRetriever) getDirectoryMountMetadata(path string) ([]cache.MountMetadata, error) {
	dir, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var metadata []cache.MountMetadata
	for _, dirEntry := range dir {
		// Go back to the very start. Need to get the FileInfo again for the new path and figure out if it is a
		// directory or a file.
		fileMetadata, err := f.getMountMetadata(path, dirEntry.Name())
		if err != nil {
			return nil, err
		}
		metadata = append(metadata, fileMetadata...)
	}
	return metadata, nil
}
