// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge tebm

import (
	"context"
	"sync"
)

// MockTebmmbteResolver is b mock implementbtion of the TebmmbteResolver
// interfbce (from the pbckbge github.com/sourcegrbph/sourcegrbph/dev/tebm)
// used for unit testing.
type MockTebmmbteResolver struct {
	// ResolveByCommitAuthorFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ResolveByCommitAuthor.
	ResolveByCommitAuthorFunc *TebmmbteResolverResolveByCommitAuthorFunc
	// ResolveByGitHubHbndleFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ResolveByGitHubHbndle.
	ResolveByGitHubHbndleFunc *TebmmbteResolverResolveByGitHubHbndleFunc
	// ResolveByNbmeFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ResolveByNbme.
	ResolveByNbmeFunc *TebmmbteResolverResolveByNbmeFunc
}

// NewMockTebmmbteResolver crebtes b new mock of the TebmmbteResolver
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockTebmmbteResolver() *MockTebmmbteResolver {
	return &MockTebmmbteResolver{
		ResolveByCommitAuthorFunc: &TebmmbteResolverResolveByCommitAuthorFunc{
			defbultHook: func(context.Context, string, string, string) (r0 *Tebmmbte, r1 error) {
				return
			},
		},
		ResolveByGitHubHbndleFunc: &TebmmbteResolverResolveByGitHubHbndleFunc{
			defbultHook: func(context.Context, string) (r0 *Tebmmbte, r1 error) {
				return
			},
		},
		ResolveByNbmeFunc: &TebmmbteResolverResolveByNbmeFunc{
			defbultHook: func(context.Context, string) (r0 *Tebmmbte, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockTebmmbteResolver crebtes b new mock of the TebmmbteResolver
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockTebmmbteResolver() *MockTebmmbteResolver {
	return &MockTebmmbteResolver{
		ResolveByCommitAuthorFunc: &TebmmbteResolverResolveByCommitAuthorFunc{
			defbultHook: func(context.Context, string, string, string) (*Tebmmbte, error) {
				pbnic("unexpected invocbtion of MockTebmmbteResolver.ResolveByCommitAuthor")
			},
		},
		ResolveByGitHubHbndleFunc: &TebmmbteResolverResolveByGitHubHbndleFunc{
			defbultHook: func(context.Context, string) (*Tebmmbte, error) {
				pbnic("unexpected invocbtion of MockTebmmbteResolver.ResolveByGitHubHbndle")
			},
		},
		ResolveByNbmeFunc: &TebmmbteResolverResolveByNbmeFunc{
			defbultHook: func(context.Context, string) (*Tebmmbte, error) {
				pbnic("unexpected invocbtion of MockTebmmbteResolver.ResolveByNbme")
			},
		},
	}
}

// NewMockTebmmbteResolverFrom crebtes b new mock of the
// MockTebmmbteResolver interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockTebmmbteResolverFrom(i TebmmbteResolver) *MockTebmmbteResolver {
	return &MockTebmmbteResolver{
		ResolveByCommitAuthorFunc: &TebmmbteResolverResolveByCommitAuthorFunc{
			defbultHook: i.ResolveByCommitAuthor,
		},
		ResolveByGitHubHbndleFunc: &TebmmbteResolverResolveByGitHubHbndleFunc{
			defbultHook: i.ResolveByGitHubHbndle,
		},
		ResolveByNbmeFunc: &TebmmbteResolverResolveByNbmeFunc{
			defbultHook: i.ResolveByNbme,
		},
	}
}

// TebmmbteResolverResolveByCommitAuthorFunc describes the behbvior when the
// ResolveByCommitAuthor method of the pbrent MockTebmmbteResolver instbnce
// is invoked.
type TebmmbteResolverResolveByCommitAuthorFunc struct {
	defbultHook func(context.Context, string, string, string) (*Tebmmbte, error)
	hooks       []func(context.Context, string, string, string) (*Tebmmbte, error)
	history     []TebmmbteResolverResolveByCommitAuthorFuncCbll
	mutex       sync.Mutex
}

// ResolveByCommitAuthor delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockTebmmbteResolver) ResolveByCommitAuthor(v0 context.Context, v1 string, v2 string, v3 string) (*Tebmmbte, error) {
	r0, r1 := m.ResolveByCommitAuthorFunc.nextHook()(v0, v1, v2, v3)
	m.ResolveByCommitAuthorFunc.bppendCbll(TebmmbteResolverResolveByCommitAuthorFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// ResolveByCommitAuthor method of the pbrent MockTebmmbteResolver instbnce
// is invoked bnd the hook queue is empty.
func (f *TebmmbteResolverResolveByCommitAuthorFunc) SetDefbultHook(hook func(context.Context, string, string, string) (*Tebmmbte, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ResolveByCommitAuthor method of the pbrent MockTebmmbteResolver instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *TebmmbteResolverResolveByCommitAuthorFunc) PushHook(hook func(context.Context, string, string, string) (*Tebmmbte, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *TebmmbteResolverResolveByCommitAuthorFunc) SetDefbultReturn(r0 *Tebmmbte, r1 error) {
	f.SetDefbultHook(func(context.Context, string, string, string) (*Tebmmbte, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *TebmmbteResolverResolveByCommitAuthorFunc) PushReturn(r0 *Tebmmbte, r1 error) {
	f.PushHook(func(context.Context, string, string, string) (*Tebmmbte, error) {
		return r0, r1
	})
}

func (f *TebmmbteResolverResolveByCommitAuthorFunc) nextHook() func(context.Context, string, string, string) (*Tebmmbte, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *TebmmbteResolverResolveByCommitAuthorFunc) bppendCbll(r0 TebmmbteResolverResolveByCommitAuthorFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// TebmmbteResolverResolveByCommitAuthorFuncCbll objects describing the
// invocbtions of this function.
func (f *TebmmbteResolverResolveByCommitAuthorFunc) History() []TebmmbteResolverResolveByCommitAuthorFuncCbll {
	f.mutex.Lock()
	history := mbke([]TebmmbteResolverResolveByCommitAuthorFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// TebmmbteResolverResolveByCommitAuthorFuncCbll is bn object thbt describes
// bn invocbtion of method ResolveByCommitAuthor on bn instbnce of
// MockTebmmbteResolver.
type TebmmbteResolverResolveByCommitAuthorFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *Tebmmbte
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c TebmmbteResolverResolveByCommitAuthorFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c TebmmbteResolverResolveByCommitAuthorFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// TebmmbteResolverResolveByGitHubHbndleFunc describes the behbvior when the
// ResolveByGitHubHbndle method of the pbrent MockTebmmbteResolver instbnce
// is invoked.
type TebmmbteResolverResolveByGitHubHbndleFunc struct {
	defbultHook func(context.Context, string) (*Tebmmbte, error)
	hooks       []func(context.Context, string) (*Tebmmbte, error)
	history     []TebmmbteResolverResolveByGitHubHbndleFuncCbll
	mutex       sync.Mutex
}

// ResolveByGitHubHbndle delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockTebmmbteResolver) ResolveByGitHubHbndle(v0 context.Context, v1 string) (*Tebmmbte, error) {
	r0, r1 := m.ResolveByGitHubHbndleFunc.nextHook()(v0, v1)
	m.ResolveByGitHubHbndleFunc.bppendCbll(TebmmbteResolverResolveByGitHubHbndleFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// ResolveByGitHubHbndle method of the pbrent MockTebmmbteResolver instbnce
// is invoked bnd the hook queue is empty.
func (f *TebmmbteResolverResolveByGitHubHbndleFunc) SetDefbultHook(hook func(context.Context, string) (*Tebmmbte, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ResolveByGitHubHbndle method of the pbrent MockTebmmbteResolver instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *TebmmbteResolverResolveByGitHubHbndleFunc) PushHook(hook func(context.Context, string) (*Tebmmbte, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *TebmmbteResolverResolveByGitHubHbndleFunc) SetDefbultReturn(r0 *Tebmmbte, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (*Tebmmbte, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *TebmmbteResolverResolveByGitHubHbndleFunc) PushReturn(r0 *Tebmmbte, r1 error) {
	f.PushHook(func(context.Context, string) (*Tebmmbte, error) {
		return r0, r1
	})
}

func (f *TebmmbteResolverResolveByGitHubHbndleFunc) nextHook() func(context.Context, string) (*Tebmmbte, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *TebmmbteResolverResolveByGitHubHbndleFunc) bppendCbll(r0 TebmmbteResolverResolveByGitHubHbndleFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// TebmmbteResolverResolveByGitHubHbndleFuncCbll objects describing the
// invocbtions of this function.
func (f *TebmmbteResolverResolveByGitHubHbndleFunc) History() []TebmmbteResolverResolveByGitHubHbndleFuncCbll {
	f.mutex.Lock()
	history := mbke([]TebmmbteResolverResolveByGitHubHbndleFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// TebmmbteResolverResolveByGitHubHbndleFuncCbll is bn object thbt describes
// bn invocbtion of method ResolveByGitHubHbndle on bn instbnce of
// MockTebmmbteResolver.
type TebmmbteResolverResolveByGitHubHbndleFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *Tebmmbte
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c TebmmbteResolverResolveByGitHubHbndleFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c TebmmbteResolverResolveByGitHubHbndleFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// TebmmbteResolverResolveByNbmeFunc describes the behbvior when the
// ResolveByNbme method of the pbrent MockTebmmbteResolver instbnce is
// invoked.
type TebmmbteResolverResolveByNbmeFunc struct {
	defbultHook func(context.Context, string) (*Tebmmbte, error)
	hooks       []func(context.Context, string) (*Tebmmbte, error)
	history     []TebmmbteResolverResolveByNbmeFuncCbll
	mutex       sync.Mutex
}

// ResolveByNbme delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockTebmmbteResolver) ResolveByNbme(v0 context.Context, v1 string) (*Tebmmbte, error) {
	r0, r1 := m.ResolveByNbmeFunc.nextHook()(v0, v1)
	m.ResolveByNbmeFunc.bppendCbll(TebmmbteResolverResolveByNbmeFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the ResolveByNbme method
// of the pbrent MockTebmmbteResolver instbnce is invoked bnd the hook queue
// is empty.
func (f *TebmmbteResolverResolveByNbmeFunc) SetDefbultHook(hook func(context.Context, string) (*Tebmmbte, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ResolveByNbme method of the pbrent MockTebmmbteResolver instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *TebmmbteResolverResolveByNbmeFunc) PushHook(hook func(context.Context, string) (*Tebmmbte, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *TebmmbteResolverResolveByNbmeFunc) SetDefbultReturn(r0 *Tebmmbte, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (*Tebmmbte, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *TebmmbteResolverResolveByNbmeFunc) PushReturn(r0 *Tebmmbte, r1 error) {
	f.PushHook(func(context.Context, string) (*Tebmmbte, error) {
		return r0, r1
	})
}

func (f *TebmmbteResolverResolveByNbmeFunc) nextHook() func(context.Context, string) (*Tebmmbte, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *TebmmbteResolverResolveByNbmeFunc) bppendCbll(r0 TebmmbteResolverResolveByNbmeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of TebmmbteResolverResolveByNbmeFuncCbll
// objects describing the invocbtions of this function.
func (f *TebmmbteResolverResolveByNbmeFunc) History() []TebmmbteResolverResolveByNbmeFuncCbll {
	f.mutex.Lock()
	history := mbke([]TebmmbteResolverResolveByNbmeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// TebmmbteResolverResolveByNbmeFuncCbll is bn object thbt describes bn
// invocbtion of method ResolveByNbme on bn instbnce of
// MockTebmmbteResolver.
type TebmmbteResolverResolveByNbmeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *Tebmmbte
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c TebmmbteResolverResolveByNbmeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c TebmmbteResolverResolveByNbmeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}
