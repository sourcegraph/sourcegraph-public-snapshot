// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge grbphql

import (
	"context"
	"sync"

	codenbv "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv"
	shbred1 "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	shbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
)

// MockAutoIndexingService is b mock implementbtion of the
// AutoIndexingService interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/trbnsport/grbphql)
// used for unit testing.
type MockAutoIndexingService struct {
	// QueueRepoRevFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method QueueRepoRev.
	QueueRepoRevFunc *AutoIndexingServiceQueueRepoRevFunc
}

// NewMockAutoIndexingService crebtes b new mock of the AutoIndexingService
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockAutoIndexingService() *MockAutoIndexingService {
	return &MockAutoIndexingService{
		QueueRepoRevFunc: &AutoIndexingServiceQueueRepoRevFunc{
			defbultHook: func(context.Context, int, string) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockAutoIndexingService crebtes b new mock of the
// AutoIndexingService interfbce. All methods pbnic on invocbtion, unless
// overwritten.
func NewStrictMockAutoIndexingService() *MockAutoIndexingService {
	return &MockAutoIndexingService{
		QueueRepoRevFunc: &AutoIndexingServiceQueueRepoRevFunc{
			defbultHook: func(context.Context, int, string) error {
				pbnic("unexpected invocbtion of MockAutoIndexingService.QueueRepoRev")
			},
		},
	}
}

// NewMockAutoIndexingServiceFrom crebtes b new mock of the
// MockAutoIndexingService interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockAutoIndexingServiceFrom(i AutoIndexingService) *MockAutoIndexingService {
	return &MockAutoIndexingService{
		QueueRepoRevFunc: &AutoIndexingServiceQueueRepoRevFunc{
			defbultHook: i.QueueRepoRev,
		},
	}
}

// AutoIndexingServiceQueueRepoRevFunc describes the behbvior when the
// QueueRepoRev method of the pbrent MockAutoIndexingService instbnce is
// invoked.
type AutoIndexingServiceQueueRepoRevFunc struct {
	defbultHook func(context.Context, int, string) error
	hooks       []func(context.Context, int, string) error
	history     []AutoIndexingServiceQueueRepoRevFuncCbll
	mutex       sync.Mutex
}

// QueueRepoRev delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockAutoIndexingService) QueueRepoRev(v0 context.Context, v1 int, v2 string) error {
	r0 := m.QueueRepoRevFunc.nextHook()(v0, v1, v2)
	m.QueueRepoRevFunc.bppendCbll(AutoIndexingServiceQueueRepoRevFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the QueueRepoRev method
// of the pbrent MockAutoIndexingService instbnce is invoked bnd the hook
// queue is empty.
func (f *AutoIndexingServiceQueueRepoRevFunc) SetDefbultHook(hook func(context.Context, int, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// QueueRepoRev method of the pbrent MockAutoIndexingService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *AutoIndexingServiceQueueRepoRevFunc) PushHook(hook func(context.Context, int, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *AutoIndexingServiceQueueRepoRevFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *AutoIndexingServiceQueueRepoRevFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, string) error {
		return r0
	})
}

func (f *AutoIndexingServiceQueueRepoRevFunc) nextHook() func(context.Context, int, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *AutoIndexingServiceQueueRepoRevFunc) bppendCbll(r0 AutoIndexingServiceQueueRepoRevFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of AutoIndexingServiceQueueRepoRevFuncCbll
// objects describing the invocbtions of this function.
func (f *AutoIndexingServiceQueueRepoRevFunc) History() []AutoIndexingServiceQueueRepoRevFuncCbll {
	f.mutex.Lock()
	history := mbke([]AutoIndexingServiceQueueRepoRevFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// AutoIndexingServiceQueueRepoRevFuncCbll is bn object thbt describes bn
// invocbtion of method QueueRepoRev on bn instbnce of
// MockAutoIndexingService.
type AutoIndexingServiceQueueRepoRevFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c AutoIndexingServiceQueueRepoRevFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c AutoIndexingServiceQueueRepoRevFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockCodeNbvService is b mock implementbtion of the CodeNbvService
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/trbnsport/grbphql)
// used for unit testing.
type MockCodeNbvService struct {
	// GetClosestDumpsForBlobFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetClosestDumpsForBlob.
	GetClosestDumpsForBlobFunc *CodeNbvServiceGetClosestDumpsForBlobFunc
	// GetDibgnosticsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetDibgnostics.
	GetDibgnosticsFunc *CodeNbvServiceGetDibgnosticsFunc
	// GetHoverFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method GetHover.
	GetHoverFunc *CodeNbvServiceGetHoverFunc
	// GetRbngesFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetRbnges.
	GetRbngesFunc *CodeNbvServiceGetRbngesFunc
	// GetStencilFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetStencil.
	GetStencilFunc *CodeNbvServiceGetStencilFunc
	// NewGetDefinitionsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method NewGetDefinitions.
	NewGetDefinitionsFunc *CodeNbvServiceNewGetDefinitionsFunc
	// NewGetImplementbtionsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method NewGetImplementbtions.
	NewGetImplementbtionsFunc *CodeNbvServiceNewGetImplementbtionsFunc
	// NewGetPrototypesFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method NewGetPrototypes.
	NewGetPrototypesFunc *CodeNbvServiceNewGetPrototypesFunc
	// NewGetReferencesFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method NewGetReferences.
	NewGetReferencesFunc *CodeNbvServiceNewGetReferencesFunc
	// SnbpshotForDocumentFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method SnbpshotForDocument.
	SnbpshotForDocumentFunc *CodeNbvServiceSnbpshotForDocumentFunc
	// VisibleUplobdsForPbthFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method VisibleUplobdsForPbth.
	VisibleUplobdsForPbthFunc *CodeNbvServiceVisibleUplobdsForPbthFunc
}

// NewMockCodeNbvService crebtes b new mock of the CodeNbvService interfbce.
// All methods return zero vblues for bll results, unless overwritten.
func NewMockCodeNbvService() *MockCodeNbvService {
	return &MockCodeNbvService{
		GetClosestDumpsForBlobFunc: &CodeNbvServiceGetClosestDumpsForBlobFunc{
			defbultHook: func(context.Context, int, string, string, bool, string) (r0 []shbred.Dump, r1 error) {
				return
			},
		},
		GetDibgnosticsFunc: &CodeNbvServiceGetDibgnosticsFunc{
			defbultHook: func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) (r0 []codenbv.DibgnosticAtUplobd, r1 int, r2 error) {
				return
			},
		},
		GetHoverFunc: &CodeNbvServiceGetHoverFunc{
			defbultHook: func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) (r0 string, r1 shbred1.Rbnge, r2 bool, r3 error) {
				return
			},
		},
		GetRbngesFunc: &CodeNbvServiceGetRbngesFunc{
			defbultHook: func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, int, int) (r0 []codenbv.AdjustedCodeIntelligenceRbnge, r1 error) {
				return
			},
		},
		GetStencilFunc: &CodeNbvServiceGetStencilFunc{
			defbultHook: func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) (r0 []shbred1.Rbnge, r1 error) {
				return
			},
		},
		NewGetDefinitionsFunc: &CodeNbvServiceNewGetDefinitionsFunc{
			defbultHook: func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) (r0 []shbred1.UplobdLocbtion, r1 error) {
				return
			},
		},
		NewGetImplementbtionsFunc: &CodeNbvServiceNewGetImplementbtionsFunc{
			defbultHook: func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) (r0 []shbred1.UplobdLocbtion, r1 codenbv.Cursor, r2 error) {
				return
			},
		},
		NewGetPrototypesFunc: &CodeNbvServiceNewGetPrototypesFunc{
			defbultHook: func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) (r0 []shbred1.UplobdLocbtion, r1 codenbv.Cursor, r2 error) {
				return
			},
		},
		NewGetReferencesFunc: &CodeNbvServiceNewGetReferencesFunc{
			defbultHook: func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) (r0 []shbred1.UplobdLocbtion, r1 codenbv.Cursor, r2 error) {
				return
			},
		},
		SnbpshotForDocumentFunc: &CodeNbvServiceSnbpshotForDocumentFunc{
			defbultHook: func(context.Context, int, string, string, int) (r0 []shbred1.SnbpshotDbtb, r1 error) {
				return
			},
		},
		VisibleUplobdsForPbthFunc: &CodeNbvServiceVisibleUplobdsForPbthFunc{
			defbultHook: func(context.Context, codenbv.RequestStbte) (r0 []shbred.Dump, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockCodeNbvService crebtes b new mock of the CodeNbvService
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockCodeNbvService() *MockCodeNbvService {
	return &MockCodeNbvService{
		GetClosestDumpsForBlobFunc: &CodeNbvServiceGetClosestDumpsForBlobFunc{
			defbultHook: func(context.Context, int, string, string, bool, string) ([]shbred.Dump, error) {
				pbnic("unexpected invocbtion of MockCodeNbvService.GetClosestDumpsForBlob")
			},
		},
		GetDibgnosticsFunc: &CodeNbvServiceGetDibgnosticsFunc{
			defbultHook: func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]codenbv.DibgnosticAtUplobd, int, error) {
				pbnic("unexpected invocbtion of MockCodeNbvService.GetDibgnostics")
			},
		},
		GetHoverFunc: &CodeNbvServiceGetHoverFunc{
			defbultHook: func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) (string, shbred1.Rbnge, bool, error) {
				pbnic("unexpected invocbtion of MockCodeNbvService.GetHover")
			},
		},
		GetRbngesFunc: &CodeNbvServiceGetRbngesFunc{
			defbultHook: func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, int, int) ([]codenbv.AdjustedCodeIntelligenceRbnge, error) {
				pbnic("unexpected invocbtion of MockCodeNbvService.GetRbnges")
			},
		},
		GetStencilFunc: &CodeNbvServiceGetStencilFunc{
			defbultHook: func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]shbred1.Rbnge, error) {
				pbnic("unexpected invocbtion of MockCodeNbvService.GetStencil")
			},
		},
		NewGetDefinitionsFunc: &CodeNbvServiceNewGetDefinitionsFunc{
			defbultHook: func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]shbred1.UplobdLocbtion, error) {
				pbnic("unexpected invocbtion of MockCodeNbvService.NewGetDefinitions")
			},
		},
		NewGetImplementbtionsFunc: &CodeNbvServiceNewGetImplementbtionsFunc{
			defbultHook: func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error) {
				pbnic("unexpected invocbtion of MockCodeNbvService.NewGetImplementbtions")
			},
		},
		NewGetPrototypesFunc: &CodeNbvServiceNewGetPrototypesFunc{
			defbultHook: func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error) {
				pbnic("unexpected invocbtion of MockCodeNbvService.NewGetPrototypes")
			},
		},
		NewGetReferencesFunc: &CodeNbvServiceNewGetReferencesFunc{
			defbultHook: func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error) {
				pbnic("unexpected invocbtion of MockCodeNbvService.NewGetReferences")
			},
		},
		SnbpshotForDocumentFunc: &CodeNbvServiceSnbpshotForDocumentFunc{
			defbultHook: func(context.Context, int, string, string, int) ([]shbred1.SnbpshotDbtb, error) {
				pbnic("unexpected invocbtion of MockCodeNbvService.SnbpshotForDocument")
			},
		},
		VisibleUplobdsForPbthFunc: &CodeNbvServiceVisibleUplobdsForPbthFunc{
			defbultHook: func(context.Context, codenbv.RequestStbte) ([]shbred.Dump, error) {
				pbnic("unexpected invocbtion of MockCodeNbvService.VisibleUplobdsForPbth")
			},
		},
	}
}

// NewMockCodeNbvServiceFrom crebtes b new mock of the MockCodeNbvService
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockCodeNbvServiceFrom(i CodeNbvService) *MockCodeNbvService {
	return &MockCodeNbvService{
		GetClosestDumpsForBlobFunc: &CodeNbvServiceGetClosestDumpsForBlobFunc{
			defbultHook: i.GetClosestDumpsForBlob,
		},
		GetDibgnosticsFunc: &CodeNbvServiceGetDibgnosticsFunc{
			defbultHook: i.GetDibgnostics,
		},
		GetHoverFunc: &CodeNbvServiceGetHoverFunc{
			defbultHook: i.GetHover,
		},
		GetRbngesFunc: &CodeNbvServiceGetRbngesFunc{
			defbultHook: i.GetRbnges,
		},
		GetStencilFunc: &CodeNbvServiceGetStencilFunc{
			defbultHook: i.GetStencil,
		},
		NewGetDefinitionsFunc: &CodeNbvServiceNewGetDefinitionsFunc{
			defbultHook: i.NewGetDefinitions,
		},
		NewGetImplementbtionsFunc: &CodeNbvServiceNewGetImplementbtionsFunc{
			defbultHook: i.NewGetImplementbtions,
		},
		NewGetPrototypesFunc: &CodeNbvServiceNewGetPrototypesFunc{
			defbultHook: i.NewGetPrototypes,
		},
		NewGetReferencesFunc: &CodeNbvServiceNewGetReferencesFunc{
			defbultHook: i.NewGetReferences,
		},
		SnbpshotForDocumentFunc: &CodeNbvServiceSnbpshotForDocumentFunc{
			defbultHook: i.SnbpshotForDocument,
		},
		VisibleUplobdsForPbthFunc: &CodeNbvServiceVisibleUplobdsForPbthFunc{
			defbultHook: i.VisibleUplobdsForPbth,
		},
	}
}

// CodeNbvServiceGetClosestDumpsForBlobFunc describes the behbvior when the
// GetClosestDumpsForBlob method of the pbrent MockCodeNbvService instbnce
// is invoked.
type CodeNbvServiceGetClosestDumpsForBlobFunc struct {
	defbultHook func(context.Context, int, string, string, bool, string) ([]shbred.Dump, error)
	hooks       []func(context.Context, int, string, string, bool, string) ([]shbred.Dump, error)
	history     []CodeNbvServiceGetClosestDumpsForBlobFuncCbll
	mutex       sync.Mutex
}

// GetClosestDumpsForBlob delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCodeNbvService) GetClosestDumpsForBlob(v0 context.Context, v1 int, v2 string, v3 string, v4 bool, v5 string) ([]shbred.Dump, error) {
	r0, r1 := m.GetClosestDumpsForBlobFunc.nextHook()(v0, v1, v2, v3, v4, v5)
	m.GetClosestDumpsForBlobFunc.bppendCbll(CodeNbvServiceGetClosestDumpsForBlobFuncCbll{v0, v1, v2, v3, v4, v5, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetClosestDumpsForBlob method of the pbrent MockCodeNbvService instbnce
// is invoked bnd the hook queue is empty.
func (f *CodeNbvServiceGetClosestDumpsForBlobFunc) SetDefbultHook(hook func(context.Context, int, string, string, bool, string) ([]shbred.Dump, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetClosestDumpsForBlob method of the pbrent MockCodeNbvService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *CodeNbvServiceGetClosestDumpsForBlobFunc) PushHook(hook func(context.Context, int, string, string, bool, string) ([]shbred.Dump, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CodeNbvServiceGetClosestDumpsForBlobFunc) SetDefbultReturn(r0 []shbred.Dump, r1 error) {
	f.SetDefbultHook(func(context.Context, int, string, string, bool, string) ([]shbred.Dump, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CodeNbvServiceGetClosestDumpsForBlobFunc) PushReturn(r0 []shbred.Dump, r1 error) {
	f.PushHook(func(context.Context, int, string, string, bool, string) ([]shbred.Dump, error) {
		return r0, r1
	})
}

func (f *CodeNbvServiceGetClosestDumpsForBlobFunc) nextHook() func(context.Context, int, string, string, bool, string) ([]shbred.Dump, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CodeNbvServiceGetClosestDumpsForBlobFunc) bppendCbll(r0 CodeNbvServiceGetClosestDumpsForBlobFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// CodeNbvServiceGetClosestDumpsForBlobFuncCbll objects describing the
// invocbtions of this function.
func (f *CodeNbvServiceGetClosestDumpsForBlobFunc) History() []CodeNbvServiceGetClosestDumpsForBlobFuncCbll {
	f.mutex.Lock()
	history := mbke([]CodeNbvServiceGetClosestDumpsForBlobFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CodeNbvServiceGetClosestDumpsForBlobFuncCbll is bn object thbt describes
// bn invocbtion of method GetClosestDumpsForBlob on bn instbnce of
// MockCodeNbvService.
type CodeNbvServiceGetClosestDumpsForBlobFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 bool
	// Arg5 is the vblue of the 6th brgument pbssed to this method
	// invocbtion.
	Arg5 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.Dump
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c CodeNbvServiceGetClosestDumpsForBlobFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CodeNbvServiceGetClosestDumpsForBlobFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// CodeNbvServiceGetDibgnosticsFunc describes the behbvior when the
// GetDibgnostics method of the pbrent MockCodeNbvService instbnce is
// invoked.
type CodeNbvServiceGetDibgnosticsFunc struct {
	defbultHook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]codenbv.DibgnosticAtUplobd, int, error)
	hooks       []func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]codenbv.DibgnosticAtUplobd, int, error)
	history     []CodeNbvServiceGetDibgnosticsFuncCbll
	mutex       sync.Mutex
}

// GetDibgnostics delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCodeNbvService) GetDibgnostics(v0 context.Context, v1 codenbv.PositionblRequestArgs, v2 codenbv.RequestStbte) ([]codenbv.DibgnosticAtUplobd, int, error) {
	r0, r1, r2 := m.GetDibgnosticsFunc.nextHook()(v0, v1, v2)
	m.GetDibgnosticsFunc.bppendCbll(CodeNbvServiceGetDibgnosticsFuncCbll{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the GetDibgnostics
// method of the pbrent MockCodeNbvService instbnce is invoked bnd the hook
// queue is empty.
func (f *CodeNbvServiceGetDibgnosticsFunc) SetDefbultHook(hook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]codenbv.DibgnosticAtUplobd, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetDibgnostics method of the pbrent MockCodeNbvService instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *CodeNbvServiceGetDibgnosticsFunc) PushHook(hook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]codenbv.DibgnosticAtUplobd, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CodeNbvServiceGetDibgnosticsFunc) SetDefbultReturn(r0 []codenbv.DibgnosticAtUplobd, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]codenbv.DibgnosticAtUplobd, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CodeNbvServiceGetDibgnosticsFunc) PushReturn(r0 []codenbv.DibgnosticAtUplobd, r1 int, r2 error) {
	f.PushHook(func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]codenbv.DibgnosticAtUplobd, int, error) {
		return r0, r1, r2
	})
}

func (f *CodeNbvServiceGetDibgnosticsFunc) nextHook() func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]codenbv.DibgnosticAtUplobd, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CodeNbvServiceGetDibgnosticsFunc) bppendCbll(r0 CodeNbvServiceGetDibgnosticsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of CodeNbvServiceGetDibgnosticsFuncCbll
// objects describing the invocbtions of this function.
func (f *CodeNbvServiceGetDibgnosticsFunc) History() []CodeNbvServiceGetDibgnosticsFuncCbll {
	f.mutex.Lock()
	history := mbke([]CodeNbvServiceGetDibgnosticsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CodeNbvServiceGetDibgnosticsFuncCbll is bn object thbt describes bn
// invocbtion of method GetDibgnostics on bn instbnce of MockCodeNbvService.
type CodeNbvServiceGetDibgnosticsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 codenbv.PositionblRequestArgs
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 codenbv.RequestStbte
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []codenbv.DibgnosticAtUplobd
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c CodeNbvServiceGetDibgnosticsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CodeNbvServiceGetDibgnosticsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// CodeNbvServiceGetHoverFunc describes the behbvior when the GetHover
// method of the pbrent MockCodeNbvService instbnce is invoked.
type CodeNbvServiceGetHoverFunc struct {
	defbultHook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) (string, shbred1.Rbnge, bool, error)
	hooks       []func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) (string, shbred1.Rbnge, bool, error)
	history     []CodeNbvServiceGetHoverFuncCbll
	mutex       sync.Mutex
}

// GetHover delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCodeNbvService) GetHover(v0 context.Context, v1 codenbv.PositionblRequestArgs, v2 codenbv.RequestStbte) (string, shbred1.Rbnge, bool, error) {
	r0, r1, r2, r3 := m.GetHoverFunc.nextHook()(v0, v1, v2)
	m.GetHoverFunc.bppendCbll(CodeNbvServiceGetHoverFuncCbll{v0, v1, v2, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefbultHook sets function thbt is cblled when the GetHover method of
// the pbrent MockCodeNbvService instbnce is invoked bnd the hook queue is
// empty.
func (f *CodeNbvServiceGetHoverFunc) SetDefbultHook(hook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) (string, shbred1.Rbnge, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetHover method of the pbrent MockCodeNbvService instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *CodeNbvServiceGetHoverFunc) PushHook(hook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) (string, shbred1.Rbnge, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CodeNbvServiceGetHoverFunc) SetDefbultReturn(r0 string, r1 shbred1.Rbnge, r2 bool, r3 error) {
	f.SetDefbultHook(func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) (string, shbred1.Rbnge, bool, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CodeNbvServiceGetHoverFunc) PushReturn(r0 string, r1 shbred1.Rbnge, r2 bool, r3 error) {
	f.PushHook(func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) (string, shbred1.Rbnge, bool, error) {
		return r0, r1, r2, r3
	})
}

func (f *CodeNbvServiceGetHoverFunc) nextHook() func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) (string, shbred1.Rbnge, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CodeNbvServiceGetHoverFunc) bppendCbll(r0 CodeNbvServiceGetHoverFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of CodeNbvServiceGetHoverFuncCbll objects
// describing the invocbtions of this function.
func (f *CodeNbvServiceGetHoverFunc) History() []CodeNbvServiceGetHoverFuncCbll {
	f.mutex.Lock()
	history := mbke([]CodeNbvServiceGetHoverFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CodeNbvServiceGetHoverFuncCbll is bn object thbt describes bn invocbtion
// of method GetHover on bn instbnce of MockCodeNbvService.
type CodeNbvServiceGetHoverFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 codenbv.PositionblRequestArgs
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 codenbv.RequestStbte
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 shbred1.Rbnge
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 bool
	// Result3 is the vblue of the 4th result returned from this method
	// invocbtion.
	Result3 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c CodeNbvServiceGetHoverFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CodeNbvServiceGetHoverFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2, c.Result3}
}

// CodeNbvServiceGetRbngesFunc describes the behbvior when the GetRbnges
// method of the pbrent MockCodeNbvService instbnce is invoked.
type CodeNbvServiceGetRbngesFunc struct {
	defbultHook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, int, int) ([]codenbv.AdjustedCodeIntelligenceRbnge, error)
	hooks       []func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, int, int) ([]codenbv.AdjustedCodeIntelligenceRbnge, error)
	history     []CodeNbvServiceGetRbngesFuncCbll
	mutex       sync.Mutex
}

// GetRbnges delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCodeNbvService) GetRbnges(v0 context.Context, v1 codenbv.PositionblRequestArgs, v2 codenbv.RequestStbte, v3 int, v4 int) ([]codenbv.AdjustedCodeIntelligenceRbnge, error) {
	r0, r1 := m.GetRbngesFunc.nextHook()(v0, v1, v2, v3, v4)
	m.GetRbngesFunc.bppendCbll(CodeNbvServiceGetRbngesFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetRbnges method of
// the pbrent MockCodeNbvService instbnce is invoked bnd the hook queue is
// empty.
func (f *CodeNbvServiceGetRbngesFunc) SetDefbultHook(hook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, int, int) ([]codenbv.AdjustedCodeIntelligenceRbnge, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetRbnges method of the pbrent MockCodeNbvService instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *CodeNbvServiceGetRbngesFunc) PushHook(hook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, int, int) ([]codenbv.AdjustedCodeIntelligenceRbnge, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CodeNbvServiceGetRbngesFunc) SetDefbultReturn(r0 []codenbv.AdjustedCodeIntelligenceRbnge, r1 error) {
	f.SetDefbultHook(func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, int, int) ([]codenbv.AdjustedCodeIntelligenceRbnge, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CodeNbvServiceGetRbngesFunc) PushReturn(r0 []codenbv.AdjustedCodeIntelligenceRbnge, r1 error) {
	f.PushHook(func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, int, int) ([]codenbv.AdjustedCodeIntelligenceRbnge, error) {
		return r0, r1
	})
}

func (f *CodeNbvServiceGetRbngesFunc) nextHook() func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, int, int) ([]codenbv.AdjustedCodeIntelligenceRbnge, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CodeNbvServiceGetRbngesFunc) bppendCbll(r0 CodeNbvServiceGetRbngesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of CodeNbvServiceGetRbngesFuncCbll objects
// describing the invocbtions of this function.
func (f *CodeNbvServiceGetRbngesFunc) History() []CodeNbvServiceGetRbngesFuncCbll {
	f.mutex.Lock()
	history := mbke([]CodeNbvServiceGetRbngesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CodeNbvServiceGetRbngesFuncCbll is bn object thbt describes bn invocbtion
// of method GetRbnges on bn instbnce of MockCodeNbvService.
type CodeNbvServiceGetRbngesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 codenbv.PositionblRequestArgs
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 codenbv.RequestStbte
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 int
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []codenbv.AdjustedCodeIntelligenceRbnge
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c CodeNbvServiceGetRbngesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CodeNbvServiceGetRbngesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// CodeNbvServiceGetStencilFunc describes the behbvior when the GetStencil
// method of the pbrent MockCodeNbvService instbnce is invoked.
type CodeNbvServiceGetStencilFunc struct {
	defbultHook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]shbred1.Rbnge, error)
	hooks       []func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]shbred1.Rbnge, error)
	history     []CodeNbvServiceGetStencilFuncCbll
	mutex       sync.Mutex
}

// GetStencil delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCodeNbvService) GetStencil(v0 context.Context, v1 codenbv.PositionblRequestArgs, v2 codenbv.RequestStbte) ([]shbred1.Rbnge, error) {
	r0, r1 := m.GetStencilFunc.nextHook()(v0, v1, v2)
	m.GetStencilFunc.bppendCbll(CodeNbvServiceGetStencilFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetStencil method of
// the pbrent MockCodeNbvService instbnce is invoked bnd the hook queue is
// empty.
func (f *CodeNbvServiceGetStencilFunc) SetDefbultHook(hook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]shbred1.Rbnge, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetStencil method of the pbrent MockCodeNbvService instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *CodeNbvServiceGetStencilFunc) PushHook(hook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]shbred1.Rbnge, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CodeNbvServiceGetStencilFunc) SetDefbultReturn(r0 []shbred1.Rbnge, r1 error) {
	f.SetDefbultHook(func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]shbred1.Rbnge, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CodeNbvServiceGetStencilFunc) PushReturn(r0 []shbred1.Rbnge, r1 error) {
	f.PushHook(func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]shbred1.Rbnge, error) {
		return r0, r1
	})
}

func (f *CodeNbvServiceGetStencilFunc) nextHook() func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]shbred1.Rbnge, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CodeNbvServiceGetStencilFunc) bppendCbll(r0 CodeNbvServiceGetStencilFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of CodeNbvServiceGetStencilFuncCbll objects
// describing the invocbtions of this function.
func (f *CodeNbvServiceGetStencilFunc) History() []CodeNbvServiceGetStencilFuncCbll {
	f.mutex.Lock()
	history := mbke([]CodeNbvServiceGetStencilFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CodeNbvServiceGetStencilFuncCbll is bn object thbt describes bn
// invocbtion of method GetStencil on bn instbnce of MockCodeNbvService.
type CodeNbvServiceGetStencilFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 codenbv.PositionblRequestArgs
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 codenbv.RequestStbte
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred1.Rbnge
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c CodeNbvServiceGetStencilFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CodeNbvServiceGetStencilFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// CodeNbvServiceNewGetDefinitionsFunc describes the behbvior when the
// NewGetDefinitions method of the pbrent MockCodeNbvService instbnce is
// invoked.
type CodeNbvServiceNewGetDefinitionsFunc struct {
	defbultHook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]shbred1.UplobdLocbtion, error)
	hooks       []func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]shbred1.UplobdLocbtion, error)
	history     []CodeNbvServiceNewGetDefinitionsFuncCbll
	mutex       sync.Mutex
}

// NewGetDefinitions delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCodeNbvService) NewGetDefinitions(v0 context.Context, v1 codenbv.PositionblRequestArgs, v2 codenbv.RequestStbte) ([]shbred1.UplobdLocbtion, error) {
	r0, r1 := m.NewGetDefinitionsFunc.nextHook()(v0, v1, v2)
	m.NewGetDefinitionsFunc.bppendCbll(CodeNbvServiceNewGetDefinitionsFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the NewGetDefinitions
// method of the pbrent MockCodeNbvService instbnce is invoked bnd the hook
// queue is empty.
func (f *CodeNbvServiceNewGetDefinitionsFunc) SetDefbultHook(hook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]shbred1.UplobdLocbtion, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// NewGetDefinitions method of the pbrent MockCodeNbvService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *CodeNbvServiceNewGetDefinitionsFunc) PushHook(hook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]shbred1.UplobdLocbtion, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CodeNbvServiceNewGetDefinitionsFunc) SetDefbultReturn(r0 []shbred1.UplobdLocbtion, r1 error) {
	f.SetDefbultHook(func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]shbred1.UplobdLocbtion, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CodeNbvServiceNewGetDefinitionsFunc) PushReturn(r0 []shbred1.UplobdLocbtion, r1 error) {
	f.PushHook(func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]shbred1.UplobdLocbtion, error) {
		return r0, r1
	})
}

func (f *CodeNbvServiceNewGetDefinitionsFunc) nextHook() func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte) ([]shbred1.UplobdLocbtion, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CodeNbvServiceNewGetDefinitionsFunc) bppendCbll(r0 CodeNbvServiceNewGetDefinitionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of CodeNbvServiceNewGetDefinitionsFuncCbll
// objects describing the invocbtions of this function.
func (f *CodeNbvServiceNewGetDefinitionsFunc) History() []CodeNbvServiceNewGetDefinitionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]CodeNbvServiceNewGetDefinitionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CodeNbvServiceNewGetDefinitionsFuncCbll is bn object thbt describes bn
// invocbtion of method NewGetDefinitions on bn instbnce of
// MockCodeNbvService.
type CodeNbvServiceNewGetDefinitionsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 codenbv.PositionblRequestArgs
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 codenbv.RequestStbte
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred1.UplobdLocbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c CodeNbvServiceNewGetDefinitionsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CodeNbvServiceNewGetDefinitionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// CodeNbvServiceNewGetImplementbtionsFunc describes the behbvior when the
// NewGetImplementbtions method of the pbrent MockCodeNbvService instbnce is
// invoked.
type CodeNbvServiceNewGetImplementbtionsFunc struct {
	defbultHook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error)
	hooks       []func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error)
	history     []CodeNbvServiceNewGetImplementbtionsFuncCbll
	mutex       sync.Mutex
}

// NewGetImplementbtions delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCodeNbvService) NewGetImplementbtions(v0 context.Context, v1 codenbv.PositionblRequestArgs, v2 codenbv.RequestStbte, v3 codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error) {
	r0, r1, r2 := m.NewGetImplementbtionsFunc.nextHook()(v0, v1, v2, v3)
	m.NewGetImplementbtionsFunc.bppendCbll(CodeNbvServiceNewGetImplementbtionsFuncCbll{v0, v1, v2, v3, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// NewGetImplementbtions method of the pbrent MockCodeNbvService instbnce is
// invoked bnd the hook queue is empty.
func (f *CodeNbvServiceNewGetImplementbtionsFunc) SetDefbultHook(hook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// NewGetImplementbtions method of the pbrent MockCodeNbvService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *CodeNbvServiceNewGetImplementbtionsFunc) PushHook(hook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CodeNbvServiceNewGetImplementbtionsFunc) SetDefbultReturn(r0 []shbred1.UplobdLocbtion, r1 codenbv.Cursor, r2 error) {
	f.SetDefbultHook(func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CodeNbvServiceNewGetImplementbtionsFunc) PushReturn(r0 []shbred1.UplobdLocbtion, r1 codenbv.Cursor, r2 error) {
	f.PushHook(func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error) {
		return r0, r1, r2
	})
}

func (f *CodeNbvServiceNewGetImplementbtionsFunc) nextHook() func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CodeNbvServiceNewGetImplementbtionsFunc) bppendCbll(r0 CodeNbvServiceNewGetImplementbtionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of CodeNbvServiceNewGetImplementbtionsFuncCbll
// objects describing the invocbtions of this function.
func (f *CodeNbvServiceNewGetImplementbtionsFunc) History() []CodeNbvServiceNewGetImplementbtionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]CodeNbvServiceNewGetImplementbtionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CodeNbvServiceNewGetImplementbtionsFuncCbll is bn object thbt describes
// bn invocbtion of method NewGetImplementbtions on bn instbnce of
// MockCodeNbvService.
type CodeNbvServiceNewGetImplementbtionsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 codenbv.PositionblRequestArgs
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 codenbv.RequestStbte
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 codenbv.Cursor
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred1.UplobdLocbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 codenbv.Cursor
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c CodeNbvServiceNewGetImplementbtionsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CodeNbvServiceNewGetImplementbtionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// CodeNbvServiceNewGetPrototypesFunc describes the behbvior when the
// NewGetPrototypes method of the pbrent MockCodeNbvService instbnce is
// invoked.
type CodeNbvServiceNewGetPrototypesFunc struct {
	defbultHook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error)
	hooks       []func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error)
	history     []CodeNbvServiceNewGetPrototypesFuncCbll
	mutex       sync.Mutex
}

// NewGetPrototypes delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCodeNbvService) NewGetPrototypes(v0 context.Context, v1 codenbv.PositionblRequestArgs, v2 codenbv.RequestStbte, v3 codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error) {
	r0, r1, r2 := m.NewGetPrototypesFunc.nextHook()(v0, v1, v2, v3)
	m.NewGetPrototypesFunc.bppendCbll(CodeNbvServiceNewGetPrototypesFuncCbll{v0, v1, v2, v3, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the NewGetPrototypes
// method of the pbrent MockCodeNbvService instbnce is invoked bnd the hook
// queue is empty.
func (f *CodeNbvServiceNewGetPrototypesFunc) SetDefbultHook(hook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// NewGetPrototypes method of the pbrent MockCodeNbvService instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *CodeNbvServiceNewGetPrototypesFunc) PushHook(hook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CodeNbvServiceNewGetPrototypesFunc) SetDefbultReturn(r0 []shbred1.UplobdLocbtion, r1 codenbv.Cursor, r2 error) {
	f.SetDefbultHook(func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CodeNbvServiceNewGetPrototypesFunc) PushReturn(r0 []shbred1.UplobdLocbtion, r1 codenbv.Cursor, r2 error) {
	f.PushHook(func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error) {
		return r0, r1, r2
	})
}

func (f *CodeNbvServiceNewGetPrototypesFunc) nextHook() func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CodeNbvServiceNewGetPrototypesFunc) bppendCbll(r0 CodeNbvServiceNewGetPrototypesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of CodeNbvServiceNewGetPrototypesFuncCbll
// objects describing the invocbtions of this function.
func (f *CodeNbvServiceNewGetPrototypesFunc) History() []CodeNbvServiceNewGetPrototypesFuncCbll {
	f.mutex.Lock()
	history := mbke([]CodeNbvServiceNewGetPrototypesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CodeNbvServiceNewGetPrototypesFuncCbll is bn object thbt describes bn
// invocbtion of method NewGetPrototypes on bn instbnce of
// MockCodeNbvService.
type CodeNbvServiceNewGetPrototypesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 codenbv.PositionblRequestArgs
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 codenbv.RequestStbte
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 codenbv.Cursor
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred1.UplobdLocbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 codenbv.Cursor
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c CodeNbvServiceNewGetPrototypesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CodeNbvServiceNewGetPrototypesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// CodeNbvServiceNewGetReferencesFunc describes the behbvior when the
// NewGetReferences method of the pbrent MockCodeNbvService instbnce is
// invoked.
type CodeNbvServiceNewGetReferencesFunc struct {
	defbultHook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error)
	hooks       []func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error)
	history     []CodeNbvServiceNewGetReferencesFuncCbll
	mutex       sync.Mutex
}

// NewGetReferences delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCodeNbvService) NewGetReferences(v0 context.Context, v1 codenbv.PositionblRequestArgs, v2 codenbv.RequestStbte, v3 codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error) {
	r0, r1, r2 := m.NewGetReferencesFunc.nextHook()(v0, v1, v2, v3)
	m.NewGetReferencesFunc.bppendCbll(CodeNbvServiceNewGetReferencesFuncCbll{v0, v1, v2, v3, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the NewGetReferences
// method of the pbrent MockCodeNbvService instbnce is invoked bnd the hook
// queue is empty.
func (f *CodeNbvServiceNewGetReferencesFunc) SetDefbultHook(hook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// NewGetReferences method of the pbrent MockCodeNbvService instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *CodeNbvServiceNewGetReferencesFunc) PushHook(hook func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CodeNbvServiceNewGetReferencesFunc) SetDefbultReturn(r0 []shbred1.UplobdLocbtion, r1 codenbv.Cursor, r2 error) {
	f.SetDefbultHook(func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CodeNbvServiceNewGetReferencesFunc) PushReturn(r0 []shbred1.UplobdLocbtion, r1 codenbv.Cursor, r2 error) {
	f.PushHook(func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error) {
		return r0, r1, r2
	})
}

func (f *CodeNbvServiceNewGetReferencesFunc) nextHook() func(context.Context, codenbv.PositionblRequestArgs, codenbv.RequestStbte, codenbv.Cursor) ([]shbred1.UplobdLocbtion, codenbv.Cursor, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CodeNbvServiceNewGetReferencesFunc) bppendCbll(r0 CodeNbvServiceNewGetReferencesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of CodeNbvServiceNewGetReferencesFuncCbll
// objects describing the invocbtions of this function.
func (f *CodeNbvServiceNewGetReferencesFunc) History() []CodeNbvServiceNewGetReferencesFuncCbll {
	f.mutex.Lock()
	history := mbke([]CodeNbvServiceNewGetReferencesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CodeNbvServiceNewGetReferencesFuncCbll is bn object thbt describes bn
// invocbtion of method NewGetReferences on bn instbnce of
// MockCodeNbvService.
type CodeNbvServiceNewGetReferencesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 codenbv.PositionblRequestArgs
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 codenbv.RequestStbte
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 codenbv.Cursor
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred1.UplobdLocbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 codenbv.Cursor
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c CodeNbvServiceNewGetReferencesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CodeNbvServiceNewGetReferencesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// CodeNbvServiceSnbpshotForDocumentFunc describes the behbvior when the
// SnbpshotForDocument method of the pbrent MockCodeNbvService instbnce is
// invoked.
type CodeNbvServiceSnbpshotForDocumentFunc struct {
	defbultHook func(context.Context, int, string, string, int) ([]shbred1.SnbpshotDbtb, error)
	hooks       []func(context.Context, int, string, string, int) ([]shbred1.SnbpshotDbtb, error)
	history     []CodeNbvServiceSnbpshotForDocumentFuncCbll
	mutex       sync.Mutex
}

// SnbpshotForDocument delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCodeNbvService) SnbpshotForDocument(v0 context.Context, v1 int, v2 string, v3 string, v4 int) ([]shbred1.SnbpshotDbtb, error) {
	r0, r1 := m.SnbpshotForDocumentFunc.nextHook()(v0, v1, v2, v3, v4)
	m.SnbpshotForDocumentFunc.bppendCbll(CodeNbvServiceSnbpshotForDocumentFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the SnbpshotForDocument
// method of the pbrent MockCodeNbvService instbnce is invoked bnd the hook
// queue is empty.
func (f *CodeNbvServiceSnbpshotForDocumentFunc) SetDefbultHook(hook func(context.Context, int, string, string, int) ([]shbred1.SnbpshotDbtb, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SnbpshotForDocument method of the pbrent MockCodeNbvService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *CodeNbvServiceSnbpshotForDocumentFunc) PushHook(hook func(context.Context, int, string, string, int) ([]shbred1.SnbpshotDbtb, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CodeNbvServiceSnbpshotForDocumentFunc) SetDefbultReturn(r0 []shbred1.SnbpshotDbtb, r1 error) {
	f.SetDefbultHook(func(context.Context, int, string, string, int) ([]shbred1.SnbpshotDbtb, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CodeNbvServiceSnbpshotForDocumentFunc) PushReturn(r0 []shbred1.SnbpshotDbtb, r1 error) {
	f.PushHook(func(context.Context, int, string, string, int) ([]shbred1.SnbpshotDbtb, error) {
		return r0, r1
	})
}

func (f *CodeNbvServiceSnbpshotForDocumentFunc) nextHook() func(context.Context, int, string, string, int) ([]shbred1.SnbpshotDbtb, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CodeNbvServiceSnbpshotForDocumentFunc) bppendCbll(r0 CodeNbvServiceSnbpshotForDocumentFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of CodeNbvServiceSnbpshotForDocumentFuncCbll
// objects describing the invocbtions of this function.
func (f *CodeNbvServiceSnbpshotForDocumentFunc) History() []CodeNbvServiceSnbpshotForDocumentFuncCbll {
	f.mutex.Lock()
	history := mbke([]CodeNbvServiceSnbpshotForDocumentFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CodeNbvServiceSnbpshotForDocumentFuncCbll is bn object thbt describes bn
// invocbtion of method SnbpshotForDocument on bn instbnce of
// MockCodeNbvService.
type CodeNbvServiceSnbpshotForDocumentFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 string
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred1.SnbpshotDbtb
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c CodeNbvServiceSnbpshotForDocumentFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CodeNbvServiceSnbpshotForDocumentFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// CodeNbvServiceVisibleUplobdsForPbthFunc describes the behbvior when the
// VisibleUplobdsForPbth method of the pbrent MockCodeNbvService instbnce is
// invoked.
type CodeNbvServiceVisibleUplobdsForPbthFunc struct {
	defbultHook func(context.Context, codenbv.RequestStbte) ([]shbred.Dump, error)
	hooks       []func(context.Context, codenbv.RequestStbte) ([]shbred.Dump, error)
	history     []CodeNbvServiceVisibleUplobdsForPbthFuncCbll
	mutex       sync.Mutex
}

// VisibleUplobdsForPbth delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockCodeNbvService) VisibleUplobdsForPbth(v0 context.Context, v1 codenbv.RequestStbte) ([]shbred.Dump, error) {
	r0, r1 := m.VisibleUplobdsForPbthFunc.nextHook()(v0, v1)
	m.VisibleUplobdsForPbthFunc.bppendCbll(CodeNbvServiceVisibleUplobdsForPbthFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// VisibleUplobdsForPbth method of the pbrent MockCodeNbvService instbnce is
// invoked bnd the hook queue is empty.
func (f *CodeNbvServiceVisibleUplobdsForPbthFunc) SetDefbultHook(hook func(context.Context, codenbv.RequestStbte) ([]shbred.Dump, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// VisibleUplobdsForPbth method of the pbrent MockCodeNbvService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *CodeNbvServiceVisibleUplobdsForPbthFunc) PushHook(hook func(context.Context, codenbv.RequestStbte) ([]shbred.Dump, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *CodeNbvServiceVisibleUplobdsForPbthFunc) SetDefbultReturn(r0 []shbred.Dump, r1 error) {
	f.SetDefbultHook(func(context.Context, codenbv.RequestStbte) ([]shbred.Dump, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *CodeNbvServiceVisibleUplobdsForPbthFunc) PushReturn(r0 []shbred.Dump, r1 error) {
	f.PushHook(func(context.Context, codenbv.RequestStbte) ([]shbred.Dump, error) {
		return r0, r1
	})
}

func (f *CodeNbvServiceVisibleUplobdsForPbthFunc) nextHook() func(context.Context, codenbv.RequestStbte) ([]shbred.Dump, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CodeNbvServiceVisibleUplobdsForPbthFunc) bppendCbll(r0 CodeNbvServiceVisibleUplobdsForPbthFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of CodeNbvServiceVisibleUplobdsForPbthFuncCbll
// objects describing the invocbtions of this function.
func (f *CodeNbvServiceVisibleUplobdsForPbthFunc) History() []CodeNbvServiceVisibleUplobdsForPbthFuncCbll {
	f.mutex.Lock()
	history := mbke([]CodeNbvServiceVisibleUplobdsForPbthFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CodeNbvServiceVisibleUplobdsForPbthFuncCbll is bn object thbt describes
// bn invocbtion of method VisibleUplobdsForPbth on bn instbnce of
// MockCodeNbvService.
type CodeNbvServiceVisibleUplobdsForPbthFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 codenbv.RequestStbte
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.Dump
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c CodeNbvServiceVisibleUplobdsForPbthFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c CodeNbvServiceVisibleUplobdsForPbthFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}
