// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge runner

import (
	"context"
	"sync"

	definition "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
	schembs "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/schembs"
	shbred "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/shbred"
)

// MockStore is b mock implementbtion of the Store interfbce (from the
// pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/runner)
// used for unit testing.
type MockStore struct {
	// DescribeFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Describe.
	DescribeFunc *StoreDescribeFunc
	// DoneFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Done.
	DoneFunc *StoreDoneFunc
	// DownFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Down.
	DownFunc *StoreDownFunc
	// IndexStbtusFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method IndexStbtus.
	IndexStbtusFunc *StoreIndexStbtusFunc
	// RunDDLStbtementsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method RunDDLStbtements.
	RunDDLStbtementsFunc *StoreRunDDLStbtementsFunc
	// TrbnsbctFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Trbnsbct.
	TrbnsbctFunc *StoreTrbnsbctFunc
	// TryLockFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method TryLock.
	TryLockFunc *StoreTryLockFunc
	// UpFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Up.
	UpFunc *StoreUpFunc
	// VersionsFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Versions.
	VersionsFunc *StoreVersionsFunc
	// WithMigrbtionLogFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method WithMigrbtionLog.
	WithMigrbtionLogFunc *StoreWithMigrbtionLogFunc
}

// NewMockStore crebtes b new mock of the Store interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockStore() *MockStore {
	return &MockStore{
		DescribeFunc: &StoreDescribeFunc{
			defbultHook: func(context.Context) (r0 mbp[string]schembs.SchembDescription, r1 error) {
				return
			},
		},
		DoneFunc: &StoreDoneFunc{
			defbultHook: func(error) (r0 error) {
				return
			},
		},
		DownFunc: &StoreDownFunc{
			defbultHook: func(context.Context, definition.Definition) (r0 error) {
				return
			},
		},
		IndexStbtusFunc: &StoreIndexStbtusFunc{
			defbultHook: func(context.Context, string, string) (r0 shbred.IndexStbtus, r1 bool, r2 error) {
				return
			},
		},
		RunDDLStbtementsFunc: &StoreRunDDLStbtementsFunc{
			defbultHook: func(context.Context, []string) (r0 error) {
				return
			},
		},
		TrbnsbctFunc: &StoreTrbnsbctFunc{
			defbultHook: func(context.Context) (r0 Store, r1 error) {
				return
			},
		},
		TryLockFunc: &StoreTryLockFunc{
			defbultHook: func(context.Context) (r0 bool, r1 func(err error) error, r2 error) {
				return
			},
		},
		UpFunc: &StoreUpFunc{
			defbultHook: func(context.Context, definition.Definition) (r0 error) {
				return
			},
		},
		VersionsFunc: &StoreVersionsFunc{
			defbultHook: func(context.Context) (r0 []int, r1 []int, r2 []int, r3 error) {
				return
			},
		},
		WithMigrbtionLogFunc: &StoreWithMigrbtionLogFunc{
			defbultHook: func(context.Context, definition.Definition, bool, func() error) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockStore crebtes b new mock of the Store interfbce. All methods
// pbnic on invocbtion, unless overwritten.
func NewStrictMockStore() *MockStore {
	return &MockStore{
		DescribeFunc: &StoreDescribeFunc{
			defbultHook: func(context.Context) (mbp[string]schembs.SchembDescription, error) {
				pbnic("unexpected invocbtion of MockStore.Describe")
			},
		},
		DoneFunc: &StoreDoneFunc{
			defbultHook: func(error) error {
				pbnic("unexpected invocbtion of MockStore.Done")
			},
		},
		DownFunc: &StoreDownFunc{
			defbultHook: func(context.Context, definition.Definition) error {
				pbnic("unexpected invocbtion of MockStore.Down")
			},
		},
		IndexStbtusFunc: &StoreIndexStbtusFunc{
			defbultHook: func(context.Context, string, string) (shbred.IndexStbtus, bool, error) {
				pbnic("unexpected invocbtion of MockStore.IndexStbtus")
			},
		},
		RunDDLStbtementsFunc: &StoreRunDDLStbtementsFunc{
			defbultHook: func(context.Context, []string) error {
				pbnic("unexpected invocbtion of MockStore.RunDDLStbtements")
			},
		},
		TrbnsbctFunc: &StoreTrbnsbctFunc{
			defbultHook: func(context.Context) (Store, error) {
				pbnic("unexpected invocbtion of MockStore.Trbnsbct")
			},
		},
		TryLockFunc: &StoreTryLockFunc{
			defbultHook: func(context.Context) (bool, func(err error) error, error) {
				pbnic("unexpected invocbtion of MockStore.TryLock")
			},
		},
		UpFunc: &StoreUpFunc{
			defbultHook: func(context.Context, definition.Definition) error {
				pbnic("unexpected invocbtion of MockStore.Up")
			},
		},
		VersionsFunc: &StoreVersionsFunc{
			defbultHook: func(context.Context) ([]int, []int, []int, error) {
				pbnic("unexpected invocbtion of MockStore.Versions")
			},
		},
		WithMigrbtionLogFunc: &StoreWithMigrbtionLogFunc{
			defbultHook: func(context.Context, definition.Definition, bool, func() error) error {
				pbnic("unexpected invocbtion of MockStore.WithMigrbtionLog")
			},
		},
	}
}

// NewMockStoreFrom crebtes b new mock of the MockStore interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockStoreFrom(i Store) *MockStore {
	return &MockStore{
		DescribeFunc: &StoreDescribeFunc{
			defbultHook: i.Describe,
		},
		DoneFunc: &StoreDoneFunc{
			defbultHook: i.Done,
		},
		DownFunc: &StoreDownFunc{
			defbultHook: i.Down,
		},
		IndexStbtusFunc: &StoreIndexStbtusFunc{
			defbultHook: i.IndexStbtus,
		},
		RunDDLStbtementsFunc: &StoreRunDDLStbtementsFunc{
			defbultHook: i.RunDDLStbtements,
		},
		TrbnsbctFunc: &StoreTrbnsbctFunc{
			defbultHook: i.Trbnsbct,
		},
		TryLockFunc: &StoreTryLockFunc{
			defbultHook: i.TryLock,
		},
		UpFunc: &StoreUpFunc{
			defbultHook: i.Up,
		},
		VersionsFunc: &StoreVersionsFunc{
			defbultHook: i.Versions,
		},
		WithMigrbtionLogFunc: &StoreWithMigrbtionLogFunc{
			defbultHook: i.WithMigrbtionLog,
		},
	}
}

// StoreDescribeFunc describes the behbvior when the Describe method of the
// pbrent MockStore instbnce is invoked.
type StoreDescribeFunc struct {
	defbultHook func(context.Context) (mbp[string]schembs.SchembDescription, error)
	hooks       []func(context.Context) (mbp[string]schembs.SchembDescription, error)
	history     []StoreDescribeFuncCbll
	mutex       sync.Mutex
}

// Describe delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Describe(v0 context.Context) (mbp[string]schembs.SchembDescription, error) {
	r0, r1 := m.DescribeFunc.nextHook()(v0)
	m.DescribeFunc.bppendCbll(StoreDescribeFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Describe method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreDescribeFunc) SetDefbultHook(hook func(context.Context) (mbp[string]schembs.SchembDescription, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Describe method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreDescribeFunc) PushHook(hook func(context.Context) (mbp[string]schembs.SchembDescription, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreDescribeFunc) SetDefbultReturn(r0 mbp[string]schembs.SchembDescription, r1 error) {
	f.SetDefbultHook(func(context.Context) (mbp[string]schembs.SchembDescription, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreDescribeFunc) PushReturn(r0 mbp[string]schembs.SchembDescription, r1 error) {
	f.PushHook(func(context.Context) (mbp[string]schembs.SchembDescription, error) {
		return r0, r1
	})
}

func (f *StoreDescribeFunc) nextHook() func(context.Context) (mbp[string]schembs.SchembDescription, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDescribeFunc) bppendCbll(r0 StoreDescribeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreDescribeFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreDescribeFunc) History() []StoreDescribeFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreDescribeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDescribeFuncCbll is bn object thbt describes bn invocbtion of method
// Describe on bn instbnce of MockStore.
type StoreDescribeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 mbp[string]schembs.SchembDescription
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreDescribeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreDescribeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreDoneFunc describes the behbvior when the Done method of the pbrent
// MockStore instbnce is invoked.
type StoreDoneFunc struct {
	defbultHook func(error) error
	hooks       []func(error) error
	history     []StoreDoneFuncCbll
	mutex       sync.Mutex
}

// Done delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Done(v0 error) error {
	r0 := m.DoneFunc.nextHook()(v0)
	m.DoneFunc.bppendCbll(StoreDoneFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Done method of the
// pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreDoneFunc) SetDefbultHook(hook func(error) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Done method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreDoneFunc) PushHook(hook func(error) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreDoneFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(error) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreDoneFunc) PushReturn(r0 error) {
	f.PushHook(func(error) error {
		return r0
	})
}

func (f *StoreDoneFunc) nextHook() func(error) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDoneFunc) bppendCbll(r0 StoreDoneFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreDoneFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreDoneFunc) History() []StoreDoneFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreDoneFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDoneFuncCbll is bn object thbt describes bn invocbtion of method
// Done on bn instbnce of MockStore.
type StoreDoneFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 error
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreDoneFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreDoneFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreDownFunc describes the behbvior when the Down method of the pbrent
// MockStore instbnce is invoked.
type StoreDownFunc struct {
	defbultHook func(context.Context, definition.Definition) error
	hooks       []func(context.Context, definition.Definition) error
	history     []StoreDownFuncCbll
	mutex       sync.Mutex
}

// Down delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Down(v0 context.Context, v1 definition.Definition) error {
	r0 := m.DownFunc.nextHook()(v0, v1)
	m.DownFunc.bppendCbll(StoreDownFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Down method of the
// pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreDownFunc) SetDefbultHook(hook func(context.Context, definition.Definition) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Down method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreDownFunc) PushHook(hook func(context.Context, definition.Definition) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreDownFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, definition.Definition) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreDownFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, definition.Definition) error {
		return r0
	})
}

func (f *StoreDownFunc) nextHook() func(context.Context, definition.Definition) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDownFunc) bppendCbll(r0 StoreDownFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreDownFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreDownFunc) History() []StoreDownFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreDownFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDownFuncCbll is bn object thbt describes bn invocbtion of method
// Down on bn instbnce of MockStore.
type StoreDownFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 definition.Definition
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreDownFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreDownFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreIndexStbtusFunc describes the behbvior when the IndexStbtus method
// of the pbrent MockStore instbnce is invoked.
type StoreIndexStbtusFunc struct {
	defbultHook func(context.Context, string, string) (shbred.IndexStbtus, bool, error)
	hooks       []func(context.Context, string, string) (shbred.IndexStbtus, bool, error)
	history     []StoreIndexStbtusFuncCbll
	mutex       sync.Mutex
}

// IndexStbtus delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) IndexStbtus(v0 context.Context, v1 string, v2 string) (shbred.IndexStbtus, bool, error) {
	r0, r1, r2 := m.IndexStbtusFunc.nextHook()(v0, v1, v2)
	m.IndexStbtusFunc.bppendCbll(StoreIndexStbtusFuncCbll{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the IndexStbtus method
// of the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreIndexStbtusFunc) SetDefbultHook(hook func(context.Context, string, string) (shbred.IndexStbtus, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// IndexStbtus method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreIndexStbtusFunc) PushHook(hook func(context.Context, string, string) (shbred.IndexStbtus, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreIndexStbtusFunc) SetDefbultReturn(r0 shbred.IndexStbtus, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, string, string) (shbred.IndexStbtus, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreIndexStbtusFunc) PushReturn(r0 shbred.IndexStbtus, r1 bool, r2 error) {
	f.PushHook(func(context.Context, string, string) (shbred.IndexStbtus, bool, error) {
		return r0, r1, r2
	})
}

func (f *StoreIndexStbtusFunc) nextHook() func(context.Context, string, string) (shbred.IndexStbtus, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreIndexStbtusFunc) bppendCbll(r0 StoreIndexStbtusFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreIndexStbtusFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreIndexStbtusFunc) History() []StoreIndexStbtusFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreIndexStbtusFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreIndexStbtusFuncCbll is bn object thbt describes bn invocbtion of
// method IndexStbtus on bn instbnce of MockStore.
type StoreIndexStbtusFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 shbred.IndexStbtus
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreIndexStbtusFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreIndexStbtusFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// StoreRunDDLStbtementsFunc describes the behbvior when the
// RunDDLStbtements method of the pbrent MockStore instbnce is invoked.
type StoreRunDDLStbtementsFunc struct {
	defbultHook func(context.Context, []string) error
	hooks       []func(context.Context, []string) error
	history     []StoreRunDDLStbtementsFuncCbll
	mutex       sync.Mutex
}

// RunDDLStbtements delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) RunDDLStbtements(v0 context.Context, v1 []string) error {
	r0 := m.RunDDLStbtementsFunc.nextHook()(v0, v1)
	m.RunDDLStbtementsFunc.bppendCbll(StoreRunDDLStbtementsFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the RunDDLStbtements
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreRunDDLStbtementsFunc) SetDefbultHook(hook func(context.Context, []string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RunDDLStbtements method of the pbrent MockStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreRunDDLStbtementsFunc) PushHook(hook func(context.Context, []string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreRunDDLStbtementsFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, []string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreRunDDLStbtementsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, []string) error {
		return r0
	})
}

func (f *StoreRunDDLStbtementsFunc) nextHook() func(context.Context, []string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreRunDDLStbtementsFunc) bppendCbll(r0 StoreRunDDLStbtementsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreRunDDLStbtementsFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreRunDDLStbtementsFunc) History() []StoreRunDDLStbtementsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreRunDDLStbtementsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreRunDDLStbtementsFuncCbll is bn object thbt describes bn invocbtion
// of method RunDDLStbtements on bn instbnce of MockStore.
type StoreRunDDLStbtementsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreRunDDLStbtementsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreRunDDLStbtementsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreTrbnsbctFunc describes the behbvior when the Trbnsbct method of the
// pbrent MockStore instbnce is invoked.
type StoreTrbnsbctFunc struct {
	defbultHook func(context.Context) (Store, error)
	hooks       []func(context.Context) (Store, error)
	history     []StoreTrbnsbctFuncCbll
	mutex       sync.Mutex
}

// Trbnsbct delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Trbnsbct(v0 context.Context) (Store, error) {
	r0, r1 := m.TrbnsbctFunc.nextHook()(v0)
	m.TrbnsbctFunc.bppendCbll(StoreTrbnsbctFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Trbnsbct method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreTrbnsbctFunc) SetDefbultHook(hook func(context.Context) (Store, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Trbnsbct method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreTrbnsbctFunc) PushHook(hook func(context.Context) (Store, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreTrbnsbctFunc) SetDefbultReturn(r0 Store, r1 error) {
	f.SetDefbultHook(func(context.Context) (Store, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreTrbnsbctFunc) PushReturn(r0 Store, r1 error) {
	f.PushHook(func(context.Context) (Store, error) {
		return r0, r1
	})
}

func (f *StoreTrbnsbctFunc) nextHook() func(context.Context) (Store, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreTrbnsbctFunc) bppendCbll(r0 StoreTrbnsbctFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreTrbnsbctFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreTrbnsbctFunc) History() []StoreTrbnsbctFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreTrbnsbctFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreTrbnsbctFuncCbll is bn object thbt describes bn invocbtion of method
// Trbnsbct on bn instbnce of MockStore.
type StoreTrbnsbctFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 Store
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreTrbnsbctFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreTrbnsbctFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreTryLockFunc describes the behbvior when the TryLock method of the
// pbrent MockStore instbnce is invoked.
type StoreTryLockFunc struct {
	defbultHook func(context.Context) (bool, func(err error) error, error)
	hooks       []func(context.Context) (bool, func(err error) error, error)
	history     []StoreTryLockFuncCbll
	mutex       sync.Mutex
}

// TryLock delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) TryLock(v0 context.Context) (bool, func(err error) error, error) {
	r0, r1, r2 := m.TryLockFunc.nextHook()(v0)
	m.TryLockFunc.bppendCbll(StoreTryLockFuncCbll{v0, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the TryLock method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreTryLockFunc) SetDefbultHook(hook func(context.Context) (bool, func(err error) error, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// TryLock method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreTryLockFunc) PushHook(hook func(context.Context) (bool, func(err error) error, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreTryLockFunc) SetDefbultReturn(r0 bool, r1 func(err error) error, r2 error) {
	f.SetDefbultHook(func(context.Context) (bool, func(err error) error, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreTryLockFunc) PushReturn(r0 bool, r1 func(err error) error, r2 error) {
	f.PushHook(func(context.Context) (bool, func(err error) error, error) {
		return r0, r1, r2
	})
}

func (f *StoreTryLockFunc) nextHook() func(context.Context) (bool, func(err error) error, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreTryLockFunc) bppendCbll(r0 StoreTryLockFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreTryLockFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreTryLockFunc) History() []StoreTryLockFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreTryLockFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreTryLockFuncCbll is bn object thbt describes bn invocbtion of method
// TryLock on bn instbnce of MockStore.
type StoreTryLockFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 func(err error) error
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreTryLockFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreTryLockFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// StoreUpFunc describes the behbvior when the Up method of the pbrent
// MockStore instbnce is invoked.
type StoreUpFunc struct {
	defbultHook func(context.Context, definition.Definition) error
	hooks       []func(context.Context, definition.Definition) error
	history     []StoreUpFuncCbll
	mutex       sync.Mutex
}

// Up delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Up(v0 context.Context, v1 definition.Definition) error {
	r0 := m.UpFunc.nextHook()(v0, v1)
	m.UpFunc.bppendCbll(StoreUpFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Up method of the
// pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreUpFunc) SetDefbultHook(hook func(context.Context, definition.Definition) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Up method of the pbrent MockStore instbnce invokes the hook bt the front
// of the queue bnd discbrds it. After the queue is empty, the defbult hook
// function is invoked for bny future bction.
func (f *StoreUpFunc) PushHook(hook func(context.Context, definition.Definition) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreUpFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, definition.Definition) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreUpFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, definition.Definition) error {
		return r0
	})
}

func (f *StoreUpFunc) nextHook() func(context.Context, definition.Definition) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreUpFunc) bppendCbll(r0 StoreUpFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreUpFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreUpFunc) History() []StoreUpFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreUpFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreUpFuncCbll is bn object thbt describes bn invocbtion of method Up on
// bn instbnce of MockStore.
type StoreUpFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 definition.Definition
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreUpFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreUpFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreVersionsFunc describes the behbvior when the Versions method of the
// pbrent MockStore instbnce is invoked.
type StoreVersionsFunc struct {
	defbultHook func(context.Context) ([]int, []int, []int, error)
	hooks       []func(context.Context) ([]int, []int, []int, error)
	history     []StoreVersionsFuncCbll
	mutex       sync.Mutex
}

// Versions delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Versions(v0 context.Context) ([]int, []int, []int, error) {
	r0, r1, r2, r3 := m.VersionsFunc.nextHook()(v0)
	m.VersionsFunc.bppendCbll(StoreVersionsFuncCbll{v0, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefbultHook sets function thbt is cblled when the Versions method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreVersionsFunc) SetDefbultHook(hook func(context.Context) ([]int, []int, []int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Versions method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreVersionsFunc) PushHook(hook func(context.Context) ([]int, []int, []int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreVersionsFunc) SetDefbultReturn(r0 []int, r1 []int, r2 []int, r3 error) {
	f.SetDefbultHook(func(context.Context) ([]int, []int, []int, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreVersionsFunc) PushReturn(r0 []int, r1 []int, r2 []int, r3 error) {
	f.PushHook(func(context.Context) ([]int, []int, []int, error) {
		return r0, r1, r2, r3
	})
}

func (f *StoreVersionsFunc) nextHook() func(context.Context) ([]int, []int, []int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreVersionsFunc) bppendCbll(r0 StoreVersionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreVersionsFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreVersionsFunc) History() []StoreVersionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreVersionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreVersionsFuncCbll is bn object thbt describes bn invocbtion of method
// Versions on bn instbnce of MockStore.
type StoreVersionsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 []int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 []int
	// Result3 is the vblue of the 4th result returned from this method
	// invocbtion.
	Result3 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreVersionsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreVersionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2, c.Result3}
}

// StoreWithMigrbtionLogFunc describes the behbvior when the
// WithMigrbtionLog method of the pbrent MockStore instbnce is invoked.
type StoreWithMigrbtionLogFunc struct {
	defbultHook func(context.Context, definition.Definition, bool, func() error) error
	hooks       []func(context.Context, definition.Definition, bool, func() error) error
	history     []StoreWithMigrbtionLogFuncCbll
	mutex       sync.Mutex
}

// WithMigrbtionLog delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) WithMigrbtionLog(v0 context.Context, v1 definition.Definition, v2 bool, v3 func() error) error {
	r0 := m.WithMigrbtionLogFunc.nextHook()(v0, v1, v2, v3)
	m.WithMigrbtionLogFunc.bppendCbll(StoreWithMigrbtionLogFuncCbll{v0, v1, v2, v3, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the WithMigrbtionLog
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreWithMigrbtionLogFunc) SetDefbultHook(hook func(context.Context, definition.Definition, bool, func() error) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// WithMigrbtionLog method of the pbrent MockStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreWithMigrbtionLogFunc) PushHook(hook func(context.Context, definition.Definition, bool, func() error) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreWithMigrbtionLogFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, definition.Definition, bool, func() error) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreWithMigrbtionLogFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, definition.Definition, bool, func() error) error {
		return r0
	})
}

func (f *StoreWithMigrbtionLogFunc) nextHook() func(context.Context, definition.Definition, bool, func() error) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreWithMigrbtionLogFunc) bppendCbll(r0 StoreWithMigrbtionLogFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreWithMigrbtionLogFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreWithMigrbtionLogFunc) History() []StoreWithMigrbtionLogFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreWithMigrbtionLogFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreWithMigrbtionLogFuncCbll is bn object thbt describes bn invocbtion
// of method WithMigrbtionLog on bn instbnce of MockStore.
type StoreWithMigrbtionLogFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 definition.Definition
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bool
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 func() error
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreWithMigrbtionLogFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreWithMigrbtionLogFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
