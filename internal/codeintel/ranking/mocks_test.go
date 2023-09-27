// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge rbnking

import (
	"context"
	"sync"
	"time"

	bpi "github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	store "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/store"
	shbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/shbred"
	shbred1 "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	conftypes "github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	schemb "github.com/sourcegrbph/sourcegrbph/schemb"
)

// MockStore is b mock implementbtion of the Store interfbce (from the
// pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/rbnking/internbl/store)
// used for unit testing.
type MockStore struct {
	// BumpDerivbtiveGrbphKeyFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method BumpDerivbtiveGrbphKey.
	BumpDerivbtiveGrbphKeyFunc *StoreBumpDerivbtiveGrbphKeyFunc
	// CoordinbteFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method Coordinbte.
	CoordinbteFunc *StoreCoordinbteFunc
	// CoverbgeCountsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CoverbgeCounts.
	CoverbgeCountsFunc *StoreCoverbgeCountsFunc
	// DeleteRbnkingProgressFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method DeleteRbnkingProgress.
	DeleteRbnkingProgressFunc *StoreDeleteRbnkingProgressFunc
	// DerivbtiveGrbphKeyFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method DerivbtiveGrbphKey.
	DerivbtiveGrbphKeyFunc *StoreDerivbtiveGrbphKeyFunc
	// GetDocumentRbnksFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetDocumentRbnks.
	GetDocumentRbnksFunc *StoreGetDocumentRbnksFunc
	// GetReferenceCountStbtisticsFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// GetReferenceCountStbtistics.
	GetReferenceCountStbtisticsFunc *StoreGetReferenceCountStbtisticsFunc
	// GetStbrRbnkFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetStbrRbnk.
	GetStbrRbnkFunc *StoreGetStbrRbnkFunc
	// GetUplobdsForRbnkingFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetUplobdsForRbnking.
	GetUplobdsForRbnkingFunc *StoreGetUplobdsForRbnkingFunc
	// InsertDefinitionsForRbnkingFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// InsertDefinitionsForRbnking.
	InsertDefinitionsForRbnkingFunc *StoreInsertDefinitionsForRbnkingFunc
	// InsertInitiblPbthCountsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method InsertInitiblPbthCounts.
	InsertInitiblPbthCountsFunc *StoreInsertInitiblPbthCountsFunc
	// InsertInitiblPbthRbnksFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method InsertInitiblPbthRbnks.
	InsertInitiblPbthRbnksFunc *StoreInsertInitiblPbthRbnksFunc
	// InsertPbthCountInputsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method InsertPbthCountInputs.
	InsertPbthCountInputsFunc *StoreInsertPbthCountInputsFunc
	// InsertPbthRbnksFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method InsertPbthRbnks.
	InsertPbthRbnksFunc *StoreInsertPbthRbnksFunc
	// InsertReferencesForRbnkingFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// InsertReferencesForRbnking.
	InsertReferencesForRbnkingFunc *StoreInsertReferencesForRbnkingFunc
	// LbstUpdbtedAtFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method LbstUpdbtedAt.
	LbstUpdbtedAtFunc *StoreLbstUpdbtedAtFunc
	// SoftDeleteStbleExportedUplobdsFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// SoftDeleteStbleExportedUplobds.
	SoftDeleteStbleExportedUplobdsFunc *StoreSoftDeleteStbleExportedUplobdsFunc
	// SummbriesFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method Summbries.
	SummbriesFunc *StoreSummbriesFunc
	// VbcuumAbbndonedExportedUplobdsFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// VbcuumAbbndonedExportedUplobds.
	VbcuumAbbndonedExportedUplobdsFunc *StoreVbcuumAbbndonedExportedUplobdsFunc
	// VbcuumDeletedExportedUplobdsFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// VbcuumDeletedExportedUplobds.
	VbcuumDeletedExportedUplobdsFunc *StoreVbcuumDeletedExportedUplobdsFunc
	// VbcuumStbleGrbphsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method VbcuumStbleGrbphs.
	VbcuumStbleGrbphsFunc *StoreVbcuumStbleGrbphsFunc
	// VbcuumStbleProcessedPbthsFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// VbcuumStbleProcessedPbths.
	VbcuumStbleProcessedPbthsFunc *StoreVbcuumStbleProcessedPbthsFunc
	// VbcuumStbleProcessedReferencesFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// VbcuumStbleProcessedReferences.
	VbcuumStbleProcessedReferencesFunc *StoreVbcuumStbleProcessedReferencesFunc
	// VbcuumStbleRbnksFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method VbcuumStbleRbnks.
	VbcuumStbleRbnksFunc *StoreVbcuumStbleRbnksFunc
	// WithTrbnsbctionFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method WithTrbnsbction.
	WithTrbnsbctionFunc *StoreWithTrbnsbctionFunc
}

// NewMockStore crebtes b new mock of the Store interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockStore() *MockStore {
	return &MockStore{
		BumpDerivbtiveGrbphKeyFunc: &StoreBumpDerivbtiveGrbphKeyFunc{
			defbultHook: func(context.Context) (r0 error) {
				return
			},
		},
		CoordinbteFunc: &StoreCoordinbteFunc{
			defbultHook: func(context.Context, string) (r0 error) {
				return
			},
		},
		CoverbgeCountsFunc: &StoreCoverbgeCountsFunc{
			defbultHook: func(context.Context, string) (r0 shbred.CoverbgeCounts, r1 error) {
				return
			},
		},
		DeleteRbnkingProgressFunc: &StoreDeleteRbnkingProgressFunc{
			defbultHook: func(context.Context, string) (r0 error) {
				return
			},
		},
		DerivbtiveGrbphKeyFunc: &StoreDerivbtiveGrbphKeyFunc{
			defbultHook: func(context.Context) (r0 string, r1 time.Time, r2 bool, r3 error) {
				return
			},
		},
		GetDocumentRbnksFunc: &StoreGetDocumentRbnksFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (r0 mbp[string]flobt64, r1 bool, r2 error) {
				return
			},
		},
		GetReferenceCountStbtisticsFunc: &StoreGetReferenceCountStbtisticsFunc{
			defbultHook: func(context.Context) (r0 flobt64, r1 error) {
				return
			},
		},
		GetStbrRbnkFunc: &StoreGetStbrRbnkFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (r0 flobt64, r1 error) {
				return
			},
		},
		GetUplobdsForRbnkingFunc: &StoreGetUplobdsForRbnkingFunc{
			defbultHook: func(context.Context, string, string, int) (r0 []shbred1.ExportedUplobd, r1 error) {
				return
			},
		},
		InsertDefinitionsForRbnkingFunc: &StoreInsertDefinitionsForRbnkingFunc{
			defbultHook: func(context.Context, string, chbn shbred.RbnkingDefinitions) (r0 error) {
				return
			},
		},
		InsertInitiblPbthCountsFunc: &StoreInsertInitiblPbthCountsFunc{
			defbultHook: func(context.Context, string, int) (r0 int, r1 int, r2 error) {
				return
			},
		},
		InsertInitiblPbthRbnksFunc: &StoreInsertInitiblPbthRbnksFunc{
			defbultHook: func(context.Context, int, []string, int, string) (r0 error) {
				return
			},
		},
		InsertPbthCountInputsFunc: &StoreInsertPbthCountInputsFunc{
			defbultHook: func(context.Context, string, int) (r0 int, r1 int, r2 error) {
				return
			},
		},
		InsertPbthRbnksFunc: &StoreInsertPbthRbnksFunc{
			defbultHook: func(context.Context, string, int) (r0 int, r1 int, r2 error) {
				return
			},
		},
		InsertReferencesForRbnkingFunc: &StoreInsertReferencesForRbnkingFunc{
			defbultHook: func(context.Context, string, int, int, chbn [16]byte) (r0 error) {
				return
			},
		},
		LbstUpdbtedAtFunc: &StoreLbstUpdbtedAtFunc{
			defbultHook: func(context.Context, []bpi.RepoID) (r0 mbp[bpi.RepoID]time.Time, r1 error) {
				return
			},
		},
		SoftDeleteStbleExportedUplobdsFunc: &StoreSoftDeleteStbleExportedUplobdsFunc{
			defbultHook: func(context.Context, string) (r0 int, r1 int, r2 error) {
				return
			},
		},
		SummbriesFunc: &StoreSummbriesFunc{
			defbultHook: func(context.Context) (r0 []shbred.Summbry, r1 error) {
				return
			},
		},
		VbcuumAbbndonedExportedUplobdsFunc: &StoreVbcuumAbbndonedExportedUplobdsFunc{
			defbultHook: func(context.Context, string, int) (r0 int, r1 error) {
				return
			},
		},
		VbcuumDeletedExportedUplobdsFunc: &StoreVbcuumDeletedExportedUplobdsFunc{
			defbultHook: func(context.Context, string) (r0 int, r1 error) {
				return
			},
		},
		VbcuumStbleGrbphsFunc: &StoreVbcuumStbleGrbphsFunc{
			defbultHook: func(context.Context, string, int) (r0 int, r1 error) {
				return
			},
		},
		VbcuumStbleProcessedPbthsFunc: &StoreVbcuumStbleProcessedPbthsFunc{
			defbultHook: func(context.Context, string, int) (r0 int, r1 error) {
				return
			},
		},
		VbcuumStbleProcessedReferencesFunc: &StoreVbcuumStbleProcessedReferencesFunc{
			defbultHook: func(context.Context, string, int) (r0 int, r1 error) {
				return
			},
		},
		VbcuumStbleRbnksFunc: &StoreVbcuumStbleRbnksFunc{
			defbultHook: func(context.Context, string) (r0 int, r1 int, r2 error) {
				return
			},
		},
		WithTrbnsbctionFunc: &StoreWithTrbnsbctionFunc{
			defbultHook: func(context.Context, func(tx store.Store) error) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockStore crebtes b new mock of the Store interfbce. All methods
// pbnic on invocbtion, unless overwritten.
func NewStrictMockStore() *MockStore {
	return &MockStore{
		BumpDerivbtiveGrbphKeyFunc: &StoreBumpDerivbtiveGrbphKeyFunc{
			defbultHook: func(context.Context) error {
				pbnic("unexpected invocbtion of MockStore.BumpDerivbtiveGrbphKey")
			},
		},
		CoordinbteFunc: &StoreCoordinbteFunc{
			defbultHook: func(context.Context, string) error {
				pbnic("unexpected invocbtion of MockStore.Coordinbte")
			},
		},
		CoverbgeCountsFunc: &StoreCoverbgeCountsFunc{
			defbultHook: func(context.Context, string) (shbred.CoverbgeCounts, error) {
				pbnic("unexpected invocbtion of MockStore.CoverbgeCounts")
			},
		},
		DeleteRbnkingProgressFunc: &StoreDeleteRbnkingProgressFunc{
			defbultHook: func(context.Context, string) error {
				pbnic("unexpected invocbtion of MockStore.DeleteRbnkingProgress")
			},
		},
		DerivbtiveGrbphKeyFunc: &StoreDerivbtiveGrbphKeyFunc{
			defbultHook: func(context.Context) (string, time.Time, bool, error) {
				pbnic("unexpected invocbtion of MockStore.DerivbtiveGrbphKey")
			},
		},
		GetDocumentRbnksFunc: &StoreGetDocumentRbnksFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (mbp[string]flobt64, bool, error) {
				pbnic("unexpected invocbtion of MockStore.GetDocumentRbnks")
			},
		},
		GetReferenceCountStbtisticsFunc: &StoreGetReferenceCountStbtisticsFunc{
			defbultHook: func(context.Context) (flobt64, error) {
				pbnic("unexpected invocbtion of MockStore.GetReferenceCountStbtistics")
			},
		},
		GetStbrRbnkFunc: &StoreGetStbrRbnkFunc{
			defbultHook: func(context.Context, bpi.RepoNbme) (flobt64, error) {
				pbnic("unexpected invocbtion of MockStore.GetStbrRbnk")
			},
		},
		GetUplobdsForRbnkingFunc: &StoreGetUplobdsForRbnkingFunc{
			defbultHook: func(context.Context, string, string, int) ([]shbred1.ExportedUplobd, error) {
				pbnic("unexpected invocbtion of MockStore.GetUplobdsForRbnking")
			},
		},
		InsertDefinitionsForRbnkingFunc: &StoreInsertDefinitionsForRbnkingFunc{
			defbultHook: func(context.Context, string, chbn shbred.RbnkingDefinitions) error {
				pbnic("unexpected invocbtion of MockStore.InsertDefinitionsForRbnking")
			},
		},
		InsertInitiblPbthCountsFunc: &StoreInsertInitiblPbthCountsFunc{
			defbultHook: func(context.Context, string, int) (int, int, error) {
				pbnic("unexpected invocbtion of MockStore.InsertInitiblPbthCounts")
			},
		},
		InsertInitiblPbthRbnksFunc: &StoreInsertInitiblPbthRbnksFunc{
			defbultHook: func(context.Context, int, []string, int, string) error {
				pbnic("unexpected invocbtion of MockStore.InsertInitiblPbthRbnks")
			},
		},
		InsertPbthCountInputsFunc: &StoreInsertPbthCountInputsFunc{
			defbultHook: func(context.Context, string, int) (int, int, error) {
				pbnic("unexpected invocbtion of MockStore.InsertPbthCountInputs")
			},
		},
		InsertPbthRbnksFunc: &StoreInsertPbthRbnksFunc{
			defbultHook: func(context.Context, string, int) (int, int, error) {
				pbnic("unexpected invocbtion of MockStore.InsertPbthRbnks")
			},
		},
		InsertReferencesForRbnkingFunc: &StoreInsertReferencesForRbnkingFunc{
			defbultHook: func(context.Context, string, int, int, chbn [16]byte) error {
				pbnic("unexpected invocbtion of MockStore.InsertReferencesForRbnking")
			},
		},
		LbstUpdbtedAtFunc: &StoreLbstUpdbtedAtFunc{
			defbultHook: func(context.Context, []bpi.RepoID) (mbp[bpi.RepoID]time.Time, error) {
				pbnic("unexpected invocbtion of MockStore.LbstUpdbtedAt")
			},
		},
		SoftDeleteStbleExportedUplobdsFunc: &StoreSoftDeleteStbleExportedUplobdsFunc{
			defbultHook: func(context.Context, string) (int, int, error) {
				pbnic("unexpected invocbtion of MockStore.SoftDeleteStbleExportedUplobds")
			},
		},
		SummbriesFunc: &StoreSummbriesFunc{
			defbultHook: func(context.Context) ([]shbred.Summbry, error) {
				pbnic("unexpected invocbtion of MockStore.Summbries")
			},
		},
		VbcuumAbbndonedExportedUplobdsFunc: &StoreVbcuumAbbndonedExportedUplobdsFunc{
			defbultHook: func(context.Context, string, int) (int, error) {
				pbnic("unexpected invocbtion of MockStore.VbcuumAbbndonedExportedUplobds")
			},
		},
		VbcuumDeletedExportedUplobdsFunc: &StoreVbcuumDeletedExportedUplobdsFunc{
			defbultHook: func(context.Context, string) (int, error) {
				pbnic("unexpected invocbtion of MockStore.VbcuumDeletedExportedUplobds")
			},
		},
		VbcuumStbleGrbphsFunc: &StoreVbcuumStbleGrbphsFunc{
			defbultHook: func(context.Context, string, int) (int, error) {
				pbnic("unexpected invocbtion of MockStore.VbcuumStbleGrbphs")
			},
		},
		VbcuumStbleProcessedPbthsFunc: &StoreVbcuumStbleProcessedPbthsFunc{
			defbultHook: func(context.Context, string, int) (int, error) {
				pbnic("unexpected invocbtion of MockStore.VbcuumStbleProcessedPbths")
			},
		},
		VbcuumStbleProcessedReferencesFunc: &StoreVbcuumStbleProcessedReferencesFunc{
			defbultHook: func(context.Context, string, int) (int, error) {
				pbnic("unexpected invocbtion of MockStore.VbcuumStbleProcessedReferences")
			},
		},
		VbcuumStbleRbnksFunc: &StoreVbcuumStbleRbnksFunc{
			defbultHook: func(context.Context, string) (int, int, error) {
				pbnic("unexpected invocbtion of MockStore.VbcuumStbleRbnks")
			},
		},
		WithTrbnsbctionFunc: &StoreWithTrbnsbctionFunc{
			defbultHook: func(context.Context, func(tx store.Store) error) error {
				pbnic("unexpected invocbtion of MockStore.WithTrbnsbction")
			},
		},
	}
}

// NewMockStoreFrom crebtes b new mock of the MockStore interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockStoreFrom(i store.Store) *MockStore {
	return &MockStore{
		BumpDerivbtiveGrbphKeyFunc: &StoreBumpDerivbtiveGrbphKeyFunc{
			defbultHook: i.BumpDerivbtiveGrbphKey,
		},
		CoordinbteFunc: &StoreCoordinbteFunc{
			defbultHook: i.Coordinbte,
		},
		CoverbgeCountsFunc: &StoreCoverbgeCountsFunc{
			defbultHook: i.CoverbgeCounts,
		},
		DeleteRbnkingProgressFunc: &StoreDeleteRbnkingProgressFunc{
			defbultHook: i.DeleteRbnkingProgress,
		},
		DerivbtiveGrbphKeyFunc: &StoreDerivbtiveGrbphKeyFunc{
			defbultHook: i.DerivbtiveGrbphKey,
		},
		GetDocumentRbnksFunc: &StoreGetDocumentRbnksFunc{
			defbultHook: i.GetDocumentRbnks,
		},
		GetReferenceCountStbtisticsFunc: &StoreGetReferenceCountStbtisticsFunc{
			defbultHook: i.GetReferenceCountStbtistics,
		},
		GetStbrRbnkFunc: &StoreGetStbrRbnkFunc{
			defbultHook: i.GetStbrRbnk,
		},
		GetUplobdsForRbnkingFunc: &StoreGetUplobdsForRbnkingFunc{
			defbultHook: i.GetUplobdsForRbnking,
		},
		InsertDefinitionsForRbnkingFunc: &StoreInsertDefinitionsForRbnkingFunc{
			defbultHook: i.InsertDefinitionsForRbnking,
		},
		InsertInitiblPbthCountsFunc: &StoreInsertInitiblPbthCountsFunc{
			defbultHook: i.InsertInitiblPbthCounts,
		},
		InsertInitiblPbthRbnksFunc: &StoreInsertInitiblPbthRbnksFunc{
			defbultHook: i.InsertInitiblPbthRbnks,
		},
		InsertPbthCountInputsFunc: &StoreInsertPbthCountInputsFunc{
			defbultHook: i.InsertPbthCountInputs,
		},
		InsertPbthRbnksFunc: &StoreInsertPbthRbnksFunc{
			defbultHook: i.InsertPbthRbnks,
		},
		InsertReferencesForRbnkingFunc: &StoreInsertReferencesForRbnkingFunc{
			defbultHook: i.InsertReferencesForRbnking,
		},
		LbstUpdbtedAtFunc: &StoreLbstUpdbtedAtFunc{
			defbultHook: i.LbstUpdbtedAt,
		},
		SoftDeleteStbleExportedUplobdsFunc: &StoreSoftDeleteStbleExportedUplobdsFunc{
			defbultHook: i.SoftDeleteStbleExportedUplobds,
		},
		SummbriesFunc: &StoreSummbriesFunc{
			defbultHook: i.Summbries,
		},
		VbcuumAbbndonedExportedUplobdsFunc: &StoreVbcuumAbbndonedExportedUplobdsFunc{
			defbultHook: i.VbcuumAbbndonedExportedUplobds,
		},
		VbcuumDeletedExportedUplobdsFunc: &StoreVbcuumDeletedExportedUplobdsFunc{
			defbultHook: i.VbcuumDeletedExportedUplobds,
		},
		VbcuumStbleGrbphsFunc: &StoreVbcuumStbleGrbphsFunc{
			defbultHook: i.VbcuumStbleGrbphs,
		},
		VbcuumStbleProcessedPbthsFunc: &StoreVbcuumStbleProcessedPbthsFunc{
			defbultHook: i.VbcuumStbleProcessedPbths,
		},
		VbcuumStbleProcessedReferencesFunc: &StoreVbcuumStbleProcessedReferencesFunc{
			defbultHook: i.VbcuumStbleProcessedReferences,
		},
		VbcuumStbleRbnksFunc: &StoreVbcuumStbleRbnksFunc{
			defbultHook: i.VbcuumStbleRbnks,
		},
		WithTrbnsbctionFunc: &StoreWithTrbnsbctionFunc{
			defbultHook: i.WithTrbnsbction,
		},
	}
}

// StoreBumpDerivbtiveGrbphKeyFunc describes the behbvior when the
// BumpDerivbtiveGrbphKey method of the pbrent MockStore instbnce is
// invoked.
type StoreBumpDerivbtiveGrbphKeyFunc struct {
	defbultHook func(context.Context) error
	hooks       []func(context.Context) error
	history     []StoreBumpDerivbtiveGrbphKeyFuncCbll
	mutex       sync.Mutex
}

// BumpDerivbtiveGrbphKey delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) BumpDerivbtiveGrbphKey(v0 context.Context) error {
	r0 := m.BumpDerivbtiveGrbphKeyFunc.nextHook()(v0)
	m.BumpDerivbtiveGrbphKeyFunc.bppendCbll(StoreBumpDerivbtiveGrbphKeyFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// BumpDerivbtiveGrbphKey method of the pbrent MockStore instbnce is invoked
// bnd the hook queue is empty.
func (f *StoreBumpDerivbtiveGrbphKeyFunc) SetDefbultHook(hook func(context.Context) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// BumpDerivbtiveGrbphKey method of the pbrent MockStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *StoreBumpDerivbtiveGrbphKeyFunc) PushHook(hook func(context.Context) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreBumpDerivbtiveGrbphKeyFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreBumpDerivbtiveGrbphKeyFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context) error {
		return r0
	})
}

func (f *StoreBumpDerivbtiveGrbphKeyFunc) nextHook() func(context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreBumpDerivbtiveGrbphKeyFunc) bppendCbll(r0 StoreBumpDerivbtiveGrbphKeyFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreBumpDerivbtiveGrbphKeyFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreBumpDerivbtiveGrbphKeyFunc) History() []StoreBumpDerivbtiveGrbphKeyFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreBumpDerivbtiveGrbphKeyFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreBumpDerivbtiveGrbphKeyFuncCbll is bn object thbt describes bn
// invocbtion of method BumpDerivbtiveGrbphKey on bn instbnce of MockStore.
type StoreBumpDerivbtiveGrbphKeyFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreBumpDerivbtiveGrbphKeyFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreBumpDerivbtiveGrbphKeyFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreCoordinbteFunc describes the behbvior when the Coordinbte method of
// the pbrent MockStore instbnce is invoked.
type StoreCoordinbteFunc struct {
	defbultHook func(context.Context, string) error
	hooks       []func(context.Context, string) error
	history     []StoreCoordinbteFuncCbll
	mutex       sync.Mutex
}

// Coordinbte delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Coordinbte(v0 context.Context, v1 string) error {
	r0 := m.CoordinbteFunc.nextHook()(v0, v1)
	m.CoordinbteFunc.bppendCbll(StoreCoordinbteFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Coordinbte method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreCoordinbteFunc) SetDefbultHook(hook func(context.Context, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Coordinbte method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreCoordinbteFunc) PushHook(hook func(context.Context, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreCoordinbteFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreCoordinbteFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string) error {
		return r0
	})
}

func (f *StoreCoordinbteFunc) nextHook() func(context.Context, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreCoordinbteFunc) bppendCbll(r0 StoreCoordinbteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreCoordinbteFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreCoordinbteFunc) History() []StoreCoordinbteFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreCoordinbteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreCoordinbteFuncCbll is bn object thbt describes bn invocbtion of
// method Coordinbte on bn instbnce of MockStore.
type StoreCoordinbteFuncCbll struct {
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
func (c StoreCoordinbteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreCoordinbteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreCoverbgeCountsFunc describes the behbvior when the CoverbgeCounts
// method of the pbrent MockStore instbnce is invoked.
type StoreCoverbgeCountsFunc struct {
	defbultHook func(context.Context, string) (shbred.CoverbgeCounts, error)
	hooks       []func(context.Context, string) (shbred.CoverbgeCounts, error)
	history     []StoreCoverbgeCountsFuncCbll
	mutex       sync.Mutex
}

// CoverbgeCounts delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) CoverbgeCounts(v0 context.Context, v1 string) (shbred.CoverbgeCounts, error) {
	r0, r1 := m.CoverbgeCountsFunc.nextHook()(v0, v1)
	m.CoverbgeCountsFunc.bppendCbll(StoreCoverbgeCountsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CoverbgeCounts
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreCoverbgeCountsFunc) SetDefbultHook(hook func(context.Context, string) (shbred.CoverbgeCounts, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CoverbgeCounts method of the pbrent MockStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreCoverbgeCountsFunc) PushHook(hook func(context.Context, string) (shbred.CoverbgeCounts, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreCoverbgeCountsFunc) SetDefbultReturn(r0 shbred.CoverbgeCounts, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (shbred.CoverbgeCounts, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreCoverbgeCountsFunc) PushReturn(r0 shbred.CoverbgeCounts, r1 error) {
	f.PushHook(func(context.Context, string) (shbred.CoverbgeCounts, error) {
		return r0, r1
	})
}

func (f *StoreCoverbgeCountsFunc) nextHook() func(context.Context, string) (shbred.CoverbgeCounts, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreCoverbgeCountsFunc) bppendCbll(r0 StoreCoverbgeCountsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreCoverbgeCountsFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreCoverbgeCountsFunc) History() []StoreCoverbgeCountsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreCoverbgeCountsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreCoverbgeCountsFuncCbll is bn object thbt describes bn invocbtion of
// method CoverbgeCounts on bn instbnce of MockStore.
type StoreCoverbgeCountsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 shbred.CoverbgeCounts
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreCoverbgeCountsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreCoverbgeCountsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreDeleteRbnkingProgressFunc describes the behbvior when the
// DeleteRbnkingProgress method of the pbrent MockStore instbnce is invoked.
type StoreDeleteRbnkingProgressFunc struct {
	defbultHook func(context.Context, string) error
	hooks       []func(context.Context, string) error
	history     []StoreDeleteRbnkingProgressFuncCbll
	mutex       sync.Mutex
}

// DeleteRbnkingProgress delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) DeleteRbnkingProgress(v0 context.Context, v1 string) error {
	r0 := m.DeleteRbnkingProgressFunc.nextHook()(v0, v1)
	m.DeleteRbnkingProgressFunc.bppendCbll(StoreDeleteRbnkingProgressFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// DeleteRbnkingProgress method of the pbrent MockStore instbnce is invoked
// bnd the hook queue is empty.
func (f *StoreDeleteRbnkingProgressFunc) SetDefbultHook(hook func(context.Context, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DeleteRbnkingProgress method of the pbrent MockStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *StoreDeleteRbnkingProgressFunc) PushHook(hook func(context.Context, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreDeleteRbnkingProgressFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreDeleteRbnkingProgressFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string) error {
		return r0
	})
}

func (f *StoreDeleteRbnkingProgressFunc) nextHook() func(context.Context, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDeleteRbnkingProgressFunc) bppendCbll(r0 StoreDeleteRbnkingProgressFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreDeleteRbnkingProgressFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreDeleteRbnkingProgressFunc) History() []StoreDeleteRbnkingProgressFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreDeleteRbnkingProgressFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDeleteRbnkingProgressFuncCbll is bn object thbt describes bn
// invocbtion of method DeleteRbnkingProgress on bn instbnce of MockStore.
type StoreDeleteRbnkingProgressFuncCbll struct {
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
func (c StoreDeleteRbnkingProgressFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreDeleteRbnkingProgressFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreDerivbtiveGrbphKeyFunc describes the behbvior when the
// DerivbtiveGrbphKey method of the pbrent MockStore instbnce is invoked.
type StoreDerivbtiveGrbphKeyFunc struct {
	defbultHook func(context.Context) (string, time.Time, bool, error)
	hooks       []func(context.Context) (string, time.Time, bool, error)
	history     []StoreDerivbtiveGrbphKeyFuncCbll
	mutex       sync.Mutex
}

// DerivbtiveGrbphKey delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) DerivbtiveGrbphKey(v0 context.Context) (string, time.Time, bool, error) {
	r0, r1, r2, r3 := m.DerivbtiveGrbphKeyFunc.nextHook()(v0)
	m.DerivbtiveGrbphKeyFunc.bppendCbll(StoreDerivbtiveGrbphKeyFuncCbll{v0, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefbultHook sets function thbt is cblled when the DerivbtiveGrbphKey
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreDerivbtiveGrbphKeyFunc) SetDefbultHook(hook func(context.Context) (string, time.Time, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DerivbtiveGrbphKey method of the pbrent MockStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *StoreDerivbtiveGrbphKeyFunc) PushHook(hook func(context.Context) (string, time.Time, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreDerivbtiveGrbphKeyFunc) SetDefbultReturn(r0 string, r1 time.Time, r2 bool, r3 error) {
	f.SetDefbultHook(func(context.Context) (string, time.Time, bool, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreDerivbtiveGrbphKeyFunc) PushReturn(r0 string, r1 time.Time, r2 bool, r3 error) {
	f.PushHook(func(context.Context) (string, time.Time, bool, error) {
		return r0, r1, r2, r3
	})
}

func (f *StoreDerivbtiveGrbphKeyFunc) nextHook() func(context.Context) (string, time.Time, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreDerivbtiveGrbphKeyFunc) bppendCbll(r0 StoreDerivbtiveGrbphKeyFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreDerivbtiveGrbphKeyFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreDerivbtiveGrbphKeyFunc) History() []StoreDerivbtiveGrbphKeyFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreDerivbtiveGrbphKeyFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreDerivbtiveGrbphKeyFuncCbll is bn object thbt describes bn invocbtion
// of method DerivbtiveGrbphKey on bn instbnce of MockStore.
type StoreDerivbtiveGrbphKeyFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 time.Time
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 bool
	// Result3 is the vblue of the 4th result returned from this method
	// invocbtion.
	Result3 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreDerivbtiveGrbphKeyFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreDerivbtiveGrbphKeyFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2, c.Result3}
}

// StoreGetDocumentRbnksFunc describes the behbvior when the
// GetDocumentRbnks method of the pbrent MockStore instbnce is invoked.
type StoreGetDocumentRbnksFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme) (mbp[string]flobt64, bool, error)
	hooks       []func(context.Context, bpi.RepoNbme) (mbp[string]flobt64, bool, error)
	history     []StoreGetDocumentRbnksFuncCbll
	mutex       sync.Mutex
}

// GetDocumentRbnks delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) GetDocumentRbnks(v0 context.Context, v1 bpi.RepoNbme) (mbp[string]flobt64, bool, error) {
	r0, r1, r2 := m.GetDocumentRbnksFunc.nextHook()(v0, v1)
	m.GetDocumentRbnksFunc.bppendCbll(StoreGetDocumentRbnksFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the GetDocumentRbnks
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreGetDocumentRbnksFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme) (mbp[string]flobt64, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetDocumentRbnks method of the pbrent MockStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreGetDocumentRbnksFunc) PushHook(hook func(context.Context, bpi.RepoNbme) (mbp[string]flobt64, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGetDocumentRbnksFunc) SetDefbultReturn(r0 mbp[string]flobt64, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme) (mbp[string]flobt64, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGetDocumentRbnksFunc) PushReturn(r0 mbp[string]flobt64, r1 bool, r2 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme) (mbp[string]flobt64, bool, error) {
		return r0, r1, r2
	})
}

func (f *StoreGetDocumentRbnksFunc) nextHook() func(context.Context, bpi.RepoNbme) (mbp[string]flobt64, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetDocumentRbnksFunc) bppendCbll(r0 StoreGetDocumentRbnksFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreGetDocumentRbnksFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreGetDocumentRbnksFunc) History() []StoreGetDocumentRbnksFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGetDocumentRbnksFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetDocumentRbnksFuncCbll is bn object thbt describes bn invocbtion
// of method GetDocumentRbnks on bn instbnce of MockStore.
type StoreGetDocumentRbnksFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 mbp[string]flobt64
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGetDocumentRbnksFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGetDocumentRbnksFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// StoreGetReferenceCountStbtisticsFunc describes the behbvior when the
// GetReferenceCountStbtistics method of the pbrent MockStore instbnce is
// invoked.
type StoreGetReferenceCountStbtisticsFunc struct {
	defbultHook func(context.Context) (flobt64, error)
	hooks       []func(context.Context) (flobt64, error)
	history     []StoreGetReferenceCountStbtisticsFuncCbll
	mutex       sync.Mutex
}

// GetReferenceCountStbtistics delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) GetReferenceCountStbtistics(v0 context.Context) (flobt64, error) {
	r0, r1 := m.GetReferenceCountStbtisticsFunc.nextHook()(v0)
	m.GetReferenceCountStbtisticsFunc.bppendCbll(StoreGetReferenceCountStbtisticsFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetReferenceCountStbtistics method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreGetReferenceCountStbtisticsFunc) SetDefbultHook(hook func(context.Context) (flobt64, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetReferenceCountStbtistics method of the pbrent MockStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *StoreGetReferenceCountStbtisticsFunc) PushHook(hook func(context.Context) (flobt64, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGetReferenceCountStbtisticsFunc) SetDefbultReturn(r0 flobt64, r1 error) {
	f.SetDefbultHook(func(context.Context) (flobt64, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGetReferenceCountStbtisticsFunc) PushReturn(r0 flobt64, r1 error) {
	f.PushHook(func(context.Context) (flobt64, error) {
		return r0, r1
	})
}

func (f *StoreGetReferenceCountStbtisticsFunc) nextHook() func(context.Context) (flobt64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetReferenceCountStbtisticsFunc) bppendCbll(r0 StoreGetReferenceCountStbtisticsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreGetReferenceCountStbtisticsFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreGetReferenceCountStbtisticsFunc) History() []StoreGetReferenceCountStbtisticsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGetReferenceCountStbtisticsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetReferenceCountStbtisticsFuncCbll is bn object thbt describes bn
// invocbtion of method GetReferenceCountStbtistics on bn instbnce of
// MockStore.
type StoreGetReferenceCountStbtisticsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 flobt64
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGetReferenceCountStbtisticsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGetReferenceCountStbtisticsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreGetStbrRbnkFunc describes the behbvior when the GetStbrRbnk method
// of the pbrent MockStore instbnce is invoked.
type StoreGetStbrRbnkFunc struct {
	defbultHook func(context.Context, bpi.RepoNbme) (flobt64, error)
	hooks       []func(context.Context, bpi.RepoNbme) (flobt64, error)
	history     []StoreGetStbrRbnkFuncCbll
	mutex       sync.Mutex
}

// GetStbrRbnk delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) GetStbrRbnk(v0 context.Context, v1 bpi.RepoNbme) (flobt64, error) {
	r0, r1 := m.GetStbrRbnkFunc.nextHook()(v0, v1)
	m.GetStbrRbnkFunc.bppendCbll(StoreGetStbrRbnkFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetStbrRbnk method
// of the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreGetStbrRbnkFunc) SetDefbultHook(hook func(context.Context, bpi.RepoNbme) (flobt64, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetStbrRbnk method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreGetStbrRbnkFunc) PushHook(hook func(context.Context, bpi.RepoNbme) (flobt64, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGetStbrRbnkFunc) SetDefbultReturn(r0 flobt64, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoNbme) (flobt64, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGetStbrRbnkFunc) PushReturn(r0 flobt64, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoNbme) (flobt64, error) {
		return r0, r1
	})
}

func (f *StoreGetStbrRbnkFunc) nextHook() func(context.Context, bpi.RepoNbme) (flobt64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetStbrRbnkFunc) bppendCbll(r0 StoreGetStbrRbnkFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreGetStbrRbnkFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreGetStbrRbnkFunc) History() []StoreGetStbrRbnkFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGetStbrRbnkFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetStbrRbnkFuncCbll is bn object thbt describes bn invocbtion of
// method GetStbrRbnk on bn instbnce of MockStore.
type StoreGetStbrRbnkFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoNbme
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 flobt64
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGetStbrRbnkFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGetStbrRbnkFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreGetUplobdsForRbnkingFunc describes the behbvior when the
// GetUplobdsForRbnking method of the pbrent MockStore instbnce is invoked.
type StoreGetUplobdsForRbnkingFunc struct {
	defbultHook func(context.Context, string, string, int) ([]shbred1.ExportedUplobd, error)
	hooks       []func(context.Context, string, string, int) ([]shbred1.ExportedUplobd, error)
	history     []StoreGetUplobdsForRbnkingFuncCbll
	mutex       sync.Mutex
}

// GetUplobdsForRbnking delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) GetUplobdsForRbnking(v0 context.Context, v1 string, v2 string, v3 int) ([]shbred1.ExportedUplobd, error) {
	r0, r1 := m.GetUplobdsForRbnkingFunc.nextHook()(v0, v1, v2, v3)
	m.GetUplobdsForRbnkingFunc.bppendCbll(StoreGetUplobdsForRbnkingFuncCbll{v0, v1, v2, v3, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetUplobdsForRbnking
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreGetUplobdsForRbnkingFunc) SetDefbultHook(hook func(context.Context, string, string, int) ([]shbred1.ExportedUplobd, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetUplobdsForRbnking method of the pbrent MockStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *StoreGetUplobdsForRbnkingFunc) PushHook(hook func(context.Context, string, string, int) ([]shbred1.ExportedUplobd, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreGetUplobdsForRbnkingFunc) SetDefbultReturn(r0 []shbred1.ExportedUplobd, r1 error) {
	f.SetDefbultHook(func(context.Context, string, string, int) ([]shbred1.ExportedUplobd, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreGetUplobdsForRbnkingFunc) PushReturn(r0 []shbred1.ExportedUplobd, r1 error) {
	f.PushHook(func(context.Context, string, string, int) ([]shbred1.ExportedUplobd, error) {
		return r0, r1
	})
}

func (f *StoreGetUplobdsForRbnkingFunc) nextHook() func(context.Context, string, string, int) ([]shbred1.ExportedUplobd, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreGetUplobdsForRbnkingFunc) bppendCbll(r0 StoreGetUplobdsForRbnkingFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreGetUplobdsForRbnkingFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreGetUplobdsForRbnkingFunc) History() []StoreGetUplobdsForRbnkingFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreGetUplobdsForRbnkingFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreGetUplobdsForRbnkingFuncCbll is bn object thbt describes bn
// invocbtion of method GetUplobdsForRbnking on bn instbnce of MockStore.
type StoreGetUplobdsForRbnkingFuncCbll struct {
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
	Arg3 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred1.ExportedUplobd
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreGetUplobdsForRbnkingFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreGetUplobdsForRbnkingFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreInsertDefinitionsForRbnkingFunc describes the behbvior when the
// InsertDefinitionsForRbnking method of the pbrent MockStore instbnce is
// invoked.
type StoreInsertDefinitionsForRbnkingFunc struct {
	defbultHook func(context.Context, string, chbn shbred.RbnkingDefinitions) error
	hooks       []func(context.Context, string, chbn shbred.RbnkingDefinitions) error
	history     []StoreInsertDefinitionsForRbnkingFuncCbll
	mutex       sync.Mutex
}

// InsertDefinitionsForRbnking delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) InsertDefinitionsForRbnking(v0 context.Context, v1 string, v2 chbn shbred.RbnkingDefinitions) error {
	r0 := m.InsertDefinitionsForRbnkingFunc.nextHook()(v0, v1, v2)
	m.InsertDefinitionsForRbnkingFunc.bppendCbll(StoreInsertDefinitionsForRbnkingFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// InsertDefinitionsForRbnking method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreInsertDefinitionsForRbnkingFunc) SetDefbultHook(hook func(context.Context, string, chbn shbred.RbnkingDefinitions) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// InsertDefinitionsForRbnking method of the pbrent MockStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *StoreInsertDefinitionsForRbnkingFunc) PushHook(hook func(context.Context, string, chbn shbred.RbnkingDefinitions) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreInsertDefinitionsForRbnkingFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string, chbn shbred.RbnkingDefinitions) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreInsertDefinitionsForRbnkingFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string, chbn shbred.RbnkingDefinitions) error {
		return r0
	})
}

func (f *StoreInsertDefinitionsForRbnkingFunc) nextHook() func(context.Context, string, chbn shbred.RbnkingDefinitions) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreInsertDefinitionsForRbnkingFunc) bppendCbll(r0 StoreInsertDefinitionsForRbnkingFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreInsertDefinitionsForRbnkingFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreInsertDefinitionsForRbnkingFunc) History() []StoreInsertDefinitionsForRbnkingFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreInsertDefinitionsForRbnkingFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreInsertDefinitionsForRbnkingFuncCbll is bn object thbt describes bn
// invocbtion of method InsertDefinitionsForRbnking on bn instbnce of
// MockStore.
type StoreInsertDefinitionsForRbnkingFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 chbn shbred.RbnkingDefinitions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreInsertDefinitionsForRbnkingFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreInsertDefinitionsForRbnkingFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreInsertInitiblPbthCountsFunc describes the behbvior when the
// InsertInitiblPbthCounts method of the pbrent MockStore instbnce is
// invoked.
type StoreInsertInitiblPbthCountsFunc struct {
	defbultHook func(context.Context, string, int) (int, int, error)
	hooks       []func(context.Context, string, int) (int, int, error)
	history     []StoreInsertInitiblPbthCountsFuncCbll
	mutex       sync.Mutex
}

// InsertInitiblPbthCounts delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) InsertInitiblPbthCounts(v0 context.Context, v1 string, v2 int) (int, int, error) {
	r0, r1, r2 := m.InsertInitiblPbthCountsFunc.nextHook()(v0, v1, v2)
	m.InsertInitiblPbthCountsFunc.bppendCbll(StoreInsertInitiblPbthCountsFuncCbll{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// InsertInitiblPbthCounts method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreInsertInitiblPbthCountsFunc) SetDefbultHook(hook func(context.Context, string, int) (int, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// InsertInitiblPbthCounts method of the pbrent MockStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *StoreInsertInitiblPbthCountsFunc) PushHook(hook func(context.Context, string, int) (int, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreInsertInitiblPbthCountsFunc) SetDefbultReturn(r0 int, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, string, int) (int, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreInsertInitiblPbthCountsFunc) PushReturn(r0 int, r1 int, r2 error) {
	f.PushHook(func(context.Context, string, int) (int, int, error) {
		return r0, r1, r2
	})
}

func (f *StoreInsertInitiblPbthCountsFunc) nextHook() func(context.Context, string, int) (int, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreInsertInitiblPbthCountsFunc) bppendCbll(r0 StoreInsertInitiblPbthCountsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreInsertInitiblPbthCountsFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreInsertInitiblPbthCountsFunc) History() []StoreInsertInitiblPbthCountsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreInsertInitiblPbthCountsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreInsertInitiblPbthCountsFuncCbll is bn object thbt describes bn
// invocbtion of method InsertInitiblPbthCounts on bn instbnce of MockStore.
type StoreInsertInitiblPbthCountsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreInsertInitiblPbthCountsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreInsertInitiblPbthCountsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// StoreInsertInitiblPbthRbnksFunc describes the behbvior when the
// InsertInitiblPbthRbnks method of the pbrent MockStore instbnce is
// invoked.
type StoreInsertInitiblPbthRbnksFunc struct {
	defbultHook func(context.Context, int, []string, int, string) error
	hooks       []func(context.Context, int, []string, int, string) error
	history     []StoreInsertInitiblPbthRbnksFuncCbll
	mutex       sync.Mutex
}

// InsertInitiblPbthRbnks delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) InsertInitiblPbthRbnks(v0 context.Context, v1 int, v2 []string, v3 int, v4 string) error {
	r0 := m.InsertInitiblPbthRbnksFunc.nextHook()(v0, v1, v2, v3, v4)
	m.InsertInitiblPbthRbnksFunc.bppendCbll(StoreInsertInitiblPbthRbnksFuncCbll{v0, v1, v2, v3, v4, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// InsertInitiblPbthRbnks method of the pbrent MockStore instbnce is invoked
// bnd the hook queue is empty.
func (f *StoreInsertInitiblPbthRbnksFunc) SetDefbultHook(hook func(context.Context, int, []string, int, string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// InsertInitiblPbthRbnks method of the pbrent MockStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *StoreInsertInitiblPbthRbnksFunc) PushHook(hook func(context.Context, int, []string, int, string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreInsertInitiblPbthRbnksFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int, []string, int, string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreInsertInitiblPbthRbnksFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, []string, int, string) error {
		return r0
	})
}

func (f *StoreInsertInitiblPbthRbnksFunc) nextHook() func(context.Context, int, []string, int, string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreInsertInitiblPbthRbnksFunc) bppendCbll(r0 StoreInsertInitiblPbthRbnksFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreInsertInitiblPbthRbnksFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreInsertInitiblPbthRbnksFunc) History() []StoreInsertInitiblPbthRbnksFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreInsertInitiblPbthRbnksFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreInsertInitiblPbthRbnksFuncCbll is bn object thbt describes bn
// invocbtion of method InsertInitiblPbthRbnks on bn instbnce of MockStore.
type StoreInsertInitiblPbthRbnksFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 []string
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 int
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreInsertInitiblPbthRbnksFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreInsertInitiblPbthRbnksFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreInsertPbthCountInputsFunc describes the behbvior when the
// InsertPbthCountInputs method of the pbrent MockStore instbnce is invoked.
type StoreInsertPbthCountInputsFunc struct {
	defbultHook func(context.Context, string, int) (int, int, error)
	hooks       []func(context.Context, string, int) (int, int, error)
	history     []StoreInsertPbthCountInputsFuncCbll
	mutex       sync.Mutex
}

// InsertPbthCountInputs delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) InsertPbthCountInputs(v0 context.Context, v1 string, v2 int) (int, int, error) {
	r0, r1, r2 := m.InsertPbthCountInputsFunc.nextHook()(v0, v1, v2)
	m.InsertPbthCountInputsFunc.bppendCbll(StoreInsertPbthCountInputsFuncCbll{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// InsertPbthCountInputs method of the pbrent MockStore instbnce is invoked
// bnd the hook queue is empty.
func (f *StoreInsertPbthCountInputsFunc) SetDefbultHook(hook func(context.Context, string, int) (int, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// InsertPbthCountInputs method of the pbrent MockStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *StoreInsertPbthCountInputsFunc) PushHook(hook func(context.Context, string, int) (int, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreInsertPbthCountInputsFunc) SetDefbultReturn(r0 int, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, string, int) (int, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreInsertPbthCountInputsFunc) PushReturn(r0 int, r1 int, r2 error) {
	f.PushHook(func(context.Context, string, int) (int, int, error) {
		return r0, r1, r2
	})
}

func (f *StoreInsertPbthCountInputsFunc) nextHook() func(context.Context, string, int) (int, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreInsertPbthCountInputsFunc) bppendCbll(r0 StoreInsertPbthCountInputsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreInsertPbthCountInputsFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreInsertPbthCountInputsFunc) History() []StoreInsertPbthCountInputsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreInsertPbthCountInputsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreInsertPbthCountInputsFuncCbll is bn object thbt describes bn
// invocbtion of method InsertPbthCountInputs on bn instbnce of MockStore.
type StoreInsertPbthCountInputsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreInsertPbthCountInputsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreInsertPbthCountInputsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// StoreInsertPbthRbnksFunc describes the behbvior when the InsertPbthRbnks
// method of the pbrent MockStore instbnce is invoked.
type StoreInsertPbthRbnksFunc struct {
	defbultHook func(context.Context, string, int) (int, int, error)
	hooks       []func(context.Context, string, int) (int, int, error)
	history     []StoreInsertPbthRbnksFuncCbll
	mutex       sync.Mutex
}

// InsertPbthRbnks delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) InsertPbthRbnks(v0 context.Context, v1 string, v2 int) (int, int, error) {
	r0, r1, r2 := m.InsertPbthRbnksFunc.nextHook()(v0, v1, v2)
	m.InsertPbthRbnksFunc.bppendCbll(StoreInsertPbthRbnksFuncCbll{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the InsertPbthRbnks
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreInsertPbthRbnksFunc) SetDefbultHook(hook func(context.Context, string, int) (int, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// InsertPbthRbnks method of the pbrent MockStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreInsertPbthRbnksFunc) PushHook(hook func(context.Context, string, int) (int, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreInsertPbthRbnksFunc) SetDefbultReturn(r0 int, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, string, int) (int, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreInsertPbthRbnksFunc) PushReturn(r0 int, r1 int, r2 error) {
	f.PushHook(func(context.Context, string, int) (int, int, error) {
		return r0, r1, r2
	})
}

func (f *StoreInsertPbthRbnksFunc) nextHook() func(context.Context, string, int) (int, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreInsertPbthRbnksFunc) bppendCbll(r0 StoreInsertPbthRbnksFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreInsertPbthRbnksFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreInsertPbthRbnksFunc) History() []StoreInsertPbthRbnksFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreInsertPbthRbnksFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreInsertPbthRbnksFuncCbll is bn object thbt describes bn invocbtion of
// method InsertPbthRbnks on bn instbnce of MockStore.
type StoreInsertPbthRbnksFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreInsertPbthRbnksFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreInsertPbthRbnksFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// StoreInsertReferencesForRbnkingFunc describes the behbvior when the
// InsertReferencesForRbnking method of the pbrent MockStore instbnce is
// invoked.
type StoreInsertReferencesForRbnkingFunc struct {
	defbultHook func(context.Context, string, int, int, chbn [16]byte) error
	hooks       []func(context.Context, string, int, int, chbn [16]byte) error
	history     []StoreInsertReferencesForRbnkingFuncCbll
	mutex       sync.Mutex
}

// InsertReferencesForRbnking delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) InsertReferencesForRbnking(v0 context.Context, v1 string, v2 int, v3 int, v4 chbn [16]byte) error {
	r0 := m.InsertReferencesForRbnkingFunc.nextHook()(v0, v1, v2, v3, v4)
	m.InsertReferencesForRbnkingFunc.bppendCbll(StoreInsertReferencesForRbnkingFuncCbll{v0, v1, v2, v3, v4, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// InsertReferencesForRbnking method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreInsertReferencesForRbnkingFunc) SetDefbultHook(hook func(context.Context, string, int, int, chbn [16]byte) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// InsertReferencesForRbnking method of the pbrent MockStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *StoreInsertReferencesForRbnkingFunc) PushHook(hook func(context.Context, string, int, int, chbn [16]byte) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreInsertReferencesForRbnkingFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string, int, int, chbn [16]byte) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreInsertReferencesForRbnkingFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string, int, int, chbn [16]byte) error {
		return r0
	})
}

func (f *StoreInsertReferencesForRbnkingFunc) nextHook() func(context.Context, string, int, int, chbn [16]byte) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreInsertReferencesForRbnkingFunc) bppendCbll(r0 StoreInsertReferencesForRbnkingFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreInsertReferencesForRbnkingFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreInsertReferencesForRbnkingFunc) History() []StoreInsertReferencesForRbnkingFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreInsertReferencesForRbnkingFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreInsertReferencesForRbnkingFuncCbll is bn object thbt describes bn
// invocbtion of method InsertReferencesForRbnking on bn instbnce of
// MockStore.
type StoreInsertReferencesForRbnkingFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Arg3 is the vblue of the 4th brgument pbssed to this method
	// invocbtion.
	Arg3 int
	// Arg4 is the vblue of the 5th brgument pbssed to this method
	// invocbtion.
	Arg4 chbn [16]byte
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreInsertReferencesForRbnkingFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreInsertReferencesForRbnkingFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// StoreLbstUpdbtedAtFunc describes the behbvior when the LbstUpdbtedAt
// method of the pbrent MockStore instbnce is invoked.
type StoreLbstUpdbtedAtFunc struct {
	defbultHook func(context.Context, []bpi.RepoID) (mbp[bpi.RepoID]time.Time, error)
	hooks       []func(context.Context, []bpi.RepoID) (mbp[bpi.RepoID]time.Time, error)
	history     []StoreLbstUpdbtedAtFuncCbll
	mutex       sync.Mutex
}

// LbstUpdbtedAt delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) LbstUpdbtedAt(v0 context.Context, v1 []bpi.RepoID) (mbp[bpi.RepoID]time.Time, error) {
	r0, r1 := m.LbstUpdbtedAtFunc.nextHook()(v0, v1)
	m.LbstUpdbtedAtFunc.bppendCbll(StoreLbstUpdbtedAtFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the LbstUpdbtedAt method
// of the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreLbstUpdbtedAtFunc) SetDefbultHook(hook func(context.Context, []bpi.RepoID) (mbp[bpi.RepoID]time.Time, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// LbstUpdbtedAt method of the pbrent MockStore instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreLbstUpdbtedAtFunc) PushHook(hook func(context.Context, []bpi.RepoID) (mbp[bpi.RepoID]time.Time, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreLbstUpdbtedAtFunc) SetDefbultReturn(r0 mbp[bpi.RepoID]time.Time, r1 error) {
	f.SetDefbultHook(func(context.Context, []bpi.RepoID) (mbp[bpi.RepoID]time.Time, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreLbstUpdbtedAtFunc) PushReturn(r0 mbp[bpi.RepoID]time.Time, r1 error) {
	f.PushHook(func(context.Context, []bpi.RepoID) (mbp[bpi.RepoID]time.Time, error) {
		return r0, r1
	})
}

func (f *StoreLbstUpdbtedAtFunc) nextHook() func(context.Context, []bpi.RepoID) (mbp[bpi.RepoID]time.Time, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreLbstUpdbtedAtFunc) bppendCbll(r0 StoreLbstUpdbtedAtFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreLbstUpdbtedAtFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreLbstUpdbtedAtFunc) History() []StoreLbstUpdbtedAtFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreLbstUpdbtedAtFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreLbstUpdbtedAtFuncCbll is bn object thbt describes bn invocbtion of
// method LbstUpdbtedAt on bn instbnce of MockStore.
type StoreLbstUpdbtedAtFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []bpi.RepoID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 mbp[bpi.RepoID]time.Time
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreLbstUpdbtedAtFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreLbstUpdbtedAtFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreSoftDeleteStbleExportedUplobdsFunc describes the behbvior when the
// SoftDeleteStbleExportedUplobds method of the pbrent MockStore instbnce is
// invoked.
type StoreSoftDeleteStbleExportedUplobdsFunc struct {
	defbultHook func(context.Context, string) (int, int, error)
	hooks       []func(context.Context, string) (int, int, error)
	history     []StoreSoftDeleteStbleExportedUplobdsFuncCbll
	mutex       sync.Mutex
}

// SoftDeleteStbleExportedUplobds delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) SoftDeleteStbleExportedUplobds(v0 context.Context, v1 string) (int, int, error) {
	r0, r1, r2 := m.SoftDeleteStbleExportedUplobdsFunc.nextHook()(v0, v1)
	m.SoftDeleteStbleExportedUplobdsFunc.bppendCbll(StoreSoftDeleteStbleExportedUplobdsFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// SoftDeleteStbleExportedUplobds method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreSoftDeleteStbleExportedUplobdsFunc) SetDefbultHook(hook func(context.Context, string) (int, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SoftDeleteStbleExportedUplobds method of the pbrent MockStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *StoreSoftDeleteStbleExportedUplobdsFunc) PushHook(hook func(context.Context, string) (int, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreSoftDeleteStbleExportedUplobdsFunc) SetDefbultReturn(r0 int, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, string) (int, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreSoftDeleteStbleExportedUplobdsFunc) PushReturn(r0 int, r1 int, r2 error) {
	f.PushHook(func(context.Context, string) (int, int, error) {
		return r0, r1, r2
	})
}

func (f *StoreSoftDeleteStbleExportedUplobdsFunc) nextHook() func(context.Context, string) (int, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreSoftDeleteStbleExportedUplobdsFunc) bppendCbll(r0 StoreSoftDeleteStbleExportedUplobdsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreSoftDeleteStbleExportedUplobdsFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreSoftDeleteStbleExportedUplobdsFunc) History() []StoreSoftDeleteStbleExportedUplobdsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreSoftDeleteStbleExportedUplobdsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreSoftDeleteStbleExportedUplobdsFuncCbll is bn object thbt describes
// bn invocbtion of method SoftDeleteStbleExportedUplobds on bn instbnce of
// MockStore.
type StoreSoftDeleteStbleExportedUplobdsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreSoftDeleteStbleExportedUplobdsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreSoftDeleteStbleExportedUplobdsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// StoreSummbriesFunc describes the behbvior when the Summbries method of
// the pbrent MockStore instbnce is invoked.
type StoreSummbriesFunc struct {
	defbultHook func(context.Context) ([]shbred.Summbry, error)
	hooks       []func(context.Context) ([]shbred.Summbry, error)
	history     []StoreSummbriesFuncCbll
	mutex       sync.Mutex
}

// Summbries delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) Summbries(v0 context.Context) ([]shbred.Summbry, error) {
	r0, r1 := m.SummbriesFunc.nextHook()(v0)
	m.SummbriesFunc.bppendCbll(StoreSummbriesFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Summbries method of
// the pbrent MockStore instbnce is invoked bnd the hook queue is empty.
func (f *StoreSummbriesFunc) SetDefbultHook(hook func(context.Context) ([]shbred.Summbry, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Summbries method of the pbrent MockStore instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *StoreSummbriesFunc) PushHook(hook func(context.Context) ([]shbred.Summbry, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreSummbriesFunc) SetDefbultReturn(r0 []shbred.Summbry, r1 error) {
	f.SetDefbultHook(func(context.Context) ([]shbred.Summbry, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreSummbriesFunc) PushReturn(r0 []shbred.Summbry, r1 error) {
	f.PushHook(func(context.Context) ([]shbred.Summbry, error) {
		return r0, r1
	})
}

func (f *StoreSummbriesFunc) nextHook() func(context.Context) ([]shbred.Summbry, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreSummbriesFunc) bppendCbll(r0 StoreSummbriesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreSummbriesFuncCbll objects describing
// the invocbtions of this function.
func (f *StoreSummbriesFunc) History() []StoreSummbriesFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreSummbriesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreSummbriesFuncCbll is bn object thbt describes bn invocbtion of
// method Summbries on bn instbnce of MockStore.
type StoreSummbriesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.Summbry
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreSummbriesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreSummbriesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreVbcuumAbbndonedExportedUplobdsFunc describes the behbvior when the
// VbcuumAbbndonedExportedUplobds method of the pbrent MockStore instbnce is
// invoked.
type StoreVbcuumAbbndonedExportedUplobdsFunc struct {
	defbultHook func(context.Context, string, int) (int, error)
	hooks       []func(context.Context, string, int) (int, error)
	history     []StoreVbcuumAbbndonedExportedUplobdsFuncCbll
	mutex       sync.Mutex
}

// VbcuumAbbndonedExportedUplobds delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) VbcuumAbbndonedExportedUplobds(v0 context.Context, v1 string, v2 int) (int, error) {
	r0, r1 := m.VbcuumAbbndonedExportedUplobdsFunc.nextHook()(v0, v1, v2)
	m.VbcuumAbbndonedExportedUplobdsFunc.bppendCbll(StoreVbcuumAbbndonedExportedUplobdsFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// VbcuumAbbndonedExportedUplobds method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreVbcuumAbbndonedExportedUplobdsFunc) SetDefbultHook(hook func(context.Context, string, int) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// VbcuumAbbndonedExportedUplobds method of the pbrent MockStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *StoreVbcuumAbbndonedExportedUplobdsFunc) PushHook(hook func(context.Context, string, int) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreVbcuumAbbndonedExportedUplobdsFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context, string, int) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreVbcuumAbbndonedExportedUplobdsFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, string, int) (int, error) {
		return r0, r1
	})
}

func (f *StoreVbcuumAbbndonedExportedUplobdsFunc) nextHook() func(context.Context, string, int) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreVbcuumAbbndonedExportedUplobdsFunc) bppendCbll(r0 StoreVbcuumAbbndonedExportedUplobdsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreVbcuumAbbndonedExportedUplobdsFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreVbcuumAbbndonedExportedUplobdsFunc) History() []StoreVbcuumAbbndonedExportedUplobdsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreVbcuumAbbndonedExportedUplobdsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreVbcuumAbbndonedExportedUplobdsFuncCbll is bn object thbt describes
// bn invocbtion of method VbcuumAbbndonedExportedUplobds on bn instbnce of
// MockStore.
type StoreVbcuumAbbndonedExportedUplobdsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreVbcuumAbbndonedExportedUplobdsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreVbcuumAbbndonedExportedUplobdsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreVbcuumDeletedExportedUplobdsFunc describes the behbvior when the
// VbcuumDeletedExportedUplobds method of the pbrent MockStore instbnce is
// invoked.
type StoreVbcuumDeletedExportedUplobdsFunc struct {
	defbultHook func(context.Context, string) (int, error)
	hooks       []func(context.Context, string) (int, error)
	history     []StoreVbcuumDeletedExportedUplobdsFuncCbll
	mutex       sync.Mutex
}

// VbcuumDeletedExportedUplobds delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) VbcuumDeletedExportedUplobds(v0 context.Context, v1 string) (int, error) {
	r0, r1 := m.VbcuumDeletedExportedUplobdsFunc.nextHook()(v0, v1)
	m.VbcuumDeletedExportedUplobdsFunc.bppendCbll(StoreVbcuumDeletedExportedUplobdsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// VbcuumDeletedExportedUplobds method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreVbcuumDeletedExportedUplobdsFunc) SetDefbultHook(hook func(context.Context, string) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// VbcuumDeletedExportedUplobds method of the pbrent MockStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *StoreVbcuumDeletedExportedUplobdsFunc) PushHook(hook func(context.Context, string) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreVbcuumDeletedExportedUplobdsFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context, string) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreVbcuumDeletedExportedUplobdsFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, string) (int, error) {
		return r0, r1
	})
}

func (f *StoreVbcuumDeletedExportedUplobdsFunc) nextHook() func(context.Context, string) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreVbcuumDeletedExportedUplobdsFunc) bppendCbll(r0 StoreVbcuumDeletedExportedUplobdsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreVbcuumDeletedExportedUplobdsFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreVbcuumDeletedExportedUplobdsFunc) History() []StoreVbcuumDeletedExportedUplobdsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreVbcuumDeletedExportedUplobdsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreVbcuumDeletedExportedUplobdsFuncCbll is bn object thbt describes bn
// invocbtion of method VbcuumDeletedExportedUplobds on bn instbnce of
// MockStore.
type StoreVbcuumDeletedExportedUplobdsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreVbcuumDeletedExportedUplobdsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreVbcuumDeletedExportedUplobdsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreVbcuumStbleGrbphsFunc describes the behbvior when the
// VbcuumStbleGrbphs method of the pbrent MockStore instbnce is invoked.
type StoreVbcuumStbleGrbphsFunc struct {
	defbultHook func(context.Context, string, int) (int, error)
	hooks       []func(context.Context, string, int) (int, error)
	history     []StoreVbcuumStbleGrbphsFuncCbll
	mutex       sync.Mutex
}

// VbcuumStbleGrbphs delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) VbcuumStbleGrbphs(v0 context.Context, v1 string, v2 int) (int, error) {
	r0, r1 := m.VbcuumStbleGrbphsFunc.nextHook()(v0, v1, v2)
	m.VbcuumStbleGrbphsFunc.bppendCbll(StoreVbcuumStbleGrbphsFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the VbcuumStbleGrbphs
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreVbcuumStbleGrbphsFunc) SetDefbultHook(hook func(context.Context, string, int) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// VbcuumStbleGrbphs method of the pbrent MockStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *StoreVbcuumStbleGrbphsFunc) PushHook(hook func(context.Context, string, int) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreVbcuumStbleGrbphsFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context, string, int) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreVbcuumStbleGrbphsFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, string, int) (int, error) {
		return r0, r1
	})
}

func (f *StoreVbcuumStbleGrbphsFunc) nextHook() func(context.Context, string, int) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreVbcuumStbleGrbphsFunc) bppendCbll(r0 StoreVbcuumStbleGrbphsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreVbcuumStbleGrbphsFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreVbcuumStbleGrbphsFunc) History() []StoreVbcuumStbleGrbphsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreVbcuumStbleGrbphsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreVbcuumStbleGrbphsFuncCbll is bn object thbt describes bn invocbtion
// of method VbcuumStbleGrbphs on bn instbnce of MockStore.
type StoreVbcuumStbleGrbphsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreVbcuumStbleGrbphsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreVbcuumStbleGrbphsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreVbcuumStbleProcessedPbthsFunc describes the behbvior when the
// VbcuumStbleProcessedPbths method of the pbrent MockStore instbnce is
// invoked.
type StoreVbcuumStbleProcessedPbthsFunc struct {
	defbultHook func(context.Context, string, int) (int, error)
	hooks       []func(context.Context, string, int) (int, error)
	history     []StoreVbcuumStbleProcessedPbthsFuncCbll
	mutex       sync.Mutex
}

// VbcuumStbleProcessedPbths delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) VbcuumStbleProcessedPbths(v0 context.Context, v1 string, v2 int) (int, error) {
	r0, r1 := m.VbcuumStbleProcessedPbthsFunc.nextHook()(v0, v1, v2)
	m.VbcuumStbleProcessedPbthsFunc.bppendCbll(StoreVbcuumStbleProcessedPbthsFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// VbcuumStbleProcessedPbths method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreVbcuumStbleProcessedPbthsFunc) SetDefbultHook(hook func(context.Context, string, int) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// VbcuumStbleProcessedPbths method of the pbrent MockStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *StoreVbcuumStbleProcessedPbthsFunc) PushHook(hook func(context.Context, string, int) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreVbcuumStbleProcessedPbthsFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context, string, int) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreVbcuumStbleProcessedPbthsFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, string, int) (int, error) {
		return r0, r1
	})
}

func (f *StoreVbcuumStbleProcessedPbthsFunc) nextHook() func(context.Context, string, int) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreVbcuumStbleProcessedPbthsFunc) bppendCbll(r0 StoreVbcuumStbleProcessedPbthsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreVbcuumStbleProcessedPbthsFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreVbcuumStbleProcessedPbthsFunc) History() []StoreVbcuumStbleProcessedPbthsFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreVbcuumStbleProcessedPbthsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreVbcuumStbleProcessedPbthsFuncCbll is bn object thbt describes bn
// invocbtion of method VbcuumStbleProcessedPbths on bn instbnce of
// MockStore.
type StoreVbcuumStbleProcessedPbthsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreVbcuumStbleProcessedPbthsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreVbcuumStbleProcessedPbthsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreVbcuumStbleProcessedReferencesFunc describes the behbvior when the
// VbcuumStbleProcessedReferences method of the pbrent MockStore instbnce is
// invoked.
type StoreVbcuumStbleProcessedReferencesFunc struct {
	defbultHook func(context.Context, string, int) (int, error)
	hooks       []func(context.Context, string, int) (int, error)
	history     []StoreVbcuumStbleProcessedReferencesFuncCbll
	mutex       sync.Mutex
}

// VbcuumStbleProcessedReferences delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) VbcuumStbleProcessedReferences(v0 context.Context, v1 string, v2 int) (int, error) {
	r0, r1 := m.VbcuumStbleProcessedReferencesFunc.nextHook()(v0, v1, v2)
	m.VbcuumStbleProcessedReferencesFunc.bppendCbll(StoreVbcuumStbleProcessedReferencesFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// VbcuumStbleProcessedReferences method of the pbrent MockStore instbnce is
// invoked bnd the hook queue is empty.
func (f *StoreVbcuumStbleProcessedReferencesFunc) SetDefbultHook(hook func(context.Context, string, int) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// VbcuumStbleProcessedReferences method of the pbrent MockStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *StoreVbcuumStbleProcessedReferencesFunc) PushHook(hook func(context.Context, string, int) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreVbcuumStbleProcessedReferencesFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context, string, int) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreVbcuumStbleProcessedReferencesFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, string, int) (int, error) {
		return r0, r1
	})
}

func (f *StoreVbcuumStbleProcessedReferencesFunc) nextHook() func(context.Context, string, int) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreVbcuumStbleProcessedReferencesFunc) bppendCbll(r0 StoreVbcuumStbleProcessedReferencesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreVbcuumStbleProcessedReferencesFuncCbll
// objects describing the invocbtions of this function.
func (f *StoreVbcuumStbleProcessedReferencesFunc) History() []StoreVbcuumStbleProcessedReferencesFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreVbcuumStbleProcessedReferencesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreVbcuumStbleProcessedReferencesFuncCbll is bn object thbt describes
// bn invocbtion of method VbcuumStbleProcessedReferences on bn instbnce of
// MockStore.
type StoreVbcuumStbleProcessedReferencesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreVbcuumStbleProcessedReferencesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreVbcuumStbleProcessedReferencesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// StoreVbcuumStbleRbnksFunc describes the behbvior when the
// VbcuumStbleRbnks method of the pbrent MockStore instbnce is invoked.
type StoreVbcuumStbleRbnksFunc struct {
	defbultHook func(context.Context, string) (int, int, error)
	hooks       []func(context.Context, string) (int, int, error)
	history     []StoreVbcuumStbleRbnksFuncCbll
	mutex       sync.Mutex
}

// VbcuumStbleRbnks delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) VbcuumStbleRbnks(v0 context.Context, v1 string) (int, int, error) {
	r0, r1, r2 := m.VbcuumStbleRbnksFunc.nextHook()(v0, v1)
	m.VbcuumStbleRbnksFunc.bppendCbll(StoreVbcuumStbleRbnksFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the VbcuumStbleRbnks
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreVbcuumStbleRbnksFunc) SetDefbultHook(hook func(context.Context, string) (int, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// VbcuumStbleRbnks method of the pbrent MockStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreVbcuumStbleRbnksFunc) PushHook(hook func(context.Context, string) (int, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreVbcuumStbleRbnksFunc) SetDefbultReturn(r0 int, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, string) (int, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreVbcuumStbleRbnksFunc) PushReturn(r0 int, r1 int, r2 error) {
	f.PushHook(func(context.Context, string) (int, int, error) {
		return r0, r1, r2
	})
}

func (f *StoreVbcuumStbleRbnksFunc) nextHook() func(context.Context, string) (int, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreVbcuumStbleRbnksFunc) bppendCbll(r0 StoreVbcuumStbleRbnksFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreVbcuumStbleRbnksFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreVbcuumStbleRbnksFunc) History() []StoreVbcuumStbleRbnksFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreVbcuumStbleRbnksFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreVbcuumStbleRbnksFuncCbll is bn object thbt describes bn invocbtion
// of method VbcuumStbleRbnks on bn instbnce of MockStore.
type StoreVbcuumStbleRbnksFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreVbcuumStbleRbnksFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreVbcuumStbleRbnksFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// StoreWithTrbnsbctionFunc describes the behbvior when the WithTrbnsbction
// method of the pbrent MockStore instbnce is invoked.
type StoreWithTrbnsbctionFunc struct {
	defbultHook func(context.Context, func(tx store.Store) error) error
	hooks       []func(context.Context, func(tx store.Store) error) error
	history     []StoreWithTrbnsbctionFuncCbll
	mutex       sync.Mutex
}

// WithTrbnsbction delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockStore) WithTrbnsbction(v0 context.Context, v1 func(tx store.Store) error) error {
	r0 := m.WithTrbnsbctionFunc.nextHook()(v0, v1)
	m.WithTrbnsbctionFunc.bppendCbll(StoreWithTrbnsbctionFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the WithTrbnsbction
// method of the pbrent MockStore instbnce is invoked bnd the hook queue is
// empty.
func (f *StoreWithTrbnsbctionFunc) SetDefbultHook(hook func(context.Context, func(tx store.Store) error) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// WithTrbnsbction method of the pbrent MockStore instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *StoreWithTrbnsbctionFunc) PushHook(hook func(context.Context, func(tx store.Store) error) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *StoreWithTrbnsbctionFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, func(tx store.Store) error) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *StoreWithTrbnsbctionFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, func(tx store.Store) error) error {
		return r0
	})
}

func (f *StoreWithTrbnsbctionFunc) nextHook() func(context.Context, func(tx store.Store) error) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *StoreWithTrbnsbctionFunc) bppendCbll(r0 StoreWithTrbnsbctionFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of StoreWithTrbnsbctionFuncCbll objects
// describing the invocbtions of this function.
func (f *StoreWithTrbnsbctionFunc) History() []StoreWithTrbnsbctionFuncCbll {
	f.mutex.Lock()
	history := mbke([]StoreWithTrbnsbctionFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// StoreWithTrbnsbctionFuncCbll is bn object thbt describes bn invocbtion of
// method WithTrbnsbction on bn instbnce of MockStore.
type StoreWithTrbnsbctionFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 func(tx store.Store) error
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c StoreWithTrbnsbctionFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c StoreWithTrbnsbctionFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockSiteConfigQuerier is b mock implementbtion of the SiteConfigQuerier
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes) used for unit
// testing.
type MockSiteConfigQuerier struct {
	// SiteConfigFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method SiteConfig.
	SiteConfigFunc *SiteConfigQuerierSiteConfigFunc
}

// NewMockSiteConfigQuerier crebtes b new mock of the SiteConfigQuerier
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockSiteConfigQuerier() *MockSiteConfigQuerier {
	return &MockSiteConfigQuerier{
		SiteConfigFunc: &SiteConfigQuerierSiteConfigFunc{
			defbultHook: func() (r0 schemb.SiteConfigurbtion) {
				return
			},
		},
	}
}

// NewStrictMockSiteConfigQuerier crebtes b new mock of the
// SiteConfigQuerier interfbce. All methods pbnic on invocbtion, unless
// overwritten.
func NewStrictMockSiteConfigQuerier() *MockSiteConfigQuerier {
	return &MockSiteConfigQuerier{
		SiteConfigFunc: &SiteConfigQuerierSiteConfigFunc{
			defbultHook: func() schemb.SiteConfigurbtion {
				pbnic("unexpected invocbtion of MockSiteConfigQuerier.SiteConfig")
			},
		},
	}
}

// NewMockSiteConfigQuerierFrom crebtes b new mock of the
// MockSiteConfigQuerier interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockSiteConfigQuerierFrom(i conftypes.SiteConfigQuerier) *MockSiteConfigQuerier {
	return &MockSiteConfigQuerier{
		SiteConfigFunc: &SiteConfigQuerierSiteConfigFunc{
			defbultHook: i.SiteConfig,
		},
	}
}

// SiteConfigQuerierSiteConfigFunc describes the behbvior when the
// SiteConfig method of the pbrent MockSiteConfigQuerier instbnce is
// invoked.
type SiteConfigQuerierSiteConfigFunc struct {
	defbultHook func() schemb.SiteConfigurbtion
	hooks       []func() schemb.SiteConfigurbtion
	history     []SiteConfigQuerierSiteConfigFuncCbll
	mutex       sync.Mutex
}

// SiteConfig delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockSiteConfigQuerier) SiteConfig() schemb.SiteConfigurbtion {
	r0 := m.SiteConfigFunc.nextHook()()
	m.SiteConfigFunc.bppendCbll(SiteConfigQuerierSiteConfigFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the SiteConfig method of
// the pbrent MockSiteConfigQuerier instbnce is invoked bnd the hook queue
// is empty.
func (f *SiteConfigQuerierSiteConfigFunc) SetDefbultHook(hook func() schemb.SiteConfigurbtion) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SiteConfig method of the pbrent MockSiteConfigQuerier instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *SiteConfigQuerierSiteConfigFunc) PushHook(hook func() schemb.SiteConfigurbtion) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *SiteConfigQuerierSiteConfigFunc) SetDefbultReturn(r0 schemb.SiteConfigurbtion) {
	f.SetDefbultHook(func() schemb.SiteConfigurbtion {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *SiteConfigQuerierSiteConfigFunc) PushReturn(r0 schemb.SiteConfigurbtion) {
	f.PushHook(func() schemb.SiteConfigurbtion {
		return r0
	})
}

func (f *SiteConfigQuerierSiteConfigFunc) nextHook() func() schemb.SiteConfigurbtion {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *SiteConfigQuerierSiteConfigFunc) bppendCbll(r0 SiteConfigQuerierSiteConfigFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of SiteConfigQuerierSiteConfigFuncCbll objects
// describing the invocbtions of this function.
func (f *SiteConfigQuerierSiteConfigFunc) History() []SiteConfigQuerierSiteConfigFuncCbll {
	f.mutex.Lock()
	history := mbke([]SiteConfigQuerierSiteConfigFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// SiteConfigQuerierSiteConfigFuncCbll is bn object thbt describes bn
// invocbtion of method SiteConfig on bn instbnce of MockSiteConfigQuerier.
type SiteConfigQuerierSiteConfigFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 schemb.SiteConfigurbtion
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c SiteConfigQuerierSiteConfigFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c SiteConfigQuerierSiteConfigFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
