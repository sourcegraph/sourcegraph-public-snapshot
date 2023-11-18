package uploadstore

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/iterator"
)

type lazyStore struct {
	store       Store
	m           sync.RWMutex
	initialized bool
}

var _ Store = &gcsStore{}

func newLazyStore(store Store) Store {
	return &lazyStore{store: store}
}

func (s *lazyStore) Init(ctx context.Context) error {
	return s.initOnce(ctx)
}

func (s *lazyStore) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := s.initOnce(ctx); err != nil {
		return nil, err
	}

	return s.store.Get(ctx, key)
}

func (s *lazyStore) List(ctx context.Context, prefix string) (*iterator.Iterator[string], error) {
	if err := s.initOnce(ctx); err != nil {
		return nil, err
	}

	return s.store.List(ctx, prefix)
}

func (s *lazyStore) Upload(ctx context.Context, key string, r io.Reader) (int64, error) {
	if err := s.initOnce(ctx); err != nil {
		return 0, err
	}

	return s.store.Upload(ctx, key, r)
}

func (s *lazyStore) Compose(ctx context.Context, destination string, sources ...string) (int64, error) {
	if err := s.initOnce(ctx); err != nil {
		return 0, err
	}

	return s.store.Compose(ctx, destination, sources...)
}

func (s *lazyStore) Delete(ctx context.Context, key string) error {
	if err := s.initOnce(ctx); err != nil {
		return err
	}

	return s.store.Delete(ctx, key)
}

func (s *lazyStore) ExpireObjects(ctx context.Context, prefix string, maxAge time.Duration) error {
	if err := s.initOnce(ctx); err != nil {
		return err
	}

	return s.store.ExpireObjects(ctx, prefix, maxAge)
}

// initOnce serializes access to the underlying store's Init method. If the
// Init method completes successfully, all future calls to this function will
// no-op.
func (s *lazyStore) initOnce(ctx context.Context) error {
	s.m.RLock()
	initialized := s.initialized
	s.m.RUnlock()
	if initialized {
		return nil
	}

	s.m.Lock()
	defer s.m.Unlock()
	if s.initialized {
		return nil
	}

	if err := s.store.Init(ctx); err != nil {
		return err
	}

	s.initialized = true
	return nil
}
