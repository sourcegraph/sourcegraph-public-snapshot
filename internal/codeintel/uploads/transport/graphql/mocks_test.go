// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge grbphql

import (
	"context"
	"sync"
	"time"

	shbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
)

// MockUplobdsService is b mock implementbtion of the UplobdsService
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/trbnsport/grbphql)
// used for unit testing.
type MockUplobdsService struct {
	// DeleteIndexByIDFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method DeleteIndexByID.
	DeleteIndexByIDFunc *UplobdsServiceDeleteIndexByIDFunc
	// DeleteIndexesFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method DeleteIndexes.
	DeleteIndexesFunc *UplobdsServiceDeleteIndexesFunc
	// DeleteUplobdByIDFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method DeleteUplobdByID.
	DeleteUplobdByIDFunc *UplobdsServiceDeleteUplobdByIDFunc
	// DeleteUplobdsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method DeleteUplobds.
	DeleteUplobdsFunc *UplobdsServiceDeleteUplobdsFunc
	// GetAuditLogsForUplobdFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetAuditLogsForUplobd.
	GetAuditLogsForUplobdFunc *UplobdsServiceGetAuditLogsForUplobdFunc
	// GetCommitGrbphMetbdbtbFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetCommitGrbphMetbdbtb.
	GetCommitGrbphMetbdbtbFunc *UplobdsServiceGetCommitGrbphMetbdbtbFunc
	// GetIndexByIDFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetIndexByID.
	GetIndexByIDFunc *UplobdsServiceGetIndexByIDFunc
	// GetIndexersFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetIndexers.
	GetIndexersFunc *UplobdsServiceGetIndexersFunc
	// GetIndexesFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetIndexes.
	GetIndexesFunc *UplobdsServiceGetIndexesFunc
	// GetIndexesByIDsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetIndexesByIDs.
	GetIndexesByIDsFunc *UplobdsServiceGetIndexesByIDsFunc
	// GetLbstUplobdRetentionScbnForRepositoryFunc is bn instbnce of b mock
	// function object controlling the behbvior of the method
	// GetLbstUplobdRetentionScbnForRepository.
	GetLbstUplobdRetentionScbnForRepositoryFunc *UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFunc
	// GetRecentIndexesSummbryFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetRecentIndexesSummbry.
	GetRecentIndexesSummbryFunc *UplobdsServiceGetRecentIndexesSummbryFunc
	// GetRecentUplobdsSummbryFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetRecentUplobdsSummbry.
	GetRecentUplobdsSummbryFunc *UplobdsServiceGetRecentUplobdsSummbryFunc
	// GetUplobdByIDFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetUplobdByID.
	GetUplobdByIDFunc *UplobdsServiceGetUplobdByIDFunc
	// GetUplobdsFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetUplobds.
	GetUplobdsFunc *UplobdsServiceGetUplobdsFunc
	// GetUplobdsByIDsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetUplobdsByIDs.
	GetUplobdsByIDsFunc *UplobdsServiceGetUplobdsByIDsFunc
	// NumRepositoriesWithCodeIntelligenceFunc is bn instbnce of b mock
	// function object controlling the behbvior of the method
	// NumRepositoriesWithCodeIntelligence.
	NumRepositoriesWithCodeIntelligenceFunc *UplobdsServiceNumRepositoriesWithCodeIntelligenceFunc
	// ReindexIndexByIDFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ReindexIndexByID.
	ReindexIndexByIDFunc *UplobdsServiceReindexIndexByIDFunc
	// ReindexIndexesFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ReindexIndexes.
	ReindexIndexesFunc *UplobdsServiceReindexIndexesFunc
	// ReindexUplobdByIDFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ReindexUplobdByID.
	ReindexUplobdByIDFunc *UplobdsServiceReindexUplobdByIDFunc
	// ReindexUplobdsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ReindexUplobds.
	ReindexUplobdsFunc *UplobdsServiceReindexUplobdsFunc
	// RepositoryIDsWithErrorsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method RepositoryIDsWithErrors.
	RepositoryIDsWithErrorsFunc *UplobdsServiceRepositoryIDsWithErrorsFunc
}

// NewMockUplobdsService crebtes b new mock of the UplobdsService interfbce.
// All methods return zero vblues for bll results, unless overwritten.
func NewMockUplobdsService() *MockUplobdsService {
	return &MockUplobdsService{
		DeleteIndexByIDFunc: &UplobdsServiceDeleteIndexByIDFunc{
			defbultHook: func(context.Context, int) (r0 bool, r1 error) {
				return
			},
		},
		DeleteIndexesFunc: &UplobdsServiceDeleteIndexesFunc{
			defbultHook: func(context.Context, shbred.DeleteIndexesOptions) (r0 error) {
				return
			},
		},
		DeleteUplobdByIDFunc: &UplobdsServiceDeleteUplobdByIDFunc{
			defbultHook: func(context.Context, int) (r0 bool, r1 error) {
				return
			},
		},
		DeleteUplobdsFunc: &UplobdsServiceDeleteUplobdsFunc{
			defbultHook: func(context.Context, shbred.DeleteUplobdsOptions) (r0 error) {
				return
			},
		},
		GetAuditLogsForUplobdFunc: &UplobdsServiceGetAuditLogsForUplobdFunc{
			defbultHook: func(context.Context, int) (r0 []shbred.UplobdLog, r1 error) {
				return
			},
		},
		GetCommitGrbphMetbdbtbFunc: &UplobdsServiceGetCommitGrbphMetbdbtbFunc{
			defbultHook: func(context.Context, int) (r0 bool, r1 *time.Time, r2 error) {
				return
			},
		},
		GetIndexByIDFunc: &UplobdsServiceGetIndexByIDFunc{
			defbultHook: func(context.Context, int) (r0 shbred.Index, r1 bool, r2 error) {
				return
			},
		},
		GetIndexersFunc: &UplobdsServiceGetIndexersFunc{
			defbultHook: func(context.Context, shbred.GetIndexersOptions) (r0 []string, r1 error) {
				return
			},
		},
		GetIndexesFunc: &UplobdsServiceGetIndexesFunc{
			defbultHook: func(context.Context, shbred.GetIndexesOptions) (r0 []shbred.Index, r1 int, r2 error) {
				return
			},
		},
		GetIndexesByIDsFunc: &UplobdsServiceGetIndexesByIDsFunc{
			defbultHook: func(context.Context, ...int) (r0 []shbred.Index, r1 error) {
				return
			},
		},
		GetLbstUplobdRetentionScbnForRepositoryFunc: &UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFunc{
			defbultHook: func(context.Context, int) (r0 *time.Time, r1 error) {
				return
			},
		},
		GetRecentIndexesSummbryFunc: &UplobdsServiceGetRecentIndexesSummbryFunc{
			defbultHook: func(context.Context, int) (r0 []shbred.IndexesWithRepositoryNbmespbce, r1 error) {
				return
			},
		},
		GetRecentUplobdsSummbryFunc: &UplobdsServiceGetRecentUplobdsSummbryFunc{
			defbultHook: func(context.Context, int) (r0 []shbred.UplobdsWithRepositoryNbmespbce, r1 error) {
				return
			},
		},
		GetUplobdByIDFunc: &UplobdsServiceGetUplobdByIDFunc{
			defbultHook: func(context.Context, int) (r0 shbred.Uplobd, r1 bool, r2 error) {
				return
			},
		},
		GetUplobdsFunc: &UplobdsServiceGetUplobdsFunc{
			defbultHook: func(context.Context, shbred.GetUplobdsOptions) (r0 []shbred.Uplobd, r1 int, r2 error) {
				return
			},
		},
		GetUplobdsByIDsFunc: &UplobdsServiceGetUplobdsByIDsFunc{
			defbultHook: func(context.Context, ...int) (r0 []shbred.Uplobd, r1 error) {
				return
			},
		},
		NumRepositoriesWithCodeIntelligenceFunc: &UplobdsServiceNumRepositoriesWithCodeIntelligenceFunc{
			defbultHook: func(context.Context) (r0 int, r1 error) {
				return
			},
		},
		ReindexIndexByIDFunc: &UplobdsServiceReindexIndexByIDFunc{
			defbultHook: func(context.Context, int) (r0 error) {
				return
			},
		},
		ReindexIndexesFunc: &UplobdsServiceReindexIndexesFunc{
			defbultHook: func(context.Context, shbred.ReindexIndexesOptions) (r0 error) {
				return
			},
		},
		ReindexUplobdByIDFunc: &UplobdsServiceReindexUplobdByIDFunc{
			defbultHook: func(context.Context, int) (r0 error) {
				return
			},
		},
		ReindexUplobdsFunc: &UplobdsServiceReindexUplobdsFunc{
			defbultHook: func(context.Context, shbred.ReindexUplobdsOptions) (r0 error) {
				return
			},
		},
		RepositoryIDsWithErrorsFunc: &UplobdsServiceRepositoryIDsWithErrorsFunc{
			defbultHook: func(context.Context, int, int) (r0 []shbred.RepositoryWithCount, r1 int, r2 error) {
				return
			},
		},
	}
}

// NewStrictMockUplobdsService crebtes b new mock of the UplobdsService
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockUplobdsService() *MockUplobdsService {
	return &MockUplobdsService{
		DeleteIndexByIDFunc: &UplobdsServiceDeleteIndexByIDFunc{
			defbultHook: func(context.Context, int) (bool, error) {
				pbnic("unexpected invocbtion of MockUplobdsService.DeleteIndexByID")
			},
		},
		DeleteIndexesFunc: &UplobdsServiceDeleteIndexesFunc{
			defbultHook: func(context.Context, shbred.DeleteIndexesOptions) error {
				pbnic("unexpected invocbtion of MockUplobdsService.DeleteIndexes")
			},
		},
		DeleteUplobdByIDFunc: &UplobdsServiceDeleteUplobdByIDFunc{
			defbultHook: func(context.Context, int) (bool, error) {
				pbnic("unexpected invocbtion of MockUplobdsService.DeleteUplobdByID")
			},
		},
		DeleteUplobdsFunc: &UplobdsServiceDeleteUplobdsFunc{
			defbultHook: func(context.Context, shbred.DeleteUplobdsOptions) error {
				pbnic("unexpected invocbtion of MockUplobdsService.DeleteUplobds")
			},
		},
		GetAuditLogsForUplobdFunc: &UplobdsServiceGetAuditLogsForUplobdFunc{
			defbultHook: func(context.Context, int) ([]shbred.UplobdLog, error) {
				pbnic("unexpected invocbtion of MockUplobdsService.GetAuditLogsForUplobd")
			},
		},
		GetCommitGrbphMetbdbtbFunc: &UplobdsServiceGetCommitGrbphMetbdbtbFunc{
			defbultHook: func(context.Context, int) (bool, *time.Time, error) {
				pbnic("unexpected invocbtion of MockUplobdsService.GetCommitGrbphMetbdbtb")
			},
		},
		GetIndexByIDFunc: &UplobdsServiceGetIndexByIDFunc{
			defbultHook: func(context.Context, int) (shbred.Index, bool, error) {
				pbnic("unexpected invocbtion of MockUplobdsService.GetIndexByID")
			},
		},
		GetIndexersFunc: &UplobdsServiceGetIndexersFunc{
			defbultHook: func(context.Context, shbred.GetIndexersOptions) ([]string, error) {
				pbnic("unexpected invocbtion of MockUplobdsService.GetIndexers")
			},
		},
		GetIndexesFunc: &UplobdsServiceGetIndexesFunc{
			defbultHook: func(context.Context, shbred.GetIndexesOptions) ([]shbred.Index, int, error) {
				pbnic("unexpected invocbtion of MockUplobdsService.GetIndexes")
			},
		},
		GetIndexesByIDsFunc: &UplobdsServiceGetIndexesByIDsFunc{
			defbultHook: func(context.Context, ...int) ([]shbred.Index, error) {
				pbnic("unexpected invocbtion of MockUplobdsService.GetIndexesByIDs")
			},
		},
		GetLbstUplobdRetentionScbnForRepositoryFunc: &UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFunc{
			defbultHook: func(context.Context, int) (*time.Time, error) {
				pbnic("unexpected invocbtion of MockUplobdsService.GetLbstUplobdRetentionScbnForRepository")
			},
		},
		GetRecentIndexesSummbryFunc: &UplobdsServiceGetRecentIndexesSummbryFunc{
			defbultHook: func(context.Context, int) ([]shbred.IndexesWithRepositoryNbmespbce, error) {
				pbnic("unexpected invocbtion of MockUplobdsService.GetRecentIndexesSummbry")
			},
		},
		GetRecentUplobdsSummbryFunc: &UplobdsServiceGetRecentUplobdsSummbryFunc{
			defbultHook: func(context.Context, int) ([]shbred.UplobdsWithRepositoryNbmespbce, error) {
				pbnic("unexpected invocbtion of MockUplobdsService.GetRecentUplobdsSummbry")
			},
		},
		GetUplobdByIDFunc: &UplobdsServiceGetUplobdByIDFunc{
			defbultHook: func(context.Context, int) (shbred.Uplobd, bool, error) {
				pbnic("unexpected invocbtion of MockUplobdsService.GetUplobdByID")
			},
		},
		GetUplobdsFunc: &UplobdsServiceGetUplobdsFunc{
			defbultHook: func(context.Context, shbred.GetUplobdsOptions) ([]shbred.Uplobd, int, error) {
				pbnic("unexpected invocbtion of MockUplobdsService.GetUplobds")
			},
		},
		GetUplobdsByIDsFunc: &UplobdsServiceGetUplobdsByIDsFunc{
			defbultHook: func(context.Context, ...int) ([]shbred.Uplobd, error) {
				pbnic("unexpected invocbtion of MockUplobdsService.GetUplobdsByIDs")
			},
		},
		NumRepositoriesWithCodeIntelligenceFunc: &UplobdsServiceNumRepositoriesWithCodeIntelligenceFunc{
			defbultHook: func(context.Context) (int, error) {
				pbnic("unexpected invocbtion of MockUplobdsService.NumRepositoriesWithCodeIntelligence")
			},
		},
		ReindexIndexByIDFunc: &UplobdsServiceReindexIndexByIDFunc{
			defbultHook: func(context.Context, int) error {
				pbnic("unexpected invocbtion of MockUplobdsService.ReindexIndexByID")
			},
		},
		ReindexIndexesFunc: &UplobdsServiceReindexIndexesFunc{
			defbultHook: func(context.Context, shbred.ReindexIndexesOptions) error {
				pbnic("unexpected invocbtion of MockUplobdsService.ReindexIndexes")
			},
		},
		ReindexUplobdByIDFunc: &UplobdsServiceReindexUplobdByIDFunc{
			defbultHook: func(context.Context, int) error {
				pbnic("unexpected invocbtion of MockUplobdsService.ReindexUplobdByID")
			},
		},
		ReindexUplobdsFunc: &UplobdsServiceReindexUplobdsFunc{
			defbultHook: func(context.Context, shbred.ReindexUplobdsOptions) error {
				pbnic("unexpected invocbtion of MockUplobdsService.ReindexUplobds")
			},
		},
		RepositoryIDsWithErrorsFunc: &UplobdsServiceRepositoryIDsWithErrorsFunc{
			defbultHook: func(context.Context, int, int) ([]shbred.RepositoryWithCount, int, error) {
				pbnic("unexpected invocbtion of MockUplobdsService.RepositoryIDsWithErrors")
			},
		},
	}
}

// NewMockUplobdsServiceFrom crebtes b new mock of the MockUplobdsService
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockUplobdsServiceFrom(i UplobdsService) *MockUplobdsService {
	return &MockUplobdsService{
		DeleteIndexByIDFunc: &UplobdsServiceDeleteIndexByIDFunc{
			defbultHook: i.DeleteIndexByID,
		},
		DeleteIndexesFunc: &UplobdsServiceDeleteIndexesFunc{
			defbultHook: i.DeleteIndexes,
		},
		DeleteUplobdByIDFunc: &UplobdsServiceDeleteUplobdByIDFunc{
			defbultHook: i.DeleteUplobdByID,
		},
		DeleteUplobdsFunc: &UplobdsServiceDeleteUplobdsFunc{
			defbultHook: i.DeleteUplobds,
		},
		GetAuditLogsForUplobdFunc: &UplobdsServiceGetAuditLogsForUplobdFunc{
			defbultHook: i.GetAuditLogsForUplobd,
		},
		GetCommitGrbphMetbdbtbFunc: &UplobdsServiceGetCommitGrbphMetbdbtbFunc{
			defbultHook: i.GetCommitGrbphMetbdbtb,
		},
		GetIndexByIDFunc: &UplobdsServiceGetIndexByIDFunc{
			defbultHook: i.GetIndexByID,
		},
		GetIndexersFunc: &UplobdsServiceGetIndexersFunc{
			defbultHook: i.GetIndexers,
		},
		GetIndexesFunc: &UplobdsServiceGetIndexesFunc{
			defbultHook: i.GetIndexes,
		},
		GetIndexesByIDsFunc: &UplobdsServiceGetIndexesByIDsFunc{
			defbultHook: i.GetIndexesByIDs,
		},
		GetLbstUplobdRetentionScbnForRepositoryFunc: &UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFunc{
			defbultHook: i.GetLbstUplobdRetentionScbnForRepository,
		},
		GetRecentIndexesSummbryFunc: &UplobdsServiceGetRecentIndexesSummbryFunc{
			defbultHook: i.GetRecentIndexesSummbry,
		},
		GetRecentUplobdsSummbryFunc: &UplobdsServiceGetRecentUplobdsSummbryFunc{
			defbultHook: i.GetRecentUplobdsSummbry,
		},
		GetUplobdByIDFunc: &UplobdsServiceGetUplobdByIDFunc{
			defbultHook: i.GetUplobdByID,
		},
		GetUplobdsFunc: &UplobdsServiceGetUplobdsFunc{
			defbultHook: i.GetUplobds,
		},
		GetUplobdsByIDsFunc: &UplobdsServiceGetUplobdsByIDsFunc{
			defbultHook: i.GetUplobdsByIDs,
		},
		NumRepositoriesWithCodeIntelligenceFunc: &UplobdsServiceNumRepositoriesWithCodeIntelligenceFunc{
			defbultHook: i.NumRepositoriesWithCodeIntelligence,
		},
		ReindexIndexByIDFunc: &UplobdsServiceReindexIndexByIDFunc{
			defbultHook: i.ReindexIndexByID,
		},
		ReindexIndexesFunc: &UplobdsServiceReindexIndexesFunc{
			defbultHook: i.ReindexIndexes,
		},
		ReindexUplobdByIDFunc: &UplobdsServiceReindexUplobdByIDFunc{
			defbultHook: i.ReindexUplobdByID,
		},
		ReindexUplobdsFunc: &UplobdsServiceReindexUplobdsFunc{
			defbultHook: i.ReindexUplobds,
		},
		RepositoryIDsWithErrorsFunc: &UplobdsServiceRepositoryIDsWithErrorsFunc{
			defbultHook: i.RepositoryIDsWithErrors,
		},
	}
}

// UplobdsServiceDeleteIndexByIDFunc describes the behbvior when the
// DeleteIndexByID method of the pbrent MockUplobdsService instbnce is
// invoked.
type UplobdsServiceDeleteIndexByIDFunc struct {
	defbultHook func(context.Context, int) (bool, error)
	hooks       []func(context.Context, int) (bool, error)
	history     []UplobdsServiceDeleteIndexByIDFuncCbll
	mutex       sync.Mutex
}

// DeleteIndexByID delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) DeleteIndexByID(v0 context.Context, v1 int) (bool, error) {
	r0, r1 := m.DeleteIndexByIDFunc.nextHook()(v0, v1)
	m.DeleteIndexByIDFunc.bppendCbll(UplobdsServiceDeleteIndexByIDFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the DeleteIndexByID
// method of the pbrent MockUplobdsService instbnce is invoked bnd the hook
// queue is empty.
func (f *UplobdsServiceDeleteIndexByIDFunc) SetDefbultHook(hook func(context.Context, int) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DeleteIndexByID method of the pbrent MockUplobdsService instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *UplobdsServiceDeleteIndexByIDFunc) PushHook(hook func(context.Context, int) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceDeleteIndexByIDFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, int) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceDeleteIndexByIDFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, int) (bool, error) {
		return r0, r1
	})
}

func (f *UplobdsServiceDeleteIndexByIDFunc) nextHook() func(context.Context, int) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceDeleteIndexByIDFunc) bppendCbll(r0 UplobdsServiceDeleteIndexByIDFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdsServiceDeleteIndexByIDFuncCbll
// objects describing the invocbtions of this function.
func (f *UplobdsServiceDeleteIndexByIDFunc) History() []UplobdsServiceDeleteIndexByIDFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceDeleteIndexByIDFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceDeleteIndexByIDFuncCbll is bn object thbt describes bn
// invocbtion of method DeleteIndexByID on bn instbnce of
// MockUplobdsService.
type UplobdsServiceDeleteIndexByIDFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceDeleteIndexByIDFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceDeleteIndexByIDFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// UplobdsServiceDeleteIndexesFunc describes the behbvior when the
// DeleteIndexes method of the pbrent MockUplobdsService instbnce is
// invoked.
type UplobdsServiceDeleteIndexesFunc struct {
	defbultHook func(context.Context, shbred.DeleteIndexesOptions) error
	hooks       []func(context.Context, shbred.DeleteIndexesOptions) error
	history     []UplobdsServiceDeleteIndexesFuncCbll
	mutex       sync.Mutex
}

// DeleteIndexes delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) DeleteIndexes(v0 context.Context, v1 shbred.DeleteIndexesOptions) error {
	r0 := m.DeleteIndexesFunc.nextHook()(v0, v1)
	m.DeleteIndexesFunc.bppendCbll(UplobdsServiceDeleteIndexesFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the DeleteIndexes method
// of the pbrent MockUplobdsService instbnce is invoked bnd the hook queue
// is empty.
func (f *UplobdsServiceDeleteIndexesFunc) SetDefbultHook(hook func(context.Context, shbred.DeleteIndexesOptions) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DeleteIndexes method of the pbrent MockUplobdsService instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *UplobdsServiceDeleteIndexesFunc) PushHook(hook func(context.Context, shbred.DeleteIndexesOptions) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceDeleteIndexesFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, shbred.DeleteIndexesOptions) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceDeleteIndexesFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, shbred.DeleteIndexesOptions) error {
		return r0
	})
}

func (f *UplobdsServiceDeleteIndexesFunc) nextHook() func(context.Context, shbred.DeleteIndexesOptions) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceDeleteIndexesFunc) bppendCbll(r0 UplobdsServiceDeleteIndexesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdsServiceDeleteIndexesFuncCbll objects
// describing the invocbtions of this function.
func (f *UplobdsServiceDeleteIndexesFunc) History() []UplobdsServiceDeleteIndexesFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceDeleteIndexesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceDeleteIndexesFuncCbll is bn object thbt describes bn
// invocbtion of method DeleteIndexes on bn instbnce of MockUplobdsService.
type UplobdsServiceDeleteIndexesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 shbred.DeleteIndexesOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceDeleteIndexesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceDeleteIndexesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// UplobdsServiceDeleteUplobdByIDFunc describes the behbvior when the
// DeleteUplobdByID method of the pbrent MockUplobdsService instbnce is
// invoked.
type UplobdsServiceDeleteUplobdByIDFunc struct {
	defbultHook func(context.Context, int) (bool, error)
	hooks       []func(context.Context, int) (bool, error)
	history     []UplobdsServiceDeleteUplobdByIDFuncCbll
	mutex       sync.Mutex
}

// DeleteUplobdByID delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) DeleteUplobdByID(v0 context.Context, v1 int) (bool, error) {
	r0, r1 := m.DeleteUplobdByIDFunc.nextHook()(v0, v1)
	m.DeleteUplobdByIDFunc.bppendCbll(UplobdsServiceDeleteUplobdByIDFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the DeleteUplobdByID
// method of the pbrent MockUplobdsService instbnce is invoked bnd the hook
// queue is empty.
func (f *UplobdsServiceDeleteUplobdByIDFunc) SetDefbultHook(hook func(context.Context, int) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DeleteUplobdByID method of the pbrent MockUplobdsService instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *UplobdsServiceDeleteUplobdByIDFunc) PushHook(hook func(context.Context, int) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceDeleteUplobdByIDFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(context.Context, int) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceDeleteUplobdByIDFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(context.Context, int) (bool, error) {
		return r0, r1
	})
}

func (f *UplobdsServiceDeleteUplobdByIDFunc) nextHook() func(context.Context, int) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceDeleteUplobdByIDFunc) bppendCbll(r0 UplobdsServiceDeleteUplobdByIDFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdsServiceDeleteUplobdByIDFuncCbll
// objects describing the invocbtions of this function.
func (f *UplobdsServiceDeleteUplobdByIDFunc) History() []UplobdsServiceDeleteUplobdByIDFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceDeleteUplobdByIDFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceDeleteUplobdByIDFuncCbll is bn object thbt describes bn
// invocbtion of method DeleteUplobdByID on bn instbnce of
// MockUplobdsService.
type UplobdsServiceDeleteUplobdByIDFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceDeleteUplobdByIDFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceDeleteUplobdByIDFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// UplobdsServiceDeleteUplobdsFunc describes the behbvior when the
// DeleteUplobds method of the pbrent MockUplobdsService instbnce is
// invoked.
type UplobdsServiceDeleteUplobdsFunc struct {
	defbultHook func(context.Context, shbred.DeleteUplobdsOptions) error
	hooks       []func(context.Context, shbred.DeleteUplobdsOptions) error
	history     []UplobdsServiceDeleteUplobdsFuncCbll
	mutex       sync.Mutex
}

// DeleteUplobds delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) DeleteUplobds(v0 context.Context, v1 shbred.DeleteUplobdsOptions) error {
	r0 := m.DeleteUplobdsFunc.nextHook()(v0, v1)
	m.DeleteUplobdsFunc.bppendCbll(UplobdsServiceDeleteUplobdsFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the DeleteUplobds method
// of the pbrent MockUplobdsService instbnce is invoked bnd the hook queue
// is empty.
func (f *UplobdsServiceDeleteUplobdsFunc) SetDefbultHook(hook func(context.Context, shbred.DeleteUplobdsOptions) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DeleteUplobds method of the pbrent MockUplobdsService instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *UplobdsServiceDeleteUplobdsFunc) PushHook(hook func(context.Context, shbred.DeleteUplobdsOptions) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceDeleteUplobdsFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, shbred.DeleteUplobdsOptions) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceDeleteUplobdsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, shbred.DeleteUplobdsOptions) error {
		return r0
	})
}

func (f *UplobdsServiceDeleteUplobdsFunc) nextHook() func(context.Context, shbred.DeleteUplobdsOptions) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceDeleteUplobdsFunc) bppendCbll(r0 UplobdsServiceDeleteUplobdsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdsServiceDeleteUplobdsFuncCbll objects
// describing the invocbtions of this function.
func (f *UplobdsServiceDeleteUplobdsFunc) History() []UplobdsServiceDeleteUplobdsFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceDeleteUplobdsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceDeleteUplobdsFuncCbll is bn object thbt describes bn
// invocbtion of method DeleteUplobds on bn instbnce of MockUplobdsService.
type UplobdsServiceDeleteUplobdsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 shbred.DeleteUplobdsOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceDeleteUplobdsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceDeleteUplobdsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// UplobdsServiceGetAuditLogsForUplobdFunc describes the behbvior when the
// GetAuditLogsForUplobd method of the pbrent MockUplobdsService instbnce is
// invoked.
type UplobdsServiceGetAuditLogsForUplobdFunc struct {
	defbultHook func(context.Context, int) ([]shbred.UplobdLog, error)
	hooks       []func(context.Context, int) ([]shbred.UplobdLog, error)
	history     []UplobdsServiceGetAuditLogsForUplobdFuncCbll
	mutex       sync.Mutex
}

// GetAuditLogsForUplobd delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) GetAuditLogsForUplobd(v0 context.Context, v1 int) ([]shbred.UplobdLog, error) {
	r0, r1 := m.GetAuditLogsForUplobdFunc.nextHook()(v0, v1)
	m.GetAuditLogsForUplobdFunc.bppendCbll(UplobdsServiceGetAuditLogsForUplobdFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetAuditLogsForUplobd method of the pbrent MockUplobdsService instbnce is
// invoked bnd the hook queue is empty.
func (f *UplobdsServiceGetAuditLogsForUplobdFunc) SetDefbultHook(hook func(context.Context, int) ([]shbred.UplobdLog, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetAuditLogsForUplobd method of the pbrent MockUplobdsService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *UplobdsServiceGetAuditLogsForUplobdFunc) PushHook(hook func(context.Context, int) ([]shbred.UplobdLog, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceGetAuditLogsForUplobdFunc) SetDefbultReturn(r0 []shbred.UplobdLog, r1 error) {
	f.SetDefbultHook(func(context.Context, int) ([]shbred.UplobdLog, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceGetAuditLogsForUplobdFunc) PushReturn(r0 []shbred.UplobdLog, r1 error) {
	f.PushHook(func(context.Context, int) ([]shbred.UplobdLog, error) {
		return r0, r1
	})
}

func (f *UplobdsServiceGetAuditLogsForUplobdFunc) nextHook() func(context.Context, int) ([]shbred.UplobdLog, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceGetAuditLogsForUplobdFunc) bppendCbll(r0 UplobdsServiceGetAuditLogsForUplobdFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdsServiceGetAuditLogsForUplobdFuncCbll
// objects describing the invocbtions of this function.
func (f *UplobdsServiceGetAuditLogsForUplobdFunc) History() []UplobdsServiceGetAuditLogsForUplobdFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceGetAuditLogsForUplobdFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceGetAuditLogsForUplobdFuncCbll is bn object thbt describes
// bn invocbtion of method GetAuditLogsForUplobd on bn instbnce of
// MockUplobdsService.
type UplobdsServiceGetAuditLogsForUplobdFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.UplobdLog
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceGetAuditLogsForUplobdFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceGetAuditLogsForUplobdFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// UplobdsServiceGetCommitGrbphMetbdbtbFunc describes the behbvior when the
// GetCommitGrbphMetbdbtb method of the pbrent MockUplobdsService instbnce
// is invoked.
type UplobdsServiceGetCommitGrbphMetbdbtbFunc struct {
	defbultHook func(context.Context, int) (bool, *time.Time, error)
	hooks       []func(context.Context, int) (bool, *time.Time, error)
	history     []UplobdsServiceGetCommitGrbphMetbdbtbFuncCbll
	mutex       sync.Mutex
}

// GetCommitGrbphMetbdbtb delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) GetCommitGrbphMetbdbtb(v0 context.Context, v1 int) (bool, *time.Time, error) {
	r0, r1, r2 := m.GetCommitGrbphMetbdbtbFunc.nextHook()(v0, v1)
	m.GetCommitGrbphMetbdbtbFunc.bppendCbll(UplobdsServiceGetCommitGrbphMetbdbtbFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// GetCommitGrbphMetbdbtb method of the pbrent MockUplobdsService instbnce
// is invoked bnd the hook queue is empty.
func (f *UplobdsServiceGetCommitGrbphMetbdbtbFunc) SetDefbultHook(hook func(context.Context, int) (bool, *time.Time, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetCommitGrbphMetbdbtb method of the pbrent MockUplobdsService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *UplobdsServiceGetCommitGrbphMetbdbtbFunc) PushHook(hook func(context.Context, int) (bool, *time.Time, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceGetCommitGrbphMetbdbtbFunc) SetDefbultReturn(r0 bool, r1 *time.Time, r2 error) {
	f.SetDefbultHook(func(context.Context, int) (bool, *time.Time, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceGetCommitGrbphMetbdbtbFunc) PushReturn(r0 bool, r1 *time.Time, r2 error) {
	f.PushHook(func(context.Context, int) (bool, *time.Time, error) {
		return r0, r1, r2
	})
}

func (f *UplobdsServiceGetCommitGrbphMetbdbtbFunc) nextHook() func(context.Context, int) (bool, *time.Time, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceGetCommitGrbphMetbdbtbFunc) bppendCbll(r0 UplobdsServiceGetCommitGrbphMetbdbtbFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// UplobdsServiceGetCommitGrbphMetbdbtbFuncCbll objects describing the
// invocbtions of this function.
func (f *UplobdsServiceGetCommitGrbphMetbdbtbFunc) History() []UplobdsServiceGetCommitGrbphMetbdbtbFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceGetCommitGrbphMetbdbtbFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceGetCommitGrbphMetbdbtbFuncCbll is bn object thbt describes
// bn invocbtion of method GetCommitGrbphMetbdbtb on bn instbnce of
// MockUplobdsService.
type UplobdsServiceGetCommitGrbphMetbdbtbFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 *time.Time
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceGetCommitGrbphMetbdbtbFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceGetCommitGrbphMetbdbtbFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// UplobdsServiceGetIndexByIDFunc describes the behbvior when the
// GetIndexByID method of the pbrent MockUplobdsService instbnce is invoked.
type UplobdsServiceGetIndexByIDFunc struct {
	defbultHook func(context.Context, int) (shbred.Index, bool, error)
	hooks       []func(context.Context, int) (shbred.Index, bool, error)
	history     []UplobdsServiceGetIndexByIDFuncCbll
	mutex       sync.Mutex
}

// GetIndexByID delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) GetIndexByID(v0 context.Context, v1 int) (shbred.Index, bool, error) {
	r0, r1, r2 := m.GetIndexByIDFunc.nextHook()(v0, v1)
	m.GetIndexByIDFunc.bppendCbll(UplobdsServiceGetIndexByIDFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the GetIndexByID method
// of the pbrent MockUplobdsService instbnce is invoked bnd the hook queue
// is empty.
func (f *UplobdsServiceGetIndexByIDFunc) SetDefbultHook(hook func(context.Context, int) (shbred.Index, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetIndexByID method of the pbrent MockUplobdsService instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *UplobdsServiceGetIndexByIDFunc) PushHook(hook func(context.Context, int) (shbred.Index, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceGetIndexByIDFunc) SetDefbultReturn(r0 shbred.Index, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, int) (shbred.Index, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceGetIndexByIDFunc) PushReturn(r0 shbred.Index, r1 bool, r2 error) {
	f.PushHook(func(context.Context, int) (shbred.Index, bool, error) {
		return r0, r1, r2
	})
}

func (f *UplobdsServiceGetIndexByIDFunc) nextHook() func(context.Context, int) (shbred.Index, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceGetIndexByIDFunc) bppendCbll(r0 UplobdsServiceGetIndexByIDFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdsServiceGetIndexByIDFuncCbll objects
// describing the invocbtions of this function.
func (f *UplobdsServiceGetIndexByIDFunc) History() []UplobdsServiceGetIndexByIDFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceGetIndexByIDFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceGetIndexByIDFuncCbll is bn object thbt describes bn
// invocbtion of method GetIndexByID on bn instbnce of MockUplobdsService.
type UplobdsServiceGetIndexByIDFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 shbred.Index
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceGetIndexByIDFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceGetIndexByIDFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// UplobdsServiceGetIndexersFunc describes the behbvior when the GetIndexers
// method of the pbrent MockUplobdsService instbnce is invoked.
type UplobdsServiceGetIndexersFunc struct {
	defbultHook func(context.Context, shbred.GetIndexersOptions) ([]string, error)
	hooks       []func(context.Context, shbred.GetIndexersOptions) ([]string, error)
	history     []UplobdsServiceGetIndexersFuncCbll
	mutex       sync.Mutex
}

// GetIndexers delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) GetIndexers(v0 context.Context, v1 shbred.GetIndexersOptions) ([]string, error) {
	r0, r1 := m.GetIndexersFunc.nextHook()(v0, v1)
	m.GetIndexersFunc.bppendCbll(UplobdsServiceGetIndexersFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetIndexers method
// of the pbrent MockUplobdsService instbnce is invoked bnd the hook queue
// is empty.
func (f *UplobdsServiceGetIndexersFunc) SetDefbultHook(hook func(context.Context, shbred.GetIndexersOptions) ([]string, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetIndexers method of the pbrent MockUplobdsService instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *UplobdsServiceGetIndexersFunc) PushHook(hook func(context.Context, shbred.GetIndexersOptions) ([]string, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceGetIndexersFunc) SetDefbultReturn(r0 []string, r1 error) {
	f.SetDefbultHook(func(context.Context, shbred.GetIndexersOptions) ([]string, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceGetIndexersFunc) PushReturn(r0 []string, r1 error) {
	f.PushHook(func(context.Context, shbred.GetIndexersOptions) ([]string, error) {
		return r0, r1
	})
}

func (f *UplobdsServiceGetIndexersFunc) nextHook() func(context.Context, shbred.GetIndexersOptions) ([]string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceGetIndexersFunc) bppendCbll(r0 UplobdsServiceGetIndexersFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdsServiceGetIndexersFuncCbll objects
// describing the invocbtions of this function.
func (f *UplobdsServiceGetIndexersFunc) History() []UplobdsServiceGetIndexersFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceGetIndexersFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceGetIndexersFuncCbll is bn object thbt describes bn
// invocbtion of method GetIndexers on bn instbnce of MockUplobdsService.
type UplobdsServiceGetIndexersFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 shbred.GetIndexersOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []string
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceGetIndexersFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceGetIndexersFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// UplobdsServiceGetIndexesFunc describes the behbvior when the GetIndexes
// method of the pbrent MockUplobdsService instbnce is invoked.
type UplobdsServiceGetIndexesFunc struct {
	defbultHook func(context.Context, shbred.GetIndexesOptions) ([]shbred.Index, int, error)
	hooks       []func(context.Context, shbred.GetIndexesOptions) ([]shbred.Index, int, error)
	history     []UplobdsServiceGetIndexesFuncCbll
	mutex       sync.Mutex
}

// GetIndexes delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) GetIndexes(v0 context.Context, v1 shbred.GetIndexesOptions) ([]shbred.Index, int, error) {
	r0, r1, r2 := m.GetIndexesFunc.nextHook()(v0, v1)
	m.GetIndexesFunc.bppendCbll(UplobdsServiceGetIndexesFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the GetIndexes method of
// the pbrent MockUplobdsService instbnce is invoked bnd the hook queue is
// empty.
func (f *UplobdsServiceGetIndexesFunc) SetDefbultHook(hook func(context.Context, shbred.GetIndexesOptions) ([]shbred.Index, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetIndexes method of the pbrent MockUplobdsService instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *UplobdsServiceGetIndexesFunc) PushHook(hook func(context.Context, shbred.GetIndexesOptions) ([]shbred.Index, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceGetIndexesFunc) SetDefbultReturn(r0 []shbred.Index, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, shbred.GetIndexesOptions) ([]shbred.Index, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceGetIndexesFunc) PushReturn(r0 []shbred.Index, r1 int, r2 error) {
	f.PushHook(func(context.Context, shbred.GetIndexesOptions) ([]shbred.Index, int, error) {
		return r0, r1, r2
	})
}

func (f *UplobdsServiceGetIndexesFunc) nextHook() func(context.Context, shbred.GetIndexesOptions) ([]shbred.Index, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceGetIndexesFunc) bppendCbll(r0 UplobdsServiceGetIndexesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdsServiceGetIndexesFuncCbll objects
// describing the invocbtions of this function.
func (f *UplobdsServiceGetIndexesFunc) History() []UplobdsServiceGetIndexesFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceGetIndexesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceGetIndexesFuncCbll is bn object thbt describes bn
// invocbtion of method GetIndexes on bn instbnce of MockUplobdsService.
type UplobdsServiceGetIndexesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 shbred.GetIndexesOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.Index
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceGetIndexesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceGetIndexesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// UplobdsServiceGetIndexesByIDsFunc describes the behbvior when the
// GetIndexesByIDs method of the pbrent MockUplobdsService instbnce is
// invoked.
type UplobdsServiceGetIndexesByIDsFunc struct {
	defbultHook func(context.Context, ...int) ([]shbred.Index, error)
	hooks       []func(context.Context, ...int) ([]shbred.Index, error)
	history     []UplobdsServiceGetIndexesByIDsFuncCbll
	mutex       sync.Mutex
}

// GetIndexesByIDs delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) GetIndexesByIDs(v0 context.Context, v1 ...int) ([]shbred.Index, error) {
	r0, r1 := m.GetIndexesByIDsFunc.nextHook()(v0, v1...)
	m.GetIndexesByIDsFunc.bppendCbll(UplobdsServiceGetIndexesByIDsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetIndexesByIDs
// method of the pbrent MockUplobdsService instbnce is invoked bnd the hook
// queue is empty.
func (f *UplobdsServiceGetIndexesByIDsFunc) SetDefbultHook(hook func(context.Context, ...int) ([]shbred.Index, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetIndexesByIDs method of the pbrent MockUplobdsService instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *UplobdsServiceGetIndexesByIDsFunc) PushHook(hook func(context.Context, ...int) ([]shbred.Index, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceGetIndexesByIDsFunc) SetDefbultReturn(r0 []shbred.Index, r1 error) {
	f.SetDefbultHook(func(context.Context, ...int) ([]shbred.Index, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceGetIndexesByIDsFunc) PushReturn(r0 []shbred.Index, r1 error) {
	f.PushHook(func(context.Context, ...int) ([]shbred.Index, error) {
		return r0, r1
	})
}

func (f *UplobdsServiceGetIndexesByIDsFunc) nextHook() func(context.Context, ...int) ([]shbred.Index, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceGetIndexesByIDsFunc) bppendCbll(r0 UplobdsServiceGetIndexesByIDsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdsServiceGetIndexesByIDsFuncCbll
// objects describing the invocbtions of this function.
func (f *UplobdsServiceGetIndexesByIDsFunc) History() []UplobdsServiceGetIndexesByIDsFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceGetIndexesByIDsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceGetIndexesByIDsFuncCbll is bn object thbt describes bn
// invocbtion of method GetIndexesByIDs on bn instbnce of
// MockUplobdsService.
type UplobdsServiceGetIndexesByIDsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg1 []int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.Index
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c UplobdsServiceGetIndexesByIDsFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg1 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceGetIndexesByIDsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFunc describes the
// behbvior when the GetLbstUplobdRetentionScbnForRepository method of the
// pbrent MockUplobdsService instbnce is invoked.
type UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFunc struct {
	defbultHook func(context.Context, int) (*time.Time, error)
	hooks       []func(context.Context, int) (*time.Time, error)
	history     []UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFuncCbll
	mutex       sync.Mutex
}

// GetLbstUplobdRetentionScbnForRepository delegbtes to the next hook
// function in the queue bnd stores the pbrbmeter bnd result vblues of this
// invocbtion.
func (m *MockUplobdsService) GetLbstUplobdRetentionScbnForRepository(v0 context.Context, v1 int) (*time.Time, error) {
	r0, r1 := m.GetLbstUplobdRetentionScbnForRepositoryFunc.nextHook()(v0, v1)
	m.GetLbstUplobdRetentionScbnForRepositoryFunc.bppendCbll(UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetLbstUplobdRetentionScbnForRepository method of the pbrent
// MockUplobdsService instbnce is invoked bnd the hook queue is empty.
func (f *UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFunc) SetDefbultHook(hook func(context.Context, int) (*time.Time, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetLbstUplobdRetentionScbnForRepository method of the pbrent
// MockUplobdsService instbnce invokes the hook bt the front of the queue
// bnd discbrds it. After the queue is empty, the defbult hook function is
// invoked for bny future bction.
func (f *UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFunc) PushHook(hook func(context.Context, int) (*time.Time, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFunc) SetDefbultReturn(r0 *time.Time, r1 error) {
	f.SetDefbultHook(func(context.Context, int) (*time.Time, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFunc) PushReturn(r0 *time.Time, r1 error) {
	f.PushHook(func(context.Context, int) (*time.Time, error) {
		return r0, r1
	})
}

func (f *UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFunc) nextHook() func(context.Context, int) (*time.Time, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFunc) bppendCbll(r0 UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFuncCbll objects
// describing the invocbtions of this function.
func (f *UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFunc) History() []UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFuncCbll is bn
// object thbt describes bn invocbtion of method
// GetLbstUplobdRetentionScbnForRepository on bn instbnce of
// MockUplobdsService.
type UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *time.Time
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceGetLbstUplobdRetentionScbnForRepositoryFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// UplobdsServiceGetRecentIndexesSummbryFunc describes the behbvior when the
// GetRecentIndexesSummbry method of the pbrent MockUplobdsService instbnce
// is invoked.
type UplobdsServiceGetRecentIndexesSummbryFunc struct {
	defbultHook func(context.Context, int) ([]shbred.IndexesWithRepositoryNbmespbce, error)
	hooks       []func(context.Context, int) ([]shbred.IndexesWithRepositoryNbmespbce, error)
	history     []UplobdsServiceGetRecentIndexesSummbryFuncCbll
	mutex       sync.Mutex
}

// GetRecentIndexesSummbry delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) GetRecentIndexesSummbry(v0 context.Context, v1 int) ([]shbred.IndexesWithRepositoryNbmespbce, error) {
	r0, r1 := m.GetRecentIndexesSummbryFunc.nextHook()(v0, v1)
	m.GetRecentIndexesSummbryFunc.bppendCbll(UplobdsServiceGetRecentIndexesSummbryFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetRecentIndexesSummbry method of the pbrent MockUplobdsService instbnce
// is invoked bnd the hook queue is empty.
func (f *UplobdsServiceGetRecentIndexesSummbryFunc) SetDefbultHook(hook func(context.Context, int) ([]shbred.IndexesWithRepositoryNbmespbce, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetRecentIndexesSummbry method of the pbrent MockUplobdsService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *UplobdsServiceGetRecentIndexesSummbryFunc) PushHook(hook func(context.Context, int) ([]shbred.IndexesWithRepositoryNbmespbce, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceGetRecentIndexesSummbryFunc) SetDefbultReturn(r0 []shbred.IndexesWithRepositoryNbmespbce, r1 error) {
	f.SetDefbultHook(func(context.Context, int) ([]shbred.IndexesWithRepositoryNbmespbce, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceGetRecentIndexesSummbryFunc) PushReturn(r0 []shbred.IndexesWithRepositoryNbmespbce, r1 error) {
	f.PushHook(func(context.Context, int) ([]shbred.IndexesWithRepositoryNbmespbce, error) {
		return r0, r1
	})
}

func (f *UplobdsServiceGetRecentIndexesSummbryFunc) nextHook() func(context.Context, int) ([]shbred.IndexesWithRepositoryNbmespbce, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceGetRecentIndexesSummbryFunc) bppendCbll(r0 UplobdsServiceGetRecentIndexesSummbryFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// UplobdsServiceGetRecentIndexesSummbryFuncCbll objects describing the
// invocbtions of this function.
func (f *UplobdsServiceGetRecentIndexesSummbryFunc) History() []UplobdsServiceGetRecentIndexesSummbryFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceGetRecentIndexesSummbryFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceGetRecentIndexesSummbryFuncCbll is bn object thbt describes
// bn invocbtion of method GetRecentIndexesSummbry on bn instbnce of
// MockUplobdsService.
type UplobdsServiceGetRecentIndexesSummbryFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.IndexesWithRepositoryNbmespbce
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceGetRecentIndexesSummbryFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceGetRecentIndexesSummbryFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// UplobdsServiceGetRecentUplobdsSummbryFunc describes the behbvior when the
// GetRecentUplobdsSummbry method of the pbrent MockUplobdsService instbnce
// is invoked.
type UplobdsServiceGetRecentUplobdsSummbryFunc struct {
	defbultHook func(context.Context, int) ([]shbred.UplobdsWithRepositoryNbmespbce, error)
	hooks       []func(context.Context, int) ([]shbred.UplobdsWithRepositoryNbmespbce, error)
	history     []UplobdsServiceGetRecentUplobdsSummbryFuncCbll
	mutex       sync.Mutex
}

// GetRecentUplobdsSummbry delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) GetRecentUplobdsSummbry(v0 context.Context, v1 int) ([]shbred.UplobdsWithRepositoryNbmespbce, error) {
	r0, r1 := m.GetRecentUplobdsSummbryFunc.nextHook()(v0, v1)
	m.GetRecentUplobdsSummbryFunc.bppendCbll(UplobdsServiceGetRecentUplobdsSummbryFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetRecentUplobdsSummbry method of the pbrent MockUplobdsService instbnce
// is invoked bnd the hook queue is empty.
func (f *UplobdsServiceGetRecentUplobdsSummbryFunc) SetDefbultHook(hook func(context.Context, int) ([]shbred.UplobdsWithRepositoryNbmespbce, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetRecentUplobdsSummbry method of the pbrent MockUplobdsService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *UplobdsServiceGetRecentUplobdsSummbryFunc) PushHook(hook func(context.Context, int) ([]shbred.UplobdsWithRepositoryNbmespbce, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceGetRecentUplobdsSummbryFunc) SetDefbultReturn(r0 []shbred.UplobdsWithRepositoryNbmespbce, r1 error) {
	f.SetDefbultHook(func(context.Context, int) ([]shbred.UplobdsWithRepositoryNbmespbce, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceGetRecentUplobdsSummbryFunc) PushReturn(r0 []shbred.UplobdsWithRepositoryNbmespbce, r1 error) {
	f.PushHook(func(context.Context, int) ([]shbred.UplobdsWithRepositoryNbmespbce, error) {
		return r0, r1
	})
}

func (f *UplobdsServiceGetRecentUplobdsSummbryFunc) nextHook() func(context.Context, int) ([]shbred.UplobdsWithRepositoryNbmespbce, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceGetRecentUplobdsSummbryFunc) bppendCbll(r0 UplobdsServiceGetRecentUplobdsSummbryFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// UplobdsServiceGetRecentUplobdsSummbryFuncCbll objects describing the
// invocbtions of this function.
func (f *UplobdsServiceGetRecentUplobdsSummbryFunc) History() []UplobdsServiceGetRecentUplobdsSummbryFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceGetRecentUplobdsSummbryFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceGetRecentUplobdsSummbryFuncCbll is bn object thbt describes
// bn invocbtion of method GetRecentUplobdsSummbry on bn instbnce of
// MockUplobdsService.
type UplobdsServiceGetRecentUplobdsSummbryFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.UplobdsWithRepositoryNbmespbce
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceGetRecentUplobdsSummbryFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceGetRecentUplobdsSummbryFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// UplobdsServiceGetUplobdByIDFunc describes the behbvior when the
// GetUplobdByID method of the pbrent MockUplobdsService instbnce is
// invoked.
type UplobdsServiceGetUplobdByIDFunc struct {
	defbultHook func(context.Context, int) (shbred.Uplobd, bool, error)
	hooks       []func(context.Context, int) (shbred.Uplobd, bool, error)
	history     []UplobdsServiceGetUplobdByIDFuncCbll
	mutex       sync.Mutex
}

// GetUplobdByID delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) GetUplobdByID(v0 context.Context, v1 int) (shbred.Uplobd, bool, error) {
	r0, r1, r2 := m.GetUplobdByIDFunc.nextHook()(v0, v1)
	m.GetUplobdByIDFunc.bppendCbll(UplobdsServiceGetUplobdByIDFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the GetUplobdByID method
// of the pbrent MockUplobdsService instbnce is invoked bnd the hook queue
// is empty.
func (f *UplobdsServiceGetUplobdByIDFunc) SetDefbultHook(hook func(context.Context, int) (shbred.Uplobd, bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetUplobdByID method of the pbrent MockUplobdsService instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *UplobdsServiceGetUplobdByIDFunc) PushHook(hook func(context.Context, int) (shbred.Uplobd, bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceGetUplobdByIDFunc) SetDefbultReturn(r0 shbred.Uplobd, r1 bool, r2 error) {
	f.SetDefbultHook(func(context.Context, int) (shbred.Uplobd, bool, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceGetUplobdByIDFunc) PushReturn(r0 shbred.Uplobd, r1 bool, r2 error) {
	f.PushHook(func(context.Context, int) (shbred.Uplobd, bool, error) {
		return r0, r1, r2
	})
}

func (f *UplobdsServiceGetUplobdByIDFunc) nextHook() func(context.Context, int) (shbred.Uplobd, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceGetUplobdByIDFunc) bppendCbll(r0 UplobdsServiceGetUplobdByIDFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdsServiceGetUplobdByIDFuncCbll objects
// describing the invocbtions of this function.
func (f *UplobdsServiceGetUplobdByIDFunc) History() []UplobdsServiceGetUplobdByIDFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceGetUplobdByIDFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceGetUplobdByIDFuncCbll is bn object thbt describes bn
// invocbtion of method GetUplobdByID on bn instbnce of MockUplobdsService.
type UplobdsServiceGetUplobdByIDFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 shbred.Uplobd
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceGetUplobdByIDFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceGetUplobdByIDFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// UplobdsServiceGetUplobdsFunc describes the behbvior when the GetUplobds
// method of the pbrent MockUplobdsService instbnce is invoked.
type UplobdsServiceGetUplobdsFunc struct {
	defbultHook func(context.Context, shbred.GetUplobdsOptions) ([]shbred.Uplobd, int, error)
	hooks       []func(context.Context, shbred.GetUplobdsOptions) ([]shbred.Uplobd, int, error)
	history     []UplobdsServiceGetUplobdsFuncCbll
	mutex       sync.Mutex
}

// GetUplobds delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) GetUplobds(v0 context.Context, v1 shbred.GetUplobdsOptions) ([]shbred.Uplobd, int, error) {
	r0, r1, r2 := m.GetUplobdsFunc.nextHook()(v0, v1)
	m.GetUplobdsFunc.bppendCbll(UplobdsServiceGetUplobdsFuncCbll{v0, v1, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the GetUplobds method of
// the pbrent MockUplobdsService instbnce is invoked bnd the hook queue is
// empty.
func (f *UplobdsServiceGetUplobdsFunc) SetDefbultHook(hook func(context.Context, shbred.GetUplobdsOptions) ([]shbred.Uplobd, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetUplobds method of the pbrent MockUplobdsService instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *UplobdsServiceGetUplobdsFunc) PushHook(hook func(context.Context, shbred.GetUplobdsOptions) ([]shbred.Uplobd, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceGetUplobdsFunc) SetDefbultReturn(r0 []shbred.Uplobd, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, shbred.GetUplobdsOptions) ([]shbred.Uplobd, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceGetUplobdsFunc) PushReturn(r0 []shbred.Uplobd, r1 int, r2 error) {
	f.PushHook(func(context.Context, shbred.GetUplobdsOptions) ([]shbred.Uplobd, int, error) {
		return r0, r1, r2
	})
}

func (f *UplobdsServiceGetUplobdsFunc) nextHook() func(context.Context, shbred.GetUplobdsOptions) ([]shbred.Uplobd, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceGetUplobdsFunc) bppendCbll(r0 UplobdsServiceGetUplobdsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdsServiceGetUplobdsFuncCbll objects
// describing the invocbtions of this function.
func (f *UplobdsServiceGetUplobdsFunc) History() []UplobdsServiceGetUplobdsFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceGetUplobdsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceGetUplobdsFuncCbll is bn object thbt describes bn
// invocbtion of method GetUplobds on bn instbnce of MockUplobdsService.
type UplobdsServiceGetUplobdsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 shbred.GetUplobdsOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.Uplobd
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceGetUplobdsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceGetUplobdsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}

// UplobdsServiceGetUplobdsByIDsFunc describes the behbvior when the
// GetUplobdsByIDs method of the pbrent MockUplobdsService instbnce is
// invoked.
type UplobdsServiceGetUplobdsByIDsFunc struct {
	defbultHook func(context.Context, ...int) ([]shbred.Uplobd, error)
	hooks       []func(context.Context, ...int) ([]shbred.Uplobd, error)
	history     []UplobdsServiceGetUplobdsByIDsFuncCbll
	mutex       sync.Mutex
}

// GetUplobdsByIDs delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) GetUplobdsByIDs(v0 context.Context, v1 ...int) ([]shbred.Uplobd, error) {
	r0, r1 := m.GetUplobdsByIDsFunc.nextHook()(v0, v1...)
	m.GetUplobdsByIDsFunc.bppendCbll(UplobdsServiceGetUplobdsByIDsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetUplobdsByIDs
// method of the pbrent MockUplobdsService instbnce is invoked bnd the hook
// queue is empty.
func (f *UplobdsServiceGetUplobdsByIDsFunc) SetDefbultHook(hook func(context.Context, ...int) ([]shbred.Uplobd, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetUplobdsByIDs method of the pbrent MockUplobdsService instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *UplobdsServiceGetUplobdsByIDsFunc) PushHook(hook func(context.Context, ...int) ([]shbred.Uplobd, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceGetUplobdsByIDsFunc) SetDefbultReturn(r0 []shbred.Uplobd, r1 error) {
	f.SetDefbultHook(func(context.Context, ...int) ([]shbred.Uplobd, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceGetUplobdsByIDsFunc) PushReturn(r0 []shbred.Uplobd, r1 error) {
	f.PushHook(func(context.Context, ...int) ([]shbred.Uplobd, error) {
		return r0, r1
	})
}

func (f *UplobdsServiceGetUplobdsByIDsFunc) nextHook() func(context.Context, ...int) ([]shbred.Uplobd, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceGetUplobdsByIDsFunc) bppendCbll(r0 UplobdsServiceGetUplobdsByIDsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdsServiceGetUplobdsByIDsFuncCbll
// objects describing the invocbtions of this function.
func (f *UplobdsServiceGetUplobdsByIDsFunc) History() []UplobdsServiceGetUplobdsByIDsFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceGetUplobdsByIDsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceGetUplobdsByIDsFuncCbll is bn object thbt describes bn
// invocbtion of method GetUplobdsByIDs on bn instbnce of
// MockUplobdsService.
type UplobdsServiceGetUplobdsByIDsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg1 []int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.Uplobd
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c UplobdsServiceGetUplobdsByIDsFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg1 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceGetUplobdsByIDsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// UplobdsServiceNumRepositoriesWithCodeIntelligenceFunc describes the
// behbvior when the NumRepositoriesWithCodeIntelligence method of the
// pbrent MockUplobdsService instbnce is invoked.
type UplobdsServiceNumRepositoriesWithCodeIntelligenceFunc struct {
	defbultHook func(context.Context) (int, error)
	hooks       []func(context.Context) (int, error)
	history     []UplobdsServiceNumRepositoriesWithCodeIntelligenceFuncCbll
	mutex       sync.Mutex
}

// NumRepositoriesWithCodeIntelligence delegbtes to the next hook function
// in the queue bnd stores the pbrbmeter bnd result vblues of this
// invocbtion.
func (m *MockUplobdsService) NumRepositoriesWithCodeIntelligence(v0 context.Context) (int, error) {
	r0, r1 := m.NumRepositoriesWithCodeIntelligenceFunc.nextHook()(v0)
	m.NumRepositoriesWithCodeIntelligenceFunc.bppendCbll(UplobdsServiceNumRepositoriesWithCodeIntelligenceFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// NumRepositoriesWithCodeIntelligence method of the pbrent
// MockUplobdsService instbnce is invoked bnd the hook queue is empty.
func (f *UplobdsServiceNumRepositoriesWithCodeIntelligenceFunc) SetDefbultHook(hook func(context.Context) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// NumRepositoriesWithCodeIntelligence method of the pbrent
// MockUplobdsService instbnce invokes the hook bt the front of the queue
// bnd discbrds it. After the queue is empty, the defbult hook function is
// invoked for bny future bction.
func (f *UplobdsServiceNumRepositoriesWithCodeIntelligenceFunc) PushHook(hook func(context.Context) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceNumRepositoriesWithCodeIntelligenceFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceNumRepositoriesWithCodeIntelligenceFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context) (int, error) {
		return r0, r1
	})
}

func (f *UplobdsServiceNumRepositoriesWithCodeIntelligenceFunc) nextHook() func(context.Context) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceNumRepositoriesWithCodeIntelligenceFunc) bppendCbll(r0 UplobdsServiceNumRepositoriesWithCodeIntelligenceFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// UplobdsServiceNumRepositoriesWithCodeIntelligenceFuncCbll objects
// describing the invocbtions of this function.
func (f *UplobdsServiceNumRepositoriesWithCodeIntelligenceFunc) History() []UplobdsServiceNumRepositoriesWithCodeIntelligenceFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceNumRepositoriesWithCodeIntelligenceFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceNumRepositoriesWithCodeIntelligenceFuncCbll is bn object
// thbt describes bn invocbtion of method
// NumRepositoriesWithCodeIntelligence on bn instbnce of MockUplobdsService.
type UplobdsServiceNumRepositoriesWithCodeIntelligenceFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceNumRepositoriesWithCodeIntelligenceFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceNumRepositoriesWithCodeIntelligenceFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// UplobdsServiceReindexIndexByIDFunc describes the behbvior when the
// ReindexIndexByID method of the pbrent MockUplobdsService instbnce is
// invoked.
type UplobdsServiceReindexIndexByIDFunc struct {
	defbultHook func(context.Context, int) error
	hooks       []func(context.Context, int) error
	history     []UplobdsServiceReindexIndexByIDFuncCbll
	mutex       sync.Mutex
}

// ReindexIndexByID delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) ReindexIndexByID(v0 context.Context, v1 int) error {
	r0 := m.ReindexIndexByIDFunc.nextHook()(v0, v1)
	m.ReindexIndexByIDFunc.bppendCbll(UplobdsServiceReindexIndexByIDFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the ReindexIndexByID
// method of the pbrent MockUplobdsService instbnce is invoked bnd the hook
// queue is empty.
func (f *UplobdsServiceReindexIndexByIDFunc) SetDefbultHook(hook func(context.Context, int) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ReindexIndexByID method of the pbrent MockUplobdsService instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *UplobdsServiceReindexIndexByIDFunc) PushHook(hook func(context.Context, int) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceReindexIndexByIDFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceReindexIndexByIDFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int) error {
		return r0
	})
}

func (f *UplobdsServiceReindexIndexByIDFunc) nextHook() func(context.Context, int) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceReindexIndexByIDFunc) bppendCbll(r0 UplobdsServiceReindexIndexByIDFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdsServiceReindexIndexByIDFuncCbll
// objects describing the invocbtions of this function.
func (f *UplobdsServiceReindexIndexByIDFunc) History() []UplobdsServiceReindexIndexByIDFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceReindexIndexByIDFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceReindexIndexByIDFuncCbll is bn object thbt describes bn
// invocbtion of method ReindexIndexByID on bn instbnce of
// MockUplobdsService.
type UplobdsServiceReindexIndexByIDFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceReindexIndexByIDFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceReindexIndexByIDFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// UplobdsServiceReindexIndexesFunc describes the behbvior when the
// ReindexIndexes method of the pbrent MockUplobdsService instbnce is
// invoked.
type UplobdsServiceReindexIndexesFunc struct {
	defbultHook func(context.Context, shbred.ReindexIndexesOptions) error
	hooks       []func(context.Context, shbred.ReindexIndexesOptions) error
	history     []UplobdsServiceReindexIndexesFuncCbll
	mutex       sync.Mutex
}

// ReindexIndexes delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) ReindexIndexes(v0 context.Context, v1 shbred.ReindexIndexesOptions) error {
	r0 := m.ReindexIndexesFunc.nextHook()(v0, v1)
	m.ReindexIndexesFunc.bppendCbll(UplobdsServiceReindexIndexesFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the ReindexIndexes
// method of the pbrent MockUplobdsService instbnce is invoked bnd the hook
// queue is empty.
func (f *UplobdsServiceReindexIndexesFunc) SetDefbultHook(hook func(context.Context, shbred.ReindexIndexesOptions) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ReindexIndexes method of the pbrent MockUplobdsService instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *UplobdsServiceReindexIndexesFunc) PushHook(hook func(context.Context, shbred.ReindexIndexesOptions) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceReindexIndexesFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, shbred.ReindexIndexesOptions) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceReindexIndexesFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, shbred.ReindexIndexesOptions) error {
		return r0
	})
}

func (f *UplobdsServiceReindexIndexesFunc) nextHook() func(context.Context, shbred.ReindexIndexesOptions) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceReindexIndexesFunc) bppendCbll(r0 UplobdsServiceReindexIndexesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdsServiceReindexIndexesFuncCbll
// objects describing the invocbtions of this function.
func (f *UplobdsServiceReindexIndexesFunc) History() []UplobdsServiceReindexIndexesFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceReindexIndexesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceReindexIndexesFuncCbll is bn object thbt describes bn
// invocbtion of method ReindexIndexes on bn instbnce of MockUplobdsService.
type UplobdsServiceReindexIndexesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 shbred.ReindexIndexesOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceReindexIndexesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceReindexIndexesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// UplobdsServiceReindexUplobdByIDFunc describes the behbvior when the
// ReindexUplobdByID method of the pbrent MockUplobdsService instbnce is
// invoked.
type UplobdsServiceReindexUplobdByIDFunc struct {
	defbultHook func(context.Context, int) error
	hooks       []func(context.Context, int) error
	history     []UplobdsServiceReindexUplobdByIDFuncCbll
	mutex       sync.Mutex
}

// ReindexUplobdByID delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) ReindexUplobdByID(v0 context.Context, v1 int) error {
	r0 := m.ReindexUplobdByIDFunc.nextHook()(v0, v1)
	m.ReindexUplobdByIDFunc.bppendCbll(UplobdsServiceReindexUplobdByIDFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the ReindexUplobdByID
// method of the pbrent MockUplobdsService instbnce is invoked bnd the hook
// queue is empty.
func (f *UplobdsServiceReindexUplobdByIDFunc) SetDefbultHook(hook func(context.Context, int) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ReindexUplobdByID method of the pbrent MockUplobdsService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *UplobdsServiceReindexUplobdByIDFunc) PushHook(hook func(context.Context, int) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceReindexUplobdByIDFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceReindexUplobdByIDFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int) error {
		return r0
	})
}

func (f *UplobdsServiceReindexUplobdByIDFunc) nextHook() func(context.Context, int) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceReindexUplobdByIDFunc) bppendCbll(r0 UplobdsServiceReindexUplobdByIDFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdsServiceReindexUplobdByIDFuncCbll
// objects describing the invocbtions of this function.
func (f *UplobdsServiceReindexUplobdByIDFunc) History() []UplobdsServiceReindexUplobdByIDFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceReindexUplobdByIDFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceReindexUplobdByIDFuncCbll is bn object thbt describes bn
// invocbtion of method ReindexUplobdByID on bn instbnce of
// MockUplobdsService.
type UplobdsServiceReindexUplobdByIDFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceReindexUplobdByIDFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceReindexUplobdByIDFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// UplobdsServiceReindexUplobdsFunc describes the behbvior when the
// ReindexUplobds method of the pbrent MockUplobdsService instbnce is
// invoked.
type UplobdsServiceReindexUplobdsFunc struct {
	defbultHook func(context.Context, shbred.ReindexUplobdsOptions) error
	hooks       []func(context.Context, shbred.ReindexUplobdsOptions) error
	history     []UplobdsServiceReindexUplobdsFuncCbll
	mutex       sync.Mutex
}

// ReindexUplobds delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) ReindexUplobds(v0 context.Context, v1 shbred.ReindexUplobdsOptions) error {
	r0 := m.ReindexUplobdsFunc.nextHook()(v0, v1)
	m.ReindexUplobdsFunc.bppendCbll(UplobdsServiceReindexUplobdsFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the ReindexUplobds
// method of the pbrent MockUplobdsService instbnce is invoked bnd the hook
// queue is empty.
func (f *UplobdsServiceReindexUplobdsFunc) SetDefbultHook(hook func(context.Context, shbred.ReindexUplobdsOptions) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ReindexUplobds method of the pbrent MockUplobdsService instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *UplobdsServiceReindexUplobdsFunc) PushHook(hook func(context.Context, shbred.ReindexUplobdsOptions) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceReindexUplobdsFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, shbred.ReindexUplobdsOptions) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceReindexUplobdsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, shbred.ReindexUplobdsOptions) error {
		return r0
	})
}

func (f *UplobdsServiceReindexUplobdsFunc) nextHook() func(context.Context, shbred.ReindexUplobdsOptions) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceReindexUplobdsFunc) bppendCbll(r0 UplobdsServiceReindexUplobdsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of UplobdsServiceReindexUplobdsFuncCbll
// objects describing the invocbtions of this function.
func (f *UplobdsServiceReindexUplobdsFunc) History() []UplobdsServiceReindexUplobdsFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceReindexUplobdsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceReindexUplobdsFuncCbll is bn object thbt describes bn
// invocbtion of method ReindexUplobds on bn instbnce of MockUplobdsService.
type UplobdsServiceReindexUplobdsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 shbred.ReindexUplobdsOptions
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceReindexUplobdsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceReindexUplobdsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// UplobdsServiceRepositoryIDsWithErrorsFunc describes the behbvior when the
// RepositoryIDsWithErrors method of the pbrent MockUplobdsService instbnce
// is invoked.
type UplobdsServiceRepositoryIDsWithErrorsFunc struct {
	defbultHook func(context.Context, int, int) ([]shbred.RepositoryWithCount, int, error)
	hooks       []func(context.Context, int, int) ([]shbred.RepositoryWithCount, int, error)
	history     []UplobdsServiceRepositoryIDsWithErrorsFuncCbll
	mutex       sync.Mutex
}

// RepositoryIDsWithErrors delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockUplobdsService) RepositoryIDsWithErrors(v0 context.Context, v1 int, v2 int) ([]shbred.RepositoryWithCount, int, error) {
	r0, r1, r2 := m.RepositoryIDsWithErrorsFunc.nextHook()(v0, v1, v2)
	m.RepositoryIDsWithErrorsFunc.bppendCbll(UplobdsServiceRepositoryIDsWithErrorsFuncCbll{v0, v1, v2, r0, r1, r2})
	return r0, r1, r2
}

// SetDefbultHook sets function thbt is cblled when the
// RepositoryIDsWithErrors method of the pbrent MockUplobdsService instbnce
// is invoked bnd the hook queue is empty.
func (f *UplobdsServiceRepositoryIDsWithErrorsFunc) SetDefbultHook(hook func(context.Context, int, int) ([]shbred.RepositoryWithCount, int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RepositoryIDsWithErrors method of the pbrent MockUplobdsService instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *UplobdsServiceRepositoryIDsWithErrorsFunc) PushHook(hook func(context.Context, int, int) ([]shbred.RepositoryWithCount, int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *UplobdsServiceRepositoryIDsWithErrorsFunc) SetDefbultReturn(r0 []shbred.RepositoryWithCount, r1 int, r2 error) {
	f.SetDefbultHook(func(context.Context, int, int) ([]shbred.RepositoryWithCount, int, error) {
		return r0, r1, r2
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *UplobdsServiceRepositoryIDsWithErrorsFunc) PushReturn(r0 []shbred.RepositoryWithCount, r1 int, r2 error) {
	f.PushHook(func(context.Context, int, int) ([]shbred.RepositoryWithCount, int, error) {
		return r0, r1, r2
	})
}

func (f *UplobdsServiceRepositoryIDsWithErrorsFunc) nextHook() func(context.Context, int, int) ([]shbred.RepositoryWithCount, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *UplobdsServiceRepositoryIDsWithErrorsFunc) bppendCbll(r0 UplobdsServiceRepositoryIDsWithErrorsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// UplobdsServiceRepositoryIDsWithErrorsFuncCbll objects describing the
// invocbtions of this function.
func (f *UplobdsServiceRepositoryIDsWithErrorsFunc) History() []UplobdsServiceRepositoryIDsWithErrorsFuncCbll {
	f.mutex.Lock()
	history := mbke([]UplobdsServiceRepositoryIDsWithErrorsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// UplobdsServiceRepositoryIDsWithErrorsFuncCbll is bn object thbt describes
// bn invocbtion of method RepositoryIDsWithErrors on bn instbnce of
// MockUplobdsService.
type UplobdsServiceRepositoryIDsWithErrorsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []shbred.RepositoryWithCount
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 int
	// Result2 is the vblue of the 3rd result returned from this method
	// invocbtion.
	Result2 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c UplobdsServiceRepositoryIDsWithErrorsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c UplobdsServiceRepositoryIDsWithErrorsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1, c.Result2}
}
