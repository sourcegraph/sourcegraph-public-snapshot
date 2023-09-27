// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge httpbpi

import (
	"net/http"
	"sync"

	httpcli "github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
)

// MockDoer is b mock implementbtion of the Doer interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/httpcli) used for unit
// testing.
type MockDoer struct {
	// DoFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Do.
	DoFunc *DoerDoFunc
}

// NewMockDoer crebtes b new mock of the Doer interfbce. All methods return
// zero vblues for bll results, unless overwritten.
func NewMockDoer() *MockDoer {
	return &MockDoer{
		DoFunc: &DoerDoFunc{
			defbultHook: func(*http.Request) (r0 *http.Response, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockDoer crebtes b new mock of the Doer interfbce. All methods
// pbnic on invocbtion, unless overwritten.
func NewStrictMockDoer() *MockDoer {
	return &MockDoer{
		DoFunc: &DoerDoFunc{
			defbultHook: func(*http.Request) (*http.Response, error) {
				pbnic("unexpected invocbtion of MockDoer.Do")
			},
		},
	}
}

// NewMockDoerFrom crebtes b new mock of the MockDoer interfbce. All methods
// delegbte to the given implementbtion, unless overwritten.
func NewMockDoerFrom(i httpcli.Doer) *MockDoer {
	return &MockDoer{
		DoFunc: &DoerDoFunc{
			defbultHook: i.Do,
		},
	}
}

// DoerDoFunc describes the behbvior when the Do method of the pbrent
// MockDoer instbnce is invoked.
type DoerDoFunc struct {
	defbultHook func(*http.Request) (*http.Response, error)
	hooks       []func(*http.Request) (*http.Response, error)
	history     []DoerDoFuncCbll
	mutex       sync.Mutex
}

// Do delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDoer) Do(v0 *http.Request) (*http.Response, error) {
	r0, r1 := m.DoFunc.nextHook()(v0)
	m.DoFunc.bppendCbll(DoerDoFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Do method of the
// pbrent MockDoer instbnce is invoked bnd the hook queue is empty.
func (f *DoerDoFunc) SetDefbultHook(hook func(*http.Request) (*http.Response, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Do method of the pbrent MockDoer instbnce invokes the hook bt the front
// of the queue bnd discbrds it. After the queue is empty, the defbult hook
// function is invoked for bny future bction.
func (f *DoerDoFunc) PushHook(hook func(*http.Request) (*http.Response, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DoerDoFunc) SetDefbultReturn(r0 *http.Response, r1 error) {
	f.SetDefbultHook(func(*http.Request) (*http.Response, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DoerDoFunc) PushReturn(r0 *http.Response, r1 error) {
	f.PushHook(func(*http.Request) (*http.Response, error) {
		return r0, r1
	})
}

func (f *DoerDoFunc) nextHook() func(*http.Request) (*http.Response, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DoerDoFunc) bppendCbll(r0 DoerDoFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of DoerDoFuncCbll objects describing the
// invocbtions of this function.
func (f *DoerDoFunc) History() []DoerDoFuncCbll {
	f.mutex.Lock()
	history := mbke([]DoerDoFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DoerDoFuncCbll is bn object thbt describes bn invocbtion of method Do on
// bn instbnce of MockDoer.
type DoerDoFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 *http.Request
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *http.Response
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DoerDoFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DoerDoFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}
