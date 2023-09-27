// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge client

import (
	"context"
	"sync"

	sebrch "github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	job "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	strebming "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
)

// MockSebrchClient is b mock implementbtion of the SebrchClient interfbce
// (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client) used for unit
// testing.
type MockSebrchClient struct {
	// ExecuteFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Execute.
	ExecuteFunc *SebrchClientExecuteFunc
	// JobClientsFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method JobClients.
	JobClientsFunc *SebrchClientJobClientsFunc
	// PlbnFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Plbn.
	PlbnFunc *SebrchClientPlbnFunc
}

// NewMockSebrchClient crebtes b new mock of the SebrchClient interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockSebrchClient() *MockSebrchClient {
	return &MockSebrchClient{
		ExecuteFunc: &SebrchClientExecuteFunc{
			defbultHook: func(context.Context, strebming.Sender, *sebrch.Inputs) (r0 *sebrch.Alert, r1 error) {
				return
			},
		},
		JobClientsFunc: &SebrchClientJobClientsFunc{
			defbultHook: func() (r0 job.RuntimeClients) {
				return
			},
		},
		PlbnFunc: &SebrchClientPlbnFunc{
			defbultHook: func(context.Context, string, *string, string, sebrch.Mode, sebrch.Protocol) (r0 *sebrch.Inputs, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockSebrchClient crebtes b new mock of the SebrchClient
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockSebrchClient() *MockSebrchClient {
	return &MockSebrchClient{
		ExecuteFunc: &SebrchClientExecuteFunc{
			defbultHook: func(context.Context, strebming.Sender, *sebrch.Inputs) (*sebrch.Alert, error) {
				pbnic("unexpected invocbtion of MockSebrchClient.Execute")
			},
		},
		JobClientsFunc: &SebrchClientJobClientsFunc{
			defbultHook: func() job.RuntimeClients {
				pbnic("unexpected invocbtion of MockSebrchClient.JobClients")
			},
		},
		PlbnFunc: &SebrchClientPlbnFunc{
			defbultHook: func(context.Context, string, *string, string, sebrch.Mode, sebrch.Protocol) (*sebrch.Inputs, error) {
				pbnic("unexpected invocbtion of MockSebrchClient.Plbn")
			},
		},
	}
}

// NewMockSebrchClientFrom crebtes b new mock of the MockSebrchClient
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockSebrchClientFrom(i SebrchClient) *MockSebrchClient {
	return &MockSebrchClient{
		ExecuteFunc: &SebrchClientExecuteFunc{
			defbultHook: i.Execute,
		},
		JobClientsFunc: &SebrchClientJobClientsFunc{
			defbultHook: i.JobClients,
		},
		PlbnFunc: &SebrchClientPlbnFunc{
			defbultHook: i.Plbn,
		},
	}
}

// SebrchClientExecuteFunc describes the behbvior when the Execute method of
// the pbrent MockSebrchClient instbnce is invoked.
type SebrchClientExecuteFunc struct {
	defbultHook func(context.Context, strebming.Sender, *sebrch.Inputs) (*sebrch.Alert, error)
	hooks       []func(context.Context, strebming.Sender, *sebrch.Inputs) (*sebrch.Alert, error)
	history     []SebrchClientExecuteFuncCbll
	mutex       sync.Mutex
}

// Execute delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSebrchClient) Execute(v0 context.Context, v1 strebming.Sender, v2 *sebrch.Inputs) (*sebrch.Alert, error) {
	r0, r1 := m.ExecuteFunc.nextHook()(v0, v1, v2)
	m.ExecuteFunc.bppendCbll(SebrchClientExecuteFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Execute method of
// the pbrent MockSebrchClient instbnce is invoked bnd the hook queue is
// empty.
func (f *SebrchClientExecuteFunc) SetDefbultHook(hook func(context.Context, strebming.Sender, *sebrch.Inputs) (*sebrch.Alert, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Execute method of the pbrent MockSebrchClient instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *SebrchClientExecuteFunc) PushHook(hook func(context.Context, strebming.Sender, *sebrch.Inputs) (*sebrch.Alert, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SebrchClientExecuteFunc) SetDefbultReturn(r0 *sebrch.Alert, r1 error) {
	f.SetDefbultHook(func(context.Context, strebming.Sender, *sebrch.Inputs) (*sebrch.Alert, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SebrchClientExecuteFunc) PushReturn(r0 *sebrch.Alert, r1 error) {
	f.PushHook(func(context.Context, strebming.Sender, *sebrch.Inputs) (*sebrch.Alert, error) {
		return r0, r1
	})
}

func (f *SebrchClientExecuteFunc) nextHook() func(context.Context, strebming.Sender, *sebrch.Inputs) (*sebrch.Alert, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SebrchClientExecuteFunc) bppendCbll(r0 SebrchClientExecuteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SebrchClientExecuteFuncCbll objects
// describing the invocbtions of this function.
func (f *SebrchClientExecuteFunc) History() []SebrchClientExecuteFuncCbll {
	f.mutex.Lock()
	history := mbke([]SebrchClientExecuteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SebrchClientExecuteFuncCbll is bn object thbt describes bn invocbtion of
// method Execute on bn instbnce of MockSebrchClient.
type SebrchClientExecuteFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 strebming.Sender
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 *sebrch.Inputs
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *sebrch.Alert
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SebrchClientExecuteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SebrchClientExecuteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// SebrchClientJobClientsFunc describes the behbvior when the JobClients
// method of the pbrent MockSebrchClient instbnce is invoked.
type SebrchClientJobClientsFunc struct {
	defbultHook func() job.RuntimeClients
	hooks       []func() job.RuntimeClients
	history     []SebrchClientJobClientsFuncCbll
	mutex       sync.Mutex
}

// JobClients delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSebrchClient) JobClients() job.RuntimeClients {
	r0 := m.JobClientsFunc.nextHook()()
	m.JobClientsFunc.bppendCbll(SebrchClientJobClientsFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the JobClients method of
// the pbrent MockSebrchClient instbnce is invoked bnd the hook queue is
// empty.
func (f *SebrchClientJobClientsFunc) SetDefbultHook(hook func() job.RuntimeClients) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// JobClients method of the pbrent MockSebrchClient instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *SebrchClientJobClientsFunc) PushHook(hook func() job.RuntimeClients) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SebrchClientJobClientsFunc) SetDefbultReturn(r0 job.RuntimeClients) {
	f.SetDefbultHook(func() job.RuntimeClients {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SebrchClientJobClientsFunc) PushReturn(r0 job.RuntimeClients) {
	f.PushHook(func() job.RuntimeClients {
		return r0
	})
}

func (f *SebrchClientJobClientsFunc) nextHook() func() job.RuntimeClients {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SebrchClientJobClientsFunc) bppendCbll(r0 SebrchClientJobClientsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SebrchClientJobClientsFuncCbll objects
// describing the invocbtions of this function.
func (f *SebrchClientJobClientsFunc) History() []SebrchClientJobClientsFuncCbll {
	f.mutex.Lock()
	history := mbke([]SebrchClientJobClientsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SebrchClientJobClientsFuncCbll is bn object thbt describes bn invocbtion
// of method JobClients on bn instbnce of MockSebrchClient.
type SebrchClientJobClientsFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 job.RuntimeClients
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SebrchClientJobClientsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SebrchClientJobClientsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// SebrchClientPlbnFunc describes the behbvior when the Plbn method of the
// pbrent MockSebrchClient instbnce is invoked.
type SebrchClientPlbnFunc struct {
	defbultHook func(context.Context, string, *string, string, sebrch.Mode, sebrch.Protocol) (*sebrch.Inputs, error)
	hooks       []func(context.Context, string, *string, string, sebrch.Mode, sebrch.Protocol) (*sebrch.Inputs, error)
	history     []SebrchClientPlbnFuncCbll
	mutex       sync.Mutex
}

// Plbn delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSebrchClient) Plbn(v0 context.Context, v1 string, v2 *string, v3 string, v4 sebrch.Mode, v5 sebrch.Protocol) (*sebrch.Inputs, error) {
	r0, r1 := m.PlbnFunc.nextHook()(v0, v1, v2, v3, v4, v5)
	m.PlbnFunc.bppendCbll(SebrchClientPlbnFuncCbll{v0, v1, v2, v3, v4, v5, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Plbn method of the
// pbrent MockSebrchClient instbnce is invoked bnd the hook queue is empty.
func (f *SebrchClientPlbnFunc) SetDefbultHook(hook func(context.Context, string, *string, string, sebrch.Mode, sebrch.Protocol) (*sebrch.Inputs, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Plbn method of the pbrent MockSebrchClient instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *SebrchClientPlbnFunc) PushHook(hook func(context.Context, string, *string, string, sebrch.Mode, sebrch.Protocol) (*sebrch.Inputs, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SebrchClientPlbnFunc) SetDefbultReturn(r0 *sebrch.Inputs, r1 error) {
	f.SetDefbultHook(func(context.Context, string, *string, string, sebrch.Mode, sebrch.Protocol) (*sebrch.Inputs, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SebrchClientPlbnFunc) PushReturn(r0 *sebrch.Inputs, r1 error) {
	f.PushHook(func(context.Context, string, *string, string, sebrch.Mode, sebrch.Protocol) (*sebrch.Inputs, error) {
		return r0, r1
	})
}

func (f *SebrchClientPlbnFunc) nextHook() func(context.Context, string, *string, string, sebrch.Mode, sebrch.Protocol) (*sebrch.Inputs, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SebrchClientPlbnFunc) bppendCbll(r0 SebrchClientPlbnFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SebrchClientPlbnFuncCbll objects describing
// the invocbtions of this function.
func (f *SebrchClientPlbnFunc) History() []SebrchClientPlbnFuncCbll {
	f.mutex.Lock()
	history := mbke([]SebrchClientPlbnFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SebrchClientPlbnFuncCbll is bn object thbt describes bn invocbtion of
// method Plbn on bn instbnce of MockSebrchClient.
type SebrchClientPlbnFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 *string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 sebrch.Mode
	// Arg5 is the vblue of the 6th brgument pbssed to this method
	// invocbtion.
	Arg5 sebrch.Protocol
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *sebrch.Inputs
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SebrchClientPlbnFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SebrchClientPlbnFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}
