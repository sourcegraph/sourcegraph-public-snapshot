// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge mocks

import (
	"context"
	"io"
	"sync"
	"time"

	uplobdstore "github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore"
	iterbtor "github.com/sourcegrbph/sourcegrbph/lib/iterbtor"
)

// MockStore is b mock implementbtion of the Store interfbce (from the
// pbckbge github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore) used for
// unit testing.
type MockStore struct {
	// ComposeFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Compose.
	ComposeFunc *StoreComposeFunc
	// DeleteFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Delete.
	DeleteFunc *StoreDeleteFunc
	// ExpireObjectsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ExpireObjects.
	ExpireObjectsFunc *StoreExpireObjectsFunc
	// GetFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Get.
	GetFunc *StoreGetFunc
	// InitFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Init.
	InitFunc *StoreInitFunc
	// ListFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method List.
	ListFunc *StoreListFunc
	// UplobdFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Uplobd.
	UplobdFunc *StoreUplobdFunc
}

// NewMockStore crebtes b new mock of the Store interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockStore() *MockStore {
	return &MockStore{
		ComposeFunc: &StoreComposeFunc{
			defbultHook: func(context.Context, string, ...string) (r0 int64, r1 error) {
				return
			},
		},
		DeleteFunc: &StoreDeleteFunc{
			defbultHook: func(context.Context, string) (r0 error) {
				return
			},
		},
		ExpireObjectsFunc: &StoreExpireObjectsFunc{
			defbultHook: func(context.Context, string, time.Durbtion) (r0 error) {
				return
			},
		},
		GetFunc: &StoreGetFunc{
			defbultHook: func(context.Context, string) (r0 io.RebdCloser, r1 error) {
				return
			},
		},
		InitFunc: &StoreInitFunc{
			defbultHook: func(context.Context) (r0 error) {
				return
			},
		},
		ListFunc: &StoreListFunc{
			defbultHook: func(context.Context, string) (r0 *iterbtor.Iterbtor[string], r1 error) {
				return
			},
		},
		UplobdFunc: &StoreUplobdFunc{
			defbultHook: func(context.Context, string, io.Rebder) (r0 int64, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockStore crebtes b new mock of the Store interfbce. All methods
// pbnic on invocbtion, unless overwritten.
func NewStrictMockStore() *MockStore {
	return &MockStore{
		ComposeFunc: &StoreComposeFunc{
			defbultHook: func(context.Context, string, ...string) (int64, error) {
				pbnic("unexpected invocbtion of MockStore.Compose")
			},
		},
		DeleteFunc: &StoreDeleteFunc{
			defbultHook: func(context.Context, string) error {
				pbnic("unexpected invocbtion of MockStore.Delete")
			},
		},
		ExpireObjectsFunc: &StoreExpireObjectsFunc{
			defbultHook: func(context.Context, string, time.Durbtion) error {
				pbnic("unexpected invocbtion of MockStore.ExpireObjects")
			},
		},
		GetFunc: &StoreGetFunc{
			defbultHook: func(context.Context, string) (io.RebdCloser, error) {
				pbnic("unexpected invocbtion of MockStore.Get")
			},
		},
		InitFunc: &StoreInitFunc{
			defbultHook: func(context.Context) error {
				pbnic("unexpected invocbtion of MockStore.Init")
			},
		},
		ListFunc: &StoreListFunc{
			defbultHook: func(context.Context, string) (*iterbtor.Iterbtor[string], error) {
				pbnic("unexpected invocbtion of MockStore.List")
			},
		},
		UplobdFunc: &StoreUplobdFunc{
			defbultHook: func(context.Context, string, io.Rebder) (int64, error) {
				pbnic("unexpected invocbtion of MockStore.Uplobd")
			},
		},
	}
}

// NewMockStoreFrom crebtes b new mock of the MockStore interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockStoreFrom(i uplobdstore.Store) *MockStore {
	return &MockStore{
		ComposeFunc: &StoreComposeFunc{
			defbultHook: i.Compose,
		},
		DeleteFunc: &StoreDeleteFunc{
			defbultHook: i.Delete,
		},
		ExpireObjectsFunc: &StoreExpireObjectsFunc{
			defbultHook: i.ExpireObjects,
		},
		GetFunc: &StoreGetFunc{
			defbultHook: i.Get,
		},
		InitFunc: &StoreInitFunc{
			defbultHook: i.Init,
		},
		ListFunc: &StoreListFunc{
			defbultHook: i.List,
		},
		UplobdFunc: &StoreUplobdFunc{
			defbultHook: i.Uplobd,
		},
	}
}

// StoreComposeFunc describes the behbvior when the Compose method of the
// pbrent MockStore instbnce is invoked.
type StoreComposeFunc struct {
	defbultHook func(context.Context, string, ...string) (int64, error)
	hooks       []func(context.Context, string, ...string) (int64, error)
	history     []StoreComposeFuncCbll
	mutex       sync.Mutex
}

// Compose delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Compose(v0 context.Context, v1 string, v2 ...string) (int64, error) {
	r0, r1 := m.ComposeFunc.nextHook()(v0, v1, v2...)
	m.ComposeFunc.bppendCbll(StoreComposeFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Compose method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreComposeFunc) SetDefbultHook(hook func(context.Context, string, ...string) (int64, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Compose method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreComposeFunc) PushHook(hook func(context.Context, string, ...string) (int64, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreComposeFunc) SetDefbultReturn(r0 int64, r1 error) {
	f.SetDefbultHook(func(context.Context, string, ...string) (int64, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreComposeFunc) PushReturn(r0 int64, r1 error) {
	f.PushHook(func(context.Context, string, ...string) (int64, error) {
		return r0, r1
	})
}

func (f *StoreComposeFunc) nextHook() func(context.Context, string, ...string) (int64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreComposeFunc) bppendCbll(r0 StoreComposeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreComposeFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreComposeFunc) History() []StoreComposeFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreComposeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreComposeFuncCbll is bn object thbt describes bn invocbtion of method
// Compose on bn instbnce of MockStore.
type StoreComposeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg2 []string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int64
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c StoreComposeFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg2 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0, c.Arg1}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreComposeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreDeleteFunc describes the behbvior when the Delete method of the
// pbrent MockStore instbnce is invoked.
type StoreDeleteFunc struct {
	defbultHook func(context.Context, string) error
	hooks       []func(context.Context, string) error
	history     []StoreDeleteFuncCbll
	mutex       sync.Mutex
}

// Delete delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Delete(v0 context.Context, v1 string) error {
	r0 := m.DeleteFunc.nextHook()(v0, v1)
	m.DeleteFunc.bppendCbll(StoreDeleteFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Delete method of the
// pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreDeleteFunc) SetDefbultHook(hook func(context.Context, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Delete method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreDeleteFunc) PushHook(hook func(context.Context, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreDeleteFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreDeleteFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string) error {
		return r0
	})
}

func (f *StoreDeleteFunc) nextHook() func(context.Context, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDeleteFunc) bppendCbll(r0 StoreDeleteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreDeleteFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreDeleteFunc) History() []StoreDeleteFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreDeleteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDeleteFuncCbll is bn object thbt describes bn invocbtion of method
// Delete on bn instbnce of MockStore.
type StoreDeleteFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreDeleteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreDeleteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreExpireObjectsFunc describes the behbvior when the ExpireObjects
// method of the pbrent MockStore instbnce is invoked.
type StoreExpireObjectsFunc struct {
	defbultHook func(context.Context, string, time.Durbtion) error
	hooks       []func(context.Context, string, time.Durbtion) error
	history     []StoreExpireObjectsFuncCbll
	mutex       sync.Mutex
}

// ExpireObjects delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) ExpireObjects(v0 context.Context, v1 string, v2 time.Durbtion) error {
	r0 := m.ExpireObjectsFunc.nextHook()(v0, v1, v2)
	m.ExpireObjectsFunc.bppendCbll(StoreExpireObjectsFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the ExpireObjects method
// of the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreExpireObjectsFunc) SetDefbultHook(hook func(context.Context, string, time.Durbtion) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ExpireObjects method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreExpireObjectsFunc) PushHook(hook func(context.Context, string, time.Durbtion) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreExpireObjectsFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string, time.Durbtion) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreExpireObjectsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string, time.Durbtion) error {
		return r0
	})
}

func (f *StoreExpireObjectsFunc) nextHook() func(context.Context, string, time.Durbtion) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreExpireObjectsFunc) bppendCbll(r0 StoreExpireObjectsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreExpireObjectsFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreExpireObjectsFunc) History() []StoreExpireObjectsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreExpireObjectsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreExpireObjectsFuncCbll is bn object thbt describes bn invocbtion of
// method ExpireObjects on bn instbnce of MockStore.
type StoreExpireObjectsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 time.Durbtion
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreExpireObjectsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreExpireObjectsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreGetFunc describes the behbvior when the Get method of the pbrent
// MockStore instbnce is invoked.
type StoreGetFunc struct {
	defbultHook func(context.Context, string) (io.RebdCloser, error)
	hooks       []func(context.Context, string) (io.RebdCloser, error)
	history     []StoreGetFuncCbll
	mutex       sync.Mutex
}

// Get delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Get(v0 context.Context, v1 string) (io.RebdCloser, error) {
	r0, r1 := m.GetFunc.nextHook()(v0, v1)
	m.GetFunc.bppendCbll(StoreGetFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Get method of the
// pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreGetFunc) SetDefbultHook(hook func(context.Context, string) (io.RebdCloser, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Get method of the pbrent MockStore instbnce invokes the hook bt the front
// of the queue bnd discbrds it. After the queue is empty, the defbult hook
// function is invoked for bny future bction.
func (f *StoreGetFunc) PushHook(hook func(context.Context, string) (io.RebdCloser, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGetFunc) SetDefbultReturn(r0 io.RebdCloser, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (io.RebdCloser, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGetFunc) PushReturn(r0 io.RebdCloser, r1 error) {
	f.PushHook(func(context.Context, string) (io.RebdCloser, error) {
		return r0, r1
	})
}

func (f *StoreGetFunc) nextHook() func(context.Context, string) (io.RebdCloser, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetFunc) bppendCbll(r0 StoreGetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreGetFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreGetFunc) History() []StoreGetFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetFuncCbll is bn object thbt describes bn invocbtion of method Get
// on bn instbnce of MockStore.
type StoreGetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 io.RebdCloser
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreInitFunc describes the behbvior when the Init method of the pbrent
// MockStore instbnce is invoked.
type StoreInitFunc struct {
	defbultHook func(context.Context) error
	hooks       []func(context.Context) error
	history     []StoreInitFuncCbll
	mutex       sync.Mutex
}

// Init delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Init(v0 context.Context) error {
	r0 := m.InitFunc.nextHook()(v0)
	m.InitFunc.bppendCbll(StoreInitFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Init method of the
// pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreInitFunc) SetDefbultHook(hook func(context.Context) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Init method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreInitFunc) PushHook(hook func(context.Context) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreInitFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreInitFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context) error {
		return r0
	})
}

func (f *StoreInitFunc) nextHook() func(context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreInitFunc) bppendCbll(r0 StoreInitFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreInitFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreInitFunc) History() []StoreInitFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreInitFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreInitFuncCbll is bn object thbt describes bn invocbtion of method
// Init on bn instbnce of MockStore.
type StoreInitFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreInitFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreInitFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreListFunc describes the behbvior when the List method of the pbrent
// MockStore instbnce is invoked.
type StoreListFunc struct {
	defbultHook func(context.Context, string) (*iterbtor.Iterbtor[string], error)
	hooks       []func(context.Context, string) (*iterbtor.Iterbtor[string], error)
	history     []StoreListFuncCbll
	mutex       sync.Mutex
}

// List delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) List(v0 context.Context, v1 string) (*iterbtor.Iterbtor[string], error) {
	r0, r1 := m.ListFunc.nextHook()(v0, v1)
	m.ListFunc.bppendCbll(StoreListFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the List method of the
// pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreListFunc) SetDefbultHook(hook func(context.Context, string) (*iterbtor.Iterbtor[string], error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// List method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreListFunc) PushHook(hook func(context.Context, string) (*iterbtor.Iterbtor[string], error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreListFunc) SetDefbultReturn(r0 *iterbtor.Iterbtor[string], r1 error) {
	f.SetDefbultHook(func(context.Context, string) (*iterbtor.Iterbtor[string], error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreListFunc) PushReturn(r0 *iterbtor.Iterbtor[string], r1 error) {
	f.PushHook(func(context.Context, string) (*iterbtor.Iterbtor[string], error) {
		return r0, r1
	})
}

func (f *StoreListFunc) nextHook() func(context.Context, string) (*iterbtor.Iterbtor[string], error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreListFunc) bppendCbll(r0 StoreListFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreListFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreListFunc) History() []StoreListFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreListFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreListFuncCbll is bn object thbt describes bn invocbtion of method
// List on bn instbnce of MockStore.
type StoreListFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *iterbtor.Iterbtor[string]
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreListFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreListFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreUplobdFunc describes the behbvior when the Uplobd method of the
// pbrent MockStore instbnce is invoked.
type StoreUplobdFunc struct {
	defbultHook func(context.Context, string, io.Rebder) (int64, error)
	hooks       []func(context.Context, string, io.Rebder) (int64, error)
	history     []StoreUplobdFuncCbll
	mutex       sync.Mutex
}

// Uplobd delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Uplobd(v0 context.Context, v1 string, v2 io.Rebder) (int64, error) {
	r0, r1 := m.UplobdFunc.nextHook()(v0, v1, v2)
	m.UplobdFunc.bppendCbll(StoreUplobdFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Uplobd method of the
// pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreUplobdFunc) SetDefbultHook(hook func(context.Context, string, io.Rebder) (int64, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Uplobd method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreUplobdFunc) PushHook(hook func(context.Context, string, io.Rebder) (int64, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreUplobdFunc) SetDefbultReturn(r0 int64, r1 error) {
	f.SetDefbultHook(func(context.Context, string, io.Rebder) (int64, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreUplobdFunc) PushReturn(r0 int64, r1 error) {
	f.PushHook(func(context.Context, string, io.Rebder) (int64, error) {
		return r0, r1
	})
}

func (f *StoreUplobdFunc) nextHook() func(context.Context, string, io.Rebder) (int64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreUplobdFunc) bppendCbll(r0 StoreUplobdFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreUplobdFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreUplobdFunc) History() []StoreUplobdFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreUplobdFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreUplobdFuncCbll is bn object thbt describes bn invocbtion of method
// Uplobd on bn instbnce of MockStore.
type StoreUplobdFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 io.Rebder
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int64
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreUplobdFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreUplobdFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}
