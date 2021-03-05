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

type gcsAPIShim struct{ client *storage.Client }
type bucketHandleShim struct{ handle *storage.BucketHandle }
type objectHandleShim struct{ handle *storage.ObjectHandle }

type composerShim struct {
	handle  *storage.ObjectHandle
	sources []*storage.ObjectHandle
}

var _ gcsAPI = &gcsAPIShim{}
var _ gcsBucketHandle = &bucketHandleShim{}
var _ gcsObjectHandle = &objectHandleShim{}
var _ gcsComposer = &composerShim{}

func (s *gcsAPIShim) Bucket(name string) gcsBucketHandle {
	return &bucketHandleShim{handle: s.client.Bucket(name)}
}

func (s *bucketHandleShim) Attrs(ctx context.Context) (*storage.BucketAttrs, error) {
	return s.handle.Attrs(ctx)
}

func (s *bucketHandleShim) Create(ctx context.Context, projectID string, attrs *storage.BucketAttrs) error {
	return s.handle.Create(ctx, projectID, attrs)
}

func (s *bucketHandleShim) Update(ctx context.Context, attrs storage.BucketAttrsToUpdate) error {
	_, err := s.handle.Update(ctx, attrs)
	return err
}

func (s *bucketHandleShim) Object(name string) gcsObjectHandle {
	return &objectHandleShim{handle: s.handle.Object(name)}
}

func (s *objectHandleShim) Delete(ctx context.Context) error {
	return s.handle.Delete(ctx)
}

func (s *objectHandleShim) NewRangeReader(ctx context.Context, offset, length int64) (io.ReadCloser, error) {
	return s.handle.NewRangeReader(ctx, offset, length)
}

func (s *objectHandleShim) NewWriter(ctx context.Context) io.WriteCloser {
	return s.handle.NewWriter(ctx)
}

func (s *objectHandleShim) ComposerFrom(sources ...gcsObjectHandle) gcsComposer {
	var handles []*storage.ObjectHandle
	for _, source := range sources {
		if shim, ok := source.(*objectHandleShim); ok {
			handles = append(handles, shim.handle)
		}
	}

	return &composerShim{handle: s.handle, sources: handles}
}

func (s *composerShim) Run(ctx context.Context) (*storage.ObjectAttrs, error) {
	for len(s.sources) > 32 {
		if _, err := s.handle.ComposerFrom(s.sources[:32]...).Run(ctx); err != nil {
			return nil, err
		}

		s.sources = append([]*storage.ObjectHandle{s.handle}, s.sources[32:]...)
	}

	return s.handle.ComposerFrom(s.sources...).Run(ctx)
}
