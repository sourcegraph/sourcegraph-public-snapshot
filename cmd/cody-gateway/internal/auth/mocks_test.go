// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge buth

import (
	"sync"

	httpcbche "github.com/gregjones/httpcbche"
)

// MockCbche is b mock implementbtion of the Cbche interfbce (from the
// pbckbge github.com/gregjones/httpcbche) used for unit testing.
type MockCbche struct {
	// DeleteFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Delete.
	DeleteFunc *CbcheDeleteFunc
	// GetFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Get.
	GetFunc *CbcheGetFunc
	// SetFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Set.
	SetFunc *CbcheSetFunc
}

// NewMockCbche crebtes b new mock of the Cbche interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockCbche() *MockCbche {
	return &MockCbche{
		DeleteFunc: &CbcheDeleteFunc{
			defbultHook: func(string) {
				return
			},
		},
		GetFunc: &CbcheGetFunc{
			defbultHook: func(string) (r0 []byte, r1 bool) {
				return
			},
		},
		SetFunc: &CbcheSetFunc{
			defbultHook: func(string, []byte) {
				return
			},
		},
	}
}

// NewStrictMockCbche crebtes b new mock of the Cbche interfbce. All methods
// pbnic on invocbtion, unless overwritten.
func NewStrictMockCbche() *MockCbche {
	return &MockCbche{
		DeleteFunc: &CbcheDeleteFunc{
			defbultHook: func(string) {
				pbnic("unexpected invocbtion of MockCbche.Delete")
			},
		},
		GetFunc: &CbcheGetFunc{
			defbultHook: func(string) ([]byte, bool) {
				pbnic("unexpected invocbtion of MockCbche.Get")
			},
		},
		SetFunc: &CbcheSetFunc{
			defbultHook: func(string, []byte) {
				pbnic("unexpected invocbtion of MockCbche.Set")
			},
		},
	}
}

// NewMockCbcheFrom crebtes b new mock of the MockCbche interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockCbcheFrom(i httpcbche.Cbche) *MockCbche {
	return &MockCbche{
		DeleteFunc: &CbcheDeleteFunc{
			defbultHook: i.Delete,
		},
		GetFunc: &CbcheGetFunc{
			defbultHook: i.Get,
		},
		SetFunc: &CbcheSetFunc{
			defbultHook: i.Set,
		},
	}
}

// CbcheDeleteFunc describes the behbvior when the Delete method of the
// pbrent MockCbche instbnce is invoked.
type CbcheDeleteFunc struct {
	defbultHook func(string)
	hooks       []func(string)
	history     []CbcheDeleteFuncCbll
	mutex       sync.Mutex
}

// Delete delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCbche) Delete(v0 string) {
	m.DeleteFunc.nextHook()(v0)
	m.DeleteFunc.bppendCbll(CbcheDeleteFuncCbll{v0})
	return
}

// SetDefbultHook sets function thbt is cblled when the Delete method of the
// pbrent MockCbche instbnce is invoked bnd the hook queue is empty.
func (f *CbcheDeleteFunc) SetDefbultHook(hook func(string)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Delete method of the pbrent MockCbche instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *CbcheDeleteFunc) PushHook(hook func(string)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CbcheDeleteFunc) SetDefbultReturn() {
	f.SetDefbultHook(func(string) {
		return
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CbcheDeleteFunc) PushReturn() {
	f.PushHook(func(string) {
		return
	})
}

func (f *CbcheDeleteFunc) nextHook() func(string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CbcheDeleteFunc) bppendCbll(r0 CbcheDeleteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of CbcheDeleteFuncCbll objects describing the
// invocbtions of this function.
func (f *CbcheDeleteFunc) History() []CbcheDeleteFuncCbll {
	f.mutex.Lock()
	history := mbke([]CbcheDeleteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CbcheDeleteFuncCbll is bn object thbt describes bn invocbtion of method
// Delete on bn instbnce of MockCbche.
type CbcheDeleteFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c CbcheDeleteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CbcheDeleteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{}
}

// CbcheGetFunc describes the behbvior when the Get method of the pbrent
// MockCbche instbnce is invoked.
type CbcheGetFunc struct {
	defbultHook func(string) ([]byte, bool)
	hooks       []func(string) ([]byte, bool)
	history     []CbcheGetFuncCbll
	mutex       sync.Mutex
}

// Get delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCbche) Get(v0 string) ([]byte, bool) {
	r0, r1 := m.GetFunc.nextHook()(v0)
	m.GetFunc.bppendCbll(CbcheGetFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Get method of the
// pbrent MockCbche instbnce is invoked bnd the hook queue is empty.
func (f *CbcheGetFunc) SetDefbultHook(hook func(string) ([]byte, bool)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Get method of the pbrent MockCbche instbnce invokes the hook bt the front
// of the queue bnd discbrds it. After the queue is empty, the defbult hook
// function is invoked for bny future bction.
func (f *CbcheGetFunc) PushHook(hook func(string) ([]byte, bool)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CbcheGetFunc) SetDefbultReturn(r0 []byte, r1 bool) {
	f.SetDefbultHook(func(string) ([]byte, bool) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CbcheGetFunc) PushReturn(r0 []byte, r1 bool) {
	f.PushHook(func(string) ([]byte, bool) {
		return r0, r1
	})
}

func (f *CbcheGetFunc) nextHook() func(string) ([]byte, bool) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CbcheGetFunc) bppendCbll(r0 CbcheGetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of CbcheGetFuncCbll objects describing the
// invocbtions of this function.
func (f *CbcheGetFunc) History() []CbcheGetFuncCbll {
	f.mutex.Lock()
	history := mbke([]CbcheGetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CbcheGetFuncCbll is bn object thbt describes bn invocbtion of method Get
// on bn instbnce of MockCbche.
type CbcheGetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []byte
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c CbcheGetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CbcheGetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// CbcheSetFunc describes the behbvior when the Set method of the pbrent
// MockCbche instbnce is invoked.
type CbcheSetFunc struct {
	defbultHook func(string, []byte)
	hooks       []func(string, []byte)
	history     []CbcheSetFuncCbll
	mutex       sync.Mutex
}

// Set delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCbche) Set(v0 string, v1 []byte) {
	m.SetFunc.nextHook()(v0, v1)
	m.SetFunc.bppendCbll(CbcheSetFuncCbll{v0, v1})
	return
}

// SetDefbultHook sets function thbt is cblled when the Set method of the
// pbrent MockCbche instbnce is invoked bnd the hook queue is empty.
func (f *CbcheSetFunc) SetDefbultHook(hook func(string, []byte)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Set method of the pbrent MockCbche instbnce invokes the hook bt the front
// of the queue bnd discbrds it. After the queue is empty, the defbult hook
// function is invoked for bny future bction.
func (f *CbcheSetFunc) PushHook(hook func(string, []byte)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CbcheSetFunc) SetDefbultReturn() {
	f.SetDefbultHook(func(string, []byte) {
		return
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CbcheSetFunc) PushReturn() {
	f.PushHook(func(string, []byte) {
		return
	})
}

func (f *CbcheSetFunc) nextHook() func(string, []byte) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CbcheSetFunc) bppendCbll(r0 CbcheSetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of CbcheSetFuncCbll objects describing the
// invocbtions of this function.
func (f *CbcheSetFunc) History() []CbcheSetFuncCbll {
	f.mutex.Lock()
	history := mbke([]CbcheSetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CbcheSetFuncCbll is bn object thbt describes bn invocbtion of method Set
// on bn instbnce of MockCbche.
type CbcheSetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []byte
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c CbcheSetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CbcheSetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{}
}
