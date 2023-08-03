package mockstore

import "github.com/sourcegraph/sourcegraph/internal/database/basestore"

type MockStore[T basestore.ShareableStore] struct {
	*basestore.Store
}

type MockableStore[T basestore.ShareableStore] interface {
	basestore.ShareableStore
	With(basestore.ShareableStore) T
}

func (ms *MockStore[T]) With(other basestore.ShareableStore) T {
	if s := get[T](other); s != nil {
		return *s
	}

	return ms.With(other)
}

type mockedStore struct {
	basestore.ShareableStore
	mockedShareableStore basestore.ShareableStore
}

// Get fetches the mocked interface T from the provided DB.
// If no mocked interface is found, nil is returned.
func get[T basestore.ShareableStore](s basestore.ShareableStore) *T {
	switch v := s.(type) {
	case *mockedStore:
		if t, ok := v.mockedShareableStore.(T); ok {
			return &t
		}
		return get[T](v.ShareableStore)
	}
	return nil
}

func With[K basestore.ShareableStore, T MockableStore[K]](ms T) MockOption {
	return func(store basestore.ShareableStore) basestore.ShareableStore {
		return &mockedStore{
			ShareableStore:       store,
			mockedShareableStore: ms,
		}
	}
}

type MockOption func(basestore.ShareableStore) basestore.ShareableStore

func NewMockableShareableStore(s basestore.ShareableStore, options ...MockOption) basestore.ShareableStore {
	for _, option := range options {
		s = option(s)
	}

	return s
}
