// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge policies

import (
	"context"
	"sync"

	store "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/internbl/store"
	shbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/shbred"
)

// MockStore is b mock implementbtion of the Store interfbce (from the
// pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies/internbl/store)
// used for unit testing.
type MockStore struct {
	// CrebteConfigurbtionPolicyFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// CrebteConfigurbtionPolicy.
	CrebteConfigurbtionPolicyFunc *StoreCrebteConfigurbtionPolicyFunc
	// DeleteConfigurbtionPolicyByIDFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// DeleteConfigurbtionPolicyByID.
	DeleteConfigurbtionPolicyByIDFunc *StoreDeleteConfigurbtionPolicyByIDFunc
	// GetConfigurbtionPoliciesFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetConfigurbtionPolicies.
	GetConfigurbtionPoliciesFunc *StoreGetConfigurbtionPoliciesFunc
	// GetConfigurbtionPolicyByIDFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// GetConfigurbtionPolicyByID.
	GetConfigurbtionPolicyByIDFunc *StoreGetConfigurbtionPolicyByIDFunc
	// GetRepoIDsByGlobPbtternsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetRepoIDsByGlobPbtterns.
	GetRepoIDsByGlobPbtternsFunc *StoreGetRepoIDsByGlobPbtternsFunc
	// RepoCountFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method RepoCount.
	RepoCountFunc *StoreRepoCountFunc
	// SelectPoliciesForRepositoryMembershipUpdbteFunc is bn instbnce of b
	// mock function object controlling the behbvior of the method
	// SelectPoliciesForRepositoryMembershipUpdbte.
	SelectPoliciesForRepositoryMembershipUpdbteFunc *StoreSelectPoliciesForRepositoryMembershipUpdbteFunc
	// UpdbteConfigurbtionPolicyFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// UpdbteConfigurbtionPolicy.
	UpdbteConfigurbtionPolicyFunc *StoreUpdbteConfigurbtionPolicyFunc
	// UpdbteReposMbtchingPbtternsFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// UpdbteReposMbtchingPbtterns.
	UpdbteReposMbtchingPbtternsFunc *StoreUpdbteReposMbtchingPbtternsFunc
}

// NewMockStore crebtes b new mock of the Store interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockStore() *MockStore {
	return &MockStore{
		CrebteConfigurbtionPolicyFunc: &StoreCrebteConfigurbtionPolicyFunc{
			defbultHook: func(context.Context, shbred.ConfigurbtionPolicy) (r0 shbred.ConfigurbtionPolicy, r1 error) {
				return
			},
		},
		DeleteConfigurbtionPolicyByIDFunc: &StoreDeleteConfigurbtionPolicyByIDFunc{
			defbultHook: func(context.Context, int) (r0 error) {
				return
			},
		},
		GetConfigurbtionPoliciesFunc: &StoreGetConfigurbtionPoliciesFunc{
			defbultHook: func(context.Context, shbred.GetConfigurbtionPoliciesOptions) (r0 []shbred.ConfigurbtionPolicy, r1 int, r2 error) {
				return
			},
		},
		GetConfigurbtionPolicyByIDFunc: &StoreGetConfigurbtionPolicyByIDFunc{
			defbultHook: func(context.Context, int) (r0 shbred.ConfigurbtionPolicy, r1 bool, r2 error) {
				return
			},
		},
		GetRepoIDsByGlobPbtternsFunc: &StoreGetRepoIDsByGlobPbtternsFunc{
			defbultHook: func(context.Context, []string, int, int) (r0 []int, r1 int, r2 error) {
				return
			},
		},
		RepoCountFunc: &StoreRepoCountFunc{
			defbultHook: func(context.Context) (r0 int, r1 error) {
				return
			},
		},
		SelectPoliciesForRepositoryMembershipUpdbteFunc: &StoreSelectPoliciesForRepositoryMembershipUpdbteFunc{
			defbultHook: func(context.Context, int) (r0 []shbred.ConfigurbtionPolicy, r1 error) {
				return
			},
		},
		UpdbteConfigurbtionPolicyFunc: &StoreUpdbteConfigurbtionPolicyFunc{
			defbultHook: func(context.Context, shbred.ConfigurbtionPolicy) (r0 error) {
				return
			},
		},
		UpdbteReposMbtchingPbtternsFunc: &StoreUpdbteReposMbtchingPbtternsFunc{
			defbultHook: func(context.Context, []string, int, *int) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockStore crebtes b new mock of the Store interfbce. All methods
// pbnic on invocbtion, unless overwritten.
func NewStrictMockStore() *MockStore {
	return &MockStore{
		CrebteConfigurbtionPolicyFunc: &StoreCrebteConfigurbtionPolicyFunc{
			defbultHook: func(context.Context, shbred.ConfigurbtionPolicy) (shbred.ConfigurbtionPolicy, error) {
				pbnic("unexpected invocbtion of MockStore.CrebteConfigurbtionPolicy")
			},
		},
		DeleteConfigurbtionPolicyByIDFunc: &StoreDeleteConfigurbtionPolicyByIDFunc{
			defbultHook: func(context.Context, int) error {
				pbnic("unexpected invocbtion of MockStore.DeleteConfigurbtionPolicyByID")
			},
		},
		GetConfigurbtionPoliciesFunc: &StoreGetConfigurbtionPoliciesFunc{
			defbultHook: func(context.Context, shbred.GetConfigurbtionPoliciesOptions) ([]shbred.ConfigurbtionPolicy, int, error) {
				pbnic("unexpected invocbtion of MockStore.GetConfigurbtionPolicies")
			},
		},
		GetConfigurbtionPolicyByIDFunc: &StoreGetConfigurbtionPolicyByIDFunc{
			defbultHook: func(context.Context, int) (shbred.ConfigurbtionPolicy, bool, error) {
				pbnic("unexpected invocbtion of MockStore.GetConfigurbtionPolicyByID")
			},
		},
		GetRepoIDsByGlobPbtternsFunc: &StoreGetRepoIDsByGlobPbtternsFunc{
			defbultHook: func(context.Context, []string, int, int) ([]int, int, error) {
				pbnic("unexpected invocbtion of MockStore.GetRepoIDsByGlobPbtterns")
			},
		},
		RepoCountFunc: &StoreRepoCountFunc{
			defbultHook: func(context.Context) (int, error) {
				pbnic("unexpected invocbtion of MockStore.RepoCount")
			},
		},
		SelectPoliciesForRepositoryMembershipUpdbteFunc: &StoreSelectPoliciesForRepositoryMembershipUpdbteFunc{
			defbultHook: func(context.Context, int) ([]shbred.ConfigurbtionPolicy, error) {
				pbnic("unexpected invocbtion of MockStore.SelectPoliciesForRepositoryMembershipUpdbte")
			},
		},
		UpdbteConfigurbtionPolicyFunc: &StoreUpdbteConfigurbtionPolicyFunc{
			defbultHook: func(context.Context, shbred.ConfigurbtionPolicy) error {
				pbnic("unexpected invocbtion of MockStore.UpdbteConfigurbtionPolicy")
			},
		},
		UpdbteReposMbtchingPbtternsFunc: &StoreUpdbteReposMbtchingPbtternsFunc{
			defbultHook: func(context.Context, []string, int, *int) error {
				pbnic("unexpected invocbtion of MockStore.UpdbteReposMbtchingPbtterns")
			},
		},
	}
}

// NewMockStoreFrom crebtes b new mock of the MockStore interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockStoreFrom(i store.Store) *MockStore {
	return &MockStore{
		CrebteConfigurbtionPolicyFunc: &StoreCrebteConfigurbtionPolicyFunc{
			defbultHook: i.CrebteConfigurbtionPolicy,
		},
		DeleteConfigurbtionPolicyByIDFunc: &StoreDeleteConfigurbtionPolicyByIDFunc{
			defbultHook: i.DeleteConfigurbtionPolicyByID,
		},
		GetConfigurbtionPoliciesFunc: &StoreGetConfigurbtionPoliciesFunc{
			defbultHook: i.GetConfigurbtionPolicies,
		},
		GetConfigurbtionPolicyByIDFunc: &StoreGetConfigurbtionPolicyByIDFunc{
			defbultHook: i.GetConfigurbtionPolicyByID,
		},
		GetRepoIDsByGlobPbtternsFunc: &StoreGetRepoIDsByGlobPbtternsFunc{
			defbultHook: i.GetRepoIDsByGlobPbtterns,
		},
		RepoCountFunc: &StoreRepoCountFunc{
			defbultHook: i.RepoCount,
		},
		SelectPoliciesForRepositoryMembershipUpdbteFunc: &StoreSelectPoliciesForRepositoryMembershipUpdbteFunc{
			defbultHook: i.SelectPoliciesForRepositoryMembershipUpdbte,
		},
		UpdbteConfigurbtionPolicyFunc: &StoreUpdbteConfigurbtionPolicyFunc{
			defbultHook: i.UpdbteConfigurbtionPolicy,
		},
		UpdbteReposMbtchingPbtternsFunc: &StoreUpdbteReposMbtchingPbtternsFunc{
			defbultHook: i.UpdbteReposMbtchingPbtterns,
		},
	}
}

// StoreCrebteConfigurbtionPolicyFunc describes the behbvior when the
// CrebteConfigurbtionPolicy method of the pbrent MockStore instbnce is
// invoked.
type StoreCrebteConfigurbtionPolicyFunc struct {
	defbultHook func(context.Context, shbred.ConfigurbtionPolicy) (shbred.ConfigurbtionPolicy, error)
	hooks       []func(context.Context, shbred.ConfigurbtionPolicy) (shbred.ConfigurbtionPolicy, error)
	history     []StoreCrebteConfigurbtionPolicyFuncCbll
	mutex       sync.Mutex
}

// CrebteConfigurbtionPolicy delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) CrebteConfigurbtionPolicy(v0 context.Context, v1 shbred.ConfigurbtionPolicy) (shbred.ConfigurbtionPolicy, error) {
	r0, r1 := m.CrebteConfigurbtionPolicyFunc.nextHook()(v0, v1)
	m.CrebteConfigurbtionPolicyFunc.bppendCbll(StoreCrebteConfigurbtionPolicyFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// CrebteConfigurbtionPolicy method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreCrebteConfigurbtionPolicyFunc) SetDefbultHook(hook func(context.Context, shbred.ConfigurbtionPolicy) (shbred.ConfigurbtionPolicy, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CrebteConfigurbtionPolicy method of the pbrent MockStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *StoreCrebteConfigurbtionPolicyFunc) PushHook(hook func(context.Context, shbred.ConfigurbtionPolicy) (shbred.ConfigurbtionPolicy, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreCrebteConfigurbtionPolicyFunc) SetDefbultReturn(r0 shbred.ConfigurbtionPolicy, r1 error) {
	f.SetDefbultHook(func(context.Context, shbred.ConfigurbtionPolicy) (shbred.ConfigurbtionPolicy, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreCrebteConfigurbtionPolicyFunc) PushReturn(r0 shbred.ConfigurbtionPolicy, r1 error) {
	f.PushHook(func(context.Context, shbred.ConfigurbtionPolicy) (shbred.ConfigurbtionPolicy, error) {
		return r0, r1
	})
}

func (f *StoreCrebteConfigurbtionPolicyFunc) nextHook() func(context.Context, shbred.ConfigurbtionPolicy) (shbred.ConfigurbtionPolicy, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreCrebteConfigurbtionPolicyFunc) bppendCbll(r0 StoreCrebteConfigurbtionPolicyFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreCrebteConfigurbtionPolicyFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreCrebteConfigurbtionPolicyFunc) History() []StoreCrebteConfigurbtionPolicyFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreCrebteConfigurbtionPolicyFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreCrebteConfigurbtionPolicyFuncCbll is bn object thbt describes bn
// invocbtion of method CrebteConfigurbtionPolicy on bn instbnce of
// MockStore.
type StoreCrebteConfigurbtionPolicyFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 shbred.ConfigurbtionPolicy
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 shbred.ConfigurbtionPolicy
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreCrebteConfigurbtionPolicyFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreCrebteConfigurbtionPolicyFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreDeleteConfigurbtionPolicyByIDFunc describes the behbvior when the
// DeleteConfigurbtionPolicyByID method of the pbrent MockStore instbnce is
// invoked.
type StoreDeleteConfigurbtionPolicyByIDFunc struct {
	defbultHook func(context.Context, int) error
	hooks       []func(context.Context, int) error
	history     []StoreDeleteConfigurbtionPolicyByIDFuncCbll
	mutex       sync.Mutex
}

// DeleteConfigurbtionPolicyByID delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) DeleteConfigurbtionPolicyByID(v0 context.Context, v1 int) error {
	r0 := m.DeleteConfigurbtionPolicyByIDFunc.nextHook()(v0, v1)
	m.DeleteConfigurbtionPolicyByIDFunc.bppendCbll(StoreDeleteConfigurbtionPolicyByIDFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// DeleteConfigurbtionPolicyByID method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreDeleteConfigurbtionPolicyByIDFunc) SetDefbultHook(hook func(context.Context, int) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DeleteConfigurbtionPolicyByID method of the pbrent MockStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *StoreDeleteConfigurbtionPolicyByIDFunc) PushHook(hook func(context.Context, int) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreDeleteConfigurbtionPolicyByIDFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreDeleteConfigurbtionPolicyByIDFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int) error {
		return r0
	})
}

func (f *StoreDeleteConfigurbtionPolicyByIDFunc) nextHook() func(context.Context, int) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDeleteConfigurbtionPolicyByIDFunc) bppendCbll(r0 StoreDeleteConfigurbtionPolicyByIDFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreDeleteConfigurbtionPolicyByIDFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreDeleteConfigurbtionPolicyByIDFunc) History() []StoreDeleteConfigurbtionPolicyByIDFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreDeleteConfigurbtionPolicyByIDFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDeleteConfigurbtionPolicyByIDFuncCbll is bn object thbt describes bn
// invocbtion of method DeleteConfigurbtionPolicyByID on bn instbnce of
// MockStore.
type StoreDeleteConfigurbtionPolicyByIDFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreDeleteConfigurbtionPolicyByIDFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreDeleteConfigurbtionPolicyByIDFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreGetConfigurbtionPoliciesFunc describes the behbvior when the
// GetConfigurbtionPolicies method of the pbrent MockStore instbnce is
// invoked.
type StoreGetConfigurbtionPoliciesFunc struct {
	defbultHook func(context.Context, shbred.GetConfigurbtionPoliciesOptions) ([]shbred.ConfigurbtionPolicy, int, error)
	hooks       []func(context.Context, shbred.GetConfigurbtionPoliciesOptions) ([]shbred.ConfigurbtionPolicy, int, error)
	history     []StoreGetConfigurbtionPoliciesFuncCbll
	mutex       sync.Mutex
}

// GetConfigurbtionPolicies delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) GetConfigurbtionPolicies(v0 context.Context, v1 shbred.GetConfigurbtionPoliciesOptions) ([]shbred.ConfigurbtionPolicy, int, error) {
	r0, r1, r2 := m.GetConfigurbtionPoliciesFunc.nextHook()(v0, v1)
	m.GetConfigurbtionPoliciesFunc.bppendCbll(StoreGetConfigurbtionPoliciesFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// GetConfigurbtionPolicies method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreGetConfigurbtionPoliciesFunc) SetDefbultHook(hook func(context.Context, shbred.GetConfigurbtionPoliciesOptions) ([]shbred.ConfigurbtionPolicy, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetConfigurbtionPolicies method of the pbrent MockStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *StoreGetConfigurbtionPoliciesFunc) PushHook(hook func(context.Context, shbred.GetConfigurbtionPoliciesOptions) ([]shbred.ConfigurbtionPolicy, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGetConfigurbtionPoliciesFunc) SetDefbultReturn(r0 []shbred.ConfigurbtionPolicy, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, shbred.GetConfigurbtionPoliciesOptions) ([]shbred.ConfigurbtionPolicy, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGetConfigurbtionPoliciesFunc) PushReturn(r0 []shbred.ConfigurbtionPolicy, r1 int, r2 error) {
	f.PushHook(func(context.Context, shbred.GetConfigurbtionPoliciesOptions) ([]shbred.ConfigurbtionPolicy, int, error) {
		return r0, r1, r2
	})
}

func (f *StoreGetConfigurbtionPoliciesFunc) nextHook() func(context.Context, shbred.GetConfigurbtionPoliciesOptions) ([]shbred.ConfigurbtionPolicy, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetConfigurbtionPoliciesFunc) bppendCbll(r0 StoreGetConfigurbtionPoliciesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreGetConfigurbtionPoliciesFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreGetConfigurbtionPoliciesFunc) History() []StoreGetConfigurbtionPoliciesFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGetConfigurbtionPoliciesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetConfigurbtionPoliciesFuncCbll is bn object thbt describes bn
// invocbtion of method GetConfigurbtionPolicies on bn instbnce of
// MockStore.
type StoreGetConfigurbtionPoliciesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 shbred.GetConfigurbtionPoliciesOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.ConfigurbtionPolicy
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGetConfigurbtionPoliciesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGetConfigurbtionPoliciesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// StoreGetConfigurbtionPolicyByIDFunc describes the behbvior when the
// GetConfigurbtionPolicyByID method of the pbrent MockStore instbnce is
// invoked.
type StoreGetConfigurbtionPolicyByIDFunc struct {
	defbultHook func(context.Context, int) (shbred.ConfigurbtionPolicy, bool, error)
	hooks       []func(context.Context, int) (shbred.ConfigurbtionPolicy, bool, error)
	history     []StoreGetConfigurbtionPolicyByIDFuncCbll
	mutex       sync.Mutex
}

// GetConfigurbtionPolicyByID delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) GetConfigurbtionPolicyByID(v0 context.Context, v1 int) (shbred.ConfigurbtionPolicy, bool, error) {
	r0, r1, r2 := m.GetConfigurbtionPolicyByIDFunc.nextHook()(v0, v1)
	m.GetConfigurbtionPolicyByIDFunc.bppendCbll(StoreGetConfigurbtionPolicyByIDFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// GetConfigurbtionPolicyByID method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreGetConfigurbtionPolicyByIDFunc) SetDefbultHook(hook func(context.Context, int) (shbred.ConfigurbtionPolicy, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetConfigurbtionPolicyByID method of the pbrent MockStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *StoreGetConfigurbtionPolicyByIDFunc) PushHook(hook func(context.Context, int) (shbred.ConfigurbtionPolicy, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGetConfigurbtionPolicyByIDFunc) SetDefbultReturn(r0 shbred.ConfigurbtionPolicy, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, int) (shbred.ConfigurbtionPolicy, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGetConfigurbtionPolicyByIDFunc) PushReturn(r0 shbred.ConfigurbtionPolicy, r1 bool, r2 error) {
	f.PushHook(func(context.Context, int) (shbred.ConfigurbtionPolicy, bool, error) {
		return r0, r1, r2
	})
}

func (f *StoreGetConfigurbtionPolicyByIDFunc) nextHook() func(context.Context, int) (shbred.ConfigurbtionPolicy, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetConfigurbtionPolicyByIDFunc) bppendCbll(r0 StoreGetConfigurbtionPolicyByIDFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreGetConfigurbtionPolicyByIDFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreGetConfigurbtionPolicyByIDFunc) History() []StoreGetConfigurbtionPolicyByIDFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGetConfigurbtionPolicyByIDFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetConfigurbtionPolicyByIDFuncCbll is bn object thbt describes bn
// invocbtion of method GetConfigurbtionPolicyByID on bn instbnce of
// MockStore.
type StoreGetConfigurbtionPolicyByIDFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 shbred.ConfigurbtionPolicy
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGetConfigurbtionPolicyByIDFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGetConfigurbtionPolicyByIDFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// StoreGetRepoIDsByGlobPbtternsFunc describes the behbvior when the
// GetRepoIDsByGlobPbtterns method of the pbrent MockStore instbnce is
// invoked.
type StoreGetRepoIDsByGlobPbtternsFunc struct {
	defbultHook func(context.Context, []string, int, int) ([]int, int, error)
	hooks       []func(context.Context, []string, int, int) ([]int, int, error)
	history     []StoreGetRepoIDsByGlobPbtternsFuncCbll
	mutex       sync.Mutex
}

// GetRepoIDsByGlobPbtterns delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) GetRepoIDsByGlobPbtterns(v0 context.Context, v1 []string, v2 int, v3 int) ([]int, int, error) {
	r0, r1, r2 := m.GetRepoIDsByGlobPbtternsFunc.nextHook()(v0, v1, v2, v3)
	m.GetRepoIDsByGlobPbtternsFunc.bppendCbll(StoreGetRepoIDsByGlobPbtternsFuncCbll{v0, v1, v2, v3, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// GetRepoIDsByGlobPbtterns method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreGetRepoIDsByGlobPbtternsFunc) SetDefbultHook(hook func(context.Context, []string, int, int) ([]int, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetRepoIDsByGlobPbtterns method of the pbrent MockStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *StoreGetRepoIDsByGlobPbtternsFunc) PushHook(hook func(context.Context, []string, int, int) ([]int, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGetRepoIDsByGlobPbtternsFunc) SetDefbultReturn(r0 []int, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, []string, int, int) ([]int, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGetRepoIDsByGlobPbtternsFunc) PushReturn(r0 []int, r1 int, r2 error) {
	f.PushHook(func(context.Context, []string, int, int) ([]int, int, error) {
		return r0, r1, r2
	})
}

func (f *StoreGetRepoIDsByGlobPbtternsFunc) nextHook() func(context.Context, []string, int, int) ([]int, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetRepoIDsByGlobPbtternsFunc) bppendCbll(r0 StoreGetRepoIDsByGlobPbtternsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreGetRepoIDsByGlobPbtternsFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreGetRepoIDsByGlobPbtternsFunc) History() []StoreGetRepoIDsByGlobPbtternsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGetRepoIDsByGlobPbtternsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetRepoIDsByGlobPbtternsFuncCbll is bn object thbt describes bn
// invocbtion of method GetRepoIDsByGlobPbtterns on bn instbnce of
// MockStore.
type StoreGetRepoIDsByGlobPbtternsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGetRepoIDsByGlobPbtternsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGetRepoIDsByGlobPbtternsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// StoreRepoCountFunc describes the behbvior when the RepoCount method of
// the pbrent MockStore instbnce is invoked.
type StoreRepoCountFunc struct {
	defbultHook func(context.Context) (int, error)
	hooks       []func(context.Context) (int, error)
	history     []StoreRepoCountFuncCbll
	mutex       sync.Mutex
}

// RepoCount delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) RepoCount(v0 context.Context) (int, error) {
	r0, r1 := m.RepoCountFunc.nextHook()(v0)
	m.RepoCountFunc.bppendCbll(StoreRepoCountFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the RepoCount method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreRepoCountFunc) SetDefbultHook(hook func(context.Context) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RepoCount method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreRepoCountFunc) PushHook(hook func(context.Context) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreRepoCountFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreRepoCountFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context) (int, error) {
		return r0, r1
	})
}

func (f *StoreRepoCountFunc) nextHook() func(context.Context) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreRepoCountFunc) bppendCbll(r0 StoreRepoCountFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreRepoCountFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreRepoCountFunc) History() []StoreRepoCountFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreRepoCountFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreRepoCountFuncCbll is bn object thbt describes bn invocbtion of
// method RepoCount on bn instbnce of MockStore.
type StoreRepoCountFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreRepoCountFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreRepoCountFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreSelectPoliciesForRepositoryMembershipUpdbteFunc describes the
// behbvior when the SelectPoliciesForRepositoryMembershipUpdbte method of
// the pbrent MockStore instbnce is invoked.
type StoreSelectPoliciesForRepositoryMembershipUpdbteFunc struct {
	defbultHook func(context.Context, int) ([]shbred.ConfigurbtionPolicy, error)
	hooks       []func(context.Context, int) ([]shbred.ConfigurbtionPolicy, error)
	history     []StoreSelectPoliciesForRepositoryMembershipUpdbteFuncCbll
	mutex       sync.Mutex
}

// SelectPoliciesForRepositoryMembershipUpdbte delegbtes to the next hook
// function in the queue bnd stores the pbrbmeter bnd result vblues of this
// invocbtion.
func (m *MockStore) SelectPoliciesForRepositoryMembershipUpdbte(v0 context.Context, v1 int) ([]shbred.ConfigurbtionPolicy, error) {
	r0, r1 := m.SelectPoliciesForRepositoryMembershipUpdbteFunc.nextHook()(v0, v1)
	m.SelectPoliciesForRepositoryMembershipUpdbteFunc.bppendCbll(StoreSelectPoliciesForRepositoryMembershipUpdbteFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// SelectPoliciesForRepositoryMembershipUpdbte method of the pbrent
// MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreSelectPoliciesForRepositoryMembershipUpdbteFunc) SetDefbultHook(hook func(context.Context, int) ([]shbred.ConfigurbtionPolicy, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SelectPoliciesForRepositoryMembershipUpdbte method of the pbrent
// MockStore instbnce invokes the hook bt the front of the queue bnd
// discbrds it. After the queue is empty, the defbult hook function is
// invoked for bny future bction.
func (f *StoreSelectPoliciesForRepositoryMembershipUpdbteFunc) PushHook(hook func(context.Context, int) ([]shbred.ConfigurbtionPolicy, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreSelectPoliciesForRepositoryMembershipUpdbteFunc) SetDefbultReturn(r0 []shbred.ConfigurbtionPolicy, r1 error) {
	f.SetDefbultHook(func(context.Context, int) ([]shbred.ConfigurbtionPolicy, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreSelectPoliciesForRepositoryMembershipUpdbteFunc) PushReturn(r0 []shbred.ConfigurbtionPolicy, r1 error) {
	f.PushHook(func(context.Context, int) ([]shbred.ConfigurbtionPolicy, error) {
		return r0, r1
	})
}

func (f *StoreSelectPoliciesForRepositoryMembershipUpdbteFunc) nextHook() func(context.Context, int) ([]shbred.ConfigurbtionPolicy, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreSelectPoliciesForRepositoryMembershipUpdbteFunc) bppendCbll(r0 StoreSelectPoliciesForRepositoryMembershipUpdbteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// StoreSelectPoliciesForRepositoryMembershipUpdbteFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreSelectPoliciesForRepositoryMembershipUpdbteFunc) History() []StoreSelectPoliciesForRepositoryMembershipUpdbteFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreSelectPoliciesForRepositoryMembershipUpdbteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreSelectPoliciesForRepositoryMembershipUpdbteFuncCbll is bn object
// thbt describes bn invocbtion of method
// SelectPoliciesForRepositoryMembershipUpdbte on bn instbnce of MockStore.
type StoreSelectPoliciesForRepositoryMembershipUpdbteFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.ConfigurbtionPolicy
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreSelectPoliciesForRepositoryMembershipUpdbteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreSelectPoliciesForRepositoryMembershipUpdbteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreUpdbteConfigurbtionPolicyFunc describes the behbvior when the
// UpdbteConfigurbtionPolicy method of the pbrent MockStore instbnce is
// invoked.
type StoreUpdbteConfigurbtionPolicyFunc struct {
	defbultHook func(context.Context, shbred.ConfigurbtionPolicy) error
	hooks       []func(context.Context, shbred.ConfigurbtionPolicy) error
	history     []StoreUpdbteConfigurbtionPolicyFuncCbll
	mutex       sync.Mutex
}

// UpdbteConfigurbtionPolicy delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) UpdbteConfigurbtionPolicy(v0 context.Context, v1 shbred.ConfigurbtionPolicy) error {
	r0 := m.UpdbteConfigurbtionPolicyFunc.nextHook()(v0, v1)
	m.UpdbteConfigurbtionPolicyFunc.bppendCbll(StoreUpdbteConfigurbtionPolicyFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// UpdbteConfigurbtionPolicy method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreUpdbteConfigurbtionPolicyFunc) SetDefbultHook(hook func(context.Context, shbred.ConfigurbtionPolicy) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpdbteConfigurbtionPolicy method of the pbrent MockStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *StoreUpdbteConfigurbtionPolicyFunc) PushHook(hook func(context.Context, shbred.ConfigurbtionPolicy) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreUpdbteConfigurbtionPolicyFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, shbred.ConfigurbtionPolicy) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreUpdbteConfigurbtionPolicyFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, shbred.ConfigurbtionPolicy) error {
		return r0
	})
}

func (f *StoreUpdbteConfigurbtionPolicyFunc) nextHook() func(context.Context, shbred.ConfigurbtionPolicy) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreUpdbteConfigurbtionPolicyFunc) bppendCbll(r0 StoreUpdbteConfigurbtionPolicyFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreUpdbteConfigurbtionPolicyFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreUpdbteConfigurbtionPolicyFunc) History() []StoreUpdbteConfigurbtionPolicyFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreUpdbteConfigurbtionPolicyFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreUpdbteConfigurbtionPolicyFuncCbll is bn object thbt describes bn
// invocbtion of method UpdbteConfigurbtionPolicy on bn instbnce of
// MockStore.
type StoreUpdbteConfigurbtionPolicyFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 shbred.ConfigurbtionPolicy
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreUpdbteConfigurbtionPolicyFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreUpdbteConfigurbtionPolicyFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreUpdbteReposMbtchingPbtternsFunc describes the behbvior when the
// UpdbteReposMbtchingPbtterns method of the pbrent MockStore instbnce is
// invoked.
type StoreUpdbteReposMbtchingPbtternsFunc struct {
	defbultHook func(context.Context, []string, int, *int) error
	hooks       []func(context.Context, []string, int, *int) error
	history     []StoreUpdbteReposMbtchingPbtternsFuncCbll
	mutex       sync.Mutex
}

// UpdbteReposMbtchingPbtterns delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) UpdbteReposMbtchingPbtterns(v0 context.Context, v1 []string, v2 int, v3 *int) error {
	r0 := m.UpdbteReposMbtchingPbtternsFunc.nextHook()(v0, v1, v2, v3)
	m.UpdbteReposMbtchingPbtternsFunc.bppendCbll(StoreUpdbteReposMbtchingPbtternsFuncCbll{v0, v1, v2, v3, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// UpdbteReposMbtchingPbtterns method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreUpdbteReposMbtchingPbtternsFunc) SetDefbultHook(hook func(context.Context, []string, int, *int) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpdbteReposMbtchingPbtterns method of the pbrent MockStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *StoreUpdbteReposMbtchingPbtternsFunc) PushHook(hook func(context.Context, []string, int, *int) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreUpdbteReposMbtchingPbtternsFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, []string, int, *int) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreUpdbteReposMbtchingPbtternsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, []string, int, *int) error {
		return r0
	})
}

func (f *StoreUpdbteReposMbtchingPbtternsFunc) nextHook() func(context.Context, []string, int, *int) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreUpdbteReposMbtchingPbtternsFunc) bppendCbll(r0 StoreUpdbteReposMbtchingPbtternsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreUpdbteReposMbtchingPbtternsFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreUpdbteReposMbtchingPbtternsFunc) History() []StoreUpdbteReposMbtchingPbtternsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreUpdbteReposMbtchingPbtternsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreUpdbteReposMbtchingPbtternsFuncCbll is bn object thbt describes bn
// invocbtion of method UpdbteReposMbtchingPbtterns on bn instbnce of
// MockStore.
type StoreUpdbteReposMbtchingPbtternsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 *int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreUpdbteReposMbtchingPbtternsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreUpdbteReposMbtchingPbtternsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockUplobdService is b mock implementbtion of the UplobdService interfbce
// (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/policies) used for
// unit testing.
type MockUplobdService struct {
	// GetCommitsVisibleToUplobdFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// GetCommitsVisibleToUplobd.
	GetCommitsVisibleToUplobdFunc *UplobdServiceGetCommitsVisibleToUplobdFunc
}

// NewMockUplobdService crebtes b new mock of the UplobdService interfbce.
// All methods return zero vblues for bll results, unless overwritten.
func NewMockUplobdService() *MockUplobdService {
	return &MockUplobdService{
		GetCommitsVisibleToUplobdFunc: &UplobdServiceGetCommitsVisibleToUplobdFunc{
			defbultHook: func(context.Context, int, int, *string) (r0 []string, r1 *string, r2 error) {
				return
			},
		},
	}
}

// NewStrictMockUplobdService crebtes b new mock of the UplobdService
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockUplobdService() *MockUplobdService {
	return &MockUplobdService{
		GetCommitsVisibleToUplobdFunc: &UplobdServiceGetCommitsVisibleToUplobdFunc{
			defbultHook: func(context.Context, int, int, *string) ([]string, *string, error) {
				pbnic("unexpected invocbtion of MockUplobdService.GetCommitsVisibleToUplobd")
			},
		},
	}
}

// NewMockUplobdServiceFrom crebtes b new mock of the MockUplobdService
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockUplobdServiceFrom(i UplobdService) *MockUplobdService {
	return &MockUplobdService{
		GetCommitsVisibleToUplobdFunc: &UplobdServiceGetCommitsVisibleToUplobdFunc{
			defbultHook: i.GetCommitsVisibleToUplobd,
		},
	}
}

// UplobdServiceGetCommitsVisibleToUplobdFunc describes the behbvior when
// the GetCommitsVisibleToUplobd method of the pbrent MockUplobdService
// instbnce is invoked.
type UplobdServiceGetCommitsVisibleToUplobdFunc struct {
	defbultHook func(context.Context, int, int, *string) ([]string, *string, error)
	hooks       []func(context.Context, int, int, *string) ([]string, *string, error)
	history     []UplobdServiceGetCommitsVisibleToUplobdFuncCbll
	mutex       sync.Mutex
}

// GetCommitsVisibleToUplobd delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdService) GetCommitsVisibleToUplobd(v0 context.Context, v1 int, v2 int, v3 *string) ([]string, *string, error) {
	r0, r1, r2 := m.GetCommitsVisibleToUplobdFunc.nextHook()(v0, v1, v2, v3)
	m.GetCommitsVisibleToUplobdFunc.bppendCbll(UplobdServiceGetCommitsVisibleToUplobdFuncCbll{v0, v1, v2, v3, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// GetCommitsVisibleToUplobd method of the pbrent MockUplobdService instbnce
// is invoked bnd the hook queue is empty.
func (f *UplobdServiceGetCommitsVisibleToUplobdFunc) SetDefbultHook(hook func(context.Context, int, int, *string) ([]string, *string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetCommitsVisibleToUplobd method of the pbrent MockUplobdService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *UplobdServiceGetCommitsVisibleToUplobdFunc) PushHook(hook func(context.Context, int, int, *string) ([]string, *string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdServiceGetCommitsVisibleToUplobdFunc) SetDefbultReturn(r0 []string, r1 *string, r2 error) {
	f.SetDefbultHook(func(context.Context, int, int, *string) ([]string, *string, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdServiceGetCommitsVisibleToUplobdFunc) PushReturn(r0 []string, r1 *string, r2 error) {
	f.PushHook(func(context.Context, int, int, *string) ([]string, *string, error) {
		return r0, r1, r2
	})
}

func (f *UplobdServiceGetCommitsVisibleToUplobdFunc) nextHook() func(context.Context, int, int, *string) ([]string, *string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdServiceGetCommitsVisibleToUplobdFunc) bppendCbll(r0 UplobdServiceGetCommitsVisibleToUplobdFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// UplobdServiceGetCommitsVisibleToUplobdFuncCbll objects describing the
// invocbtions of this function.
func (f *UplobdServiceGetCommitsVisibleToUplobdFunc) History() []UplobdServiceGetCommitsVisibleToUplobdFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdServiceGetCommitsVisibleToUplobdFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdServiceGetCommitsVisibleToUplobdFuncCbll is bn object thbt
// describes bn invocbtion of method GetCommitsVisibleToUplobd on bn
// instbnce of MockUplobdService.
type UplobdServiceGetCommitsVisibleToUplobdFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 *string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 *string
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdServiceGetCommitsVisibleToUplobdFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdServiceGetCommitsVisibleToUplobdFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}
