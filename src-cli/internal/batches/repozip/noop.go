package repozip

import "context"

type NoopArchive struct{}

func (a *NoopArchive) Ensure(context.Context) error {
	return nil
}
func (a *NoopArchive) Close() error {
	return nil
}
func (a *NoopArchive) Path() string {
	return ""
}
func (a *NoopArchive) AdditionalFilePaths() map[string]string {
	return nil
}
