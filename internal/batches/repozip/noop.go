package repozip

import "context"

// NewNoopRegistry returns a registry that returns archives that don't perform
// any operation. This is used in executor mode to disable fetching repo archives
// in favor of cloning.
func NewNoopRegistry() ArchiveRegistry {
	return &noopArchiveRegistry{}
}

type noopArchiveRegistry struct{}

func (r *noopArchiveRegistry) Checkout(repo RepoRevision, path string) Archive {
	return &noopArchive{}
}

type noopArchive struct{}

func (a *noopArchive) Ensure(context.Context) error {
	return nil
}
func (a *noopArchive) Close() error {
	return nil
}
func (a *noopArchive) Path() string {
	return ""
}
func (a *noopArchive) AdditionalFilePaths() map[string]string {
	return nil
}
