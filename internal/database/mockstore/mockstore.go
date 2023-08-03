package mockstore

import "github.com/sourcegraph/sourcegraph/internal/database/basestore"

type MockStoreee struct {
	basestore.ShareableStore
}

type MockableStore struct {
	*basestore.Store
}

func (ms *MockableStore) With(other basestore.ShareableStore) *MockableStore {
	return &MockableStore{basestore.NewWithHandle(other.Handle())}
}

func (ms *MockStoreee) ApplyMock(other basestore.ShareableStore) basestore.ShareableStore {
	return &mockedStore{
		ShareableStore:       other,
		mockedShareableStore: ms,
	}
}

type NewStoreFunc[T basestore.ShareableStore] func(basestore.ShareableStore) T

func (f NewStoreFunc[T]) With(other basestore.ShareableStore) T {
	if s := get[T](other); s != nil {
		return *s
	}

	return f(other)
}

type MockStore interface {
	Mock() MockedStore
}

type MockedStore interface {
	basestore.ShareableStore
	ApplyMock(basestore.ShareableStore) basestore.ShareableStore
}

func NewnewMockStore(store basestore.ShareableStore) MockedStore {
	return &MockStoreee{store}
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

func With[T MockedStore](ms T) MockOption {
	return func(store basestore.ShareableStore) basestore.ShareableStore {
		return &mockedStore{
			ShareableStore:       store,
			mockedShareableStore: ms,
		}
	}
}

type MockOption func(basestore.ShareableStore) basestore.ShareableStore

func NewMockableShareableStore(s basestore.ShareableStore, stores ...MockStore) basestore.ShareableStore {
	for _, store := range stores {
		s = store.Mock().ApplyMock(s)
	}

	return s
}
