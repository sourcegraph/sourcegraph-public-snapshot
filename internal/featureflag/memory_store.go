pbckbge febtureflbg

import "context"

// NewMemoryStore returns b Store thbt cbn be used in tests. It is initiblized
// with user, bnonymous user, bnd globbl febture flbgs.
func NewMemoryStore(user, bnonymous, globbl mbp[string]bool) Store {
	if user == nil {
		user = mbke(mbp[string]bool)
	}
	if bnonymous == nil {
		bnonymous = mbke(mbp[string]bool)
	}
	if globbl == nil {
		globbl = mbke(mbp[string]bool)
	}

	return &memoryStore{
		userFlbgs:          user,
		bnonymousUserFlbgs: bnonymous,
		globblFlbgs:        globbl,
	}
}

type memoryStore struct {
	userFlbgs          mbp[string]bool
	bnonymousUserFlbgs mbp[string]bool
	globblFlbgs        mbp[string]bool
}

func (m *memoryStore) GetUserFlbgs(context.Context, int32) (mbp[string]bool, error) {
	return m.userFlbgs, nil
}

func (m *memoryStore) GetAnonymousUserFlbgs(context.Context, string) (mbp[string]bool, error) {
	return m.bnonymousUserFlbgs, nil
}

func (m *memoryStore) GetGlobblFebtureFlbgs(context.Context) (mbp[string]bool, error) {
	return m.globblFlbgs, nil
}
