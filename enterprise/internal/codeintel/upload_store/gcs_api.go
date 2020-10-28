package uploadstore

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
)

type gcsAPI interface {
	Bucket(name string) gcsBucketHandle
}

type gcsBucketHandle interface {
	Attrs(ctx context.Context) (*storage.BucketAttrs, error)
	Create(ctx context.Context, projectID string, attrs *storage.BucketAttrs) error
	Update(ctx context.Context, attrs storage.BucketAttrsToUpdate) error
	Object(name string) gcsObjectHandle
}

type gcsObjectHandle interface {
	Delete(ctx context.Context) error
	NewRangeReader(ctx context.Context, offset, length int64) (io.ReadCloser, error)
	NewWriter(ctx context.Context) io.WriteCloser
	ComposerFrom(sources ...gcsObjectHandle) gcsComposer
}

type gcsComposer interface {
	Run(ctx context.Context) (*storage.ObjectAttrs, error)
}

type gcsAPIShim struct{ *storage.Client }
type bucketHandleShim struct{ *storage.BucketHandle }
type objectHandleShim struct{ *storage.ObjectHandle }
type composerShim struct{ *storage.Composer }

var _ gcsAPI = &gcsAPIShim{}
var _ gcsBucketHandle = &bucketHandleShim{}
var _ gcsObjectHandle = &objectHandleShim{}
var _ gcsComposer = &composerShim{}

func (s *gcsAPIShim) Bucket(name string) gcsBucketHandle {
	return &bucketHandleShim{s.Client.Bucket(name)}
}

func (s *bucketHandleShim) Attrs(ctx context.Context) (*storage.BucketAttrs, error) {
	return s.BucketHandle.Attrs(ctx)
}

func (s *bucketHandleShim) Create(ctx context.Context, projectID string, attrs *storage.BucketAttrs) error {
	return s.BucketHandle.Create(ctx, projectID, attrs)
}

func (s *bucketHandleShim) Update(ctx context.Context, attrs storage.BucketAttrsToUpdate) error {
	_, err := s.BucketHandle.Update(ctx, attrs)
	return err
}

func (s *bucketHandleShim) Object(name string) gcsObjectHandle {
	return &objectHandleShim{s.BucketHandle.Object(name)}
}

func (s *objectHandleShim) Delete(ctx context.Context) error {
	return s.ObjectHandle.Delete(ctx)
}

func (s *objectHandleShim) NewRangeReader(ctx context.Context, offset, length int64) (io.ReadCloser, error) {
	return s.ObjectHandle.NewRangeReader(ctx, offset, length)
}

func (s *objectHandleShim) NewWriter(ctx context.Context) io.WriteCloser {
	return s.ObjectHandle.NewWriter(ctx)
}

func (s *objectHandleShim) ComposerFrom(sources ...gcsObjectHandle) gcsComposer {
	var rawSources []*storage.ObjectHandle
	for _, source := range sources {
		if shim, ok := source.(*objectHandleShim); ok {
			rawSources = append(rawSources, shim.ObjectHandle)
		}
	}

	return &composerShim{s.ObjectHandle.ComposerFrom(rawSources...)}
}

func (s *composerShim) Run(ctx context.Context) (*storage.ObjectAttrs, error) {
	return s.Composer.Run(ctx)
}
