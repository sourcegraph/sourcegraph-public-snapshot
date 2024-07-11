// Package core sits at the bottom-most layer of internal/codeintel.
//
// Do not import other packages from internal/codeintel here. If we
// need to use certain types, then those types should be moved to this
// package, or we can use interfaces (e.g. see UploadLike below)
package core

import (
	"strings"
) // See package-level comment before adding imports.

// UploadRelPath is the path fragment that is relative
// to the upload root (which may be non-empty). In the database table
// codeintel_scip_document_lookup, the paths stored are UploadRelPath
// values.
type UploadRelPath struct {
	rawValue string
}

func NewUploadRelPathUnchecked(s string) UploadRelPath {
	return UploadRelPath{rawValue: s}
}

func (p UploadRelPath) Equal(other UploadRelPath) bool {
	return p.rawValue == other.rawValue
}

func (p UploadRelPath) RawValue() string {
	return p.rawValue
}

// RepoRelPath is the exact path to a document/location relative
// to the root of the repository, which is equivalent to
// Join(upload.Root, UploadRelPath).
type RepoRelPath struct {
	rawValue string
}

func NewRepoRelPathUnchecked(s string) RepoRelPath {
	return RepoRelPath{rawValue: s}
}

// NewRepoRelPath takes an UploadLike as the first argument instead of a
// *shared.CompletedUpload to avoid a dependency on a higher-level package.
func NewRepoRelPath(uploadLike UploadLike, p UploadRelPath) RepoRelPath {
	// TODO: We should use filepath.Join here but that breaks some tests
	return RepoRelPath{rawValue: uploadLike.GetRoot() + p.rawValue}
}

func (p RepoRelPath) RawValue() string {
	return p.rawValue
}

// NewUploadRelPath takes an UploadLike as the first argument instead of a
// *shared.CompletedUpload to avoid a dependency on a higher-level package.
func NewUploadRelPath(uploadLike UploadLike, p RepoRelPath) UploadRelPath {
	// TODO: Introduce a panic here when u.Root is not a prefix of p.rawValue
	// It seems like filepath.Rel can do the error-checking for us, but currently
	// just using strings.TrimPrefix for compatibility with old behavior.
	return NewUploadRelPathUnchecked(strings.TrimPrefix(p.rawValue, uploadLike.GetRoot()))
}

func (p RepoRelPath) Equal(other RepoRelPath) bool {
	return p.rawValue == other.rawValue
}
