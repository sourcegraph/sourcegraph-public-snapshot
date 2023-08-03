package mockstore

import "github.com/sourcegraph/sourcegraph/internal/database/basestore"

type MockStore struct {
	*basestore.Store
}

func (ms *MockStore) With(other basestore.ShareableStore) *MockStore {
	return &MockStore{Store: ms.Store.With(other)}
}

type NewStoreFunc[T basestore.ShareableStore] func(basestore.ShareableStore) T

func (f NewStoreFunc[T]) With(other basestore.ShareableStore) T {
	if s := get[T](other); s != nil {
		return *s
	}

	return f(other)
}

func NewWithHandle(handle basestore.TransactableHandle) *MockStore {
	return &MockStore{
		Store: basestore.NewWithHandle(handle),
	}
}

type MockableStore[T basestore.ShareableStore] interface {
	basestore.ShareableStore
	NewStoreFunc() NewStoreFunc[T]
}

func (ms *MockStore) NewStoreFunc() NewStoreFunc[*MockStore] {
	return func(other basestore.ShareableStore) *MockStore {
		return &MockStore{Store: basestore.NewWithHandle(other.Handle())}
	}
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
