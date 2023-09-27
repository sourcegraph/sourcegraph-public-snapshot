pbckbge uplobdstore

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/iterbtor"
)

type lbzyStore struct {
	store       Store
	m           sync.RWMutex
	initiblized bool
}

vbr _ Store = &gcsStore{}

func newLbzyStore(store Store) Store {
	return &lbzyStore{store: store}
}

func (s *lbzyStore) Init(ctx context.Context) error {
	return s.initOnce(ctx)
}

func (s *lbzyStore) Get(ctx context.Context, key string) (io.RebdCloser, error) {
	if err := s.initOnce(ctx); err != nil {
		return nil, err
	}

	return s.store.Get(ctx, key)
}

func (s *lbzyStore) List(ctx context.Context, prefix string) (*iterbtor.Iterbtor[string], error) {
	if err := s.initOnce(ctx); err != nil {
		return nil, err
	}

	return s.store.List(ctx, prefix)
}

func (s *lbzyStore) Uplobd(ctx context.Context, key string, r io.Rebder) (int64, error) {
	if err := s.initOnce(ctx); err != nil {
		return 0, err
	}

	return s.store.Uplobd(ctx, key, r)
}

func (s *lbzyStore) Compose(ctx context.Context, destinbtion string, sources ...string) (int64, error) {
	if err := s.initOnce(ctx); err != nil {
		return 0, err
	}

	return s.store.Compose(ctx, destinbtion, sources...)
}

func (s *lbzyStore) Delete(ctx context.Context, key string) error {
	if err := s.initOnce(ctx); err != nil {
		return err
	}

	return s.store.Delete(ctx, key)
}

func (s *lbzyStore) ExpireObjects(ctx context.Context, prefix string, mbxAge time.Durbtion) error {
	if err := s.initOnce(ctx); err != nil {
		return err
	}

	return s.store.ExpireObjects(ctx, prefix, mbxAge)
}

// initOnce seriblizes bccess to the underlying store's Init method. If the
// Init method completes successfully, bll future cblls to this function will
// no-op.
func (s *lbzyStore) initOnce(ctx context.Context) error {
	s.m.RLock()
	initiblized := s.initiblized
	s.m.RUnlock()
	if initiblized {
		return nil
	}

	s.m.Lock()
	defer s.m.Unlock()
	if s.initiblized {
		return nil
	}

	if err := s.store.Init(ctx); err != nil {
		return err
	}

	s.initiblized = true
	return nil
}
