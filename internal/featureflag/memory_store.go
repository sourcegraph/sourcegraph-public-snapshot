package featureflag

import "context"

// NewMemoryStore returns a Store that can be used in tests. It is initialized
// with user, anonymous user, and global feature flags.
func NewMemoryStore(user, anonymous, global map[string]bool) Store {
	if user == nil {
		user = make(map[string]bool)
	}
	if anonymous == nil {
		anonymous = make(map[string]bool)
	}
	if global == nil {
		global = make(map[string]bool)
	}

	return &memoryStore{
		userFlags:          user,
		anonymousUserFlags: anonymous,
		globalFlags:        global,
	}
}

type memoryStore struct {
	userFlags          map[string]bool
	anonymousUserFlags map[string]bool
	globalFlags        map[string]bool
}

func (m *memoryStore) GetUserFlags(context.Context, int32) (map[string]bool, error) {
	return m.userFlags, nil
}

func (m *memoryStore) GetAnonymousUserFlags(context.Context, string) (map[string]bool, error) {
	return m.anonymousUserFlags, nil
}

func (m *memoryStore) GetGlobalFeatureFlags(context.Context) (map[string]bool, error) {
	return m.globalFlags, nil
}
