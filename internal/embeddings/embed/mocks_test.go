// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge embed

import (
	"context"
	"sync"

	context1 "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/context"
)

// MockContextService is b mock implementbtion of the ContextService
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/embeddings/embed) used for
// unit testing.
type MockContextService struct {
	// SplitIntoEmbeddbbleChunksFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// SplitIntoEmbeddbbleChunks.
	SplitIntoEmbeddbbleChunksFunc *ContextServiceSplitIntoEmbeddbbleChunksFunc
}

// NewMockContextService crebtes b new mock of the ContextService interfbce.
// All methods return zero vblues for bll results, unless overwritten.
func NewMockContextService() *MockContextService {
	return &MockContextService{
		SplitIntoEmbeddbbleChunksFunc: &ContextServiceSplitIntoEmbeddbbleChunksFunc{
			defbultHook: func(context.Context, string, string, context1.SplitOptions) (r0 []context1.EmbeddbbleChunk, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockContextService crebtes b new mock of the ContextService
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockContextService() *MockContextService {
	return &MockContextService{
		SplitIntoEmbeddbbleChunksFunc: &ContextServiceSplitIntoEmbeddbbleChunksFunc{
			defbultHook: func(context.Context, string, string, context1.SplitOptions) ([]context1.EmbeddbbleChunk, error) {
				pbnic("unexpected invocbtion of MockContextService.SplitIntoEmbeddbbleChunks")
			},
		},
	}
}

// NewMockContextServiceFrom crebtes b new mock of the MockContextService
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockContextServiceFrom(i ContextService) *MockContextService {
	return &MockContextService{
		SplitIntoEmbeddbbleChunksFunc: &ContextServiceSplitIntoEmbeddbbleChunksFunc{
			defbultHook: i.SplitIntoEmbeddbbleChunks,
		},
	}
}

// ContextServiceSplitIntoEmbeddbbleChunksFunc describes the behbvior when
// the SplitIntoEmbeddbbleChunks method of the pbrent MockContextService
// instbnce is invoked.
type ContextServiceSplitIntoEmbeddbbleChunksFunc struct {
	defbultHook func(context.Context, string, string, context1.SplitOptions) ([]context1.EmbeddbbleChunk, error)
	hooks       []func(context.Context, string, string, context1.SplitOptions) ([]context1.EmbeddbbleChunk, error)
	history     []ContextServiceSplitIntoEmbeddbbleChunksFuncCbll
	mutex       sync.Mutex
}

// SplitIntoEmbeddbbleChunks delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockContextService) SplitIntoEmbeddbbleChunks(v0 context.Context, v1 string, v2 string, v3 context1.SplitOptions) ([]context1.EmbeddbbleChunk, error) {
	r0, r1 := m.SplitIntoEmbeddbbleChunksFunc.nextHook()(v0, v1, v2, v3)
	m.SplitIntoEmbeddbbleChunksFunc.bppendCbll(ContextServiceSplitIntoEmbeddbbleChunksFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// SplitIntoEmbeddbbleChunks method of the pbrent MockContextService
// instbnce is invoked bnd the hook queue is empty.
func (f *ContextServiceSplitIntoEmbeddbbleChunksFunc) SetDefbultHook(hook func(context.Context, string, string, context1.SplitOptions) ([]context1.EmbeddbbleChunk, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SplitIntoEmbeddbbleChunks method of the pbrent MockContextService
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *ContextServiceSplitIntoEmbeddbbleChunksFunc) PushHook(hook func(context.Context, string, string, context1.SplitOptions) ([]context1.EmbeddbbleChunk, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *ContextServiceSplitIntoEmbeddbbleChunksFunc) SetDefbultReturn(r0 []context1.EmbeddbbleChunk, r1 error) {
	f.SetDefbultHook(func(context.Context, string, string, context1.SplitOptions) ([]context1.EmbeddbbleChunk, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *ContextServiceSplitIntoEmbeddbbleChunksFunc) PushReturn(r0 []context1.EmbeddbbleChunk, r1 error) {
	f.PushHook(func(context.Context, string, string, context1.SplitOptions) ([]context1.EmbeddbbleChunk, error) {
		return r0, r1
	})
}

func (f *ContextServiceSplitIntoEmbeddbbleChunksFunc) nextHook() func(context.Context, string, string, context1.SplitOptions) ([]context1.EmbeddbbleChunk, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *ContextServiceSplitIntoEmbeddbbleChunksFunc) bppendCbll(r0 ContextServiceSplitIntoEmbeddbbleChunksFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// ContextServiceSplitIntoEmbeddbbleChunksFuncCbll objects describing the
// invocbtions of this function.
func (f *ContextServiceSplitIntoEmbeddbbleChunksFunc) History() []ContextServiceSplitIntoEmbeddbbleChunksFuncCbll {
	f.mutex.Lock()
	history := mbke([]ContextServiceSplitIntoEmbeddbbleChunksFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// ContextServiceSplitIntoEmbeddbbleChunksFuncCbll is bn object thbt
// describes bn invocbtion of method SplitIntoEmbeddbbleChunks on bn
// instbnce of MockContextService.
type ContextServiceSplitIntoEmbeddbbleChunksFuncCbll struct {
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
	Arg3 context1.SplitOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []context1.EmbeddbbleChunk
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c ContextServiceSplitIntoEmbeddbbleChunksFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c ContextServiceSplitIntoEmbeddbbleChunksFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}
