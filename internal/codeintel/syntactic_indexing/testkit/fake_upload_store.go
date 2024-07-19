package testkit

import (
	"bytes"
	"context"
	"io"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/object"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/iterator"
)

type fakeUploadStore struct {
	files map[string][]byte
}

var _ object.Storage = &fakeUploadStore{}

func NewFakeUploadStore() object.Storage {
	return &fakeUploadStore{files: map[string][]byte{}}
}

func (s *fakeUploadStore) Init(ctx context.Context) error {
	return nil
}

func (s *fakeUploadStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	file, ok := s.files[key]
	if !ok {
		return nil, errors.Newf("file %s not found", key)
	}
	return io.NopCloser(bytes.NewReader(file)), nil
}

func (s *fakeUploadStore) List(ctx context.Context, prefix string) (*iterator.Iterator[string], error) {
	var names []string
	for k := range s.files {
		if strings.HasPrefix(k, prefix) {
			names = append(names, k)
		}
	}

	return iterator.From(names), nil
}

func (s *fakeUploadStore) Upload(ctx context.Context, key string, r io.Reader) (int64, error) {
	file, err := io.ReadAll(r)
	if err != nil {
		return -1, errors.Newf("error reading file %s", key)
	}
	s.files[key] = file
	return int64(len(file)), nil
}

func (s *fakeUploadStore) Compose(ctx context.Context, destination string, sources ...string) (int64, error) {
	return 0, nil
}

func (s *fakeUploadStore) Delete(ctx context.Context, key string) error {
	return nil
}

func (s *fakeUploadStore) ExpireObjects(ctx context.Context, prefix string, maxAge time.Duration) error {
	return nil
}
