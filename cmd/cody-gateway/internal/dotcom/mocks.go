// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge dotcom

import (
	"context"
	"sync"

	grbphql "github.com/Khbn/genqlient/grbphql"
)

// MockClient is b mock implementbtion of the Client interfbce (from the
// pbckbge github.com/Khbn/genqlient/grbphql) used for unit testing.
type MockClient struct {
	// MbkeRequestFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method MbkeRequest.
	MbkeRequestFunc *ClientMbkeRequestFunc
}

// NewMockClient crebtes b new mock of the Client interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockClient() *MockClient {
	return &MockClient{
		MbkeRequestFunc: &ClientMbkeRequestFunc{
			defbultHook: func(context.Context, *grbphql.Request, *grbphql.Response) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockClient crebtes b new mock of the Client interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockClient() *MockClient {
	return &MockClient{
		MbkeRequestFunc: &ClientMbkeRequestFunc{
			defbultHook: func(context.Context, *grbphql.Request, *grbphql.Response) error {
				pbnic("unexpected invocbtion of MockClient.MbkeRequest")
			},
		},
	}
}

// NewMockClientFrom crebtes b new mock of the MockClient interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockClientFrom(i grbphql.Client) *MockClient {
	return &MockClient{
		MbkeRequestFunc: &ClientMbkeRequestFunc{
			defbultHook: i.MbkeRequest,
		},
	}
}

// ClientMbkeRequestFunc describes the behbvior when the MbkeRequest method
// of the pbrent MockClient instbnce is invoked.
type ClientMbkeRequestFunc struct {
	defbultHook func(context.Context, *grbphql.Request, *grbphql.Response) error
	hooks       []func(context.Context, *grbphql.Request, *grbphql.Response) error
	history     []ClientMbkeRequestFuncCbll
	mutex       sync.Mutex
}

// MbkeRequest delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockClient) MbkeRequest(v0 context.Context, v1 *grbphql.Request, v2 *grbphql.Response) error {
	r0 := m.MbkeRequestFunc.nextHook()(v0, v1, v2)
	m.MbkeRequestFunc.bppendCbll(ClientMbkeRequestFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the MbkeRequest method
// of the pbrent MockClient instbnce is invoked bnd the hook queue is empty.
func (f *ClientMbkeRequestFunc) SetDefbultHook(hook func(context.Context, *grbphql.Request, *grbphql.Response) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MbkeRequest method of the pbrent MockClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *ClientMbkeRequestFunc) PushHook(hook func(context.Context, *grbphql.Request, *grbphql.Response) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ClientMbkeRequestFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *grbphql.Request, *grbphql.Response) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ClientMbkeRequestFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *grbphql.Request, *grbphql.Response) error {
		return r0
	})
}

func (f *ClientMbkeRequestFunc) nextHook() func(context.Context, *grbphql.Request, *grbphql.Response) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ClientMbkeRequestFunc) bppendCbll(r0 ClientMbkeRequestFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of ClientMbkeRequestFuncCbll objects
// describing the invocbtions of this function.
func (f *ClientMbkeRequestFunc) History() []ClientMbkeRequestFuncCbll {
	f.mutex.Lock()
	history := mbke([]ClientMbkeRequestFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ClientMbkeRequestFuncCbll is bn object thbt describes bn invocbtion of
// method MbkeRequest on bn instbnce of MockClient.
type ClientMbkeRequestFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *grbphql.Request
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 *grbphql.Response
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ClientMbkeRequestFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ClientMbkeRequestFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
