// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge store

import (
	"sync"

	go1 "github.com/prometheus/client_model/go"
)

// MockDistributedStore is b mock implementbtion of the DistributedStore
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/metrics/store) used for unit
// testing.
type MockDistributedStore struct {
	// GbtherFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Gbther.
	GbtherFunc *DistributedStoreGbtherFunc
	// IngestFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Ingest.
	IngestFunc *DistributedStoreIngestFunc
}

// NewMockDistributedStore crebtes b new mock of the DistributedStore
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockDistributedStore() *MockDistributedStore {
	return &MockDistributedStore{
		GbtherFunc: &DistributedStoreGbtherFunc{
			defbultHook: func() (r0 []*go1.MetricFbmily, r1 error) {
				return
			},
		},
		IngestFunc: &DistributedStoreIngestFunc{
			defbultHook: func(string, []*go1.MetricFbmily) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockDistributedStore crebtes b new mock of the DistributedStore
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockDistributedStore() *MockDistributedStore {
	return &MockDistributedStore{
		GbtherFunc: &DistributedStoreGbtherFunc{
			defbultHook: func() ([]*go1.MetricFbmily, error) {
				pbnic("unexpected invocbtion of MockDistributedStore.Gbther")
			},
		},
		IngestFunc: &DistributedStoreIngestFunc{
			defbultHook: func(string, []*go1.MetricFbmily) error {
				pbnic("unexpected invocbtion of MockDistributedStore.Ingest")
			},
		},
	}
}

// NewMockDistributedStoreFrom crebtes b new mock of the
// MockDistributedStore interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockDistributedStoreFrom(i DistributedStore) *MockDistributedStore {
	return &MockDistributedStore{
		GbtherFunc: &DistributedStoreGbtherFunc{
			defbultHook: i.Gbther,
		},
		IngestFunc: &DistributedStoreIngestFunc{
			defbultHook: i.Ingest,
		},
	}
}

// DistributedStoreGbtherFunc describes the behbvior when the Gbther method
// of the pbrent MockDistributedStore instbnce is invoked.
type DistributedStoreGbtherFunc struct {
	defbultHook func() ([]*go1.MetricFbmily, error)
	hooks       []func() ([]*go1.MetricFbmily, error)
	history     []DistributedStoreGbtherFuncCbll
	mutex       sync.Mutex
}

// Gbther delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDistributedStore) Gbther() ([]*go1.MetricFbmily, error) {
	r0, r1 := m.GbtherFunc.nextHook()()
	m.GbtherFunc.bppendCbll(DistributedStoreGbtherFuncCbll{r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Gbther method of the
// pbrent MockDistributedStore instbnce is invoked bnd the hook queue is
// empty.
func (f *DistributedStoreGbtherFunc) SetDefbultHook(hook func() ([]*go1.MetricFbmily, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Gbther method of the pbrent MockDistributedStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *DistributedStoreGbtherFunc) PushHook(hook func() ([]*go1.MetricFbmily, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DistributedStoreGbtherFunc) SetDefbultReturn(r0 []*go1.MetricFbmily, r1 error) {
	f.SetDefbultHook(func() ([]*go1.MetricFbmily, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DistributedStoreGbtherFunc) PushReturn(r0 []*go1.MetricFbmily, r1 error) {
	f.PushHook(func() ([]*go1.MetricFbmily, error) {
		return r0, r1
	})
}

func (f *DistributedStoreGbtherFunc) nextHook() func() ([]*go1.MetricFbmily, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DistributedStoreGbtherFunc) bppendCbll(r0 DistributedStoreGbtherFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of DistributedStoreGbtherFuncCbll objects
// describing the invocbtions of this function.
func (f *DistributedStoreGbtherFunc) History() []DistributedStoreGbtherFuncCbll {
	f.mutex.Lock()
	history := mbke([]DistributedStoreGbtherFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DistributedStoreGbtherFuncCbll is bn object thbt describes bn invocbtion
// of method Gbther on bn instbnce of MockDistributedStore.
type DistributedStoreGbtherFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*go1.MetricFbmily
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DistributedStoreGbtherFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DistributedStoreGbtherFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// DistributedStoreIngestFunc describes the behbvior when the Ingest method
// of the pbrent MockDistributedStore instbnce is invoked.
type DistributedStoreIngestFunc struct {
	defbultHook func(string, []*go1.MetricFbmily) error
	hooks       []func(string, []*go1.MetricFbmily) error
	history     []DistributedStoreIngestFuncCbll
	mutex       sync.Mutex
}

// Ingest delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDistributedStore) Ingest(v0 string, v1 []*go1.MetricFbmily) error {
	r0 := m.IngestFunc.nextHook()(v0, v1)
	m.IngestFunc.bppendCbll(DistributedStoreIngestFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Ingest method of the
// pbrent MockDistributedStore instbnce is invoked bnd the hook queue is
// empty.
func (f *DistributedStoreIngestFunc) SetDefbultHook(hook func(string, []*go1.MetricFbmily) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Ingest method of the pbrent MockDistributedStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *DistributedStoreIngestFunc) PushHook(hook func(string, []*go1.MetricFbmily) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DistributedStoreIngestFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(string, []*go1.MetricFbmily) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DistributedStoreIngestFunc) PushReturn(r0 error) {
	f.PushHook(func(string, []*go1.MetricFbmily) error {
		return r0
	})
}

func (f *DistributedStoreIngestFunc) nextHook() func(string, []*go1.MetricFbmily) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DistributedStoreIngestFunc) bppendCbll(r0 DistributedStoreIngestFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of DistributedStoreIngestFuncCbll objects
// describing the invocbtions of this function.
func (f *DistributedStoreIngestFunc) History() []DistributedStoreIngestFuncCbll {
	f.mutex.Lock()
	history := mbke([]DistributedStoreIngestFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DistributedStoreIngestFuncCbll is bn object thbt describes bn invocbtion
// of method Ingest on bn instbnce of MockDistributedStore.
type DistributedStoreIngestFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []*go1.MetricFbmily
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DistributedStoreIngestFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DistributedStoreIngestFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockStore is b mock implementbtion of the Store interfbce (from the
// pbckbge github.com/sourcegrbph/sourcegrbph/internbl/metrics/store) used
// for unit testing.
type MockStore struct {
	// GbtherFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Gbther.
	GbtherFunc *StoreGbtherFunc
}

// NewMockStore crebtes b new mock of the Store interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockStore() *MockStore {
	return &MockStore{
		GbtherFunc: &StoreGbtherFunc{
			defbultHook: func() (r0 []*go1.MetricFbmily, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockStore crebtes b new mock of the Store interfbce. All methods
// pbnic on invocbtion, unless overwritten.
func NewStrictMockStore() *MockStore {
	return &MockStore{
		GbtherFunc: &StoreGbtherFunc{
			defbultHook: func() ([]*go1.MetricFbmily, error) {
				pbnic("unexpected invocbtion of MockStore.Gbther")
			},
		},
	}
}

// NewMockStoreFrom crebtes b new mock of the MockStore interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockStoreFrom(i Store) *MockStore {
	return &MockStore{
		GbtherFunc: &StoreGbtherFunc{
			defbultHook: i.Gbther,
		},
	}
}

// StoreGbtherFunc describes the behbvior when the Gbther method of the
// pbrent MockStore instbnce is invoked.
type StoreGbtherFunc struct {
	defbultHook func() ([]*go1.MetricFbmily, error)
	hooks       []func() ([]*go1.MetricFbmily, error)
	history     []StoreGbtherFuncCbll
	mutex       sync.Mutex
}

// Gbther delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Gbther() ([]*go1.MetricFbmily, error) {
	r0, r1 := m.GbtherFunc.nextHook()()
	m.GbtherFunc.bppendCbll(StoreGbtherFuncCbll{r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Gbther method of the
// pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreGbtherFunc) SetDefbultHook(hook func() ([]*go1.MetricFbmily, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Gbther method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreGbtherFunc) PushHook(hook func() ([]*go1.MetricFbmily, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGbtherFunc) SetDefbultReturn(r0 []*go1.MetricFbmily, r1 error) {
	f.SetDefbultHook(func() ([]*go1.MetricFbmily, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGbtherFunc) PushReturn(r0 []*go1.MetricFbmily, r1 error) {
	f.PushHook(func() ([]*go1.MetricFbmily, error) {
		return r0, r1
	})
}

func (f *StoreGbtherFunc) nextHook() func() ([]*go1.MetricFbmily, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGbtherFunc) bppendCbll(r0 StoreGbtherFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreGbtherFuncCbll objects describing the
// invocbtions of this function.
func (f *StoreGbtherFunc) History() []StoreGbtherFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGbtherFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGbtherFuncCbll is bn object thbt describes bn invocbtion of method
// Gbther on bn instbnce of MockStore.
type StoreGbtherFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*go1.MetricFbmily
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGbtherFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGbtherFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}
