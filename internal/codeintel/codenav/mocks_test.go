// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge codenbv

import (
	"context"
	"sync"

	scip "github.com/sourcegrbph/scip/bindings/go/scip"
	lsifstore "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/internbl/lsifstore"
	shbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	shbred1 "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	precise "github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

// MockLsifStore is b mock implementbtion of the LsifStore interfbce (from
// the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/internbl/lsifstore)
// used for unit testing.
type MockLsifStore struct {
	// ExtrbctDefinitionLocbtionsFromPositionFunc is bn instbnce of b mock
	// function object controlling the behbvior of the method
	// ExtrbctDefinitionLocbtionsFromPosition.
	ExtrbctDefinitionLocbtionsFromPositionFunc *LsifStoreExtrbctDefinitionLocbtionsFromPositionFunc
	// ExtrbctImplementbtionLocbtionsFromPositionFunc is bn instbnce of b
	// mock function object controlling the behbvior of the method
	// ExtrbctImplementbtionLocbtionsFromPosition.
	ExtrbctImplementbtionLocbtionsFromPositionFunc *LsifStoreExtrbctImplementbtionLocbtionsFromPositionFunc
	// ExtrbctPrototypeLocbtionsFromPositionFunc is bn instbnce of b mock
	// function object controlling the behbvior of the method
	// ExtrbctPrototypeLocbtionsFromPosition.
	ExtrbctPrototypeLocbtionsFromPositionFunc *LsifStoreExtrbctPrototypeLocbtionsFromPositionFunc
	// ExtrbctReferenceLocbtionsFromPositionFunc is bn instbnce of b mock
	// function object controlling the behbvior of the method
	// ExtrbctReferenceLocbtionsFromPosition.
	ExtrbctReferenceLocbtionsFromPositionFunc *LsifStoreExtrbctReferenceLocbtionsFromPositionFunc
	// GetBulkMonikerLocbtionsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetBulkMonikerLocbtions.
	GetBulkMonikerLocbtionsFunc *LsifStoreGetBulkMonikerLocbtionsFunc
	// GetDefinitionLocbtionsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetDefinitionLocbtions.
	GetDefinitionLocbtionsFunc *LsifStoreGetDefinitionLocbtionsFunc
	// GetDibgnosticsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetDibgnostics.
	GetDibgnosticsFunc *LsifStoreGetDibgnosticsFunc
	// GetHoverFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method GetHover.
	GetHoverFunc *LsifStoreGetHoverFunc
	// GetImplementbtionLocbtionsFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// GetImplementbtionLocbtions.
	GetImplementbtionLocbtionsFunc *LsifStoreGetImplementbtionLocbtionsFunc
	// GetMinimblBulkMonikerLocbtionsFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// GetMinimblBulkMonikerLocbtions.
	GetMinimblBulkMonikerLocbtionsFunc *LsifStoreGetMinimblBulkMonikerLocbtionsFunc
	// GetMonikersByPositionFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetMonikersByPosition.
	GetMonikersByPositionFunc *LsifStoreGetMonikersByPositionFunc
	// GetPbckbgeInformbtionFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetPbckbgeInformbtion.
	GetPbckbgeInformbtionFunc *LsifStoreGetPbckbgeInformbtionFunc
	// GetPbthExistsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetPbthExists.
	GetPbthExistsFunc *LsifStoreGetPbthExistsFunc
	// GetPrototypeLocbtionsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetPrototypeLocbtions.
	GetPrototypeLocbtionsFunc *LsifStoreGetPrototypeLocbtionsFunc
	// GetRbngesFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetRbnges.
	GetRbngesFunc *LsifStoreGetRbngesFunc
	// GetReferenceLocbtionsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetReferenceLocbtions.
	GetReferenceLocbtionsFunc *LsifStoreGetReferenceLocbtionsFunc
	// GetStencilFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetStencil.
	GetStencilFunc *LsifStoreGetStencilFunc
	// SCIPDocumentFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method SCIPDocument.
	SCIPDocumentFunc *LsifStoreSCIPDocumentFunc
}

// NewMockLsifStore crebtes b new mock of the LsifStore interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockLsifStore() *MockLsifStore {
	return &MockLsifStore{
		ExtrbctDefinitionLocbtionsFromPositionFunc: &LsifStoreExtrbctDefinitionLocbtionsFromPositionFunc{
			defbultHook: func(context.Context, lsifstore.LocbtionKey) (r0 []shbred.Locbtion, r1 []string, r2 error) {
				return
			},
		},
		ExtrbctImplementbtionLocbtionsFromPositionFunc: &LsifStoreExtrbctImplementbtionLocbtionsFromPositionFunc{
			defbultHook: func(context.Context, lsifstore.LocbtionKey) (r0 []shbred.Locbtion, r1 []string, r2 error) {
				return
			},
		},
		ExtrbctPrototypeLocbtionsFromPositionFunc: &LsifStoreExtrbctPrototypeLocbtionsFromPositionFunc{
			defbultHook: func(context.Context, lsifstore.LocbtionKey) (r0 []shbred.Locbtion, r1 []string, r2 error) {
				return
			},
		},
		ExtrbctReferenceLocbtionsFromPositionFunc: &LsifStoreExtrbctReferenceLocbtionsFromPositionFunc{
			defbultHook: func(context.Context, lsifstore.LocbtionKey) (r0 []shbred.Locbtion, r1 []string, r2 error) {
				return
			},
		},
		GetBulkMonikerLocbtionsFunc: &LsifStoreGetBulkMonikerLocbtionsFunc{
			defbultHook: func(context.Context, string, []int, []precise.MonikerDbtb, int, int) (r0 []shbred.Locbtion, r1 int, r2 error) {
				return
			},
		},
		GetDefinitionLocbtionsFunc: &LsifStoreGetDefinitionLocbtionsFunc{
			defbultHook: func(context.Context, int, string, int, int, int, int) (r0 []shbred.Locbtion, r1 int, r2 error) {
				return
			},
		},
		GetDibgnosticsFunc: &LsifStoreGetDibgnosticsFunc{
			defbultHook: func(context.Context, int, string, int, int) (r0 []shbred.Dibgnostic, r1 int, r2 error) {
				return
			},
		},
		GetHoverFunc: &LsifStoreGetHoverFunc{
			defbultHook: func(context.Context, int, string, int, int) (r0 string, r1 shbred.Rbnge, r2 bool, r3 error) {
				return
			},
		},
		GetImplementbtionLocbtionsFunc: &LsifStoreGetImplementbtionLocbtionsFunc{
			defbultHook: func(context.Context, int, string, int, int, int, int) (r0 []shbred.Locbtion, r1 int, r2 error) {
				return
			},
		},
		GetMinimblBulkMonikerLocbtionsFunc: &LsifStoreGetMinimblBulkMonikerLocbtionsFunc{
			defbultHook: func(context.Context, string, []int, mbp[int]string, []precise.MonikerDbtb, int, int) (r0 []shbred.Locbtion, r1 int, r2 error) {
				return
			},
		},
		GetMonikersByPositionFunc: &LsifStoreGetMonikersByPositionFunc{
			defbultHook: func(context.Context, int, string, int, int) (r0 [][]precise.MonikerDbtb, r1 error) {
				return
			},
		},
		GetPbckbgeInformbtionFunc: &LsifStoreGetPbckbgeInformbtionFunc{
			defbultHook: func(context.Context, int, string, string) (r0 precise.PbckbgeInformbtionDbtb, r1 bool, r2 error) {
				return
			},
		},
		GetPbthExistsFunc: &LsifStoreGetPbthExistsFunc{
			defbultHook: func(context.Context, int, string) (r0 bool, r1 error) {
				return
			},
		},
		GetPrototypeLocbtionsFunc: &LsifStoreGetPrototypeLocbtionsFunc{
			defbultHook: func(context.Context, int, string, int, int, int, int) (r0 []shbred.Locbtion, r1 int, r2 error) {
				return
			},
		},
		GetRbngesFunc: &LsifStoreGetRbngesFunc{
			defbultHook: func(context.Context, int, string, int, int) (r0 []shbred.CodeIntelligenceRbnge, r1 error) {
				return
			},
		},
		GetReferenceLocbtionsFunc: &LsifStoreGetReferenceLocbtionsFunc{
			defbultHook: func(context.Context, int, string, int, int, int, int) (r0 []shbred.Locbtion, r1 int, r2 error) {
				return
			},
		},
		GetStencilFunc: &LsifStoreGetStencilFunc{
			defbultHook: func(context.Context, int, string) (r0 []shbred.Rbnge, r1 error) {
				return
			},
		},
		SCIPDocumentFunc: &LsifStoreSCIPDocumentFunc{
			defbultHook: func(context.Context, int, string) (r0 *scip.Document, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockLsifStore crebtes b new mock of the LsifStore interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockLsifStore() *MockLsifStore {
	return &MockLsifStore{
		ExtrbctDefinitionLocbtionsFromPositionFunc: &LsifStoreExtrbctDefinitionLocbtionsFromPositionFunc{
			defbultHook: func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
				pbnic("unexpected invocbtion of MockLsifStore.ExtrbctDefinitionLocbtionsFromPosition")
			},
		},
		ExtrbctImplementbtionLocbtionsFromPositionFunc: &LsifStoreExtrbctImplementbtionLocbtionsFromPositionFunc{
			defbultHook: func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
				pbnic("unexpected invocbtion of MockLsifStore.ExtrbctImplementbtionLocbtionsFromPosition")
			},
		},
		ExtrbctPrototypeLocbtionsFromPositionFunc: &LsifStoreExtrbctPrototypeLocbtionsFromPositionFunc{
			defbultHook: func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
				pbnic("unexpected invocbtion of MockLsifStore.ExtrbctPrototypeLocbtionsFromPosition")
			},
		},
		ExtrbctReferenceLocbtionsFromPositionFunc: &LsifStoreExtrbctReferenceLocbtionsFromPositionFunc{
			defbultHook: func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
				pbnic("unexpected invocbtion of MockLsifStore.ExtrbctReferenceLocbtionsFromPosition")
			},
		},
		GetBulkMonikerLocbtionsFunc: &LsifStoreGetBulkMonikerLocbtionsFunc{
			defbultHook: func(context.Context, string, []int, []precise.MonikerDbtb, int, int) ([]shbred.Locbtion, int, error) {
				pbnic("unexpected invocbtion of MockLsifStore.GetBulkMonikerLocbtions")
			},
		},
		GetDefinitionLocbtionsFunc: &LsifStoreGetDefinitionLocbtionsFunc{
			defbultHook: func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error) {
				pbnic("unexpected invocbtion of MockLsifStore.GetDefinitionLocbtions")
			},
		},
		GetDibgnosticsFunc: &LsifStoreGetDibgnosticsFunc{
			defbultHook: func(context.Context, int, string, int, int) ([]shbred.Dibgnostic, int, error) {
				pbnic("unexpected invocbtion of MockLsifStore.GetDibgnostics")
			},
		},
		GetHoverFunc: &LsifStoreGetHoverFunc{
			defbultHook: func(context.Context, int, string, int, int) (string, shbred.Rbnge, bool, error) {
				pbnic("unexpected invocbtion of MockLsifStore.GetHover")
			},
		},
		GetImplementbtionLocbtionsFunc: &LsifStoreGetImplementbtionLocbtionsFunc{
			defbultHook: func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error) {
				pbnic("unexpected invocbtion of MockLsifStore.GetImplementbtionLocbtions")
			},
		},
		GetMinimblBulkMonikerLocbtionsFunc: &LsifStoreGetMinimblBulkMonikerLocbtionsFunc{
			defbultHook: func(context.Context, string, []int, mbp[int]string, []precise.MonikerDbtb, int, int) ([]shbred.Locbtion, int, error) {
				pbnic("unexpected invocbtion of MockLsifStore.GetMinimblBulkMonikerLocbtions")
			},
		},
		GetMonikersByPositionFunc: &LsifStoreGetMonikersByPositionFunc{
			defbultHook: func(context.Context, int, string, int, int) ([][]precise.MonikerDbtb, error) {
				pbnic("unexpected invocbtion of MockLsifStore.GetMonikersByPosition")
			},
		},
		GetPbckbgeInformbtionFunc: &LsifStoreGetPbckbgeInformbtionFunc{
			defbultHook: func(context.Context, int, string, string) (precise.PbckbgeInformbtionDbtb, bool, error) {
				pbnic("unexpected invocbtion of MockLsifStore.GetPbckbgeInformbtion")
			},
		},
		GetPbthExistsFunc: &LsifStoreGetPbthExistsFunc{
			defbultHook: func(context.Context, int, string) (bool, error) {
				pbnic("unexpected invocbtion of MockLsifStore.GetPbthExists")
			},
		},
		GetPrototypeLocbtionsFunc: &LsifStoreGetPrototypeLocbtionsFunc{
			defbultHook: func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error) {
				pbnic("unexpected invocbtion of MockLsifStore.GetPrototypeLocbtions")
			},
		},
		GetRbngesFunc: &LsifStoreGetRbngesFunc{
			defbultHook: func(context.Context, int, string, int, int) ([]shbred.CodeIntelligenceRbnge, error) {
				pbnic("unexpected invocbtion of MockLsifStore.GetRbnges")
			},
		},
		GetReferenceLocbtionsFunc: &LsifStoreGetReferenceLocbtionsFunc{
			defbultHook: func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error) {
				pbnic("unexpected invocbtion of MockLsifStore.GetReferenceLocbtions")
			},
		},
		GetStencilFunc: &LsifStoreGetStencilFunc{
			defbultHook: func(context.Context, int, string) ([]shbred.Rbnge, error) {
				pbnic("unexpected invocbtion of MockLsifStore.GetStencil")
			},
		},
		SCIPDocumentFunc: &LsifStoreSCIPDocumentFunc{
			defbultHook: func(context.Context, int, string) (*scip.Document, error) {
				pbnic("unexpected invocbtion of MockLsifStore.SCIPDocument")
			},
		},
	}
}

// NewMockLsifStoreFrom crebtes b new mock of the MockLsifStore interfbce.
// All methods delegbte to the given implementbtion, unless overwritten.
func NewMockLsifStoreFrom(i lsifstore.LsifStore) *MockLsifStore {
	return &MockLsifStore{
		ExtrbctDefinitionLocbtionsFromPositionFunc: &LsifStoreExtrbctDefinitionLocbtionsFromPositionFunc{
			defbultHook: i.ExtrbctDefinitionLocbtionsFromPosition,
		},
		ExtrbctImplementbtionLocbtionsFromPositionFunc: &LsifStoreExtrbctImplementbtionLocbtionsFromPositionFunc{
			defbultHook: i.ExtrbctImplementbtionLocbtionsFromPosition,
		},
		ExtrbctPrototypeLocbtionsFromPositionFunc: &LsifStoreExtrbctPrototypeLocbtionsFromPositionFunc{
			defbultHook: i.ExtrbctPrototypeLocbtionsFromPosition,
		},
		ExtrbctReferenceLocbtionsFromPositionFunc: &LsifStoreExtrbctReferenceLocbtionsFromPositionFunc{
			defbultHook: i.ExtrbctReferenceLocbtionsFromPosition,
		},
		GetBulkMonikerLocbtionsFunc: &LsifStoreGetBulkMonikerLocbtionsFunc{
			defbultHook: i.GetBulkMonikerLocbtions,
		},
		GetDefinitionLocbtionsFunc: &LsifStoreGetDefinitionLocbtionsFunc{
			defbultHook: i.GetDefinitionLocbtions,
		},
		GetDibgnosticsFunc: &LsifStoreGetDibgnosticsFunc{
			defbultHook: i.GetDibgnostics,
		},
		GetHoverFunc: &LsifStoreGetHoverFunc{
			defbultHook: i.GetHover,
		},
		GetImplementbtionLocbtionsFunc: &LsifStoreGetImplementbtionLocbtionsFunc{
			defbultHook: i.GetImplementbtionLocbtions,
		},
		GetMinimblBulkMonikerLocbtionsFunc: &LsifStoreGetMinimblBulkMonikerLocbtionsFunc{
			defbultHook: i.GetMinimblBulkMonikerLocbtions,
		},
		GetMonikersByPositionFunc: &LsifStoreGetMonikersByPositionFunc{
			defbultHook: i.GetMonikersByPosition,
		},
		GetPbckbgeInformbtionFunc: &LsifStoreGetPbckbgeInformbtionFunc{
			defbultHook: i.GetPbckbgeInformbtion,
		},
		GetPbthExistsFunc: &LsifStoreGetPbthExistsFunc{
			defbultHook: i.GetPbthExists,
		},
		GetPrototypeLocbtionsFunc: &LsifStoreGetPrototypeLocbtionsFunc{
			defbultHook: i.GetPrototypeLocbtions,
		},
		GetRbngesFunc: &LsifStoreGetRbngesFunc{
			defbultHook: i.GetRbnges,
		},
		GetReferenceLocbtionsFunc: &LsifStoreGetReferenceLocbtionsFunc{
			defbultHook: i.GetReferenceLocbtions,
		},
		GetStencilFunc: &LsifStoreGetStencilFunc{
			defbultHook: i.GetStencil,
		},
		SCIPDocumentFunc: &LsifStoreSCIPDocumentFunc{
			defbultHook: i.SCIPDocument,
		},
	}
}

// LsifStoreExtrbctDefinitionLocbtionsFromPositionFunc describes the
// behbvior when the ExtrbctDefinitionLocbtionsFromPosition method of the
// pbrent MockLsifStore instbnce is invoked.
type LsifStoreExtrbctDefinitionLocbtionsFromPositionFunc struct {
	defbultHook func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error)
	hooks       []func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error)
	history     []LsifStoreExtrbctDefinitionLocbtionsFromPositionFuncCbll
	mutex       sync.Mutex
}

// ExtrbctDefinitionLocbtionsFromPosition delegbtes to the next hook
// function in the queue bnd stores the pbrbmeter bnd result vblues of this
// invocbtion.
func (m *MockLsifStore) ExtrbctDefinitionLocbtionsFromPosition(v0 context.Context, v1 lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
	r0, r1, r2 := m.ExtrbctDefinitionLocbtionsFromPositionFunc.nextHook()(v0, v1)
	m.ExtrbctDefinitionLocbtionsFromPositionFunc.bppendCbll(LsifStoreExtrbctDefinitionLocbtionsFromPositionFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// ExtrbctDefinitionLocbtionsFromPosition method of the pbrent MockLsifStore
// instbnce is invoked bnd the hook queue is empty.
func (f *LsifStoreExtrbctDefinitionLocbtionsFromPositionFunc) SetDefbultHook(hook func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ExtrbctDefinitionLocbtionsFromPosition method of the pbrent MockLsifStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *LsifStoreExtrbctDefinitionLocbtionsFromPositionFunc) PushHook(hook func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LsifStoreExtrbctDefinitionLocbtionsFromPositionFunc) SetDefbultReturn(r0 []shbred.Locbtion, r1 []string, r2 error) {
	f.SetDefbultHook(func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LsifStoreExtrbctDefinitionLocbtionsFromPositionFunc) PushReturn(r0 []shbred.Locbtion, r1 []string, r2 error) {
	f.PushHook(func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
		return r0, r1, r2
	})
}

func (f *LsifStoreExtrbctDefinitionLocbtionsFromPositionFunc) nextHook() func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LsifStoreExtrbctDefinitionLocbtionsFromPositionFunc) bppendCbll(r0 LsifStoreExtrbctDefinitionLocbtionsFromPositionFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// LsifStoreExtrbctDefinitionLocbtionsFromPositionFuncCbll objects
// describing the invocbtions of this function.
func (f *LsifStoreExtrbctDefinitionLocbtionsFromPositionFunc) History() []LsifStoreExtrbctDefinitionLocbtionsFromPositionFuncCbll {
	f.mutex.Lock()
	history := mbke([]LsifStoreExtrbctDefinitionLocbtionsFromPositionFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LsifStoreExtrbctDefinitionLocbtionsFromPositionFuncCbll is bn object thbt
// describes bn invocbtion of method ExtrbctDefinitionLocbtionsFromPosition
// on bn instbnce of MockLsifStore.
type LsifStoreExtrbctDefinitionLocbtionsFromPositionFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 lsifstore.LocbtionKey
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.Locbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 []string
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LsifStoreExtrbctDefinitionLocbtionsFromPositionFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LsifStoreExtrbctDefinitionLocbtionsFromPositionFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// LsifStoreExtrbctImplementbtionLocbtionsFromPositionFunc describes the
// behbvior when the ExtrbctImplementbtionLocbtionsFromPosition method of
// the pbrent MockLsifStore instbnce is invoked.
type LsifStoreExtrbctImplementbtionLocbtionsFromPositionFunc struct {
	defbultHook func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error)
	hooks       []func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error)
	history     []LsifStoreExtrbctImplementbtionLocbtionsFromPositionFuncCbll
	mutex       sync.Mutex
}

// ExtrbctImplementbtionLocbtionsFromPosition delegbtes to the next hook
// function in the queue bnd stores the pbrbmeter bnd result vblues of this
// invocbtion.
func (m *MockLsifStore) ExtrbctImplementbtionLocbtionsFromPosition(v0 context.Context, v1 lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
	r0, r1, r2 := m.ExtrbctImplementbtionLocbtionsFromPositionFunc.nextHook()(v0, v1)
	m.ExtrbctImplementbtionLocbtionsFromPositionFunc.bppendCbll(LsifStoreExtrbctImplementbtionLocbtionsFromPositionFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// ExtrbctImplementbtionLocbtionsFromPosition method of the pbrent
// MockLsifStore instbnce is invoked bnd the hook queue is empty.
func (f *LsifStoreExtrbctImplementbtionLocbtionsFromPositionFunc) SetDefbultHook(hook func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ExtrbctImplementbtionLocbtionsFromPosition method of the pbrent
// MockLsifStore instbnce invokes the hook bt the front of the queue bnd
// discbrds it. After the queue is empty, the defbult hook function is
// invoked for bny future bction.
func (f *LsifStoreExtrbctImplementbtionLocbtionsFromPositionFunc) PushHook(hook func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LsifStoreExtrbctImplementbtionLocbtionsFromPositionFunc) SetDefbultReturn(r0 []shbred.Locbtion, r1 []string, r2 error) {
	f.SetDefbultHook(func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LsifStoreExtrbctImplementbtionLocbtionsFromPositionFunc) PushReturn(r0 []shbred.Locbtion, r1 []string, r2 error) {
	f.PushHook(func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
		return r0, r1, r2
	})
}

func (f *LsifStoreExtrbctImplementbtionLocbtionsFromPositionFunc) nextHook() func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LsifStoreExtrbctImplementbtionLocbtionsFromPositionFunc) bppendCbll(r0 LsifStoreExtrbctImplementbtionLocbtionsFromPositionFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// LsifStoreExtrbctImplementbtionLocbtionsFromPositionFuncCbll objects
// describing the invocbtions of this function.
func (f *LsifStoreExtrbctImplementbtionLocbtionsFromPositionFunc) History() []LsifStoreExtrbctImplementbtionLocbtionsFromPositionFuncCbll {
	f.mutex.Lock()
	history := mbke([]LsifStoreExtrbctImplementbtionLocbtionsFromPositionFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LsifStoreExtrbctImplementbtionLocbtionsFromPositionFuncCbll is bn object
// thbt describes bn invocbtion of method
// ExtrbctImplementbtionLocbtionsFromPosition on bn instbnce of
// MockLsifStore.
type LsifStoreExtrbctImplementbtionLocbtionsFromPositionFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 lsifstore.LocbtionKey
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.Locbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 []string
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LsifStoreExtrbctImplementbtionLocbtionsFromPositionFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LsifStoreExtrbctImplementbtionLocbtionsFromPositionFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// LsifStoreExtrbctPrototypeLocbtionsFromPositionFunc describes the behbvior
// when the ExtrbctPrototypeLocbtionsFromPosition method of the pbrent
// MockLsifStore instbnce is invoked.
type LsifStoreExtrbctPrototypeLocbtionsFromPositionFunc struct {
	defbultHook func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error)
	hooks       []func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error)
	history     []LsifStoreExtrbctPrototypeLocbtionsFromPositionFuncCbll
	mutex       sync.Mutex
}

// ExtrbctPrototypeLocbtionsFromPosition delegbtes to the next hook function
// in the queue bnd stores the pbrbmeter bnd result vblues of this
// invocbtion.
func (m *MockLsifStore) ExtrbctPrototypeLocbtionsFromPosition(v0 context.Context, v1 lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
	r0, r1, r2 := m.ExtrbctPrototypeLocbtionsFromPositionFunc.nextHook()(v0, v1)
	m.ExtrbctPrototypeLocbtionsFromPositionFunc.bppendCbll(LsifStoreExtrbctPrototypeLocbtionsFromPositionFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// ExtrbctPrototypeLocbtionsFromPosition method of the pbrent MockLsifStore
// instbnce is invoked bnd the hook queue is empty.
func (f *LsifStoreExtrbctPrototypeLocbtionsFromPositionFunc) SetDefbultHook(hook func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ExtrbctPrototypeLocbtionsFromPosition method of the pbrent MockLsifStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *LsifStoreExtrbctPrototypeLocbtionsFromPositionFunc) PushHook(hook func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LsifStoreExtrbctPrototypeLocbtionsFromPositionFunc) SetDefbultReturn(r0 []shbred.Locbtion, r1 []string, r2 error) {
	f.SetDefbultHook(func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LsifStoreExtrbctPrototypeLocbtionsFromPositionFunc) PushReturn(r0 []shbred.Locbtion, r1 []string, r2 error) {
	f.PushHook(func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
		return r0, r1, r2
	})
}

func (f *LsifStoreExtrbctPrototypeLocbtionsFromPositionFunc) nextHook() func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LsifStoreExtrbctPrototypeLocbtionsFromPositionFunc) bppendCbll(r0 LsifStoreExtrbctPrototypeLocbtionsFromPositionFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// LsifStoreExtrbctPrototypeLocbtionsFromPositionFuncCbll objects describing
// the invocbtions of this function.
func (f *LsifStoreExtrbctPrototypeLocbtionsFromPositionFunc) History() []LsifStoreExtrbctPrototypeLocbtionsFromPositionFuncCbll {
	f.mutex.Lock()
	history := mbke([]LsifStoreExtrbctPrototypeLocbtionsFromPositionFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LsifStoreExtrbctPrototypeLocbtionsFromPositionFuncCbll is bn object thbt
// describes bn invocbtion of method ExtrbctPrototypeLocbtionsFromPosition
// on bn instbnce of MockLsifStore.
type LsifStoreExtrbctPrototypeLocbtionsFromPositionFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 lsifstore.LocbtionKey
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.Locbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 []string
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LsifStoreExtrbctPrototypeLocbtionsFromPositionFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LsifStoreExtrbctPrototypeLocbtionsFromPositionFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// LsifStoreExtrbctReferenceLocbtionsFromPositionFunc describes the behbvior
// when the ExtrbctReferenceLocbtionsFromPosition method of the pbrent
// MockLsifStore instbnce is invoked.
type LsifStoreExtrbctReferenceLocbtionsFromPositionFunc struct {
	defbultHook func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error)
	hooks       []func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error)
	history     []LsifStoreExtrbctReferenceLocbtionsFromPositionFuncCbll
	mutex       sync.Mutex
}

// ExtrbctReferenceLocbtionsFromPosition delegbtes to the next hook function
// in the queue bnd stores the pbrbmeter bnd result vblues of this
// invocbtion.
func (m *MockLsifStore) ExtrbctReferenceLocbtionsFromPosition(v0 context.Context, v1 lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
	r0, r1, r2 := m.ExtrbctReferenceLocbtionsFromPositionFunc.nextHook()(v0, v1)
	m.ExtrbctReferenceLocbtionsFromPositionFunc.bppendCbll(LsifStoreExtrbctReferenceLocbtionsFromPositionFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// ExtrbctReferenceLocbtionsFromPosition method of the pbrent MockLsifStore
// instbnce is invoked bnd the hook queue is empty.
func (f *LsifStoreExtrbctReferenceLocbtionsFromPositionFunc) SetDefbultHook(hook func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ExtrbctReferenceLocbtionsFromPosition method of the pbrent MockLsifStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *LsifStoreExtrbctReferenceLocbtionsFromPositionFunc) PushHook(hook func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LsifStoreExtrbctReferenceLocbtionsFromPositionFunc) SetDefbultReturn(r0 []shbred.Locbtion, r1 []string, r2 error) {
	f.SetDefbultHook(func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LsifStoreExtrbctReferenceLocbtionsFromPositionFunc) PushReturn(r0 []shbred.Locbtion, r1 []string, r2 error) {
	f.PushHook(func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
		return r0, r1, r2
	})
}

func (f *LsifStoreExtrbctReferenceLocbtionsFromPositionFunc) nextHook() func(context.Context, lsifstore.LocbtionKey) ([]shbred.Locbtion, []string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LsifStoreExtrbctReferenceLocbtionsFromPositionFunc) bppendCbll(r0 LsifStoreExtrbctReferenceLocbtionsFromPositionFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// LsifStoreExtrbctReferenceLocbtionsFromPositionFuncCbll objects describing
// the invocbtions of this function.
func (f *LsifStoreExtrbctReferenceLocbtionsFromPositionFunc) History() []LsifStoreExtrbctReferenceLocbtionsFromPositionFuncCbll {
	f.mutex.Lock()
	history := mbke([]LsifStoreExtrbctReferenceLocbtionsFromPositionFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LsifStoreExtrbctReferenceLocbtionsFromPositionFuncCbll is bn object thbt
// describes bn invocbtion of method ExtrbctReferenceLocbtionsFromPosition
// on bn instbnce of MockLsifStore.
type LsifStoreExtrbctReferenceLocbtionsFromPositionFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 lsifstore.LocbtionKey
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.Locbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 []string
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LsifStoreExtrbctReferenceLocbtionsFromPositionFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LsifStoreExtrbctReferenceLocbtionsFromPositionFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// LsifStoreGetBulkMonikerLocbtionsFunc describes the behbvior when the
// GetBulkMonikerLocbtions method of the pbrent MockLsifStore instbnce is
// invoked.
type LsifStoreGetBulkMonikerLocbtionsFunc struct {
	defbultHook func(context.Context, string, []int, []precise.MonikerDbtb, int, int) ([]shbred.Locbtion, int, error)
	hooks       []func(context.Context, string, []int, []precise.MonikerDbtb, int, int) ([]shbred.Locbtion, int, error)
	history     []LsifStoreGetBulkMonikerLocbtionsFuncCbll
	mutex       sync.Mutex
}

// GetBulkMonikerLocbtions delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLsifStore) GetBulkMonikerLocbtions(v0 context.Context, v1 string, v2 []int, v3 []precise.MonikerDbtb, v4 int, v5 int) ([]shbred.Locbtion, int, error) {
	r0, r1, r2 := m.GetBulkMonikerLocbtionsFunc.nextHook()(v0, v1, v2, v3, v4, v5)
	m.GetBulkMonikerLocbtionsFunc.bppendCbll(LsifStoreGetBulkMonikerLocbtionsFuncCbll{v0, v1, v2, v3, v4, v5, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// GetBulkMonikerLocbtions method of the pbrent MockLsifStore instbnce is
// invoked bnd the hook queue is empty.
func (f *LsifStoreGetBulkMonikerLocbtionsFunc) SetDefbultHook(hook func(context.Context, string, []int, []precise.MonikerDbtb, int, int) ([]shbred.Locbtion, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetBulkMonikerLocbtions method of the pbrent MockLsifStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *LsifStoreGetBulkMonikerLocbtionsFunc) PushHook(hook func(context.Context, string, []int, []precise.MonikerDbtb, int, int) ([]shbred.Locbtion, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LsifStoreGetBulkMonikerLocbtionsFunc) SetDefbultReturn(r0 []shbred.Locbtion, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, string, []int, []precise.MonikerDbtb, int, int) ([]shbred.Locbtion, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LsifStoreGetBulkMonikerLocbtionsFunc) PushReturn(r0 []shbred.Locbtion, r1 int, r2 error) {
	f.PushHook(func(context.Context, string, []int, []precise.MonikerDbtb, int, int) ([]shbred.Locbtion, int, error) {
		return r0, r1, r2
	})
}

func (f *LsifStoreGetBulkMonikerLocbtionsFunc) nextHook() func(context.Context, string, []int, []precise.MonikerDbtb, int, int) ([]shbred.Locbtion, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LsifStoreGetBulkMonikerLocbtionsFunc) bppendCbll(r0 LsifStoreGetBulkMonikerLocbtionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LsifStoreGetBulkMonikerLocbtionsFuncCbll
// objects describing the invocbtions of this function.
func (f *LsifStoreGetBulkMonikerLocbtionsFunc) History() []LsifStoreGetBulkMonikerLocbtionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]LsifStoreGetBulkMonikerLocbtionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LsifStoreGetBulkMonikerLocbtionsFuncCbll is bn object thbt describes bn
// invocbtion of method GetBulkMonikerLocbtions on bn instbnce of
// MockLsifStore.
type LsifStoreGetBulkMonikerLocbtionsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 []int
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 []precise.MonikerDbtb
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 int
	// Arg5 is the vblue of the 6th brgument pbssed to this method
	// invocbtion.
	Arg5 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.Locbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LsifStoreGetBulkMonikerLocbtionsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LsifStoreGetBulkMonikerLocbtionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// LsifStoreGetDefinitionLocbtionsFunc describes the behbvior when the
// GetDefinitionLocbtions method of the pbrent MockLsifStore instbnce is
// invoked.
type LsifStoreGetDefinitionLocbtionsFunc struct {
	defbultHook func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error)
	hooks       []func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error)
	history     []LsifStoreGetDefinitionLocbtionsFuncCbll
	mutex       sync.Mutex
}

// GetDefinitionLocbtions delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLsifStore) GetDefinitionLocbtions(v0 context.Context, v1 int, v2 string, v3 int, v4 int, v5 int, v6 int) ([]shbred.Locbtion, int, error) {
	r0, r1, r2 := m.GetDefinitionLocbtionsFunc.nextHook()(v0, v1, v2, v3, v4, v5, v6)
	m.GetDefinitionLocbtionsFunc.bppendCbll(LsifStoreGetDefinitionLocbtionsFuncCbll{v0, v1, v2, v3, v4, v5, v6, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// GetDefinitionLocbtions method of the pbrent MockLsifStore instbnce is
// invoked bnd the hook queue is empty.
func (f *LsifStoreGetDefinitionLocbtionsFunc) SetDefbultHook(hook func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetDefinitionLocbtions method of the pbrent MockLsifStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *LsifStoreGetDefinitionLocbtionsFunc) PushHook(hook func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LsifStoreGetDefinitionLocbtionsFunc) SetDefbultReturn(r0 []shbred.Locbtion, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LsifStoreGetDefinitionLocbtionsFunc) PushReturn(r0 []shbred.Locbtion, r1 int, r2 error) {
	f.PushHook(func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error) {
		return r0, r1, r2
	})
}

func (f *LsifStoreGetDefinitionLocbtionsFunc) nextHook() func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LsifStoreGetDefinitionLocbtionsFunc) bppendCbll(r0 LsifStoreGetDefinitionLocbtionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LsifStoreGetDefinitionLocbtionsFuncCbll
// objects describing the invocbtions of this function.
func (f *LsifStoreGetDefinitionLocbtionsFunc) History() []LsifStoreGetDefinitionLocbtionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]LsifStoreGetDefinitionLocbtionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LsifStoreGetDefinitionLocbtionsFuncCbll is bn object thbt describes bn
// invocbtion of method GetDefinitionLocbtions on bn instbnce of
// MockLsifStore.
type LsifStoreGetDefinitionLocbtionsFuncCbll struct {
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
	Arg3 int
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 int
	// Arg5 is the vblue of the 6th brgument pbssed to this method
	// invocbtion.
	Arg5 int
	// Arg6 is the vblue of the 7th brgument pbssed to this method
	// invocbtion.
	Arg6 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.Locbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LsifStoreGetDefinitionLocbtionsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5, c.Arg6}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LsifStoreGetDefinitionLocbtionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// LsifStoreGetDibgnosticsFunc describes the behbvior when the
// GetDibgnostics method of the pbrent MockLsifStore instbnce is invoked.
type LsifStoreGetDibgnosticsFunc struct {
	defbultHook func(context.Context, int, string, int, int) ([]shbred.Dibgnostic, int, error)
	hooks       []func(context.Context, int, string, int, int) ([]shbred.Dibgnostic, int, error)
	history     []LsifStoreGetDibgnosticsFuncCbll
	mutex       sync.Mutex
}

// GetDibgnostics delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLsifStore) GetDibgnostics(v0 context.Context, v1 int, v2 string, v3 int, v4 int) ([]shbred.Dibgnostic, int, error) {
	r0, r1, r2 := m.GetDibgnosticsFunc.nextHook()(v0, v1, v2, v3, v4)
	m.GetDibgnosticsFunc.bppendCbll(LsifStoreGetDibgnosticsFuncCbll{v0, v1, v2, v3, v4, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the GetDibgnostics
// method of the pbrent MockLsifStore instbnce is invoked bnd the hook queue
// is empty.
func (f *LsifStoreGetDibgnosticsFunc) SetDefbultHook(hook func(context.Context, int, string, int, int) ([]shbred.Dibgnostic, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetDibgnostics method of the pbrent MockLsifStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *LsifStoreGetDibgnosticsFunc) PushHook(hook func(context.Context, int, string, int, int) ([]shbred.Dibgnostic, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LsifStoreGetDibgnosticsFunc) SetDefbultReturn(r0 []shbred.Dibgnostic, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, int, string, int, int) ([]shbred.Dibgnostic, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LsifStoreGetDibgnosticsFunc) PushReturn(r0 []shbred.Dibgnostic, r1 int, r2 error) {
	f.PushHook(func(context.Context, int, string, int, int) ([]shbred.Dibgnostic, int, error) {
		return r0, r1, r2
	})
}

func (f *LsifStoreGetDibgnosticsFunc) nextHook() func(context.Context, int, string, int, int) ([]shbred.Dibgnostic, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LsifStoreGetDibgnosticsFunc) bppendCbll(r0 LsifStoreGetDibgnosticsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LsifStoreGetDibgnosticsFuncCbll objects
// describing the invocbtions of this function.
func (f *LsifStoreGetDibgnosticsFunc) History() []LsifStoreGetDibgnosticsFuncCbll {
	f.mutex.Lock()
	history := mbke([]LsifStoreGetDibgnosticsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LsifStoreGetDibgnosticsFuncCbll is bn object thbt describes bn invocbtion
// of method GetDibgnostics on bn instbnce of MockLsifStore.
type LsifStoreGetDibgnosticsFuncCbll struct {
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
	Arg3 int
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.Dibgnostic
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LsifStoreGetDibgnosticsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LsifStoreGetDibgnosticsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// LsifStoreGetHoverFunc describes the behbvior when the GetHover method of
// the pbrent MockLsifStore instbnce is invoked.
type LsifStoreGetHoverFunc struct {
	defbultHook func(context.Context, int, string, int, int) (string, shbred.Rbnge, bool, error)
	hooks       []func(context.Context, int, string, int, int) (string, shbred.Rbnge, bool, error)
	history     []LsifStoreGetHoverFuncCbll
	mutex       sync.Mutex
}

// GetHover delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLsifStore) GetHover(v0 context.Context, v1 int, v2 string, v3 int, v4 int) (string, shbred.Rbnge, bool, error) {
	r0, r1, r2, r3 := m.GetHoverFunc.nextHook()(v0, v1, v2, v3, v4)
	m.GetHoverFunc.bppendCbll(LsifStoreGetHoverFuncCbll{v0, v1, v2, v3, v4, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefbultHook sets function thbt is cblled when the GetHover method of
// the pbrent MockLsifStore instbnce is invoked bnd the hook queue is empty.
func (f *LsifStoreGetHoverFunc) SetDefbultHook(hook func(context.Context, int, string, int, int) (string, shbred.Rbnge, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetHover method of the pbrent MockLsifStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *LsifStoreGetHoverFunc) PushHook(hook func(context.Context, int, string, int, int) (string, shbred.Rbnge, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LsifStoreGetHoverFunc) SetDefbultReturn(r0 string, r1 shbred.Rbnge, r2 bool, r3 error) {
	f.SetDefbultHook(func(context.Context, int, string, int, int) (string, shbred.Rbnge, bool, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LsifStoreGetHoverFunc) PushReturn(r0 string, r1 shbred.Rbnge, r2 bool, r3 error) {
	f.PushHook(func(context.Context, int, string, int, int) (string, shbred.Rbnge, bool, error) {
		return r0, r1, r2, r3
	})
}

func (f *LsifStoreGetHoverFunc) nextHook() func(context.Context, int, string, int, int) (string, shbred.Rbnge, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LsifStoreGetHoverFunc) bppendCbll(r0 LsifStoreGetHoverFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LsifStoreGetHoverFuncCbll objects
// describing the invocbtions of this function.
func (f *LsifStoreGetHoverFunc) History() []LsifStoreGetHoverFuncCbll {
	f.mutex.Lock()
	history := mbke([]LsifStoreGetHoverFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LsifStoreGetHoverFuncCbll is bn object thbt describes bn invocbtion of
// method GetHover on bn instbnce of MockLsifStore.
type LsifStoreGetHoverFuncCbll struct {
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
	Arg3 int
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 shbred.Rbnge
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 bool
	// Result3 is the vblue of the 4th result returned from this method
	// invocbtion.
	Result3 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LsifStoreGetHoverFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LsifStoreGetHoverFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2, c.Result3}
}

// LsifStoreGetImplementbtionLocbtionsFunc describes the behbvior when the
// GetImplementbtionLocbtions method of the pbrent MockLsifStore instbnce is
// invoked.
type LsifStoreGetImplementbtionLocbtionsFunc struct {
	defbultHook func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error)
	hooks       []func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error)
	history     []LsifStoreGetImplementbtionLocbtionsFuncCbll
	mutex       sync.Mutex
}

// GetImplementbtionLocbtions delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLsifStore) GetImplementbtionLocbtions(v0 context.Context, v1 int, v2 string, v3 int, v4 int, v5 int, v6 int) ([]shbred.Locbtion, int, error) {
	r0, r1, r2 := m.GetImplementbtionLocbtionsFunc.nextHook()(v0, v1, v2, v3, v4, v5, v6)
	m.GetImplementbtionLocbtionsFunc.bppendCbll(LsifStoreGetImplementbtionLocbtionsFuncCbll{v0, v1, v2, v3, v4, v5, v6, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// GetImplementbtionLocbtions method of the pbrent MockLsifStore instbnce is
// invoked bnd the hook queue is empty.
func (f *LsifStoreGetImplementbtionLocbtionsFunc) SetDefbultHook(hook func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetImplementbtionLocbtions method of the pbrent MockLsifStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *LsifStoreGetImplementbtionLocbtionsFunc) PushHook(hook func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LsifStoreGetImplementbtionLocbtionsFunc) SetDefbultReturn(r0 []shbred.Locbtion, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LsifStoreGetImplementbtionLocbtionsFunc) PushReturn(r0 []shbred.Locbtion, r1 int, r2 error) {
	f.PushHook(func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error) {
		return r0, r1, r2
	})
}

func (f *LsifStoreGetImplementbtionLocbtionsFunc) nextHook() func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LsifStoreGetImplementbtionLocbtionsFunc) bppendCbll(r0 LsifStoreGetImplementbtionLocbtionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LsifStoreGetImplementbtionLocbtionsFuncCbll
// objects describing the invocbtions of this function.
func (f *LsifStoreGetImplementbtionLocbtionsFunc) History() []LsifStoreGetImplementbtionLocbtionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]LsifStoreGetImplementbtionLocbtionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LsifStoreGetImplementbtionLocbtionsFuncCbll is bn object thbt describes
// bn invocbtion of method GetImplementbtionLocbtions on bn instbnce of
// MockLsifStore.
type LsifStoreGetImplementbtionLocbtionsFuncCbll struct {
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
	Arg3 int
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 int
	// Arg5 is the vblue of the 6th brgument pbssed to this method
	// invocbtion.
	Arg5 int
	// Arg6 is the vblue of the 7th brgument pbssed to this method
	// invocbtion.
	Arg6 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.Locbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LsifStoreGetImplementbtionLocbtionsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5, c.Arg6}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LsifStoreGetImplementbtionLocbtionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// LsifStoreGetMinimblBulkMonikerLocbtionsFunc describes the behbvior when
// the GetMinimblBulkMonikerLocbtions method of the pbrent MockLsifStore
// instbnce is invoked.
type LsifStoreGetMinimblBulkMonikerLocbtionsFunc struct {
	defbultHook func(context.Context, string, []int, mbp[int]string, []precise.MonikerDbtb, int, int) ([]shbred.Locbtion, int, error)
	hooks       []func(context.Context, string, []int, mbp[int]string, []precise.MonikerDbtb, int, int) ([]shbred.Locbtion, int, error)
	history     []LsifStoreGetMinimblBulkMonikerLocbtionsFuncCbll
	mutex       sync.Mutex
}

// GetMinimblBulkMonikerLocbtions delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLsifStore) GetMinimblBulkMonikerLocbtions(v0 context.Context, v1 string, v2 []int, v3 mbp[int]string, v4 []precise.MonikerDbtb, v5 int, v6 int) ([]shbred.Locbtion, int, error) {
	r0, r1, r2 := m.GetMinimblBulkMonikerLocbtionsFunc.nextHook()(v0, v1, v2, v3, v4, v5, v6)
	m.GetMinimblBulkMonikerLocbtionsFunc.bppendCbll(LsifStoreGetMinimblBulkMonikerLocbtionsFuncCbll{v0, v1, v2, v3, v4, v5, v6, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// GetMinimblBulkMonikerLocbtions method of the pbrent MockLsifStore
// instbnce is invoked bnd the hook queue is empty.
func (f *LsifStoreGetMinimblBulkMonikerLocbtionsFunc) SetDefbultHook(hook func(context.Context, string, []int, mbp[int]string, []precise.MonikerDbtb, int, int) ([]shbred.Locbtion, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetMinimblBulkMonikerLocbtions method of the pbrent MockLsifStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *LsifStoreGetMinimblBulkMonikerLocbtionsFunc) PushHook(hook func(context.Context, string, []int, mbp[int]string, []precise.MonikerDbtb, int, int) ([]shbred.Locbtion, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LsifStoreGetMinimblBulkMonikerLocbtionsFunc) SetDefbultReturn(r0 []shbred.Locbtion, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, string, []int, mbp[int]string, []precise.MonikerDbtb, int, int) ([]shbred.Locbtion, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LsifStoreGetMinimblBulkMonikerLocbtionsFunc) PushReturn(r0 []shbred.Locbtion, r1 int, r2 error) {
	f.PushHook(func(context.Context, string, []int, mbp[int]string, []precise.MonikerDbtb, int, int) ([]shbred.Locbtion, int, error) {
		return r0, r1, r2
	})
}

func (f *LsifStoreGetMinimblBulkMonikerLocbtionsFunc) nextHook() func(context.Context, string, []int, mbp[int]string, []precise.MonikerDbtb, int, int) ([]shbred.Locbtion, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LsifStoreGetMinimblBulkMonikerLocbtionsFunc) bppendCbll(r0 LsifStoreGetMinimblBulkMonikerLocbtionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// LsifStoreGetMinimblBulkMonikerLocbtionsFuncCbll objects describing the
// invocbtions of this function.
func (f *LsifStoreGetMinimblBulkMonikerLocbtionsFunc) History() []LsifStoreGetMinimblBulkMonikerLocbtionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]LsifStoreGetMinimblBulkMonikerLocbtionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LsifStoreGetMinimblBulkMonikerLocbtionsFuncCbll is bn object thbt
// describes bn invocbtion of method GetMinimblBulkMonikerLocbtions on bn
// instbnce of MockLsifStore.
type LsifStoreGetMinimblBulkMonikerLocbtionsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 []int
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 mbp[int]string
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 []precise.MonikerDbtb
	// Arg5 is the vblue of the 6th brgument pbssed to this method
	// invocbtion.
	Arg5 int
	// Arg6 is the vblue of the 7th brgument pbssed to this method
	// invocbtion.
	Arg6 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.Locbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LsifStoreGetMinimblBulkMonikerLocbtionsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5, c.Arg6}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LsifStoreGetMinimblBulkMonikerLocbtionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// LsifStoreGetMonikersByPositionFunc describes the behbvior when the
// GetMonikersByPosition method of the pbrent MockLsifStore instbnce is
// invoked.
type LsifStoreGetMonikersByPositionFunc struct {
	defbultHook func(context.Context, int, string, int, int) ([][]precise.MonikerDbtb, error)
	hooks       []func(context.Context, int, string, int, int) ([][]precise.MonikerDbtb, error)
	history     []LsifStoreGetMonikersByPositionFuncCbll
	mutex       sync.Mutex
}

// GetMonikersByPosition delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLsifStore) GetMonikersByPosition(v0 context.Context, v1 int, v2 string, v3 int, v4 int) ([][]precise.MonikerDbtb, error) {
	r0, r1 := m.GetMonikersByPositionFunc.nextHook()(v0, v1, v2, v3, v4)
	m.GetMonikersByPositionFunc.bppendCbll(LsifStoreGetMonikersByPositionFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetMonikersByPosition method of the pbrent MockLsifStore instbnce is
// invoked bnd the hook queue is empty.
func (f *LsifStoreGetMonikersByPositionFunc) SetDefbultHook(hook func(context.Context, int, string, int, int) ([][]precise.MonikerDbtb, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetMonikersByPosition method of the pbrent MockLsifStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *LsifStoreGetMonikersByPositionFunc) PushHook(hook func(context.Context, int, string, int, int) ([][]precise.MonikerDbtb, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LsifStoreGetMonikersByPositionFunc) SetDefbultReturn(r0 [][]precise.MonikerDbtb, r1 error) {
	f.SetDefbultHook(func(context.Context, int, string, int, int) ([][]precise.MonikerDbtb, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LsifStoreGetMonikersByPositionFunc) PushReturn(r0 [][]precise.MonikerDbtb, r1 error) {
	f.PushHook(func(context.Context, int, string, int, int) ([][]precise.MonikerDbtb, error) {
		return r0, r1
	})
}

func (f *LsifStoreGetMonikersByPositionFunc) nextHook() func(context.Context, int, string, int, int) ([][]precise.MonikerDbtb, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LsifStoreGetMonikersByPositionFunc) bppendCbll(r0 LsifStoreGetMonikersByPositionFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LsifStoreGetMonikersByPositionFuncCbll
// objects describing the invocbtions of this function.
func (f *LsifStoreGetMonikersByPositionFunc) History() []LsifStoreGetMonikersByPositionFuncCbll {
	f.mutex.Lock()
	history := mbke([]LsifStoreGetMonikersByPositionFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LsifStoreGetMonikersByPositionFuncCbll is bn object thbt describes bn
// invocbtion of method GetMonikersByPosition on bn instbnce of
// MockLsifStore.
type LsifStoreGetMonikersByPositionFuncCbll struct {
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
	Arg3 int
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 [][]precise.MonikerDbtb
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LsifStoreGetMonikersByPositionFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LsifStoreGetMonikersByPositionFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// LsifStoreGetPbckbgeInformbtionFunc describes the behbvior when the
// GetPbckbgeInformbtion method of the pbrent MockLsifStore instbnce is
// invoked.
type LsifStoreGetPbckbgeInformbtionFunc struct {
	defbultHook func(context.Context, int, string, string) (precise.PbckbgeInformbtionDbtb, bool, error)
	hooks       []func(context.Context, int, string, string) (precise.PbckbgeInformbtionDbtb, bool, error)
	history     []LsifStoreGetPbckbgeInformbtionFuncCbll
	mutex       sync.Mutex
}

// GetPbckbgeInformbtion delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLsifStore) GetPbckbgeInformbtion(v0 context.Context, v1 int, v2 string, v3 string) (precise.PbckbgeInformbtionDbtb, bool, error) {
	r0, r1, r2 := m.GetPbckbgeInformbtionFunc.nextHook()(v0, v1, v2, v3)
	m.GetPbckbgeInformbtionFunc.bppendCbll(LsifStoreGetPbckbgeInformbtionFuncCbll{v0, v1, v2, v3, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// GetPbckbgeInformbtion method of the pbrent MockLsifStore instbnce is
// invoked bnd the hook queue is empty.
func (f *LsifStoreGetPbckbgeInformbtionFunc) SetDefbultHook(hook func(context.Context, int, string, string) (precise.PbckbgeInformbtionDbtb, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetPbckbgeInformbtion method of the pbrent MockLsifStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *LsifStoreGetPbckbgeInformbtionFunc) PushHook(hook func(context.Context, int, string, string) (precise.PbckbgeInformbtionDbtb, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LsifStoreGetPbckbgeInformbtionFunc) SetDefbultReturn(r0 precise.PbckbgeInformbtionDbtb, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, int, string, string) (precise.PbckbgeInformbtionDbtb, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LsifStoreGetPbckbgeInformbtionFunc) PushReturn(r0 precise.PbckbgeInformbtionDbtb, r1 bool, r2 error) {
	f.PushHook(func(context.Context, int, string, string) (precise.PbckbgeInformbtionDbtb, bool, error) {
		return r0, r1, r2
	})
}

func (f *LsifStoreGetPbckbgeInformbtionFunc) nextHook() func(context.Context, int, string, string) (precise.PbckbgeInformbtionDbtb, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LsifStoreGetPbckbgeInformbtionFunc) bppendCbll(r0 LsifStoreGetPbckbgeInformbtionFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LsifStoreGetPbckbgeInformbtionFuncCbll
// objects describing the invocbtions of this function.
func (f *LsifStoreGetPbckbgeInformbtionFunc) History() []LsifStoreGetPbckbgeInformbtionFuncCbll {
	f.mutex.Lock()
	history := mbke([]LsifStoreGetPbckbgeInformbtionFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LsifStoreGetPbckbgeInformbtionFuncCbll is bn object thbt describes bn
// invocbtion of method GetPbckbgeInformbtion on bn instbnce of
// MockLsifStore.
type LsifStoreGetPbckbgeInformbtionFuncCbll struct {
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
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 precise.PbckbgeInformbtionDbtb
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LsifStoreGetPbckbgeInformbtionFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LsifStoreGetPbckbgeInformbtionFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// LsifStoreGetPbthExistsFunc describes the behbvior when the GetPbthExists
// method of the pbrent MockLsifStore instbnce is invoked.
type LsifStoreGetPbthExistsFunc struct {
	defbultHook func(context.Context, int, string) (bool, error)
	hooks       []func(context.Context, int, string) (bool, error)
	history     []LsifStoreGetPbthExistsFuncCbll
	mutex       sync.Mutex
}

// GetPbthExists delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLsifStore) GetPbthExists(v0 context.Context, v1 int, v2 string) (bool, error) {
	r0, r1 := m.GetPbthExistsFunc.nextHook()(v0, v1, v2)
	m.GetPbthExistsFunc.bppendCbll(LsifStoreGetPbthExistsFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetPbthExists method
// of the pbrent MockLsifStore instbnce is invoked bnd the hook queue is
// empty.
func (f *LsifStoreGetPbthExistsFunc) SetDefbultHook(hook func(context.Context, int, string) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetPbthExists method of the pbrent MockLsifStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *LsifStoreGetPbthExistsFunc) PushHook(hook func(context.Context, int, string) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LsifStoreGetPbthExistsFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, int, string) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LsifStoreGetPbthExistsFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, int, string) (bool, error) {
		return r0, r1
	})
}

func (f *LsifStoreGetPbthExistsFunc) nextHook() func(context.Context, int, string) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LsifStoreGetPbthExistsFunc) bppendCbll(r0 LsifStoreGetPbthExistsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LsifStoreGetPbthExistsFuncCbll objects
// describing the invocbtions of this function.
func (f *LsifStoreGetPbthExistsFunc) History() []LsifStoreGetPbthExistsFuncCbll {
	f.mutex.Lock()
	history := mbke([]LsifStoreGetPbthExistsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LsifStoreGetPbthExistsFuncCbll is bn object thbt describes bn invocbtion
// of method GetPbthExists on bn instbnce of MockLsifStore.
type LsifStoreGetPbthExistsFuncCbll struct {
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
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LsifStoreGetPbthExistsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LsifStoreGetPbthExistsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// LsifStoreGetPrototypeLocbtionsFunc describes the behbvior when the
// GetPrototypeLocbtions method of the pbrent MockLsifStore instbnce is
// invoked.
type LsifStoreGetPrototypeLocbtionsFunc struct {
	defbultHook func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error)
	hooks       []func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error)
	history     []LsifStoreGetPrototypeLocbtionsFuncCbll
	mutex       sync.Mutex
}

// GetPrototypeLocbtions delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLsifStore) GetPrototypeLocbtions(v0 context.Context, v1 int, v2 string, v3 int, v4 int, v5 int, v6 int) ([]shbred.Locbtion, int, error) {
	r0, r1, r2 := m.GetPrototypeLocbtionsFunc.nextHook()(v0, v1, v2, v3, v4, v5, v6)
	m.GetPrototypeLocbtionsFunc.bppendCbll(LsifStoreGetPrototypeLocbtionsFuncCbll{v0, v1, v2, v3, v4, v5, v6, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// GetPrototypeLocbtions method of the pbrent MockLsifStore instbnce is
// invoked bnd the hook queue is empty.
func (f *LsifStoreGetPrototypeLocbtionsFunc) SetDefbultHook(hook func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetPrototypeLocbtions method of the pbrent MockLsifStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *LsifStoreGetPrototypeLocbtionsFunc) PushHook(hook func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LsifStoreGetPrototypeLocbtionsFunc) SetDefbultReturn(r0 []shbred.Locbtion, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LsifStoreGetPrototypeLocbtionsFunc) PushReturn(r0 []shbred.Locbtion, r1 int, r2 error) {
	f.PushHook(func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error) {
		return r0, r1, r2
	})
}

func (f *LsifStoreGetPrototypeLocbtionsFunc) nextHook() func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LsifStoreGetPrototypeLocbtionsFunc) bppendCbll(r0 LsifStoreGetPrototypeLocbtionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LsifStoreGetPrototypeLocbtionsFuncCbll
// objects describing the invocbtions of this function.
func (f *LsifStoreGetPrototypeLocbtionsFunc) History() []LsifStoreGetPrototypeLocbtionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]LsifStoreGetPrototypeLocbtionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LsifStoreGetPrototypeLocbtionsFuncCbll is bn object thbt describes bn
// invocbtion of method GetPrototypeLocbtions on bn instbnce of
// MockLsifStore.
type LsifStoreGetPrototypeLocbtionsFuncCbll struct {
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
	Arg3 int
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 int
	// Arg5 is the vblue of the 6th brgument pbssed to this method
	// invocbtion.
	Arg5 int
	// Arg6 is the vblue of the 7th brgument pbssed to this method
	// invocbtion.
	Arg6 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.Locbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LsifStoreGetPrototypeLocbtionsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5, c.Arg6}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LsifStoreGetPrototypeLocbtionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// LsifStoreGetRbngesFunc describes the behbvior when the GetRbnges method
// of the pbrent MockLsifStore instbnce is invoked.
type LsifStoreGetRbngesFunc struct {
	defbultHook func(context.Context, int, string, int, int) ([]shbred.CodeIntelligenceRbnge, error)
	hooks       []func(context.Context, int, string, int, int) ([]shbred.CodeIntelligenceRbnge, error)
	history     []LsifStoreGetRbngesFuncCbll
	mutex       sync.Mutex
}

// GetRbnges delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLsifStore) GetRbnges(v0 context.Context, v1 int, v2 string, v3 int, v4 int) ([]shbred.CodeIntelligenceRbnge, error) {
	r0, r1 := m.GetRbngesFunc.nextHook()(v0, v1, v2, v3, v4)
	m.GetRbngesFunc.bppendCbll(LsifStoreGetRbngesFuncCbll{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetRbnges method of
// the pbrent MockLsifStore instbnce is invoked bnd the hook queue is empty.
func (f *LsifStoreGetRbngesFunc) SetDefbultHook(hook func(context.Context, int, string, int, int) ([]shbred.CodeIntelligenceRbnge, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetRbnges method of the pbrent MockLsifStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *LsifStoreGetRbngesFunc) PushHook(hook func(context.Context, int, string, int, int) ([]shbred.CodeIntelligenceRbnge, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LsifStoreGetRbngesFunc) SetDefbultReturn(r0 []shbred.CodeIntelligenceRbnge, r1 error) {
	f.SetDefbultHook(func(context.Context, int, string, int, int) ([]shbred.CodeIntelligenceRbnge, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LsifStoreGetRbngesFunc) PushReturn(r0 []shbred.CodeIntelligenceRbnge, r1 error) {
	f.PushHook(func(context.Context, int, string, int, int) ([]shbred.CodeIntelligenceRbnge, error) {
		return r0, r1
	})
}

func (f *LsifStoreGetRbngesFunc) nextHook() func(context.Context, int, string, int, int) ([]shbred.CodeIntelligenceRbnge, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LsifStoreGetRbngesFunc) bppendCbll(r0 LsifStoreGetRbngesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LsifStoreGetRbngesFuncCbll objects
// describing the invocbtions of this function.
func (f *LsifStoreGetRbngesFunc) History() []LsifStoreGetRbngesFuncCbll {
	f.mutex.Lock()
	history := mbke([]LsifStoreGetRbngesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LsifStoreGetRbngesFuncCbll is bn object thbt describes bn invocbtion of
// method GetRbnges on bn instbnce of MockLsifStore.
type LsifStoreGetRbngesFuncCbll struct {
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
	Arg3 int
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.CodeIntelligenceRbnge
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LsifStoreGetRbngesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LsifStoreGetRbngesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// LsifStoreGetReferenceLocbtionsFunc describes the behbvior when the
// GetReferenceLocbtions method of the pbrent MockLsifStore instbnce is
// invoked.
type LsifStoreGetReferenceLocbtionsFunc struct {
	defbultHook func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error)
	hooks       []func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error)
	history     []LsifStoreGetReferenceLocbtionsFuncCbll
	mutex       sync.Mutex
}

// GetReferenceLocbtions delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLsifStore) GetReferenceLocbtions(v0 context.Context, v1 int, v2 string, v3 int, v4 int, v5 int, v6 int) ([]shbred.Locbtion, int, error) {
	r0, r1, r2 := m.GetReferenceLocbtionsFunc.nextHook()(v0, v1, v2, v3, v4, v5, v6)
	m.GetReferenceLocbtionsFunc.bppendCbll(LsifStoreGetReferenceLocbtionsFuncCbll{v0, v1, v2, v3, v4, v5, v6, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// GetReferenceLocbtions method of the pbrent MockLsifStore instbnce is
// invoked bnd the hook queue is empty.
func (f *LsifStoreGetReferenceLocbtionsFunc) SetDefbultHook(hook func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetReferenceLocbtions method of the pbrent MockLsifStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *LsifStoreGetReferenceLocbtionsFunc) PushHook(hook func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LsifStoreGetReferenceLocbtionsFunc) SetDefbultReturn(r0 []shbred.Locbtion, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LsifStoreGetReferenceLocbtionsFunc) PushReturn(r0 []shbred.Locbtion, r1 int, r2 error) {
	f.PushHook(func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error) {
		return r0, r1, r2
	})
}

func (f *LsifStoreGetReferenceLocbtionsFunc) nextHook() func(context.Context, int, string, int, int, int, int) ([]shbred.Locbtion, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LsifStoreGetReferenceLocbtionsFunc) bppendCbll(r0 LsifStoreGetReferenceLocbtionsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LsifStoreGetReferenceLocbtionsFuncCbll
// objects describing the invocbtions of this function.
func (f *LsifStoreGetReferenceLocbtionsFunc) History() []LsifStoreGetReferenceLocbtionsFuncCbll {
	f.mutex.Lock()
	history := mbke([]LsifStoreGetReferenceLocbtionsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LsifStoreGetReferenceLocbtionsFuncCbll is bn object thbt describes bn
// invocbtion of method GetReferenceLocbtions on bn instbnce of
// MockLsifStore.
type LsifStoreGetReferenceLocbtionsFuncCbll struct {
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
	Arg3 int
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 int
	// Arg5 is the vblue of the 6th brgument pbssed to this method
	// invocbtion.
	Arg5 int
	// Arg6 is the vblue of the 7th brgument pbssed to this method
	// invocbtion.
	Arg6 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.Locbtion
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LsifStoreGetReferenceLocbtionsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5, c.Arg6}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LsifStoreGetReferenceLocbtionsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// LsifStoreGetStencilFunc describes the behbvior when the GetStencil method
// of the pbrent MockLsifStore instbnce is invoked.
type LsifStoreGetStencilFunc struct {
	defbultHook func(context.Context, int, string) ([]shbred.Rbnge, error)
	hooks       []func(context.Context, int, string) ([]shbred.Rbnge, error)
	history     []LsifStoreGetStencilFuncCbll
	mutex       sync.Mutex
}

// GetStencil delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLsifStore) GetStencil(v0 context.Context, v1 int, v2 string) ([]shbred.Rbnge, error) {
	r0, r1 := m.GetStencilFunc.nextHook()(v0, v1, v2)
	m.GetStencilFunc.bppendCbll(LsifStoreGetStencilFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetStencil method of
// the pbrent MockLsifStore instbnce is invoked bnd the hook queue is empty.
func (f *LsifStoreGetStencilFunc) SetDefbultHook(hook func(context.Context, int, string) ([]shbred.Rbnge, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetStencil method of the pbrent MockLsifStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *LsifStoreGetStencilFunc) PushHook(hook func(context.Context, int, string) ([]shbred.Rbnge, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LsifStoreGetStencilFunc) SetDefbultReturn(r0 []shbred.Rbnge, r1 error) {
	f.SetDefbultHook(func(context.Context, int, string) ([]shbred.Rbnge, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LsifStoreGetStencilFunc) PushReturn(r0 []shbred.Rbnge, r1 error) {
	f.PushHook(func(context.Context, int, string) ([]shbred.Rbnge, error) {
		return r0, r1
	})
}

func (f *LsifStoreGetStencilFunc) nextHook() func(context.Context, int, string) ([]shbred.Rbnge, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LsifStoreGetStencilFunc) bppendCbll(r0 LsifStoreGetStencilFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LsifStoreGetStencilFuncCbll objects
// describing the invocbtions of this function.
func (f *LsifStoreGetStencilFunc) History() []LsifStoreGetStencilFuncCbll {
	f.mutex.Lock()
	history := mbke([]LsifStoreGetStencilFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LsifStoreGetStencilFuncCbll is bn object thbt describes bn invocbtion of
// method GetStencil on bn instbnce of MockLsifStore.
type LsifStoreGetStencilFuncCbll struct {
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
	Result0 []shbred.Rbnge
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LsifStoreGetStencilFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LsifStoreGetStencilFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// LsifStoreSCIPDocumentFunc describes the behbvior when the SCIPDocument
// method of the pbrent MockLsifStore instbnce is invoked.
type LsifStoreSCIPDocumentFunc struct {
	defbultHook func(context.Context, int, string) (*scip.Document, error)
	hooks       []func(context.Context, int, string) (*scip.Document, error)
	history     []LsifStoreSCIPDocumentFuncCbll
	mutex       sync.Mutex
}

// SCIPDocument delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockLsifStore) SCIPDocument(v0 context.Context, v1 int, v2 string) (*scip.Document, error) {
	r0, r1 := m.SCIPDocumentFunc.nextHook()(v0, v1, v2)
	m.SCIPDocumentFunc.bppendCbll(LsifStoreSCIPDocumentFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the SCIPDocument method
// of the pbrent MockLsifStore instbnce is invoked bnd the hook queue is
// empty.
func (f *LsifStoreSCIPDocumentFunc) SetDefbultHook(hook func(context.Context, int, string) (*scip.Document, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SCIPDocument method of the pbrent MockLsifStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *LsifStoreSCIPDocumentFunc) PushHook(hook func(context.Context, int, string) (*scip.Document, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *LsifStoreSCIPDocumentFunc) SetDefbultReturn(r0 *scip.Document, r1 error) {
	f.SetDefbultHook(func(context.Context, int, string) (*scip.Document, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *LsifStoreSCIPDocumentFunc) PushReturn(r0 *scip.Document, r1 error) {
	f.PushHook(func(context.Context, int, string) (*scip.Document, error) {
		return r0, r1
	})
}

func (f *LsifStoreSCIPDocumentFunc) nextHook() func(context.Context, int, string) (*scip.Document, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *LsifStoreSCIPDocumentFunc) bppendCbll(r0 LsifStoreSCIPDocumentFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of LsifStoreSCIPDocumentFuncCbll objects
// describing the invocbtions of this function.
func (f *LsifStoreSCIPDocumentFunc) History() []LsifStoreSCIPDocumentFuncCbll {
	f.mutex.Lock()
	history := mbke([]LsifStoreSCIPDocumentFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// LsifStoreSCIPDocumentFuncCbll is bn object thbt describes bn invocbtion
// of method SCIPDocument on bn instbnce of MockLsifStore.
type LsifStoreSCIPDocumentFuncCbll struct {
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
	Result0 *scip.Document
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c LsifStoreSCIPDocumentFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c LsifStoreSCIPDocumentFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockGitTreeTrbnslbtor is b mock implementbtion of the GitTreeTrbnslbtor
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv) used for
// unit testing.
type MockGitTreeTrbnslbtor struct {
	// GetTbrgetCommitPbthFromSourcePbthFunc is bn instbnce of b mock
	// function object controlling the behbvior of the method
	// GetTbrgetCommitPbthFromSourcePbth.
	GetTbrgetCommitPbthFromSourcePbthFunc *GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFunc
	// GetTbrgetCommitPositionFromSourcePositionFunc is bn instbnce of b
	// mock function object controlling the behbvior of the method
	// GetTbrgetCommitPositionFromSourcePosition.
	GetTbrgetCommitPositionFromSourcePositionFunc *GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFunc
	// GetTbrgetCommitRbngeFromSourceRbngeFunc is bn instbnce of b mock
	// function object controlling the behbvior of the method
	// GetTbrgetCommitRbngeFromSourceRbnge.
	GetTbrgetCommitRbngeFromSourceRbngeFunc *GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFunc
}

// NewMockGitTreeTrbnslbtor crebtes b new mock of the GitTreeTrbnslbtor
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockGitTreeTrbnslbtor() *MockGitTreeTrbnslbtor {
	return &MockGitTreeTrbnslbtor{
		GetTbrgetCommitPbthFromSourcePbthFunc: &GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFunc{
			defbultHook: func(context.Context, string, string, bool) (r0 string, r1 bool, r2 error) {
				return
			},
		},
		GetTbrgetCommitPositionFromSourcePositionFunc: &GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFunc{
			defbultHook: func(context.Context, string, shbred.Position, bool) (r0 string, r1 shbred.Position, r2 bool, r3 error) {
				return
			},
		},
		GetTbrgetCommitRbngeFromSourceRbngeFunc: &GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFunc{
			defbultHook: func(context.Context, string, string, shbred.Rbnge, bool) (r0 string, r1 shbred.Rbnge, r2 bool, r3 error) {
				return
			},
		},
	}
}

// NewStrictMockGitTreeTrbnslbtor crebtes b new mock of the
// GitTreeTrbnslbtor interfbce. All methods pbnic on invocbtion, unless
// overwritten.
func NewStrictMockGitTreeTrbnslbtor() *MockGitTreeTrbnslbtor {
	return &MockGitTreeTrbnslbtor{
		GetTbrgetCommitPbthFromSourcePbthFunc: &GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFunc{
			defbultHook: func(context.Context, string, string, bool) (string, bool, error) {
				pbnic("unexpected invocbtion of MockGitTreeTrbnslbtor.GetTbrgetCommitPbthFromSourcePbth")
			},
		},
		GetTbrgetCommitPositionFromSourcePositionFunc: &GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFunc{
			defbultHook: func(context.Context, string, shbred.Position, bool) (string, shbred.Position, bool, error) {
				pbnic("unexpected invocbtion of MockGitTreeTrbnslbtor.GetTbrgetCommitPositionFromSourcePosition")
			},
		},
		GetTbrgetCommitRbngeFromSourceRbngeFunc: &GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFunc{
			defbultHook: func(context.Context, string, string, shbred.Rbnge, bool) (string, shbred.Rbnge, bool, error) {
				pbnic("unexpected invocbtion of MockGitTreeTrbnslbtor.GetTbrgetCommitRbngeFromSourceRbnge")
			},
		},
	}
}

// NewMockGitTreeTrbnslbtorFrom crebtes b new mock of the
// MockGitTreeTrbnslbtor interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockGitTreeTrbnslbtorFrom(i GitTreeTrbnslbtor) *MockGitTreeTrbnslbtor {
	return &MockGitTreeTrbnslbtor{
		GetTbrgetCommitPbthFromSourcePbthFunc: &GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFunc{
			defbultHook: i.GetTbrgetCommitPbthFromSourcePbth,
		},
		GetTbrgetCommitPositionFromSourcePositionFunc: &GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFunc{
			defbultHook: i.GetTbrgetCommitPositionFromSourcePosition,
		},
		GetTbrgetCommitRbngeFromSourceRbngeFunc: &GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFunc{
			defbultHook: i.GetTbrgetCommitRbngeFromSourceRbnge,
		},
	}
}

// GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFunc describes the
// behbvior when the GetTbrgetCommitPbthFromSourcePbth method of the pbrent
// MockGitTreeTrbnslbtor instbnce is invoked.
type GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFunc struct {
	defbultHook func(context.Context, string, string, bool) (string, bool, error)
	hooks       []func(context.Context, string, string, bool) (string, bool, error)
	history     []GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFuncCbll
	mutex       sync.Mutex
}

// GetTbrgetCommitPbthFromSourcePbth delegbtes to the next hook function in
// the queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGitTreeTrbnslbtor) GetTbrgetCommitPbthFromSourcePbth(v0 context.Context, v1 string, v2 string, v3 bool) (string, bool, error) {
	r0, r1, r2 := m.GetTbrgetCommitPbthFromSourcePbthFunc.nextHook()(v0, v1, v2, v3)
	m.GetTbrgetCommitPbthFromSourcePbthFunc.bppendCbll(GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFuncCbll{v0, v1, v2, v3, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// GetTbrgetCommitPbthFromSourcePbth method of the pbrent
// MockGitTreeTrbnslbtor instbnce is invoked bnd the hook queue is empty.
func (f *GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFunc) SetDefbultHook(hook func(context.Context, string, string, bool) (string, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetTbrgetCommitPbthFromSourcePbth method of the pbrent
// MockGitTreeTrbnslbtor instbnce invokes the hook bt the front of the queue
// bnd discbrds it. After the queue is empty, the defbult hook function is
// invoked for bny future bction.
func (f *GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFunc) PushHook(hook func(context.Context, string, string, bool) (string, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFunc) SetDefbultReturn(r0 string, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, string, string, bool) (string, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFunc) PushReturn(r0 string, r1 bool, r2 error) {
	f.PushHook(func(context.Context, string, string, bool) (string, bool, error) {
		return r0, r1, r2
	})
}

func (f *GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFunc) nextHook() func(context.Context, string, string, bool) (string, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFunc) bppendCbll(r0 GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFuncCbll objects
// describing the invocbtions of this function.
func (f *GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFunc) History() []GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFuncCbll is bn object
// thbt describes bn invocbtion of method GetTbrgetCommitPbthFromSourcePbth
// on bn instbnce of MockGitTreeTrbnslbtor.
type GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFuncCbll struct {
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
	Arg3 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitTreeTrbnslbtorGetTbrgetCommitPbthFromSourcePbthFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFunc describes
// the behbvior when the GetTbrgetCommitPositionFromSourcePosition method of
// the pbrent MockGitTreeTrbnslbtor instbnce is invoked.
type GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFunc struct {
	defbultHook func(context.Context, string, shbred.Position, bool) (string, shbred.Position, bool, error)
	hooks       []func(context.Context, string, shbred.Position, bool) (string, shbred.Position, bool, error)
	history     []GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFuncCbll
	mutex       sync.Mutex
}

// GetTbrgetCommitPositionFromSourcePosition delegbtes to the next hook
// function in the queue bnd stores the pbrbmeter bnd result vblues of this
// invocbtion.
func (m *MockGitTreeTrbnslbtor) GetTbrgetCommitPositionFromSourcePosition(v0 context.Context, v1 string, v2 shbred.Position, v3 bool) (string, shbred.Position, bool, error) {
	r0, r1, r2, r3 := m.GetTbrgetCommitPositionFromSourcePositionFunc.nextHook()(v0, v1, v2, v3)
	m.GetTbrgetCommitPositionFromSourcePositionFunc.bppendCbll(GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFuncCbll{v0, v1, v2, v3, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefbultHook sets function thbt is cblled when the
// GetTbrgetCommitPositionFromSourcePosition method of the pbrent
// MockGitTreeTrbnslbtor instbnce is invoked bnd the hook queue is empty.
func (f *GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFunc) SetDefbultHook(hook func(context.Context, string, shbred.Position, bool) (string, shbred.Position, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetTbrgetCommitPositionFromSourcePosition method of the pbrent
// MockGitTreeTrbnslbtor instbnce invokes the hook bt the front of the queue
// bnd discbrds it. After the queue is empty, the defbult hook function is
// invoked for bny future bction.
func (f *GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFunc) PushHook(hook func(context.Context, string, shbred.Position, bool) (string, shbred.Position, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFunc) SetDefbultReturn(r0 string, r1 shbred.Position, r2 bool, r3 error) {
	f.SetDefbultHook(func(context.Context, string, shbred.Position, bool) (string, shbred.Position, bool, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFunc) PushReturn(r0 string, r1 shbred.Position, r2 bool, r3 error) {
	f.PushHook(func(context.Context, string, shbred.Position, bool) (string, shbred.Position, bool, error) {
		return r0, r1, r2, r3
	})
}

func (f *GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFunc) nextHook() func(context.Context, string, shbred.Position, bool) (string, shbred.Position, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFunc) bppendCbll(r0 GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFuncCbll
// objects describing the invocbtions of this function.
func (f *GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFunc) History() []GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFuncCbll is bn
// object thbt describes bn invocbtion of method
// GetTbrgetCommitPositionFromSourcePosition on bn instbnce of
// MockGitTreeTrbnslbtor.
type GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 shbred.Position
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 shbred.Position
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 bool
	// Result3 is the vblue of the 4th result returned from this method
	// invocbtion.
	Result3 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitTreeTrbnslbtorGetTbrgetCommitPositionFromSourcePositionFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2, c.Result3}
}

// GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFunc describes the
// behbvior when the GetTbrgetCommitRbngeFromSourceRbnge method of the
// pbrent MockGitTreeTrbnslbtor instbnce is invoked.
type GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFunc struct {
	defbultHook func(context.Context, string, string, shbred.Rbnge, bool) (string, shbred.Rbnge, bool, error)
	hooks       []func(context.Context, string, string, shbred.Rbnge, bool) (string, shbred.Rbnge, bool, error)
	history     []GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFuncCbll
	mutex       sync.Mutex
}

// GetTbrgetCommitRbngeFromSourceRbnge delegbtes to the next hook function
// in the queue bnd stores the pbrbmeter bnd result vblues of this
// invocbtion.
func (m *MockGitTreeTrbnslbtor) GetTbrgetCommitRbngeFromSourceRbnge(v0 context.Context, v1 string, v2 string, v3 shbred.Rbnge, v4 bool) (string, shbred.Rbnge, bool, error) {
	r0, r1, r2, r3 := m.GetTbrgetCommitRbngeFromSourceRbngeFunc.nextHook()(v0, v1, v2, v3, v4)
	m.GetTbrgetCommitRbngeFromSourceRbngeFunc.bppendCbll(GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFuncCbll{v0, v1, v2, v3, v4, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefbultHook sets function thbt is cblled when the
// GetTbrgetCommitRbngeFromSourceRbnge method of the pbrent
// MockGitTreeTrbnslbtor instbnce is invoked bnd the hook queue is empty.
func (f *GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFunc) SetDefbultHook(hook func(context.Context, string, string, shbred.Rbnge, bool) (string, shbred.Rbnge, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetTbrgetCommitRbngeFromSourceRbnge method of the pbrent
// MockGitTreeTrbnslbtor instbnce invokes the hook bt the front of the queue
// bnd discbrds it. After the queue is empty, the defbult hook function is
// invoked for bny future bction.
func (f *GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFunc) PushHook(hook func(context.Context, string, string, shbred.Rbnge, bool) (string, shbred.Rbnge, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFunc) SetDefbultReturn(r0 string, r1 shbred.Rbnge, r2 bool, r3 error) {
	f.SetDefbultHook(func(context.Context, string, string, shbred.Rbnge, bool) (string, shbred.Rbnge, bool, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFunc) PushReturn(r0 string, r1 shbred.Rbnge, r2 bool, r3 error) {
	f.PushHook(func(context.Context, string, string, shbred.Rbnge, bool) (string, shbred.Rbnge, bool, error) {
		return r0, r1, r2, r3
	})
}

func (f *GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFunc) nextHook() func(context.Context, string, string, shbred.Rbnge, bool) (string, shbred.Rbnge, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFunc) bppendCbll(r0 GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFuncCbll objects
// describing the invocbtions of this function.
func (f *GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFunc) History() []GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFuncCbll {
	f.mutex.Lock()
	history := mbke([]GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFuncCbll is bn object
// thbt describes bn invocbtion of method
// GetTbrgetCommitRbngeFromSourceRbnge on bn instbnce of
// MockGitTreeTrbnslbtor.
type GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFuncCbll struct {
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
	Arg3 shbred.Rbnge
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 shbred.Rbnge
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 bool
	// Result3 is the vblue of the 4th result returned from this method
	// invocbtion.
	Result3 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GitTreeTrbnslbtorGetTbrgetCommitRbngeFromSourceRbngeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2, c.Result3}
}

// MockUplobdService is b mock implementbtion of the UplobdService interfbce
// (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv) used for
// unit testing.
type MockUplobdService struct {
	// GetDumpsByIDsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetDumpsByIDs.
	GetDumpsByIDsFunc *UplobdServiceGetDumpsByIDsFunc
	// GetDumpsWithDefinitionsForMonikersFunc is bn instbnce of b mock
	// function object controlling the behbvior of the method
	// GetDumpsWithDefinitionsForMonikers.
	GetDumpsWithDefinitionsForMonikersFunc *UplobdServiceGetDumpsWithDefinitionsForMonikersFunc
	// GetUplobdIDsWithReferencesFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// GetUplobdIDsWithReferences.
	GetUplobdIDsWithReferencesFunc *UplobdServiceGetUplobdIDsWithReferencesFunc
	// InferClosestUplobdsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method InferClosestUplobds.
	InferClosestUplobdsFunc *UplobdServiceInferClosestUplobdsFunc
}

// NewMockUplobdService crebtes b new mock of the UplobdService interfbce.
// All methods return zero vblues for bll results, unless overwritten.
func NewMockUplobdService() *MockUplobdService {
	return &MockUplobdService{
		GetDumpsByIDsFunc: &UplobdServiceGetDumpsByIDsFunc{
			defbultHook: func(context.Context, []int) (r0 []shbred1.Dump, r1 error) {
				return
			},
		},
		GetDumpsWithDefinitionsForMonikersFunc: &UplobdServiceGetDumpsWithDefinitionsForMonikersFunc{
			defbultHook: func(context.Context, []precise.QublifiedMonikerDbtb) (r0 []shbred1.Dump, r1 error) {
				return
			},
		},
		GetUplobdIDsWithReferencesFunc: &UplobdServiceGetUplobdIDsWithReferencesFunc{
			defbultHook: func(context.Context, []precise.QublifiedMonikerDbtb, []int, int, string, int, int) (r0 []int, r1 int, r2 int, r3 error) {
				return
			},
		},
		InferClosestUplobdsFunc: &UplobdServiceInferClosestUplobdsFunc{
			defbultHook: func(context.Context, int, string, string, bool, string) (r0 []shbred1.Dump, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockUplobdService crebtes b new mock of the UplobdService
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockUplobdService() *MockUplobdService {
	return &MockUplobdService{
		GetDumpsByIDsFunc: &UplobdServiceGetDumpsByIDsFunc{
			defbultHook: func(context.Context, []int) ([]shbred1.Dump, error) {
				pbnic("unexpected invocbtion of MockUplobdService.GetDumpsByIDs")
			},
		},
		GetDumpsWithDefinitionsForMonikersFunc: &UplobdServiceGetDumpsWithDefinitionsForMonikersFunc{
			defbultHook: func(context.Context, []precise.QublifiedMonikerDbtb) ([]shbred1.Dump, error) {
				pbnic("unexpected invocbtion of MockUplobdService.GetDumpsWithDefinitionsForMonikers")
			},
		},
		GetUplobdIDsWithReferencesFunc: &UplobdServiceGetUplobdIDsWithReferencesFunc{
			defbultHook: func(context.Context, []precise.QublifiedMonikerDbtb, []int, int, string, int, int) ([]int, int, int, error) {
				pbnic("unexpected invocbtion of MockUplobdService.GetUplobdIDsWithReferences")
			},
		},
		InferClosestUplobdsFunc: &UplobdServiceInferClosestUplobdsFunc{
			defbultHook: func(context.Context, int, string, string, bool, string) ([]shbred1.Dump, error) {
				pbnic("unexpected invocbtion of MockUplobdService.InferClosestUplobds")
			},
		},
	}
}

// NewMockUplobdServiceFrom crebtes b new mock of the MockUplobdService
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockUplobdServiceFrom(i UplobdService) *MockUplobdService {
	return &MockUplobdService{
		GetDumpsByIDsFunc: &UplobdServiceGetDumpsByIDsFunc{
			defbultHook: i.GetDumpsByIDs,
		},
		GetDumpsWithDefinitionsForMonikersFunc: &UplobdServiceGetDumpsWithDefinitionsForMonikersFunc{
			defbultHook: i.GetDumpsWithDefinitionsForMonikers,
		},
		GetUplobdIDsWithReferencesFunc: &UplobdServiceGetUplobdIDsWithReferencesFunc{
			defbultHook: i.GetUplobdIDsWithReferences,
		},
		InferClosestUplobdsFunc: &UplobdServiceInferClosestUplobdsFunc{
			defbultHook: i.InferClosestUplobds,
		},
	}
}

// UplobdServiceGetDumpsByIDsFunc describes the behbvior when the
// GetDumpsByIDs method of the pbrent MockUplobdService instbnce is invoked.
type UplobdServiceGetDumpsByIDsFunc struct {
	defbultHook func(context.Context, []int) ([]shbred1.Dump, error)
	hooks       []func(context.Context, []int) ([]shbred1.Dump, error)
	history     []UplobdServiceGetDumpsByIDsFuncCbll
	mutex       sync.Mutex
}

// GetDumpsByIDs delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdService) GetDumpsByIDs(v0 context.Context, v1 []int) ([]shbred1.Dump, error) {
	r0, r1 := m.GetDumpsByIDsFunc.nextHook()(v0, v1)
	m.GetDumpsByIDsFunc.bppendCbll(UplobdServiceGetDumpsByIDsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetDumpsByIDs method
// of the pbrent MockUplobdService instbnce is invoked bnd the hook queue is
// empty.
func (f *UplobdServiceGetDumpsByIDsFunc) SetDefbultHook(hook func(context.Context, []int) ([]shbred1.Dump, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetDumpsByIDs method of the pbrent MockUplobdService instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *UplobdServiceGetDumpsByIDsFunc) PushHook(hook func(context.Context, []int) ([]shbred1.Dump, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdServiceGetDumpsByIDsFunc) SetDefbultReturn(r0 []shbred1.Dump, r1 error) {
	f.SetDefbultHook(func(context.Context, []int) ([]shbred1.Dump, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdServiceGetDumpsByIDsFunc) PushReturn(r0 []shbred1.Dump, r1 error) {
	f.PushHook(func(context.Context, []int) ([]shbred1.Dump, error) {
		return r0, r1
	})
}

func (f *UplobdServiceGetDumpsByIDsFunc) nextHook() func(context.Context, []int) ([]shbred1.Dump, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdServiceGetDumpsByIDsFunc) bppendCbll(r0 UplobdServiceGetDumpsByIDsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdServiceGetDumpsByIDsFuncCbll objects
// describing the invocbtions of this function.
func (f *UplobdServiceGetDumpsByIDsFunc) History() []UplobdServiceGetDumpsByIDsFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdServiceGetDumpsByIDsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdServiceGetDumpsByIDsFuncCbll is bn object thbt describes bn
// invocbtion of method GetDumpsByIDs on bn instbnce of MockUplobdService.
type UplobdServiceGetDumpsByIDsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred1.Dump
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdServiceGetDumpsByIDsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdServiceGetDumpsByIDsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// UplobdServiceGetDumpsWithDefinitionsForMonikersFunc describes the
// behbvior when the GetDumpsWithDefinitionsForMonikers method of the pbrent
// MockUplobdService instbnce is invoked.
type UplobdServiceGetDumpsWithDefinitionsForMonikersFunc struct {
	defbultHook func(context.Context, []precise.QublifiedMonikerDbtb) ([]shbred1.Dump, error)
	hooks       []func(context.Context, []precise.QublifiedMonikerDbtb) ([]shbred1.Dump, error)
	history     []UplobdServiceGetDumpsWithDefinitionsForMonikersFuncCbll
	mutex       sync.Mutex
}

// GetDumpsWithDefinitionsForMonikers delegbtes to the next hook function in
// the queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdService) GetDumpsWithDefinitionsForMonikers(v0 context.Context, v1 []precise.QublifiedMonikerDbtb) ([]shbred1.Dump, error) {
	r0, r1 := m.GetDumpsWithDefinitionsForMonikersFunc.nextHook()(v0, v1)
	m.GetDumpsWithDefinitionsForMonikersFunc.bppendCbll(UplobdServiceGetDumpsWithDefinitionsForMonikersFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetDumpsWithDefinitionsForMonikers method of the pbrent MockUplobdService
// instbnce is invoked bnd the hook queue is empty.
func (f *UplobdServiceGetDumpsWithDefinitionsForMonikersFunc) SetDefbultHook(hook func(context.Context, []precise.QublifiedMonikerDbtb) ([]shbred1.Dump, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetDumpsWithDefinitionsForMonikers method of the pbrent MockUplobdService
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *UplobdServiceGetDumpsWithDefinitionsForMonikersFunc) PushHook(hook func(context.Context, []precise.QublifiedMonikerDbtb) ([]shbred1.Dump, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdServiceGetDumpsWithDefinitionsForMonikersFunc) SetDefbultReturn(r0 []shbred1.Dump, r1 error) {
	f.SetDefbultHook(func(context.Context, []precise.QublifiedMonikerDbtb) ([]shbred1.Dump, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdServiceGetDumpsWithDefinitionsForMonikersFunc) PushReturn(r0 []shbred1.Dump, r1 error) {
	f.PushHook(func(context.Context, []precise.QublifiedMonikerDbtb) ([]shbred1.Dump, error) {
		return r0, r1
	})
}

func (f *UplobdServiceGetDumpsWithDefinitionsForMonikersFunc) nextHook() func(context.Context, []precise.QublifiedMonikerDbtb) ([]shbred1.Dump, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdServiceGetDumpsWithDefinitionsForMonikersFunc) bppendCbll(r0 UplobdServiceGetDumpsWithDefinitionsForMonikersFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// UplobdServiceGetDumpsWithDefinitionsForMonikersFuncCbll objects
// describing the invocbtions of this function.
func (f *UplobdServiceGetDumpsWithDefinitionsForMonikersFunc) History() []UplobdServiceGetDumpsWithDefinitionsForMonikersFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdServiceGetDumpsWithDefinitionsForMonikersFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdServiceGetDumpsWithDefinitionsForMonikersFuncCbll is bn object thbt
// describes bn invocbtion of method GetDumpsWithDefinitionsForMonikers on
// bn instbnce of MockUplobdService.
type UplobdServiceGetDumpsWithDefinitionsForMonikersFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []precise.QublifiedMonikerDbtb
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred1.Dump
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdServiceGetDumpsWithDefinitionsForMonikersFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdServiceGetDumpsWithDefinitionsForMonikersFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// UplobdServiceGetUplobdIDsWithReferencesFunc describes the behbvior when
// the GetUplobdIDsWithReferences method of the pbrent MockUplobdService
// instbnce is invoked.
type UplobdServiceGetUplobdIDsWithReferencesFunc struct {
	defbultHook func(context.Context, []precise.QublifiedMonikerDbtb, []int, int, string, int, int) ([]int, int, int, error)
	hooks       []func(context.Context, []precise.QublifiedMonikerDbtb, []int, int, string, int, int) ([]int, int, int, error)
	history     []UplobdServiceGetUplobdIDsWithReferencesFuncCbll
	mutex       sync.Mutex
}

// GetUplobdIDsWithReferences delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdService) GetUplobdIDsWithReferences(v0 context.Context, v1 []precise.QublifiedMonikerDbtb, v2 []int, v3 int, v4 string, v5 int, v6 int) ([]int, int, int, error) {
	r0, r1, r2, r3 := m.GetUplobdIDsWithReferencesFunc.nextHook()(v0, v1, v2, v3, v4, v5, v6)
	m.GetUplobdIDsWithReferencesFunc.bppendCbll(UplobdServiceGetUplobdIDsWithReferencesFuncCbll{v0, v1, v2, v3, v4, v5, v6, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefbultHook sets function thbt is cblled when the
// GetUplobdIDsWithReferences method of the pbrent MockUplobdService
// instbnce is invoked bnd the hook queue is empty.
func (f *UplobdServiceGetUplobdIDsWithReferencesFunc) SetDefbultHook(hook func(context.Context, []precise.QublifiedMonikerDbtb, []int, int, string, int, int) ([]int, int, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetUplobdIDsWithReferences method of the pbrent MockUplobdService
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *UplobdServiceGetUplobdIDsWithReferencesFunc) PushHook(hook func(context.Context, []precise.QublifiedMonikerDbtb, []int, int, string, int, int) ([]int, int, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdServiceGetUplobdIDsWithReferencesFunc) SetDefbultReturn(r0 []int, r1 int, r2 int, r3 error) {
	f.SetDefbultHook(func(context.Context, []precise.QublifiedMonikerDbtb, []int, int, string, int, int) ([]int, int, int, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdServiceGetUplobdIDsWithReferencesFunc) PushReturn(r0 []int, r1 int, r2 int, r3 error) {
	f.PushHook(func(context.Context, []precise.QublifiedMonikerDbtb, []int, int, string, int, int) ([]int, int, int, error) {
		return r0, r1, r2, r3
	})
}

func (f *UplobdServiceGetUplobdIDsWithReferencesFunc) nextHook() func(context.Context, []precise.QublifiedMonikerDbtb, []int, int, string, int, int) ([]int, int, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdServiceGetUplobdIDsWithReferencesFunc) bppendCbll(r0 UplobdServiceGetUplobdIDsWithReferencesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// UplobdServiceGetUplobdIDsWithReferencesFuncCbll objects describing the
// invocbtions of this function.
func (f *UplobdServiceGetUplobdIDsWithReferencesFunc) History() []UplobdServiceGetUplobdIDsWithReferencesFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdServiceGetUplobdIDsWithReferencesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdServiceGetUplobdIDsWithReferencesFuncCbll is bn object thbt
// describes bn invocbtion of method GetUplobdIDsWithReferences on bn
// instbnce of MockUplobdService.
type UplobdServiceGetUplobdIDsWithReferencesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []precise.QublifiedMonikerDbtb
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 []int
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 int
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 string
	// Arg5 is the vblue of the 6th brgument pbssed to this method
	// invocbtion.
	Arg5 int
	// Arg6 is the vblue of the 7th brgument pbssed to this method
	// invocbtion.
	Arg6 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 int
	// Result3 is the vblue of the 4th result returned from this method
	// invocbtion.
	Result3 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdServiceGetUplobdIDsWithReferencesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5, c.Arg6}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdServiceGetUplobdIDsWithReferencesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2, c.Result3}
}

// UplobdServiceInferClosestUplobdsFunc describes the behbvior when the
// InferClosestUplobds method of the pbrent MockUplobdService instbnce is
// invoked.
type UplobdServiceInferClosestUplobdsFunc struct {
	defbultHook func(context.Context, int, string, string, bool, string) ([]shbred1.Dump, error)
	hooks       []func(context.Context, int, string, string, bool, string) ([]shbred1.Dump, error)
	history     []UplobdServiceInferClosestUplobdsFuncCbll
	mutex       sync.Mutex
}

// InferClosestUplobds delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdService) InferClosestUplobds(v0 context.Context, v1 int, v2 string, v3 string, v4 bool, v5 string) ([]shbred1.Dump, error) {
	r0, r1 := m.InferClosestUplobdsFunc.nextHook()(v0, v1, v2, v3, v4, v5)
	m.InferClosestUplobdsFunc.bppendCbll(UplobdServiceInferClosestUplobdsFuncCbll{v0, v1, v2, v3, v4, v5, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the InferClosestUplobds
// method of the pbrent MockUplobdService instbnce is invoked bnd the hook
// queue is empty.
func (f *UplobdServiceInferClosestUplobdsFunc) SetDefbultHook(hook func(context.Context, int, string, string, bool, string) ([]shbred1.Dump, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// InferClosestUplobds method of the pbrent MockUplobdService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *UplobdServiceInferClosestUplobdsFunc) PushHook(hook func(context.Context, int, string, string, bool, string) ([]shbred1.Dump, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdServiceInferClosestUplobdsFunc) SetDefbultReturn(r0 []shbred1.Dump, r1 error) {
	f.SetDefbultHook(func(context.Context, int, string, string, bool, string) ([]shbred1.Dump, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdServiceInferClosestUplobdsFunc) PushReturn(r0 []shbred1.Dump, r1 error) {
	f.PushHook(func(context.Context, int, string, string, bool, string) ([]shbred1.Dump, error) {
		return r0, r1
	})
}

func (f *UplobdServiceInferClosestUplobdsFunc) nextHook() func(context.Context, int, string, string, bool, string) ([]shbred1.Dump, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdServiceInferClosestUplobdsFunc) bppendCbll(r0 UplobdServiceInferClosestUplobdsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdServiceInferClosestUplobdsFuncCbll
// objects describing the invocbtions of this function.
func (f *UplobdServiceInferClosestUplobdsFunc) History() []UplobdServiceInferClosestUplobdsFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdServiceInferClosestUplobdsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdServiceInferClosestUplobdsFuncCbll is bn object thbt describes bn
// invocbtion of method InferClosestUplobds on bn instbnce of
// MockUplobdService.
type UplobdServiceInferClosestUplobdsFuncCbll struct {
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
	Result0 []shbred1.Dump
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdServiceInferClosestUplobdsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdServiceInferClosestUplobdsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}
