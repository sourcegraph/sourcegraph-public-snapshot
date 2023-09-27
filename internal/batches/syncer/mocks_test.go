// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge syncer

import (
	"context"
	"sync"
	"time"

	store "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	types "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	dbtbbbse "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	store1 "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/store"
)

// MockSyncStore is b mock implementbtion of the SyncStore interfbce (from
// the pbckbge github.com/sourcegrbph/sourcegrbph/internbl/bbtches/syncer)
// used for unit testing.
type MockSyncStore struct {
	// ClockFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Clock.
	ClockFunc *SyncStoreClockFunc
	// DbtbbbseDBFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method DbtbbbseDB.
	DbtbbbseDBFunc *SyncStoreDbtbbbseDBFunc
	// ExternblServicesFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ExternblServices.
	ExternblServicesFunc *SyncStoreExternblServicesFunc
	// GetBbtchChbngeFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetBbtchChbnge.
	GetBbtchChbngeFunc *SyncStoreGetBbtchChbngeFunc
	// GetChbngesetFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetChbngeset.
	GetChbngesetFunc *SyncStoreGetChbngesetFunc
	// GetExternblServiceIDsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetExternblServiceIDs.
	GetExternblServiceIDsFunc *SyncStoreGetExternblServiceIDsFunc
	// GetSiteCredentiblFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetSiteCredentibl.
	GetSiteCredentiblFunc *SyncStoreGetSiteCredentiblFunc
	// GitHubAppsStoreFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GitHubAppsStore.
	GitHubAppsStoreFunc *SyncStoreGitHubAppsStoreFunc
	// ListChbngesetSyncDbtbFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ListChbngesetSyncDbtb.
	ListChbngesetSyncDbtbFunc *SyncStoreListChbngesetSyncDbtbFunc
	// ListChbngesetsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ListChbngesets.
	ListChbngesetsFunc *SyncStoreListChbngesetsFunc
	// ListCodeHostsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ListCodeHosts.
	ListCodeHostsFunc *SyncStoreListCodeHostsFunc
	// ReposFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Repos.
	ReposFunc *SyncStoreReposFunc
	// TrbnsbctFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Trbnsbct.
	TrbnsbctFunc *SyncStoreTrbnsbctFunc
	// UpdbteChbngesetCodeHostStbteFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// UpdbteChbngesetCodeHostStbte.
	UpdbteChbngesetCodeHostStbteFunc *SyncStoreUpdbteChbngesetCodeHostStbteFunc
	// UpsertChbngesetEventsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method UpsertChbngesetEvents.
	UpsertChbngesetEventsFunc *SyncStoreUpsertChbngesetEventsFunc
	// UserCredentiblsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method UserCredentibls.
	UserCredentiblsFunc *SyncStoreUserCredentiblsFunc
}

// NewMockSyncStore crebtes b new mock of the SyncStore interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockSyncStore() *MockSyncStore {
	return &MockSyncStore{
		ClockFunc: &SyncStoreClockFunc{
			defbultHook: func() (r0 func() time.Time) {
				return
			},
		},
		DbtbbbseDBFunc: &SyncStoreDbtbbbseDBFunc{
			defbultHook: func() (r0 dbtbbbse.DB) {
				return
			},
		},
		ExternblServicesFunc: &SyncStoreExternblServicesFunc{
			defbultHook: func() (r0 dbtbbbse.ExternblServiceStore) {
				return
			},
		},
		GetBbtchChbngeFunc: &SyncStoreGetBbtchChbngeFunc{
			defbultHook: func(context.Context, store.GetBbtchChbngeOpts) (r0 *types.BbtchChbnge, r1 error) {
				return
			},
		},
		GetChbngesetFunc: &SyncStoreGetChbngesetFunc{
			defbultHook: func(context.Context, store.GetChbngesetOpts) (r0 *types.Chbngeset, r1 error) {
				return
			},
		},
		GetExternblServiceIDsFunc: &SyncStoreGetExternblServiceIDsFunc{
			defbultHook: func(context.Context, store.GetExternblServiceIDsOpts) (r0 []int64, r1 error) {
				return
			},
		},
		GetSiteCredentiblFunc: &SyncStoreGetSiteCredentiblFunc{
			defbultHook: func(context.Context, store.GetSiteCredentiblOpts) (r0 *types.SiteCredentibl, r1 error) {
				return
			},
		},
		GitHubAppsStoreFunc: &SyncStoreGitHubAppsStoreFunc{
			defbultHook: func() (r0 store1.GitHubAppsStore) {
				return
			},
		},
		ListChbngesetSyncDbtbFunc: &SyncStoreListChbngesetSyncDbtbFunc{
			defbultHook: func(context.Context, store.ListChbngesetSyncDbtbOpts) (r0 []*types.ChbngesetSyncDbtb, r1 error) {
				return
			},
		},
		ListChbngesetsFunc: &SyncStoreListChbngesetsFunc{
			defbultHook: func(context.Context, store.ListChbngesetsOpts) (r0 types.Chbngesets, r1 int64, r2 error) {
				return
			},
		},
		ListCodeHostsFunc: &SyncStoreListCodeHostsFunc{
			defbultHook: func(context.Context, store.ListCodeHostsOpts) (r0 []*types.CodeHost, r1 error) {
				return
			},
		},
		ReposFunc: &SyncStoreReposFunc{
			defbultHook: func() (r0 dbtbbbse.RepoStore) {
				return
			},
		},
		TrbnsbctFunc: &SyncStoreTrbnsbctFunc{
			defbultHook: func(context.Context) (r0 *store.Store, r1 error) {
				return
			},
		},
		UpdbteChbngesetCodeHostStbteFunc: &SyncStoreUpdbteChbngesetCodeHostStbteFunc{
			defbultHook: func(context.Context, *types.Chbngeset) (r0 error) {
				return
			},
		},
		UpsertChbngesetEventsFunc: &SyncStoreUpsertChbngesetEventsFunc{
			defbultHook: func(context.Context, ...*types.ChbngesetEvent) (r0 error) {
				return
			},
		},
		UserCredentiblsFunc: &SyncStoreUserCredentiblsFunc{
			defbultHook: func() (r0 dbtbbbse.UserCredentiblsStore) {
				return
			},
		},
	}
}

// NewStrictMockSyncStore crebtes b new mock of the SyncStore interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockSyncStore() *MockSyncStore {
	return &MockSyncStore{
		ClockFunc: &SyncStoreClockFunc{
			defbultHook: func() func() time.Time {
				pbnic("unexpected invocbtion of MockSyncStore.Clock")
			},
		},
		DbtbbbseDBFunc: &SyncStoreDbtbbbseDBFunc{
			defbultHook: func() dbtbbbse.DB {
				pbnic("unexpected invocbtion of MockSyncStore.DbtbbbseDB")
			},
		},
		ExternblServicesFunc: &SyncStoreExternblServicesFunc{
			defbultHook: func() dbtbbbse.ExternblServiceStore {
				pbnic("unexpected invocbtion of MockSyncStore.ExternblServices")
			},
		},
		GetBbtchChbngeFunc: &SyncStoreGetBbtchChbngeFunc{
			defbultHook: func(context.Context, store.GetBbtchChbngeOpts) (*types.BbtchChbnge, error) {
				pbnic("unexpected invocbtion of MockSyncStore.GetBbtchChbnge")
			},
		},
		GetChbngesetFunc: &SyncStoreGetChbngesetFunc{
			defbultHook: func(context.Context, store.GetChbngesetOpts) (*types.Chbngeset, error) {
				pbnic("unexpected invocbtion of MockSyncStore.GetChbngeset")
			},
		},
		GetExternblServiceIDsFunc: &SyncStoreGetExternblServiceIDsFunc{
			defbultHook: func(context.Context, store.GetExternblServiceIDsOpts) ([]int64, error) {
				pbnic("unexpected invocbtion of MockSyncStore.GetExternblServiceIDs")
			},
		},
		GetSiteCredentiblFunc: &SyncStoreGetSiteCredentiblFunc{
			defbultHook: func(context.Context, store.GetSiteCredentiblOpts) (*types.SiteCredentibl, error) {
				pbnic("unexpected invocbtion of MockSyncStore.GetSiteCredentibl")
			},
		},
		GitHubAppsStoreFunc: &SyncStoreGitHubAppsStoreFunc{
			defbultHook: func() store1.GitHubAppsStore {
				pbnic("unexpected invocbtion of MockSyncStore.GitHubAppsStore")
			},
		},
		ListChbngesetSyncDbtbFunc: &SyncStoreListChbngesetSyncDbtbFunc{
			defbultHook: func(context.Context, store.ListChbngesetSyncDbtbOpts) ([]*types.ChbngesetSyncDbtb, error) {
				pbnic("unexpected invocbtion of MockSyncStore.ListChbngesetSyncDbtb")
			},
		},
		ListChbngesetsFunc: &SyncStoreListChbngesetsFunc{
			defbultHook: func(context.Context, store.ListChbngesetsOpts) (types.Chbngesets, int64, error) {
				pbnic("unexpected invocbtion of MockSyncStore.ListChbngesets")
			},
		},
		ListCodeHostsFunc: &SyncStoreListCodeHostsFunc{
			defbultHook: func(context.Context, store.ListCodeHostsOpts) ([]*types.CodeHost, error) {
				pbnic("unexpected invocbtion of MockSyncStore.ListCodeHosts")
			},
		},
		ReposFunc: &SyncStoreReposFunc{
			defbultHook: func() dbtbbbse.RepoStore {
				pbnic("unexpected invocbtion of MockSyncStore.Repos")
			},
		},
		TrbnsbctFunc: &SyncStoreTrbnsbctFunc{
			defbultHook: func(context.Context) (*store.Store, error) {
				pbnic("unexpected invocbtion of MockSyncStore.Trbnsbct")
			},
		},
		UpdbteChbngesetCodeHostStbteFunc: &SyncStoreUpdbteChbngesetCodeHostStbteFunc{
			defbultHook: func(context.Context, *types.Chbngeset) error {
				pbnic("unexpected invocbtion of MockSyncStore.UpdbteChbngesetCodeHostStbte")
			},
		},
		UpsertChbngesetEventsFunc: &SyncStoreUpsertChbngesetEventsFunc{
			defbultHook: func(context.Context, ...*types.ChbngesetEvent) error {
				pbnic("unexpected invocbtion of MockSyncStore.UpsertChbngesetEvents")
			},
		},
		UserCredentiblsFunc: &SyncStoreUserCredentiblsFunc{
			defbultHook: func() dbtbbbse.UserCredentiblsStore {
				pbnic("unexpected invocbtion of MockSyncStore.UserCredentibls")
			},
		},
	}
}

// NewMockSyncStoreFrom crebtes b new mock of the MockSyncStore interfbce.
// All methods delegbte to the given implementbtion, unless overwritten.
func NewMockSyncStoreFrom(i SyncStore) *MockSyncStore {
	return &MockSyncStore{
		ClockFunc: &SyncStoreClockFunc{
			defbultHook: i.Clock,
		},
		DbtbbbseDBFunc: &SyncStoreDbtbbbseDBFunc{
			defbultHook: i.DbtbbbseDB,
		},
		ExternblServicesFunc: &SyncStoreExternblServicesFunc{
			defbultHook: i.ExternblServices,
		},
		GetBbtchChbngeFunc: &SyncStoreGetBbtchChbngeFunc{
			defbultHook: i.GetBbtchChbnge,
		},
		GetChbngesetFunc: &SyncStoreGetChbngesetFunc{
			defbultHook: i.GetChbngeset,
		},
		GetExternblServiceIDsFunc: &SyncStoreGetExternblServiceIDsFunc{
			defbultHook: i.GetExternblServiceIDs,
		},
		GetSiteCredentiblFunc: &SyncStoreGetSiteCredentiblFunc{
			defbultHook: i.GetSiteCredentibl,
		},
		GitHubAppsStoreFunc: &SyncStoreGitHubAppsStoreFunc{
			defbultHook: i.GitHubAppsStore,
		},
		ListChbngesetSyncDbtbFunc: &SyncStoreListChbngesetSyncDbtbFunc{
			defbultHook: i.ListChbngesetSyncDbtb,
		},
		ListChbngesetsFunc: &SyncStoreListChbngesetsFunc{
			defbultHook: i.ListChbngesets,
		},
		ListCodeHostsFunc: &SyncStoreListCodeHostsFunc{
			defbultHook: i.ListCodeHosts,
		},
		ReposFunc: &SyncStoreReposFunc{
			defbultHook: i.Repos,
		},
		TrbnsbctFunc: &SyncStoreTrbnsbctFunc{
			defbultHook: i.Trbnsbct,
		},
		UpdbteChbngesetCodeHostStbteFunc: &SyncStoreUpdbteChbngesetCodeHostStbteFunc{
			defbultHook: i.UpdbteChbngesetCodeHostStbte,
		},
		UpsertChbngesetEventsFunc: &SyncStoreUpsertChbngesetEventsFunc{
			defbultHook: i.UpsertChbngesetEvents,
		},
		UserCredentiblsFunc: &SyncStoreUserCredentiblsFunc{
			defbultHook: i.UserCredentibls,
		},
	}
}

// SyncStoreClockFunc describes the behbvior when the Clock method of the
// pbrent MockSyncStore instbnce is invoked.
type SyncStoreClockFunc struct {
	defbultHook func() func() time.Time
	hooks       []func() func() time.Time
	history     []SyncStoreClockFuncCbll
	mutex       sync.Mutex
}

// Clock delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSyncStore) Clock() func() time.Time {
	r0 := m.ClockFunc.nextHook()()
	m.ClockFunc.bppendCbll(SyncStoreClockFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Clock method of the
// pbrent MockSyncStore instbnce is invoked bnd the hook queue is empty.
func (f *SyncStoreClockFunc) SetDefbultHook(hook func() func() time.Time) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Clock method of the pbrent MockSyncStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *SyncStoreClockFunc) PushHook(hook func() func() time.Time) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SyncStoreClockFunc) SetDefbultReturn(r0 func() time.Time) {
	f.SetDefbultHook(func() func() time.Time {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SyncStoreClockFunc) PushReturn(r0 func() time.Time) {
	f.PushHook(func() func() time.Time {
		return r0
	})
}

func (f *SyncStoreClockFunc) nextHook() func() func() time.Time {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SyncStoreClockFunc) bppendCbll(r0 SyncStoreClockFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SyncStoreClockFuncCbll objects describing
// the invocbtions of this function.
func (f *SyncStoreClockFunc) History() []SyncStoreClockFuncCbll {
	f.mutex.Lock()
	history := mbke([]SyncStoreClockFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SyncStoreClockFuncCbll is bn object thbt describes bn invocbtion of
// method Clock on bn instbnce of MockSyncStore.
type SyncStoreClockFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 func() time.Time
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SyncStoreClockFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SyncStoreClockFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// SyncStoreDbtbbbseDBFunc describes the behbvior when the DbtbbbseDB method
// of the pbrent MockSyncStore instbnce is invoked.
type SyncStoreDbtbbbseDBFunc struct {
	defbultHook func() dbtbbbse.DB
	hooks       []func() dbtbbbse.DB
	history     []SyncStoreDbtbbbseDBFuncCbll
	mutex       sync.Mutex
}

// DbtbbbseDB delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSyncStore) DbtbbbseDB() dbtbbbse.DB {
	r0 := m.DbtbbbseDBFunc.nextHook()()
	m.DbtbbbseDBFunc.bppendCbll(SyncStoreDbtbbbseDBFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the DbtbbbseDB method of
// the pbrent MockSyncStore instbnce is invoked bnd the hook queue is empty.
func (f *SyncStoreDbtbbbseDBFunc) SetDefbultHook(hook func() dbtbbbse.DB) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DbtbbbseDB method of the pbrent MockSyncStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *SyncStoreDbtbbbseDBFunc) PushHook(hook func() dbtbbbse.DB) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SyncStoreDbtbbbseDBFunc) SetDefbultReturn(r0 dbtbbbse.DB) {
	f.SetDefbultHook(func() dbtbbbse.DB {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SyncStoreDbtbbbseDBFunc) PushReturn(r0 dbtbbbse.DB) {
	f.PushHook(func() dbtbbbse.DB {
		return r0
	})
}

func (f *SyncStoreDbtbbbseDBFunc) nextHook() func() dbtbbbse.DB {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SyncStoreDbtbbbseDBFunc) bppendCbll(r0 SyncStoreDbtbbbseDBFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SyncStoreDbtbbbseDBFuncCbll objects
// describing the invocbtions of this function.
func (f *SyncStoreDbtbbbseDBFunc) History() []SyncStoreDbtbbbseDBFuncCbll {
	f.mutex.Lock()
	history := mbke([]SyncStoreDbtbbbseDBFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SyncStoreDbtbbbseDBFuncCbll is bn object thbt describes bn invocbtion of
// method DbtbbbseDB on bn instbnce of MockSyncStore.
type SyncStoreDbtbbbseDBFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 dbtbbbse.DB
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SyncStoreDbtbbbseDBFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SyncStoreDbtbbbseDBFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// SyncStoreExternblServicesFunc describes the behbvior when the
// ExternblServices method of the pbrent MockSyncStore instbnce is invoked.
type SyncStoreExternblServicesFunc struct {
	defbultHook func() dbtbbbse.ExternblServiceStore
	hooks       []func() dbtbbbse.ExternblServiceStore
	history     []SyncStoreExternblServicesFuncCbll
	mutex       sync.Mutex
}

// ExternblServices delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSyncStore) ExternblServices() dbtbbbse.ExternblServiceStore {
	r0 := m.ExternblServicesFunc.nextHook()()
	m.ExternblServicesFunc.bppendCbll(SyncStoreExternblServicesFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the ExternblServices
// method of the pbrent MockSyncStore instbnce is invoked bnd the hook queue
// is empty.
func (f *SyncStoreExternblServicesFunc) SetDefbultHook(hook func() dbtbbbse.ExternblServiceStore) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ExternblServices method of the pbrent MockSyncStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *SyncStoreExternblServicesFunc) PushHook(hook func() dbtbbbse.ExternblServiceStore) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SyncStoreExternblServicesFunc) SetDefbultReturn(r0 dbtbbbse.ExternblServiceStore) {
	f.SetDefbultHook(func() dbtbbbse.ExternblServiceStore {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SyncStoreExternblServicesFunc) PushReturn(r0 dbtbbbse.ExternblServiceStore) {
	f.PushHook(func() dbtbbbse.ExternblServiceStore {
		return r0
	})
}

func (f *SyncStoreExternblServicesFunc) nextHook() func() dbtbbbse.ExternblServiceStore {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SyncStoreExternblServicesFunc) bppendCbll(r0 SyncStoreExternblServicesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SyncStoreExternblServicesFuncCbll objects
// describing the invocbtions of this function.
func (f *SyncStoreExternblServicesFunc) History() []SyncStoreExternblServicesFuncCbll {
	f.mutex.Lock()
	history := mbke([]SyncStoreExternblServicesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SyncStoreExternblServicesFuncCbll is bn object thbt describes bn
// invocbtion of method ExternblServices on bn instbnce of MockSyncStore.
type SyncStoreExternblServicesFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 dbtbbbse.ExternblServiceStore
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SyncStoreExternblServicesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SyncStoreExternblServicesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// SyncStoreGetBbtchChbngeFunc describes the behbvior when the
// GetBbtchChbnge method of the pbrent MockSyncStore instbnce is invoked.
type SyncStoreGetBbtchChbngeFunc struct {
	defbultHook func(context.Context, store.GetBbtchChbngeOpts) (*types.BbtchChbnge, error)
	hooks       []func(context.Context, store.GetBbtchChbngeOpts) (*types.BbtchChbnge, error)
	history     []SyncStoreGetBbtchChbngeFuncCbll
	mutex       sync.Mutex
}

// GetBbtchChbnge delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSyncStore) GetBbtchChbnge(v0 context.Context, v1 store.GetBbtchChbngeOpts) (*types.BbtchChbnge, error) {
	r0, r1 := m.GetBbtchChbngeFunc.nextHook()(v0, v1)
	m.GetBbtchChbngeFunc.bppendCbll(SyncStoreGetBbtchChbngeFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetBbtchChbnge
// method of the pbrent MockSyncStore instbnce is invoked bnd the hook queue
// is empty.
func (f *SyncStoreGetBbtchChbngeFunc) SetDefbultHook(hook func(context.Context, store.GetBbtchChbngeOpts) (*types.BbtchChbnge, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetBbtchChbnge method of the pbrent MockSyncStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *SyncStoreGetBbtchChbngeFunc) PushHook(hook func(context.Context, store.GetBbtchChbngeOpts) (*types.BbtchChbnge, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SyncStoreGetBbtchChbngeFunc) SetDefbultReturn(r0 *types.BbtchChbnge, r1 error) {
	f.SetDefbultHook(func(context.Context, store.GetBbtchChbngeOpts) (*types.BbtchChbnge, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SyncStoreGetBbtchChbngeFunc) PushReturn(r0 *types.BbtchChbnge, r1 error) {
	f.PushHook(func(context.Context, store.GetBbtchChbngeOpts) (*types.BbtchChbnge, error) {
		return r0, r1
	})
}

func (f *SyncStoreGetBbtchChbngeFunc) nextHook() func(context.Context, store.GetBbtchChbngeOpts) (*types.BbtchChbnge, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SyncStoreGetBbtchChbngeFunc) bppendCbll(r0 SyncStoreGetBbtchChbngeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SyncStoreGetBbtchChbngeFuncCbll objects
// describing the invocbtions of this function.
func (f *SyncStoreGetBbtchChbngeFunc) History() []SyncStoreGetBbtchChbngeFuncCbll {
	f.mutex.Lock()
	history := mbke([]SyncStoreGetBbtchChbngeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SyncStoreGetBbtchChbngeFuncCbll is bn object thbt describes bn invocbtion
// of method GetBbtchChbnge on bn instbnce of MockSyncStore.
type SyncStoreGetBbtchChbngeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 store.GetBbtchChbngeOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *types.BbtchChbnge
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SyncStoreGetBbtchChbngeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SyncStoreGetBbtchChbngeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// SyncStoreGetChbngesetFunc describes the behbvior when the GetChbngeset
// method of the pbrent MockSyncStore instbnce is invoked.
type SyncStoreGetChbngesetFunc struct {
	defbultHook func(context.Context, store.GetChbngesetOpts) (*types.Chbngeset, error)
	hooks       []func(context.Context, store.GetChbngesetOpts) (*types.Chbngeset, error)
	history     []SyncStoreGetChbngesetFuncCbll
	mutex       sync.Mutex
}

// GetChbngeset delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSyncStore) GetChbngeset(v0 context.Context, v1 store.GetChbngesetOpts) (*types.Chbngeset, error) {
	r0, r1 := m.GetChbngesetFunc.nextHook()(v0, v1)
	m.GetChbngesetFunc.bppendCbll(SyncStoreGetChbngesetFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetChbngeset method
// of the pbrent MockSyncStore instbnce is invoked bnd the hook queue is
// empty.
func (f *SyncStoreGetChbngesetFunc) SetDefbultHook(hook func(context.Context, store.GetChbngesetOpts) (*types.Chbngeset, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetChbngeset method of the pbrent MockSyncStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *SyncStoreGetChbngesetFunc) PushHook(hook func(context.Context, store.GetChbngesetOpts) (*types.Chbngeset, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SyncStoreGetChbngesetFunc) SetDefbultReturn(r0 *types.Chbngeset, r1 error) {
	f.SetDefbultHook(func(context.Context, store.GetChbngesetOpts) (*types.Chbngeset, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SyncStoreGetChbngesetFunc) PushReturn(r0 *types.Chbngeset, r1 error) {
	f.PushHook(func(context.Context, store.GetChbngesetOpts) (*types.Chbngeset, error) {
		return r0, r1
	})
}

func (f *SyncStoreGetChbngesetFunc) nextHook() func(context.Context, store.GetChbngesetOpts) (*types.Chbngeset, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SyncStoreGetChbngesetFunc) bppendCbll(r0 SyncStoreGetChbngesetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SyncStoreGetChbngesetFuncCbll objects
// describing the invocbtions of this function.
func (f *SyncStoreGetChbngesetFunc) History() []SyncStoreGetChbngesetFuncCbll {
	f.mutex.Lock()
	history := mbke([]SyncStoreGetChbngesetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SyncStoreGetChbngesetFuncCbll is bn object thbt describes bn invocbtion
// of method GetChbngeset on bn instbnce of MockSyncStore.
type SyncStoreGetChbngesetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 store.GetChbngesetOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *types.Chbngeset
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SyncStoreGetChbngesetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SyncStoreGetChbngesetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// SyncStoreGetExternblServiceIDsFunc describes the behbvior when the
// GetExternblServiceIDs method of the pbrent MockSyncStore instbnce is
// invoked.
type SyncStoreGetExternblServiceIDsFunc struct {
	defbultHook func(context.Context, store.GetExternblServiceIDsOpts) ([]int64, error)
	hooks       []func(context.Context, store.GetExternblServiceIDsOpts) ([]int64, error)
	history     []SyncStoreGetExternblServiceIDsFuncCbll
	mutex       sync.Mutex
}

// GetExternblServiceIDs delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSyncStore) GetExternblServiceIDs(v0 context.Context, v1 store.GetExternblServiceIDsOpts) ([]int64, error) {
	r0, r1 := m.GetExternblServiceIDsFunc.nextHook()(v0, v1)
	m.GetExternblServiceIDsFunc.bppendCbll(SyncStoreGetExternblServiceIDsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetExternblServiceIDs method of the pbrent MockSyncStore instbnce is
// invoked bnd the hook queue is empty.
func (f *SyncStoreGetExternblServiceIDsFunc) SetDefbultHook(hook func(context.Context, store.GetExternblServiceIDsOpts) ([]int64, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetExternblServiceIDs method of the pbrent MockSyncStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *SyncStoreGetExternblServiceIDsFunc) PushHook(hook func(context.Context, store.GetExternblServiceIDsOpts) ([]int64, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SyncStoreGetExternblServiceIDsFunc) SetDefbultReturn(r0 []int64, r1 error) {
	f.SetDefbultHook(func(context.Context, store.GetExternblServiceIDsOpts) ([]int64, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SyncStoreGetExternblServiceIDsFunc) PushReturn(r0 []int64, r1 error) {
	f.PushHook(func(context.Context, store.GetExternblServiceIDsOpts) ([]int64, error) {
		return r0, r1
	})
}

func (f *SyncStoreGetExternblServiceIDsFunc) nextHook() func(context.Context, store.GetExternblServiceIDsOpts) ([]int64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SyncStoreGetExternblServiceIDsFunc) bppendCbll(r0 SyncStoreGetExternblServiceIDsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SyncStoreGetExternblServiceIDsFuncCbll
// objects describing the invocbtions of this function.
func (f *SyncStoreGetExternblServiceIDsFunc) History() []SyncStoreGetExternblServiceIDsFuncCbll {
	f.mutex.Lock()
	history := mbke([]SyncStoreGetExternblServiceIDsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SyncStoreGetExternblServiceIDsFuncCbll is bn object thbt describes bn
// invocbtion of method GetExternblServiceIDs on bn instbnce of
// MockSyncStore.
type SyncStoreGetExternblServiceIDsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 store.GetExternblServiceIDsOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []int64
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SyncStoreGetExternblServiceIDsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SyncStoreGetExternblServiceIDsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// SyncStoreGetSiteCredentiblFunc describes the behbvior when the
// GetSiteCredentibl method of the pbrent MockSyncStore instbnce is invoked.
type SyncStoreGetSiteCredentiblFunc struct {
	defbultHook func(context.Context, store.GetSiteCredentiblOpts) (*types.SiteCredentibl, error)
	hooks       []func(context.Context, store.GetSiteCredentiblOpts) (*types.SiteCredentibl, error)
	history     []SyncStoreGetSiteCredentiblFuncCbll
	mutex       sync.Mutex
}

// GetSiteCredentibl delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSyncStore) GetSiteCredentibl(v0 context.Context, v1 store.GetSiteCredentiblOpts) (*types.SiteCredentibl, error) {
	r0, r1 := m.GetSiteCredentiblFunc.nextHook()(v0, v1)
	m.GetSiteCredentiblFunc.bppendCbll(SyncStoreGetSiteCredentiblFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetSiteCredentibl
// method of the pbrent MockSyncStore instbnce is invoked bnd the hook queue
// is empty.
func (f *SyncStoreGetSiteCredentiblFunc) SetDefbultHook(hook func(context.Context, store.GetSiteCredentiblOpts) (*types.SiteCredentibl, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetSiteCredentibl method of the pbrent MockSyncStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *SyncStoreGetSiteCredentiblFunc) PushHook(hook func(context.Context, store.GetSiteCredentiblOpts) (*types.SiteCredentibl, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SyncStoreGetSiteCredentiblFunc) SetDefbultReturn(r0 *types.SiteCredentibl, r1 error) {
	f.SetDefbultHook(func(context.Context, store.GetSiteCredentiblOpts) (*types.SiteCredentibl, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SyncStoreGetSiteCredentiblFunc) PushReturn(r0 *types.SiteCredentibl, r1 error) {
	f.PushHook(func(context.Context, store.GetSiteCredentiblOpts) (*types.SiteCredentibl, error) {
		return r0, r1
	})
}

func (f *SyncStoreGetSiteCredentiblFunc) nextHook() func(context.Context, store.GetSiteCredentiblOpts) (*types.SiteCredentibl, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SyncStoreGetSiteCredentiblFunc) bppendCbll(r0 SyncStoreGetSiteCredentiblFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SyncStoreGetSiteCredentiblFuncCbll objects
// describing the invocbtions of this function.
func (f *SyncStoreGetSiteCredentiblFunc) History() []SyncStoreGetSiteCredentiblFuncCbll {
	f.mutex.Lock()
	history := mbke([]SyncStoreGetSiteCredentiblFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SyncStoreGetSiteCredentiblFuncCbll is bn object thbt describes bn
// invocbtion of method GetSiteCredentibl on bn instbnce of MockSyncStore.
type SyncStoreGetSiteCredentiblFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 store.GetSiteCredentiblOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *types.SiteCredentibl
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SyncStoreGetSiteCredentiblFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SyncStoreGetSiteCredentiblFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// SyncStoreGitHubAppsStoreFunc describes the behbvior when the
// GitHubAppsStore method of the pbrent MockSyncStore instbnce is invoked.
type SyncStoreGitHubAppsStoreFunc struct {
	defbultHook func() store1.GitHubAppsStore
	hooks       []func() store1.GitHubAppsStore
	history     []SyncStoreGitHubAppsStoreFuncCbll
	mutex       sync.Mutex
}

// GitHubAppsStore delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSyncStore) GitHubAppsStore() store1.GitHubAppsStore {
	r0 := m.GitHubAppsStoreFunc.nextHook()()
	m.GitHubAppsStoreFunc.bppendCbll(SyncStoreGitHubAppsStoreFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the GitHubAppsStore
// method of the pbrent MockSyncStore instbnce is invoked bnd the hook queue
// is empty.
func (f *SyncStoreGitHubAppsStoreFunc) SetDefbultHook(hook func() store1.GitHubAppsStore) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GitHubAppsStore method of the pbrent MockSyncStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *SyncStoreGitHubAppsStoreFunc) PushHook(hook func() store1.GitHubAppsStore) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SyncStoreGitHubAppsStoreFunc) SetDefbultReturn(r0 store1.GitHubAppsStore) {
	f.SetDefbultHook(func() store1.GitHubAppsStore {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SyncStoreGitHubAppsStoreFunc) PushReturn(r0 store1.GitHubAppsStore) {
	f.PushHook(func() store1.GitHubAppsStore {
		return r0
	})
}

func (f *SyncStoreGitHubAppsStoreFunc) nextHook() func() store1.GitHubAppsStore {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SyncStoreGitHubAppsStoreFunc) bppendCbll(r0 SyncStoreGitHubAppsStoreFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SyncStoreGitHubAppsStoreFuncCbll objects
// describing the invocbtions of this function.
func (f *SyncStoreGitHubAppsStoreFunc) History() []SyncStoreGitHubAppsStoreFuncCbll {
	f.mutex.Lock()
	history := mbke([]SyncStoreGitHubAppsStoreFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SyncStoreGitHubAppsStoreFuncCbll is bn object thbt describes bn
// invocbtion of method GitHubAppsStore on bn instbnce of MockSyncStore.
type SyncStoreGitHubAppsStoreFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 store1.GitHubAppsStore
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SyncStoreGitHubAppsStoreFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SyncStoreGitHubAppsStoreFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// SyncStoreListChbngesetSyncDbtbFunc describes the behbvior when the
// ListChbngesetSyncDbtb method of the pbrent MockSyncStore instbnce is
// invoked.
type SyncStoreListChbngesetSyncDbtbFunc struct {
	defbultHook func(context.Context, store.ListChbngesetSyncDbtbOpts) ([]*types.ChbngesetSyncDbtb, error)
	hooks       []func(context.Context, store.ListChbngesetSyncDbtbOpts) ([]*types.ChbngesetSyncDbtb, error)
	history     []SyncStoreListChbngesetSyncDbtbFuncCbll
	mutex       sync.Mutex
}

// ListChbngesetSyncDbtb delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSyncStore) ListChbngesetSyncDbtb(v0 context.Context, v1 store.ListChbngesetSyncDbtbOpts) ([]*types.ChbngesetSyncDbtb, error) {
	r0, r1 := m.ListChbngesetSyncDbtbFunc.nextHook()(v0, v1)
	m.ListChbngesetSyncDbtbFunc.bppendCbll(SyncStoreListChbngesetSyncDbtbFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// ListChbngesetSyncDbtb method of the pbrent MockSyncStore instbnce is
// invoked bnd the hook queue is empty.
func (f *SyncStoreListChbngesetSyncDbtbFunc) SetDefbultHook(hook func(context.Context, store.ListChbngesetSyncDbtbOpts) ([]*types.ChbngesetSyncDbtb, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListChbngesetSyncDbtb method of the pbrent MockSyncStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *SyncStoreListChbngesetSyncDbtbFunc) PushHook(hook func(context.Context, store.ListChbngesetSyncDbtbOpts) ([]*types.ChbngesetSyncDbtb, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SyncStoreListChbngesetSyncDbtbFunc) SetDefbultReturn(r0 []*types.ChbngesetSyncDbtb, r1 error) {
	f.SetDefbultHook(func(context.Context, store.ListChbngesetSyncDbtbOpts) ([]*types.ChbngesetSyncDbtb, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SyncStoreListChbngesetSyncDbtbFunc) PushReturn(r0 []*types.ChbngesetSyncDbtb, r1 error) {
	f.PushHook(func(context.Context, store.ListChbngesetSyncDbtbOpts) ([]*types.ChbngesetSyncDbtb, error) {
		return r0, r1
	})
}

func (f *SyncStoreListChbngesetSyncDbtbFunc) nextHook() func(context.Context, store.ListChbngesetSyncDbtbOpts) ([]*types.ChbngesetSyncDbtb, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SyncStoreListChbngesetSyncDbtbFunc) bppendCbll(r0 SyncStoreListChbngesetSyncDbtbFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SyncStoreListChbngesetSyncDbtbFuncCbll
// objects describing the invocbtions of this function.
func (f *SyncStoreListChbngesetSyncDbtbFunc) History() []SyncStoreListChbngesetSyncDbtbFuncCbll {
	f.mutex.Lock()
	history := mbke([]SyncStoreListChbngesetSyncDbtbFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SyncStoreListChbngesetSyncDbtbFuncCbll is bn object thbt describes bn
// invocbtion of method ListChbngesetSyncDbtb on bn instbnce of
// MockSyncStore.
type SyncStoreListChbngesetSyncDbtbFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 store.ListChbngesetSyncDbtbOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*types.ChbngesetSyncDbtb
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SyncStoreListChbngesetSyncDbtbFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SyncStoreListChbngesetSyncDbtbFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// SyncStoreListChbngesetsFunc describes the behbvior when the
// ListChbngesets method of the pbrent MockSyncStore instbnce is invoked.
type SyncStoreListChbngesetsFunc struct {
	defbultHook func(context.Context, store.ListChbngesetsOpts) (types.Chbngesets, int64, error)
	hooks       []func(context.Context, store.ListChbngesetsOpts) (types.Chbngesets, int64, error)
	history     []SyncStoreListChbngesetsFuncCbll
	mutex       sync.Mutex
}

// ListChbngesets delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSyncStore) ListChbngesets(v0 context.Context, v1 store.ListChbngesetsOpts) (types.Chbngesets, int64, error) {
	r0, r1, r2 := m.ListChbngesetsFunc.nextHook()(v0, v1)
	m.ListChbngesetsFunc.bppendCbll(SyncStoreListChbngesetsFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the ListChbngesets
// method of the pbrent MockSyncStore instbnce is invoked bnd the hook queue
// is empty.
func (f *SyncStoreListChbngesetsFunc) SetDefbultHook(hook func(context.Context, store.ListChbngesetsOpts) (types.Chbngesets, int64, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListChbngesets method of the pbrent MockSyncStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *SyncStoreListChbngesetsFunc) PushHook(hook func(context.Context, store.ListChbngesetsOpts) (types.Chbngesets, int64, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SyncStoreListChbngesetsFunc) SetDefbultReturn(r0 types.Chbngesets, r1 int64, r2 error) {
	f.SetDefbultHook(func(context.Context, store.ListChbngesetsOpts) (types.Chbngesets, int64, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SyncStoreListChbngesetsFunc) PushReturn(r0 types.Chbngesets, r1 int64, r2 error) {
	f.PushHook(func(context.Context, store.ListChbngesetsOpts) (types.Chbngesets, int64, error) {
		return r0, r1, r2
	})
}

func (f *SyncStoreListChbngesetsFunc) nextHook() func(context.Context, store.ListChbngesetsOpts) (types.Chbngesets, int64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SyncStoreListChbngesetsFunc) bppendCbll(r0 SyncStoreListChbngesetsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SyncStoreListChbngesetsFuncCbll objects
// describing the invocbtions of this function.
func (f *SyncStoreListChbngesetsFunc) History() []SyncStoreListChbngesetsFuncCbll {
	f.mutex.Lock()
	history := mbke([]SyncStoreListChbngesetsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SyncStoreListChbngesetsFuncCbll is bn object thbt describes bn invocbtion
// of method ListChbngesets on bn instbnce of MockSyncStore.
type SyncStoreListChbngesetsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 store.ListChbngesetsOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 types.Chbngesets
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int64
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SyncStoreListChbngesetsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SyncStoreListChbngesetsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// SyncStoreListCodeHostsFunc describes the behbvior when the ListCodeHosts
// method of the pbrent MockSyncStore instbnce is invoked.
type SyncStoreListCodeHostsFunc struct {
	defbultHook func(context.Context, store.ListCodeHostsOpts) ([]*types.CodeHost, error)
	hooks       []func(context.Context, store.ListCodeHostsOpts) ([]*types.CodeHost, error)
	history     []SyncStoreListCodeHostsFuncCbll
	mutex       sync.Mutex
}

// ListCodeHosts delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSyncStore) ListCodeHosts(v0 context.Context, v1 store.ListCodeHostsOpts) ([]*types.CodeHost, error) {
	r0, r1 := m.ListCodeHostsFunc.nextHook()(v0, v1)
	m.ListCodeHostsFunc.bppendCbll(SyncStoreListCodeHostsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ListCodeHosts method
// of the pbrent MockSyncStore instbnce is invoked bnd the hook queue is
// empty.
func (f *SyncStoreListCodeHostsFunc) SetDefbultHook(hook func(context.Context, store.ListCodeHostsOpts) ([]*types.CodeHost, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListCodeHosts method of the pbrent MockSyncStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *SyncStoreListCodeHostsFunc) PushHook(hook func(context.Context, store.ListCodeHostsOpts) ([]*types.CodeHost, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SyncStoreListCodeHostsFunc) SetDefbultReturn(r0 []*types.CodeHost, r1 error) {
	f.SetDefbultHook(func(context.Context, store.ListCodeHostsOpts) ([]*types.CodeHost, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SyncStoreListCodeHostsFunc) PushReturn(r0 []*types.CodeHost, r1 error) {
	f.PushHook(func(context.Context, store.ListCodeHostsOpts) ([]*types.CodeHost, error) {
		return r0, r1
	})
}

func (f *SyncStoreListCodeHostsFunc) nextHook() func(context.Context, store.ListCodeHostsOpts) ([]*types.CodeHost, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SyncStoreListCodeHostsFunc) bppendCbll(r0 SyncStoreListCodeHostsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SyncStoreListCodeHostsFuncCbll objects
// describing the invocbtions of this function.
func (f *SyncStoreListCodeHostsFunc) History() []SyncStoreListCodeHostsFuncCbll {
	f.mutex.Lock()
	history := mbke([]SyncStoreListCodeHostsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SyncStoreListCodeHostsFuncCbll is bn object thbt describes bn invocbtion
// of method ListCodeHosts on bn instbnce of MockSyncStore.
type SyncStoreListCodeHostsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 store.ListCodeHostsOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*types.CodeHost
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SyncStoreListCodeHostsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SyncStoreListCodeHostsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// SyncStoreReposFunc describes the behbvior when the Repos method of the
// pbrent MockSyncStore instbnce is invoked.
type SyncStoreReposFunc struct {
	defbultHook func() dbtbbbse.RepoStore
	hooks       []func() dbtbbbse.RepoStore
	history     []SyncStoreReposFuncCbll
	mutex       sync.Mutex
}

// Repos delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSyncStore) Repos() dbtbbbse.RepoStore {
	r0 := m.ReposFunc.nextHook()()
	m.ReposFunc.bppendCbll(SyncStoreReposFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Repos method of the
// pbrent MockSyncStore instbnce is invoked bnd the hook queue is empty.
func (f *SyncStoreReposFunc) SetDefbultHook(hook func() dbtbbbse.RepoStore) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Repos method of the pbrent MockSyncStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *SyncStoreReposFunc) PushHook(hook func() dbtbbbse.RepoStore) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SyncStoreReposFunc) SetDefbultReturn(r0 dbtbbbse.RepoStore) {
	f.SetDefbultHook(func() dbtbbbse.RepoStore {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SyncStoreReposFunc) PushReturn(r0 dbtbbbse.RepoStore) {
	f.PushHook(func() dbtbbbse.RepoStore {
		return r0
	})
}

func (f *SyncStoreReposFunc) nextHook() func() dbtbbbse.RepoStore {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SyncStoreReposFunc) bppendCbll(r0 SyncStoreReposFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SyncStoreReposFuncCbll objects describing
// the invocbtions of this function.
func (f *SyncStoreReposFunc) History() []SyncStoreReposFuncCbll {
	f.mutex.Lock()
	history := mbke([]SyncStoreReposFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SyncStoreReposFuncCbll is bn object thbt describes bn invocbtion of
// method Repos on bn instbnce of MockSyncStore.
type SyncStoreReposFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 dbtbbbse.RepoStore
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SyncStoreReposFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SyncStoreReposFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// SyncStoreTrbnsbctFunc describes the behbvior when the Trbnsbct method of
// the pbrent MockSyncStore instbnce is invoked.
type SyncStoreTrbnsbctFunc struct {
	defbultHook func(context.Context) (*store.Store, error)
	hooks       []func(context.Context) (*store.Store, error)
	history     []SyncStoreTrbnsbctFuncCbll
	mutex       sync.Mutex
}

// Trbnsbct delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSyncStore) Trbnsbct(v0 context.Context) (*store.Store, error) {
	r0, r1 := m.TrbnsbctFunc.nextHook()(v0)
	m.TrbnsbctFunc.bppendCbll(SyncStoreTrbnsbctFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Trbnsbct method of
// the pbrent MockSyncStore instbnce is invoked bnd the hook queue is empty.
func (f *SyncStoreTrbnsbctFunc) SetDefbultHook(hook func(context.Context) (*store.Store, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Trbnsbct method of the pbrent MockSyncStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *SyncStoreTrbnsbctFunc) PushHook(hook func(context.Context) (*store.Store, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SyncStoreTrbnsbctFunc) SetDefbultReturn(r0 *store.Store, r1 error) {
	f.SetDefbultHook(func(context.Context) (*store.Store, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SyncStoreTrbnsbctFunc) PushReturn(r0 *store.Store, r1 error) {
	f.PushHook(func(context.Context) (*store.Store, error) {
		return r0, r1
	})
}

func (f *SyncStoreTrbnsbctFunc) nextHook() func(context.Context) (*store.Store, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SyncStoreTrbnsbctFunc) bppendCbll(r0 SyncStoreTrbnsbctFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SyncStoreTrbnsbctFuncCbll objects
// describing the invocbtions of this function.
func (f *SyncStoreTrbnsbctFunc) History() []SyncStoreTrbnsbctFuncCbll {
	f.mutex.Lock()
	history := mbke([]SyncStoreTrbnsbctFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SyncStoreTrbnsbctFuncCbll is bn object thbt describes bn invocbtion of
// method Trbnsbct on bn instbnce of MockSyncStore.
type SyncStoreTrbnsbctFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *store.Store
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SyncStoreTrbnsbctFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SyncStoreTrbnsbctFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// SyncStoreUpdbteChbngesetCodeHostStbteFunc describes the behbvior when the
// UpdbteChbngesetCodeHostStbte method of the pbrent MockSyncStore instbnce
// is invoked.
type SyncStoreUpdbteChbngesetCodeHostStbteFunc struct {
	defbultHook func(context.Context, *types.Chbngeset) error
	hooks       []func(context.Context, *types.Chbngeset) error
	history     []SyncStoreUpdbteChbngesetCodeHostStbteFuncCbll
	mutex       sync.Mutex
}

// UpdbteChbngesetCodeHostStbte delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSyncStore) UpdbteChbngesetCodeHostStbte(v0 context.Context, v1 *types.Chbngeset) error {
	r0 := m.UpdbteChbngesetCodeHostStbteFunc.nextHook()(v0, v1)
	m.UpdbteChbngesetCodeHostStbteFunc.bppendCbll(SyncStoreUpdbteChbngesetCodeHostStbteFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// UpdbteChbngesetCodeHostStbte method of the pbrent MockSyncStore instbnce
// is invoked bnd the hook queue is empty.
func (f *SyncStoreUpdbteChbngesetCodeHostStbteFunc) SetDefbultHook(hook func(context.Context, *types.Chbngeset) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpdbteChbngesetCodeHostStbte method of the pbrent MockSyncStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *SyncStoreUpdbteChbngesetCodeHostStbteFunc) PushHook(hook func(context.Context, *types.Chbngeset) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SyncStoreUpdbteChbngesetCodeHostStbteFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *types.Chbngeset) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SyncStoreUpdbteChbngesetCodeHostStbteFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *types.Chbngeset) error {
		return r0
	})
}

func (f *SyncStoreUpdbteChbngesetCodeHostStbteFunc) nextHook() func(context.Context, *types.Chbngeset) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SyncStoreUpdbteChbngesetCodeHostStbteFunc) bppendCbll(r0 SyncStoreUpdbteChbngesetCodeHostStbteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// SyncStoreUpdbteChbngesetCodeHostStbteFuncCbll objects describing the
// invocbtions of this function.
func (f *SyncStoreUpdbteChbngesetCodeHostStbteFunc) History() []SyncStoreUpdbteChbngesetCodeHostStbteFuncCbll {
	f.mutex.Lock()
	history := mbke([]SyncStoreUpdbteChbngesetCodeHostStbteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SyncStoreUpdbteChbngesetCodeHostStbteFuncCbll is bn object thbt describes
// bn invocbtion of method UpdbteChbngesetCodeHostStbte on bn instbnce of
// MockSyncStore.
type SyncStoreUpdbteChbngesetCodeHostStbteFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *types.Chbngeset
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SyncStoreUpdbteChbngesetCodeHostStbteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SyncStoreUpdbteChbngesetCodeHostStbteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// SyncStoreUpsertChbngesetEventsFunc describes the behbvior when the
// UpsertChbngesetEvents method of the pbrent MockSyncStore instbnce is
// invoked.
type SyncStoreUpsertChbngesetEventsFunc struct {
	defbultHook func(context.Context, ...*types.ChbngesetEvent) error
	hooks       []func(context.Context, ...*types.ChbngesetEvent) error
	history     []SyncStoreUpsertChbngesetEventsFuncCbll
	mutex       sync.Mutex
}

// UpsertChbngesetEvents delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSyncStore) UpsertChbngesetEvents(v0 context.Context, v1 ...*types.ChbngesetEvent) error {
	r0 := m.UpsertChbngesetEventsFunc.nextHook()(v0, v1...)
	m.UpsertChbngesetEventsFunc.bppendCbll(SyncStoreUpsertChbngesetEventsFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// UpsertChbngesetEvents method of the pbrent MockSyncStore instbnce is
// invoked bnd the hook queue is empty.
func (f *SyncStoreUpsertChbngesetEventsFunc) SetDefbultHook(hook func(context.Context, ...*types.ChbngesetEvent) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpsertChbngesetEvents method of the pbrent MockSyncStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *SyncStoreUpsertChbngesetEventsFunc) PushHook(hook func(context.Context, ...*types.ChbngesetEvent) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SyncStoreUpsertChbngesetEventsFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, ...*types.ChbngesetEvent) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SyncStoreUpsertChbngesetEventsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, ...*types.ChbngesetEvent) error {
		return r0
	})
}

func (f *SyncStoreUpsertChbngesetEventsFunc) nextHook() func(context.Context, ...*types.ChbngesetEvent) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SyncStoreUpsertChbngesetEventsFunc) bppendCbll(r0 SyncStoreUpsertChbngesetEventsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SyncStoreUpsertChbngesetEventsFuncCbll
// objects describing the invocbtions of this function.
func (f *SyncStoreUpsertChbngesetEventsFunc) History() []SyncStoreUpsertChbngesetEventsFuncCbll {
	f.mutex.Lock()
	history := mbke([]SyncStoreUpsertChbngesetEventsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SyncStoreUpsertChbngesetEventsFuncCbll is bn object thbt describes bn
// invocbtion of method UpsertChbngesetEvents on bn instbnce of
// MockSyncStore.
type SyncStoreUpsertChbngesetEventsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg1 []*types.ChbngesetEvent
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c SyncStoreUpsertChbngesetEventsFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg1 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SyncStoreUpsertChbngesetEventsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// SyncStoreUserCredentiblsFunc describes the behbvior when the
// UserCredentibls method of the pbrent MockSyncStore instbnce is invoked.
type SyncStoreUserCredentiblsFunc struct {
	defbultHook func() dbtbbbse.UserCredentiblsStore
	hooks       []func() dbtbbbse.UserCredentiblsStore
	history     []SyncStoreUserCredentiblsFuncCbll
	mutex       sync.Mutex
}

// UserCredentibls delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSyncStore) UserCredentibls() dbtbbbse.UserCredentiblsStore {
	r0 := m.UserCredentiblsFunc.nextHook()()
	m.UserCredentiblsFunc.bppendCbll(SyncStoreUserCredentiblsFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the UserCredentibls
// method of the pbrent MockSyncStore instbnce is invoked bnd the hook queue
// is empty.
func (f *SyncStoreUserCredentiblsFunc) SetDefbultHook(hook func() dbtbbbse.UserCredentiblsStore) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UserCredentibls method of the pbrent MockSyncStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *SyncStoreUserCredentiblsFunc) PushHook(hook func() dbtbbbse.UserCredentiblsStore) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SyncStoreUserCredentiblsFunc) SetDefbultReturn(r0 dbtbbbse.UserCredentiblsStore) {
	f.SetDefbultHook(func() dbtbbbse.UserCredentiblsStore {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SyncStoreUserCredentiblsFunc) PushReturn(r0 dbtbbbse.UserCredentiblsStore) {
	f.PushHook(func() dbtbbbse.UserCredentiblsStore {
		return r0
	})
}

func (f *SyncStoreUserCredentiblsFunc) nextHook() func() dbtbbbse.UserCredentiblsStore {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SyncStoreUserCredentiblsFunc) bppendCbll(r0 SyncStoreUserCredentiblsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SyncStoreUserCredentiblsFuncCbll objects
// describing the invocbtions of this function.
func (f *SyncStoreUserCredentiblsFunc) History() []SyncStoreUserCredentiblsFuncCbll {
	f.mutex.Lock()
	history := mbke([]SyncStoreUserCredentiblsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SyncStoreUserCredentiblsFuncCbll is bn object thbt describes bn
// invocbtion of method UserCredentibls on bn instbnce of MockSyncStore.
type SyncStoreUserCredentiblsFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 dbtbbbse.UserCredentiblsStore
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SyncStoreUserCredentiblsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SyncStoreUserCredentiblsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
