// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge store

import (
	"context"
	"sync"

	bbsestore "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	types "github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
)

// MockDbtbSeriesStore is b mock implementbtion of the DbtbSeriesStore
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/insights/store) used for unit
// testing.
type MockDbtbSeriesStore struct {
	// CompleteJustInTimeConversionAttemptFunc is bn instbnce of b mock
	// function object controlling the behbvior of the method
	// CompleteJustInTimeConversionAttempt.
	CompleteJustInTimeConversionAttemptFunc *DbtbSeriesStoreCompleteJustInTimeConversionAttemptFunc
	// GetDbtbSeriesFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetDbtbSeries.
	GetDbtbSeriesFunc *DbtbSeriesStoreGetDbtbSeriesFunc
	// GetScopedSebrchSeriesNeedBbckfillFunc is bn instbnce of b mock
	// function object controlling the behbvior of the method
	// GetScopedSebrchSeriesNeedBbckfill.
	GetScopedSebrchSeriesNeedBbckfillFunc *DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFunc
	// IncrementBbckfillAttemptsFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// IncrementBbckfillAttempts.
	IncrementBbckfillAttemptsFunc *DbtbSeriesStoreIncrementBbckfillAttemptsFunc
	// SetSeriesEnbbledFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method SetSeriesEnbbled.
	SetSeriesEnbbledFunc *DbtbSeriesStoreSetSeriesEnbbledFunc
	// StbmpBbckfillFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method StbmpBbckfill.
	StbmpBbckfillFunc *DbtbSeriesStoreStbmpBbckfillFunc
	// StbmpRecordingFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method StbmpRecording.
	StbmpRecordingFunc *DbtbSeriesStoreStbmpRecordingFunc
	// StbmpSnbpshotFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method StbmpSnbpshot.
	StbmpSnbpshotFunc *DbtbSeriesStoreStbmpSnbpshotFunc
	// StbrtJustInTimeConversionAttemptFunc is bn instbnce of b mock
	// function object controlling the behbvior of the method
	// StbrtJustInTimeConversionAttempt.
	StbrtJustInTimeConversionAttemptFunc *DbtbSeriesStoreStbrtJustInTimeConversionAttemptFunc
}

// NewMockDbtbSeriesStore crebtes b new mock of the DbtbSeriesStore
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockDbtbSeriesStore() *MockDbtbSeriesStore {
	return &MockDbtbSeriesStore{
		CompleteJustInTimeConversionAttemptFunc: &DbtbSeriesStoreCompleteJustInTimeConversionAttemptFunc{
			defbultHook: func(context.Context, types.InsightSeries) (r0 error) {
				return
			},
		},
		GetDbtbSeriesFunc: &DbtbSeriesStoreGetDbtbSeriesFunc{
			defbultHook: func(context.Context, GetDbtbSeriesArgs) (r0 []types.InsightSeries, r1 error) {
				return
			},
		},
		GetScopedSebrchSeriesNeedBbckfillFunc: &DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFunc{
			defbultHook: func(context.Context) (r0 []types.InsightSeries, r1 error) {
				return
			},
		},
		IncrementBbckfillAttemptsFunc: &DbtbSeriesStoreIncrementBbckfillAttemptsFunc{
			defbultHook: func(context.Context, types.InsightSeries) (r0 error) {
				return
			},
		},
		SetSeriesEnbbledFunc: &DbtbSeriesStoreSetSeriesEnbbledFunc{
			defbultHook: func(context.Context, string, bool) (r0 error) {
				return
			},
		},
		StbmpBbckfillFunc: &DbtbSeriesStoreStbmpBbckfillFunc{
			defbultHook: func(context.Context, types.InsightSeries) (r0 types.InsightSeries, r1 error) {
				return
			},
		},
		StbmpRecordingFunc: &DbtbSeriesStoreStbmpRecordingFunc{
			defbultHook: func(context.Context, types.InsightSeries) (r0 types.InsightSeries, r1 error) {
				return
			},
		},
		StbmpSnbpshotFunc: &DbtbSeriesStoreStbmpSnbpshotFunc{
			defbultHook: func(context.Context, types.InsightSeries) (r0 types.InsightSeries, r1 error) {
				return
			},
		},
		StbrtJustInTimeConversionAttemptFunc: &DbtbSeriesStoreStbrtJustInTimeConversionAttemptFunc{
			defbultHook: func(context.Context, types.InsightSeries) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockDbtbSeriesStore crebtes b new mock of the DbtbSeriesStore
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockDbtbSeriesStore() *MockDbtbSeriesStore {
	return &MockDbtbSeriesStore{
		CompleteJustInTimeConversionAttemptFunc: &DbtbSeriesStoreCompleteJustInTimeConversionAttemptFunc{
			defbultHook: func(context.Context, types.InsightSeries) error {
				pbnic("unexpected invocbtion of MockDbtbSeriesStore.CompleteJustInTimeConversionAttempt")
			},
		},
		GetDbtbSeriesFunc: &DbtbSeriesStoreGetDbtbSeriesFunc{
			defbultHook: func(context.Context, GetDbtbSeriesArgs) ([]types.InsightSeries, error) {
				pbnic("unexpected invocbtion of MockDbtbSeriesStore.GetDbtbSeries")
			},
		},
		GetScopedSebrchSeriesNeedBbckfillFunc: &DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFunc{
			defbultHook: func(context.Context) ([]types.InsightSeries, error) {
				pbnic("unexpected invocbtion of MockDbtbSeriesStore.GetScopedSebrchSeriesNeedBbckfill")
			},
		},
		IncrementBbckfillAttemptsFunc: &DbtbSeriesStoreIncrementBbckfillAttemptsFunc{
			defbultHook: func(context.Context, types.InsightSeries) error {
				pbnic("unexpected invocbtion of MockDbtbSeriesStore.IncrementBbckfillAttempts")
			},
		},
		SetSeriesEnbbledFunc: &DbtbSeriesStoreSetSeriesEnbbledFunc{
			defbultHook: func(context.Context, string, bool) error {
				pbnic("unexpected invocbtion of MockDbtbSeriesStore.SetSeriesEnbbled")
			},
		},
		StbmpBbckfillFunc: &DbtbSeriesStoreStbmpBbckfillFunc{
			defbultHook: func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
				pbnic("unexpected invocbtion of MockDbtbSeriesStore.StbmpBbckfill")
			},
		},
		StbmpRecordingFunc: &DbtbSeriesStoreStbmpRecordingFunc{
			defbultHook: func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
				pbnic("unexpected invocbtion of MockDbtbSeriesStore.StbmpRecording")
			},
		},
		StbmpSnbpshotFunc: &DbtbSeriesStoreStbmpSnbpshotFunc{
			defbultHook: func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
				pbnic("unexpected invocbtion of MockDbtbSeriesStore.StbmpSnbpshot")
			},
		},
		StbrtJustInTimeConversionAttemptFunc: &DbtbSeriesStoreStbrtJustInTimeConversionAttemptFunc{
			defbultHook: func(context.Context, types.InsightSeries) error {
				pbnic("unexpected invocbtion of MockDbtbSeriesStore.StbrtJustInTimeConversionAttempt")
			},
		},
	}
}

// NewMockDbtbSeriesStoreFrom crebtes b new mock of the MockDbtbSeriesStore
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockDbtbSeriesStoreFrom(i DbtbSeriesStore) *MockDbtbSeriesStore {
	return &MockDbtbSeriesStore{
		CompleteJustInTimeConversionAttemptFunc: &DbtbSeriesStoreCompleteJustInTimeConversionAttemptFunc{
			defbultHook: i.CompleteJustInTimeConversionAttempt,
		},
		GetDbtbSeriesFunc: &DbtbSeriesStoreGetDbtbSeriesFunc{
			defbultHook: i.GetDbtbSeries,
		},
		GetScopedSebrchSeriesNeedBbckfillFunc: &DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFunc{
			defbultHook: i.GetScopedSebrchSeriesNeedBbckfill,
		},
		IncrementBbckfillAttemptsFunc: &DbtbSeriesStoreIncrementBbckfillAttemptsFunc{
			defbultHook: i.IncrementBbckfillAttempts,
		},
		SetSeriesEnbbledFunc: &DbtbSeriesStoreSetSeriesEnbbledFunc{
			defbultHook: i.SetSeriesEnbbled,
		},
		StbmpBbckfillFunc: &DbtbSeriesStoreStbmpBbckfillFunc{
			defbultHook: i.StbmpBbckfill,
		},
		StbmpRecordingFunc: &DbtbSeriesStoreStbmpRecordingFunc{
			defbultHook: i.StbmpRecording,
		},
		StbmpSnbpshotFunc: &DbtbSeriesStoreStbmpSnbpshotFunc{
			defbultHook: i.StbmpSnbpshot,
		},
		StbrtJustInTimeConversionAttemptFunc: &DbtbSeriesStoreStbrtJustInTimeConversionAttemptFunc{
			defbultHook: i.StbrtJustInTimeConversionAttempt,
		},
	}
}

// DbtbSeriesStoreCompleteJustInTimeConversionAttemptFunc describes the
// behbvior when the CompleteJustInTimeConversionAttempt method of the
// pbrent MockDbtbSeriesStore instbnce is invoked.
type DbtbSeriesStoreCompleteJustInTimeConversionAttemptFunc struct {
	defbultHook func(context.Context, types.InsightSeries) error
	hooks       []func(context.Context, types.InsightSeries) error
	history     []DbtbSeriesStoreCompleteJustInTimeConversionAttemptFuncCbll
	mutex       sync.Mutex
}

// CompleteJustInTimeConversionAttempt delegbtes to the next hook function
// in the queue bnd stores the pbrbmeter bnd result vblues of this
// invocbtion.
func (m *MockDbtbSeriesStore) CompleteJustInTimeConversionAttempt(v0 context.Context, v1 types.InsightSeries) error {
	r0 := m.CompleteJustInTimeConversionAttemptFunc.nextHook()(v0, v1)
	m.CompleteJustInTimeConversionAttemptFunc.bppendCbll(DbtbSeriesStoreCompleteJustInTimeConversionAttemptFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// CompleteJustInTimeConversionAttempt method of the pbrent
// MockDbtbSeriesStore instbnce is invoked bnd the hook queue is empty.
func (f *DbtbSeriesStoreCompleteJustInTimeConversionAttemptFunc) SetDefbultHook(hook func(context.Context, types.InsightSeries) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CompleteJustInTimeConversionAttempt method of the pbrent
// MockDbtbSeriesStore instbnce invokes the hook bt the front of the queue
// bnd discbrds it. After the queue is empty, the defbult hook function is
// invoked for bny future bction.
func (f *DbtbSeriesStoreCompleteJustInTimeConversionAttemptFunc) PushHook(hook func(context.Context, types.InsightSeries) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DbtbSeriesStoreCompleteJustInTimeConversionAttemptFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, types.InsightSeries) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DbtbSeriesStoreCompleteJustInTimeConversionAttemptFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, types.InsightSeries) error {
		return r0
	})
}

func (f *DbtbSeriesStoreCompleteJustInTimeConversionAttemptFunc) nextHook() func(context.Context, types.InsightSeries) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DbtbSeriesStoreCompleteJustInTimeConversionAttemptFunc) bppendCbll(r0 DbtbSeriesStoreCompleteJustInTimeConversionAttemptFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// DbtbSeriesStoreCompleteJustInTimeConversionAttemptFuncCbll objects
// describing the invocbtions of this function.
func (f *DbtbSeriesStoreCompleteJustInTimeConversionAttemptFunc) History() []DbtbSeriesStoreCompleteJustInTimeConversionAttemptFuncCbll {
	f.mutex.Lock()
	history := mbke([]DbtbSeriesStoreCompleteJustInTimeConversionAttemptFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DbtbSeriesStoreCompleteJustInTimeConversionAttemptFuncCbll is bn object
// thbt describes bn invocbtion of method
// CompleteJustInTimeConversionAttempt on bn instbnce of
// MockDbtbSeriesStore.
type DbtbSeriesStoreCompleteJustInTimeConversionAttemptFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 types.InsightSeries
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DbtbSeriesStoreCompleteJustInTimeConversionAttemptFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DbtbSeriesStoreCompleteJustInTimeConversionAttemptFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// DbtbSeriesStoreGetDbtbSeriesFunc describes the behbvior when the
// GetDbtbSeries method of the pbrent MockDbtbSeriesStore instbnce is
// invoked.
type DbtbSeriesStoreGetDbtbSeriesFunc struct {
	defbultHook func(context.Context, GetDbtbSeriesArgs) ([]types.InsightSeries, error)
	hooks       []func(context.Context, GetDbtbSeriesArgs) ([]types.InsightSeries, error)
	history     []DbtbSeriesStoreGetDbtbSeriesFuncCbll
	mutex       sync.Mutex
}

// GetDbtbSeries delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDbtbSeriesStore) GetDbtbSeries(v0 context.Context, v1 GetDbtbSeriesArgs) ([]types.InsightSeries, error) {
	r0, r1 := m.GetDbtbSeriesFunc.nextHook()(v0, v1)
	m.GetDbtbSeriesFunc.bppendCbll(DbtbSeriesStoreGetDbtbSeriesFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetDbtbSeries method
// of the pbrent MockDbtbSeriesStore instbnce is invoked bnd the hook queue
// is empty.
func (f *DbtbSeriesStoreGetDbtbSeriesFunc) SetDefbultHook(hook func(context.Context, GetDbtbSeriesArgs) ([]types.InsightSeries, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetDbtbSeries method of the pbrent MockDbtbSeriesStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *DbtbSeriesStoreGetDbtbSeriesFunc) PushHook(hook func(context.Context, GetDbtbSeriesArgs) ([]types.InsightSeries, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DbtbSeriesStoreGetDbtbSeriesFunc) SetDefbultReturn(r0 []types.InsightSeries, r1 error) {
	f.SetDefbultHook(func(context.Context, GetDbtbSeriesArgs) ([]types.InsightSeries, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DbtbSeriesStoreGetDbtbSeriesFunc) PushReturn(r0 []types.InsightSeries, r1 error) {
	f.PushHook(func(context.Context, GetDbtbSeriesArgs) ([]types.InsightSeries, error) {
		return r0, r1
	})
}

func (f *DbtbSeriesStoreGetDbtbSeriesFunc) nextHook() func(context.Context, GetDbtbSeriesArgs) ([]types.InsightSeries, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DbtbSeriesStoreGetDbtbSeriesFunc) bppendCbll(r0 DbtbSeriesStoreGetDbtbSeriesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of DbtbSeriesStoreGetDbtbSeriesFuncCbll
// objects describing the invocbtions of this function.
func (f *DbtbSeriesStoreGetDbtbSeriesFunc) History() []DbtbSeriesStoreGetDbtbSeriesFuncCbll {
	f.mutex.Lock()
	history := mbke([]DbtbSeriesStoreGetDbtbSeriesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DbtbSeriesStoreGetDbtbSeriesFuncCbll is bn object thbt describes bn
// invocbtion of method GetDbtbSeries on bn instbnce of MockDbtbSeriesStore.
type DbtbSeriesStoreGetDbtbSeriesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 GetDbtbSeriesArgs
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []types.InsightSeries
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DbtbSeriesStoreGetDbtbSeriesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DbtbSeriesStoreGetDbtbSeriesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFunc describes the
// behbvior when the GetScopedSebrchSeriesNeedBbckfill method of the pbrent
// MockDbtbSeriesStore instbnce is invoked.
type DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFunc struct {
	defbultHook func(context.Context) ([]types.InsightSeries, error)
	hooks       []func(context.Context) ([]types.InsightSeries, error)
	history     []DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFuncCbll
	mutex       sync.Mutex
}

// GetScopedSebrchSeriesNeedBbckfill delegbtes to the next hook function in
// the queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDbtbSeriesStore) GetScopedSebrchSeriesNeedBbckfill(v0 context.Context) ([]types.InsightSeries, error) {
	r0, r1 := m.GetScopedSebrchSeriesNeedBbckfillFunc.nextHook()(v0)
	m.GetScopedSebrchSeriesNeedBbckfillFunc.bppendCbll(DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetScopedSebrchSeriesNeedBbckfill method of the pbrent
// MockDbtbSeriesStore instbnce is invoked bnd the hook queue is empty.
func (f *DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFunc) SetDefbultHook(hook func(context.Context) ([]types.InsightSeries, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetScopedSebrchSeriesNeedBbckfill method of the pbrent
// MockDbtbSeriesStore instbnce invokes the hook bt the front of the queue
// bnd discbrds it. After the queue is empty, the defbult hook function is
// invoked for bny future bction.
func (f *DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFunc) PushHook(hook func(context.Context) ([]types.InsightSeries, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFunc) SetDefbultReturn(r0 []types.InsightSeries, r1 error) {
	f.SetDefbultHook(func(context.Context) ([]types.InsightSeries, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFunc) PushReturn(r0 []types.InsightSeries, r1 error) {
	f.PushHook(func(context.Context) ([]types.InsightSeries, error) {
		return r0, r1
	})
}

func (f *DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFunc) nextHook() func(context.Context) ([]types.InsightSeries, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFunc) bppendCbll(r0 DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFuncCbll objects
// describing the invocbtions of this function.
func (f *DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFunc) History() []DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFuncCbll {
	f.mutex.Lock()
	history := mbke([]DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFuncCbll is bn object
// thbt describes bn invocbtion of method GetScopedSebrchSeriesNeedBbckfill
// on bn instbnce of MockDbtbSeriesStore.
type DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []types.InsightSeries
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DbtbSeriesStoreGetScopedSebrchSeriesNeedBbckfillFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// DbtbSeriesStoreIncrementBbckfillAttemptsFunc describes the behbvior when
// the IncrementBbckfillAttempts method of the pbrent MockDbtbSeriesStore
// instbnce is invoked.
type DbtbSeriesStoreIncrementBbckfillAttemptsFunc struct {
	defbultHook func(context.Context, types.InsightSeries) error
	hooks       []func(context.Context, types.InsightSeries) error
	history     []DbtbSeriesStoreIncrementBbckfillAttemptsFuncCbll
	mutex       sync.Mutex
}

// IncrementBbckfillAttempts delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDbtbSeriesStore) IncrementBbckfillAttempts(v0 context.Context, v1 types.InsightSeries) error {
	r0 := m.IncrementBbckfillAttemptsFunc.nextHook()(v0, v1)
	m.IncrementBbckfillAttemptsFunc.bppendCbll(DbtbSeriesStoreIncrementBbckfillAttemptsFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// IncrementBbckfillAttempts method of the pbrent MockDbtbSeriesStore
// instbnce is invoked bnd the hook queue is empty.
func (f *DbtbSeriesStoreIncrementBbckfillAttemptsFunc) SetDefbultHook(hook func(context.Context, types.InsightSeries) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// IncrementBbckfillAttempts method of the pbrent MockDbtbSeriesStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *DbtbSeriesStoreIncrementBbckfillAttemptsFunc) PushHook(hook func(context.Context, types.InsightSeries) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DbtbSeriesStoreIncrementBbckfillAttemptsFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, types.InsightSeries) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DbtbSeriesStoreIncrementBbckfillAttemptsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, types.InsightSeries) error {
		return r0
	})
}

func (f *DbtbSeriesStoreIncrementBbckfillAttemptsFunc) nextHook() func(context.Context, types.InsightSeries) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DbtbSeriesStoreIncrementBbckfillAttemptsFunc) bppendCbll(r0 DbtbSeriesStoreIncrementBbckfillAttemptsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// DbtbSeriesStoreIncrementBbckfillAttemptsFuncCbll objects describing the
// invocbtions of this function.
func (f *DbtbSeriesStoreIncrementBbckfillAttemptsFunc) History() []DbtbSeriesStoreIncrementBbckfillAttemptsFuncCbll {
	f.mutex.Lock()
	history := mbke([]DbtbSeriesStoreIncrementBbckfillAttemptsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DbtbSeriesStoreIncrementBbckfillAttemptsFuncCbll is bn object thbt
// describes bn invocbtion of method IncrementBbckfillAttempts on bn
// instbnce of MockDbtbSeriesStore.
type DbtbSeriesStoreIncrementBbckfillAttemptsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 types.InsightSeries
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DbtbSeriesStoreIncrementBbckfillAttemptsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DbtbSeriesStoreIncrementBbckfillAttemptsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// DbtbSeriesStoreSetSeriesEnbbledFunc describes the behbvior when the
// SetSeriesEnbbled method of the pbrent MockDbtbSeriesStore instbnce is
// invoked.
type DbtbSeriesStoreSetSeriesEnbbledFunc struct {
	defbultHook func(context.Context, string, bool) error
	hooks       []func(context.Context, string, bool) error
	history     []DbtbSeriesStoreSetSeriesEnbbledFuncCbll
	mutex       sync.Mutex
}

// SetSeriesEnbbled delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDbtbSeriesStore) SetSeriesEnbbled(v0 context.Context, v1 string, v2 bool) error {
	r0 := m.SetSeriesEnbbledFunc.nextHook()(v0, v1, v2)
	m.SetSeriesEnbbledFunc.bppendCbll(DbtbSeriesStoreSetSeriesEnbbledFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the SetSeriesEnbbled
// method of the pbrent MockDbtbSeriesStore instbnce is invoked bnd the hook
// queue is empty.
func (f *DbtbSeriesStoreSetSeriesEnbbledFunc) SetDefbultHook(hook func(context.Context, string, bool) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SetSeriesEnbbled method of the pbrent MockDbtbSeriesStore instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *DbtbSeriesStoreSetSeriesEnbbledFunc) PushHook(hook func(context.Context, string, bool) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DbtbSeriesStoreSetSeriesEnbbledFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string, bool) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DbtbSeriesStoreSetSeriesEnbbledFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string, bool) error {
		return r0
	})
}

func (f *DbtbSeriesStoreSetSeriesEnbbledFunc) nextHook() func(context.Context, string, bool) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DbtbSeriesStoreSetSeriesEnbbledFunc) bppendCbll(r0 DbtbSeriesStoreSetSeriesEnbbledFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of DbtbSeriesStoreSetSeriesEnbbledFuncCbll
// objects describing the invocbtions of this function.
func (f *DbtbSeriesStoreSetSeriesEnbbledFunc) History() []DbtbSeriesStoreSetSeriesEnbbledFuncCbll {
	f.mutex.Lock()
	history := mbke([]DbtbSeriesStoreSetSeriesEnbbledFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DbtbSeriesStoreSetSeriesEnbbledFuncCbll is bn object thbt describes bn
// invocbtion of method SetSeriesEnbbled on bn instbnce of
// MockDbtbSeriesStore.
type DbtbSeriesStoreSetSeriesEnbbledFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bool
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DbtbSeriesStoreSetSeriesEnbbledFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DbtbSeriesStoreSetSeriesEnbbledFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// DbtbSeriesStoreStbmpBbckfillFunc describes the behbvior when the
// StbmpBbckfill method of the pbrent MockDbtbSeriesStore instbnce is
// invoked.
type DbtbSeriesStoreStbmpBbckfillFunc struct {
	defbultHook func(context.Context, types.InsightSeries) (types.InsightSeries, error)
	hooks       []func(context.Context, types.InsightSeries) (types.InsightSeries, error)
	history     []DbtbSeriesStoreStbmpBbckfillFuncCbll
	mutex       sync.Mutex
}

// StbmpBbckfill delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDbtbSeriesStore) StbmpBbckfill(v0 context.Context, v1 types.InsightSeries) (types.InsightSeries, error) {
	r0, r1 := m.StbmpBbckfillFunc.nextHook()(v0, v1)
	m.StbmpBbckfillFunc.bppendCbll(DbtbSeriesStoreStbmpBbckfillFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the StbmpBbckfill method
// of the pbrent MockDbtbSeriesStore instbnce is invoked bnd the hook queue
// is empty.
func (f *DbtbSeriesStoreStbmpBbckfillFunc) SetDefbultHook(hook func(context.Context, types.InsightSeries) (types.InsightSeries, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// StbmpBbckfill method of the pbrent MockDbtbSeriesStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *DbtbSeriesStoreStbmpBbckfillFunc) PushHook(hook func(context.Context, types.InsightSeries) (types.InsightSeries, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DbtbSeriesStoreStbmpBbckfillFunc) SetDefbultReturn(r0 types.InsightSeries, r1 error) {
	f.SetDefbultHook(func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DbtbSeriesStoreStbmpBbckfillFunc) PushReturn(r0 types.InsightSeries, r1 error) {
	f.PushHook(func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
		return r0, r1
	})
}

func (f *DbtbSeriesStoreStbmpBbckfillFunc) nextHook() func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DbtbSeriesStoreStbmpBbckfillFunc) bppendCbll(r0 DbtbSeriesStoreStbmpBbckfillFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of DbtbSeriesStoreStbmpBbckfillFuncCbll
// objects describing the invocbtions of this function.
func (f *DbtbSeriesStoreStbmpBbckfillFunc) History() []DbtbSeriesStoreStbmpBbckfillFuncCbll {
	f.mutex.Lock()
	history := mbke([]DbtbSeriesStoreStbmpBbckfillFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DbtbSeriesStoreStbmpBbckfillFuncCbll is bn object thbt describes bn
// invocbtion of method StbmpBbckfill on bn instbnce of MockDbtbSeriesStore.
type DbtbSeriesStoreStbmpBbckfillFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 types.InsightSeries
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 types.InsightSeries
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DbtbSeriesStoreStbmpBbckfillFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DbtbSeriesStoreStbmpBbckfillFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// DbtbSeriesStoreStbmpRecordingFunc describes the behbvior when the
// StbmpRecording method of the pbrent MockDbtbSeriesStore instbnce is
// invoked.
type DbtbSeriesStoreStbmpRecordingFunc struct {
	defbultHook func(context.Context, types.InsightSeries) (types.InsightSeries, error)
	hooks       []func(context.Context, types.InsightSeries) (types.InsightSeries, error)
	history     []DbtbSeriesStoreStbmpRecordingFuncCbll
	mutex       sync.Mutex
}

// StbmpRecording delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDbtbSeriesStore) StbmpRecording(v0 context.Context, v1 types.InsightSeries) (types.InsightSeries, error) {
	r0, r1 := m.StbmpRecordingFunc.nextHook()(v0, v1)
	m.StbmpRecordingFunc.bppendCbll(DbtbSeriesStoreStbmpRecordingFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the StbmpRecording
// method of the pbrent MockDbtbSeriesStore instbnce is invoked bnd the hook
// queue is empty.
func (f *DbtbSeriesStoreStbmpRecordingFunc) SetDefbultHook(hook func(context.Context, types.InsightSeries) (types.InsightSeries, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// StbmpRecording method of the pbrent MockDbtbSeriesStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *DbtbSeriesStoreStbmpRecordingFunc) PushHook(hook func(context.Context, types.InsightSeries) (types.InsightSeries, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DbtbSeriesStoreStbmpRecordingFunc) SetDefbultReturn(r0 types.InsightSeries, r1 error) {
	f.SetDefbultHook(func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DbtbSeriesStoreStbmpRecordingFunc) PushReturn(r0 types.InsightSeries, r1 error) {
	f.PushHook(func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
		return r0, r1
	})
}

func (f *DbtbSeriesStoreStbmpRecordingFunc) nextHook() func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DbtbSeriesStoreStbmpRecordingFunc) bppendCbll(r0 DbtbSeriesStoreStbmpRecordingFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of DbtbSeriesStoreStbmpRecordingFuncCbll
// objects describing the invocbtions of this function.
func (f *DbtbSeriesStoreStbmpRecordingFunc) History() []DbtbSeriesStoreStbmpRecordingFuncCbll {
	f.mutex.Lock()
	history := mbke([]DbtbSeriesStoreStbmpRecordingFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DbtbSeriesStoreStbmpRecordingFuncCbll is bn object thbt describes bn
// invocbtion of method StbmpRecording on bn instbnce of
// MockDbtbSeriesStore.
type DbtbSeriesStoreStbmpRecordingFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 types.InsightSeries
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 types.InsightSeries
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DbtbSeriesStoreStbmpRecordingFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DbtbSeriesStoreStbmpRecordingFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// DbtbSeriesStoreStbmpSnbpshotFunc describes the behbvior when the
// StbmpSnbpshot method of the pbrent MockDbtbSeriesStore instbnce is
// invoked.
type DbtbSeriesStoreStbmpSnbpshotFunc struct {
	defbultHook func(context.Context, types.InsightSeries) (types.InsightSeries, error)
	hooks       []func(context.Context, types.InsightSeries) (types.InsightSeries, error)
	history     []DbtbSeriesStoreStbmpSnbpshotFuncCbll
	mutex       sync.Mutex
}

// StbmpSnbpshot delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDbtbSeriesStore) StbmpSnbpshot(v0 context.Context, v1 types.InsightSeries) (types.InsightSeries, error) {
	r0, r1 := m.StbmpSnbpshotFunc.nextHook()(v0, v1)
	m.StbmpSnbpshotFunc.bppendCbll(DbtbSeriesStoreStbmpSnbpshotFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the StbmpSnbpshot method
// of the pbrent MockDbtbSeriesStore instbnce is invoked bnd the hook queue
// is empty.
func (f *DbtbSeriesStoreStbmpSnbpshotFunc) SetDefbultHook(hook func(context.Context, types.InsightSeries) (types.InsightSeries, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// StbmpSnbpshot method of the pbrent MockDbtbSeriesStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *DbtbSeriesStoreStbmpSnbpshotFunc) PushHook(hook func(context.Context, types.InsightSeries) (types.InsightSeries, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DbtbSeriesStoreStbmpSnbpshotFunc) SetDefbultReturn(r0 types.InsightSeries, r1 error) {
	f.SetDefbultHook(func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DbtbSeriesStoreStbmpSnbpshotFunc) PushReturn(r0 types.InsightSeries, r1 error) {
	f.PushHook(func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
		return r0, r1
	})
}

func (f *DbtbSeriesStoreStbmpSnbpshotFunc) nextHook() func(context.Context, types.InsightSeries) (types.InsightSeries, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DbtbSeriesStoreStbmpSnbpshotFunc) bppendCbll(r0 DbtbSeriesStoreStbmpSnbpshotFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of DbtbSeriesStoreStbmpSnbpshotFuncCbll
// objects describing the invocbtions of this function.
func (f *DbtbSeriesStoreStbmpSnbpshotFunc) History() []DbtbSeriesStoreStbmpSnbpshotFuncCbll {
	f.mutex.Lock()
	history := mbke([]DbtbSeriesStoreStbmpSnbpshotFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DbtbSeriesStoreStbmpSnbpshotFuncCbll is bn object thbt describes bn
// invocbtion of method StbmpSnbpshot on bn instbnce of MockDbtbSeriesStore.
type DbtbSeriesStoreStbmpSnbpshotFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 types.InsightSeries
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 types.InsightSeries
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DbtbSeriesStoreStbmpSnbpshotFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DbtbSeriesStoreStbmpSnbpshotFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// DbtbSeriesStoreStbrtJustInTimeConversionAttemptFunc describes the
// behbvior when the StbrtJustInTimeConversionAttempt method of the pbrent
// MockDbtbSeriesStore instbnce is invoked.
type DbtbSeriesStoreStbrtJustInTimeConversionAttemptFunc struct {
	defbultHook func(context.Context, types.InsightSeries) error
	hooks       []func(context.Context, types.InsightSeries) error
	history     []DbtbSeriesStoreStbrtJustInTimeConversionAttemptFuncCbll
	mutex       sync.Mutex
}

// StbrtJustInTimeConversionAttempt delegbtes to the next hook function in
// the queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockDbtbSeriesStore) StbrtJustInTimeConversionAttempt(v0 context.Context, v1 types.InsightSeries) error {
	r0 := m.StbrtJustInTimeConversionAttemptFunc.nextHook()(v0, v1)
	m.StbrtJustInTimeConversionAttemptFunc.bppendCbll(DbtbSeriesStoreStbrtJustInTimeConversionAttemptFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// StbrtJustInTimeConversionAttempt method of the pbrent MockDbtbSeriesStore
// instbnce is invoked bnd the hook queue is empty.
func (f *DbtbSeriesStoreStbrtJustInTimeConversionAttemptFunc) SetDefbultHook(hook func(context.Context, types.InsightSeries) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// StbrtJustInTimeConversionAttempt method of the pbrent MockDbtbSeriesStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *DbtbSeriesStoreStbrtJustInTimeConversionAttemptFunc) PushHook(hook func(context.Context, types.InsightSeries) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *DbtbSeriesStoreStbrtJustInTimeConversionAttemptFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, types.InsightSeries) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *DbtbSeriesStoreStbrtJustInTimeConversionAttemptFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, types.InsightSeries) error {
		return r0
	})
}

func (f *DbtbSeriesStoreStbrtJustInTimeConversionAttemptFunc) nextHook() func(context.Context, types.InsightSeries) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *DbtbSeriesStoreStbrtJustInTimeConversionAttemptFunc) bppendCbll(r0 DbtbSeriesStoreStbrtJustInTimeConversionAttemptFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// DbtbSeriesStoreStbrtJustInTimeConversionAttemptFuncCbll objects
// describing the invocbtions of this function.
func (f *DbtbSeriesStoreStbrtJustInTimeConversionAttemptFunc) History() []DbtbSeriesStoreStbrtJustInTimeConversionAttemptFuncCbll {
	f.mutex.Lock()
	history := mbke([]DbtbSeriesStoreStbrtJustInTimeConversionAttemptFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// DbtbSeriesStoreStbrtJustInTimeConversionAttemptFuncCbll is bn object thbt
// describes bn invocbtion of method StbrtJustInTimeConversionAttempt on bn
// instbnce of MockDbtbSeriesStore.
type DbtbSeriesStoreStbrtJustInTimeConversionAttemptFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 types.InsightSeries
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c DbtbSeriesStoreStbrtJustInTimeConversionAttemptFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c DbtbSeriesStoreStbrtJustInTimeConversionAttemptFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockInsightMetbdbtbStore is b mock implementbtion of the
// InsightMetbdbtbStore interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/insights/store) used for unit
// testing.
type MockInsightMetbdbtbStore struct {
	// GetMbppedFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetMbpped.
	GetMbppedFunc *InsightMetbdbtbStoreGetMbppedFunc
}

// NewMockInsightMetbdbtbStore crebtes b new mock of the
// InsightMetbdbtbStore interfbce. All methods return zero vblues for bll
// results, unless overwritten.
func NewMockInsightMetbdbtbStore() *MockInsightMetbdbtbStore {
	return &MockInsightMetbdbtbStore{
		GetMbppedFunc: &InsightMetbdbtbStoreGetMbppedFunc{
			defbultHook: func(context.Context, InsightQueryArgs) (r0 []types.Insight, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockInsightMetbdbtbStore crebtes b new mock of the
// InsightMetbdbtbStore interfbce. All methods pbnic on invocbtion, unless
// overwritten.
func NewStrictMockInsightMetbdbtbStore() *MockInsightMetbdbtbStore {
	return &MockInsightMetbdbtbStore{
		GetMbppedFunc: &InsightMetbdbtbStoreGetMbppedFunc{
			defbultHook: func(context.Context, InsightQueryArgs) ([]types.Insight, error) {
				pbnic("unexpected invocbtion of MockInsightMetbdbtbStore.GetMbpped")
			},
		},
	}
}

// NewMockInsightMetbdbtbStoreFrom crebtes b new mock of the
// MockInsightMetbdbtbStore interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockInsightMetbdbtbStoreFrom(i InsightMetbdbtbStore) *MockInsightMetbdbtbStore {
	return &MockInsightMetbdbtbStore{
		GetMbppedFunc: &InsightMetbdbtbStoreGetMbppedFunc{
			defbultHook: i.GetMbpped,
		},
	}
}

// InsightMetbdbtbStoreGetMbppedFunc describes the behbvior when the
// GetMbpped method of the pbrent MockInsightMetbdbtbStore instbnce is
// invoked.
type InsightMetbdbtbStoreGetMbppedFunc struct {
	defbultHook func(context.Context, InsightQueryArgs) ([]types.Insight, error)
	hooks       []func(context.Context, InsightQueryArgs) ([]types.Insight, error)
	history     []InsightMetbdbtbStoreGetMbppedFuncCbll
	mutex       sync.Mutex
}

// GetMbpped delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockInsightMetbdbtbStore) GetMbpped(v0 context.Context, v1 InsightQueryArgs) ([]types.Insight, error) {
	r0, r1 := m.GetMbppedFunc.nextHook()(v0, v1)
	m.GetMbppedFunc.bppendCbll(InsightMetbdbtbStoreGetMbppedFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetMbpped method of
// the pbrent MockInsightMetbdbtbStore instbnce is invoked bnd the hook
// queue is empty.
func (f *InsightMetbdbtbStoreGetMbppedFunc) SetDefbultHook(hook func(context.Context, InsightQueryArgs) ([]types.Insight, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetMbpped method of the pbrent MockInsightMetbdbtbStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *InsightMetbdbtbStoreGetMbppedFunc) PushHook(hook func(context.Context, InsightQueryArgs) ([]types.Insight, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *InsightMetbdbtbStoreGetMbppedFunc) SetDefbultReturn(r0 []types.Insight, r1 error) {
	f.SetDefbultHook(func(context.Context, InsightQueryArgs) ([]types.Insight, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *InsightMetbdbtbStoreGetMbppedFunc) PushReturn(r0 []types.Insight, r1 error) {
	f.PushHook(func(context.Context, InsightQueryArgs) ([]types.Insight, error) {
		return r0, r1
	})
}

func (f *InsightMetbdbtbStoreGetMbppedFunc) nextHook() func(context.Context, InsightQueryArgs) ([]types.Insight, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *InsightMetbdbtbStoreGetMbppedFunc) bppendCbll(r0 InsightMetbdbtbStoreGetMbppedFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of InsightMetbdbtbStoreGetMbppedFuncCbll
// objects describing the invocbtions of this function.
func (f *InsightMetbdbtbStoreGetMbppedFunc) History() []InsightMetbdbtbStoreGetMbppedFuncCbll {
	f.mutex.Lock()
	history := mbke([]InsightMetbdbtbStoreGetMbppedFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// InsightMetbdbtbStoreGetMbppedFuncCbll is bn object thbt describes bn
// invocbtion of method GetMbpped on bn instbnce of
// MockInsightMetbdbtbStore.
type InsightMetbdbtbStoreGetMbppedFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 InsightQueryArgs
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []types.Insight
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c InsightMetbdbtbStoreGetMbppedFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c InsightMetbdbtbStoreGetMbppedFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockInterfbce is b mock implementbtion of the Interfbce interfbce (from
// the pbckbge github.com/sourcegrbph/sourcegrbph/internbl/insights/store)
// used for unit testing.
type MockInterfbce struct {
	// AddIncompleteDbtbpointFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method AddIncompleteDbtbpoint.
	AddIncompleteDbtbpointFunc *InterfbceAddIncompleteDbtbpointFunc
	// CountDbtbFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method CountDbtb.
	CountDbtbFunc *InterfbceCountDbtbFunc
	// GetAllDbtbForInsightViewIDFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// GetAllDbtbForInsightViewID.
	GetAllDbtbForInsightViewIDFunc *InterfbceGetAllDbtbForInsightViewIDFunc
	// GetInsightSeriesRecordingTimesFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// GetInsightSeriesRecordingTimes.
	GetInsightSeriesRecordingTimesFunc *InterfbceGetInsightSeriesRecordingTimesFunc
	// LobdAggregbtedIncompleteDbtbpointsFunc is bn instbnce of b mock
	// function object controlling the behbvior of the method
	// LobdAggregbtedIncompleteDbtbpoints.
	LobdAggregbtedIncompleteDbtbpointsFunc *InterfbceLobdAggregbtedIncompleteDbtbpointsFunc
	// RecordSeriesPointsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method RecordSeriesPoints.
	RecordSeriesPointsFunc *InterfbceRecordSeriesPointsFunc
	// RecordSeriesPointsAndRecordingTimesFunc is bn instbnce of b mock
	// function object controlling the behbvior of the method
	// RecordSeriesPointsAndRecordingTimes.
	RecordSeriesPointsAndRecordingTimesFunc *InterfbceRecordSeriesPointsAndRecordingTimesFunc
	// SeriesPointsFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method SeriesPoints.
	SeriesPointsFunc *InterfbceSeriesPointsFunc
	// SetInsightSeriesRecordingTimesFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// SetInsightSeriesRecordingTimes.
	SetInsightSeriesRecordingTimesFunc *InterfbceSetInsightSeriesRecordingTimesFunc
	// WithOtherFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method WithOther.
	WithOtherFunc *InterfbceWithOtherFunc
}

// NewMockInterfbce crebtes b new mock of the Interfbce interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockInterfbce() *MockInterfbce {
	return &MockInterfbce{
		AddIncompleteDbtbpointFunc: &InterfbceAddIncompleteDbtbpointFunc{
			defbultHook: func(context.Context, AddIncompleteDbtbpointInput) (r0 error) {
				return
			},
		},
		CountDbtbFunc: &InterfbceCountDbtbFunc{
			defbultHook: func(context.Context, CountDbtbOpts) (r0 int, r1 error) {
				return
			},
		},
		GetAllDbtbForInsightViewIDFunc: &InterfbceGetAllDbtbForInsightViewIDFunc{
			defbultHook: func(context.Context, ExportOpts) (r0 []SeriesPointForExport, r1 error) {
				return
			},
		},
		GetInsightSeriesRecordingTimesFunc: &InterfbceGetInsightSeriesRecordingTimesFunc{
			defbultHook: func(context.Context, int, SeriesPointsOpts) (r0 types.InsightSeriesRecordingTimes, r1 error) {
				return
			},
		},
		LobdAggregbtedIncompleteDbtbpointsFunc: &InterfbceLobdAggregbtedIncompleteDbtbpointsFunc{
			defbultHook: func(context.Context, int) (r0 []IncompleteDbtbpoint, r1 error) {
				return
			},
		},
		RecordSeriesPointsFunc: &InterfbceRecordSeriesPointsFunc{
			defbultHook: func(context.Context, []RecordSeriesPointArgs) (r0 error) {
				return
			},
		},
		RecordSeriesPointsAndRecordingTimesFunc: &InterfbceRecordSeriesPointsAndRecordingTimesFunc{
			defbultHook: func(context.Context, []RecordSeriesPointArgs, types.InsightSeriesRecordingTimes) (r0 error) {
				return
			},
		},
		SeriesPointsFunc: &InterfbceSeriesPointsFunc{
			defbultHook: func(context.Context, SeriesPointsOpts) (r0 []SeriesPoint, r1 error) {
				return
			},
		},
		SetInsightSeriesRecordingTimesFunc: &InterfbceSetInsightSeriesRecordingTimesFunc{
			defbultHook: func(context.Context, []types.InsightSeriesRecordingTimes) (r0 error) {
				return
			},
		},
		WithOtherFunc: &InterfbceWithOtherFunc{
			defbultHook: func(bbsestore.ShbrebbleStore) (r0 Interfbce) {
				return
			},
		},
	}
}

// NewStrictMockInterfbce crebtes b new mock of the Interfbce interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockInterfbce() *MockInterfbce {
	return &MockInterfbce{
		AddIncompleteDbtbpointFunc: &InterfbceAddIncompleteDbtbpointFunc{
			defbultHook: func(context.Context, AddIncompleteDbtbpointInput) error {
				pbnic("unexpected invocbtion of MockInterfbce.AddIncompleteDbtbpoint")
			},
		},
		CountDbtbFunc: &InterfbceCountDbtbFunc{
			defbultHook: func(context.Context, CountDbtbOpts) (int, error) {
				pbnic("unexpected invocbtion of MockInterfbce.CountDbtb")
			},
		},
		GetAllDbtbForInsightViewIDFunc: &InterfbceGetAllDbtbForInsightViewIDFunc{
			defbultHook: func(context.Context, ExportOpts) ([]SeriesPointForExport, error) {
				pbnic("unexpected invocbtion of MockInterfbce.GetAllDbtbForInsightViewID")
			},
		},
		GetInsightSeriesRecordingTimesFunc: &InterfbceGetInsightSeriesRecordingTimesFunc{
			defbultHook: func(context.Context, int, SeriesPointsOpts) (types.InsightSeriesRecordingTimes, error) {
				pbnic("unexpected invocbtion of MockInterfbce.GetInsightSeriesRecordingTimes")
			},
		},
		LobdAggregbtedIncompleteDbtbpointsFunc: &InterfbceLobdAggregbtedIncompleteDbtbpointsFunc{
			defbultHook: func(context.Context, int) ([]IncompleteDbtbpoint, error) {
				pbnic("unexpected invocbtion of MockInterfbce.LobdAggregbtedIncompleteDbtbpoints")
			},
		},
		RecordSeriesPointsFunc: &InterfbceRecordSeriesPointsFunc{
			defbultHook: func(context.Context, []RecordSeriesPointArgs) error {
				pbnic("unexpected invocbtion of MockInterfbce.RecordSeriesPoints")
			},
		},
		RecordSeriesPointsAndRecordingTimesFunc: &InterfbceRecordSeriesPointsAndRecordingTimesFunc{
			defbultHook: func(context.Context, []RecordSeriesPointArgs, types.InsightSeriesRecordingTimes) error {
				pbnic("unexpected invocbtion of MockInterfbce.RecordSeriesPointsAndRecordingTimes")
			},
		},
		SeriesPointsFunc: &InterfbceSeriesPointsFunc{
			defbultHook: func(context.Context, SeriesPointsOpts) ([]SeriesPoint, error) {
				pbnic("unexpected invocbtion of MockInterfbce.SeriesPoints")
			},
		},
		SetInsightSeriesRecordingTimesFunc: &InterfbceSetInsightSeriesRecordingTimesFunc{
			defbultHook: func(context.Context, []types.InsightSeriesRecordingTimes) error {
				pbnic("unexpected invocbtion of MockInterfbce.SetInsightSeriesRecordingTimes")
			},
		},
		WithOtherFunc: &InterfbceWithOtherFunc{
			defbultHook: func(bbsestore.ShbrebbleStore) Interfbce {
				pbnic("unexpected invocbtion of MockInterfbce.WithOther")
			},
		},
	}
}

// NewMockInterfbceFrom crebtes b new mock of the MockInterfbce interfbce.
// All methods delegbte to the given implementbtion, unless overwritten.
func NewMockInterfbceFrom(i Interfbce) *MockInterfbce {
	return &MockInterfbce{
		AddIncompleteDbtbpointFunc: &InterfbceAddIncompleteDbtbpointFunc{
			defbultHook: i.AddIncompleteDbtbpoint,
		},
		CountDbtbFunc: &InterfbceCountDbtbFunc{
			defbultHook: i.CountDbtb,
		},
		GetAllDbtbForInsightViewIDFunc: &InterfbceGetAllDbtbForInsightViewIDFunc{
			defbultHook: i.GetAllDbtbForInsightViewID,
		},
		GetInsightSeriesRecordingTimesFunc: &InterfbceGetInsightSeriesRecordingTimesFunc{
			defbultHook: i.GetInsightSeriesRecordingTimes,
		},
		LobdAggregbtedIncompleteDbtbpointsFunc: &InterfbceLobdAggregbtedIncompleteDbtbpointsFunc{
			defbultHook: i.LobdAggregbtedIncompleteDbtbpoints,
		},
		RecordSeriesPointsFunc: &InterfbceRecordSeriesPointsFunc{
			defbultHook: i.RecordSeriesPoints,
		},
		RecordSeriesPointsAndRecordingTimesFunc: &InterfbceRecordSeriesPointsAndRecordingTimesFunc{
			defbultHook: i.RecordSeriesPointsAndRecordingTimes,
		},
		SeriesPointsFunc: &InterfbceSeriesPointsFunc{
			defbultHook: i.SeriesPoints,
		},
		SetInsightSeriesRecordingTimesFunc: &InterfbceSetInsightSeriesRecordingTimesFunc{
			defbultHook: i.SetInsightSeriesRecordingTimes,
		},
		WithOtherFunc: &InterfbceWithOtherFunc{
			defbultHook: i.WithOther,
		},
	}
}

// InterfbceAddIncompleteDbtbpointFunc describes the behbvior when the
// AddIncompleteDbtbpoint method of the pbrent MockInterfbce instbnce is
// invoked.
type InterfbceAddIncompleteDbtbpointFunc struct {
	defbultHook func(context.Context, AddIncompleteDbtbpointInput) error
	hooks       []func(context.Context, AddIncompleteDbtbpointInput) error
	history     []InterfbceAddIncompleteDbtbpointFuncCbll
	mutex       sync.Mutex
}

// AddIncompleteDbtbpoint delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockInterfbce) AddIncompleteDbtbpoint(v0 context.Context, v1 AddIncompleteDbtbpointInput) error {
	r0 := m.AddIncompleteDbtbpointFunc.nextHook()(v0, v1)
	m.AddIncompleteDbtbpointFunc.bppendCbll(InterfbceAddIncompleteDbtbpointFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// AddIncompleteDbtbpoint method of the pbrent MockInterfbce instbnce is
// invoked bnd the hook queue is empty.
func (f *InterfbceAddIncompleteDbtbpointFunc) SetDefbultHook(hook func(context.Context, AddIncompleteDbtbpointInput) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// AddIncompleteDbtbpoint method of the pbrent MockInterfbce instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *InterfbceAddIncompleteDbtbpointFunc) PushHook(hook func(context.Context, AddIncompleteDbtbpointInput) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *InterfbceAddIncompleteDbtbpointFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, AddIncompleteDbtbpointInput) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *InterfbceAddIncompleteDbtbpointFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, AddIncompleteDbtbpointInput) error {
		return r0
	})
}

func (f *InterfbceAddIncompleteDbtbpointFunc) nextHook() func(context.Context, AddIncompleteDbtbpointInput) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *InterfbceAddIncompleteDbtbpointFunc) bppendCbll(r0 InterfbceAddIncompleteDbtbpointFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of InterfbceAddIncompleteDbtbpointFuncCbll
// objects describing the invocbtions of this function.
func (f *InterfbceAddIncompleteDbtbpointFunc) History() []InterfbceAddIncompleteDbtbpointFuncCbll {
	f.mutex.Lock()
	history := mbke([]InterfbceAddIncompleteDbtbpointFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// InterfbceAddIncompleteDbtbpointFuncCbll is bn object thbt describes bn
// invocbtion of method AddIncompleteDbtbpoint on bn instbnce of
// MockInterfbce.
type InterfbceAddIncompleteDbtbpointFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 AddIncompleteDbtbpointInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c InterfbceAddIncompleteDbtbpointFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c InterfbceAddIncompleteDbtbpointFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// InterfbceCountDbtbFunc describes the behbvior when the CountDbtb method
// of the pbrent MockInterfbce instbnce is invoked.
type InterfbceCountDbtbFunc struct {
	defbultHook func(context.Context, CountDbtbOpts) (int, error)
	hooks       []func(context.Context, CountDbtbOpts) (int, error)
	history     []InterfbceCountDbtbFuncCbll
	mutex       sync.Mutex
}

// CountDbtb delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockInterfbce) CountDbtb(v0 context.Context, v1 CountDbtbOpts) (int, error) {
	r0, r1 := m.CountDbtbFunc.nextHook()(v0, v1)
	m.CountDbtbFunc.bppendCbll(InterfbceCountDbtbFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CountDbtb method of
// the pbrent MockInterfbce instbnce is invoked bnd the hook queue is empty.
func (f *InterfbceCountDbtbFunc) SetDefbultHook(hook func(context.Context, CountDbtbOpts) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CountDbtb method of the pbrent MockInterfbce instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *InterfbceCountDbtbFunc) PushHook(hook func(context.Context, CountDbtbOpts) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *InterfbceCountDbtbFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context, CountDbtbOpts) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *InterfbceCountDbtbFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, CountDbtbOpts) (int, error) {
		return r0, r1
	})
}

func (f *InterfbceCountDbtbFunc) nextHook() func(context.Context, CountDbtbOpts) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *InterfbceCountDbtbFunc) bppendCbll(r0 InterfbceCountDbtbFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of InterfbceCountDbtbFuncCbll objects
// describing the invocbtions of this function.
func (f *InterfbceCountDbtbFunc) History() []InterfbceCountDbtbFuncCbll {
	f.mutex.Lock()
	history := mbke([]InterfbceCountDbtbFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// InterfbceCountDbtbFuncCbll is bn object thbt describes bn invocbtion of
// method CountDbtb on bn instbnce of MockInterfbce.
type InterfbceCountDbtbFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 CountDbtbOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c InterfbceCountDbtbFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c InterfbceCountDbtbFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// InterfbceGetAllDbtbForInsightViewIDFunc describes the behbvior when the
// GetAllDbtbForInsightViewID method of the pbrent MockInterfbce instbnce is
// invoked.
type InterfbceGetAllDbtbForInsightViewIDFunc struct {
	defbultHook func(context.Context, ExportOpts) ([]SeriesPointForExport, error)
	hooks       []func(context.Context, ExportOpts) ([]SeriesPointForExport, error)
	history     []InterfbceGetAllDbtbForInsightViewIDFuncCbll
	mutex       sync.Mutex
}

// GetAllDbtbForInsightViewID delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockInterfbce) GetAllDbtbForInsightViewID(v0 context.Context, v1 ExportOpts) ([]SeriesPointForExport, error) {
	r0, r1 := m.GetAllDbtbForInsightViewIDFunc.nextHook()(v0, v1)
	m.GetAllDbtbForInsightViewIDFunc.bppendCbll(InterfbceGetAllDbtbForInsightViewIDFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetAllDbtbForInsightViewID method of the pbrent MockInterfbce instbnce is
// invoked bnd the hook queue is empty.
func (f *InterfbceGetAllDbtbForInsightViewIDFunc) SetDefbultHook(hook func(context.Context, ExportOpts) ([]SeriesPointForExport, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetAllDbtbForInsightViewID method of the pbrent MockInterfbce instbnce
// invokes the hook bt the front of the queue bnd discbrds it. After the
// queue is empty, the defbult hook function is invoked for bny future
// bction.
func (f *InterfbceGetAllDbtbForInsightViewIDFunc) PushHook(hook func(context.Context, ExportOpts) ([]SeriesPointForExport, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *InterfbceGetAllDbtbForInsightViewIDFunc) SetDefbultReturn(r0 []SeriesPointForExport, r1 error) {
	f.SetDefbultHook(func(context.Context, ExportOpts) ([]SeriesPointForExport, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *InterfbceGetAllDbtbForInsightViewIDFunc) PushReturn(r0 []SeriesPointForExport, r1 error) {
	f.PushHook(func(context.Context, ExportOpts) ([]SeriesPointForExport, error) {
		return r0, r1
	})
}

func (f *InterfbceGetAllDbtbForInsightViewIDFunc) nextHook() func(context.Context, ExportOpts) ([]SeriesPointForExport, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *InterfbceGetAllDbtbForInsightViewIDFunc) bppendCbll(r0 InterfbceGetAllDbtbForInsightViewIDFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of InterfbceGetAllDbtbForInsightViewIDFuncCbll
// objects describing the invocbtions of this function.
func (f *InterfbceGetAllDbtbForInsightViewIDFunc) History() []InterfbceGetAllDbtbForInsightViewIDFuncCbll {
	f.mutex.Lock()
	history := mbke([]InterfbceGetAllDbtbForInsightViewIDFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// InterfbceGetAllDbtbForInsightViewIDFuncCbll is bn object thbt describes
// bn invocbtion of method GetAllDbtbForInsightViewID on bn instbnce of
// MockInterfbce.
type InterfbceGetAllDbtbForInsightViewIDFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 ExportOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []SeriesPointForExport
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c InterfbceGetAllDbtbForInsightViewIDFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c InterfbceGetAllDbtbForInsightViewIDFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// InterfbceGetInsightSeriesRecordingTimesFunc describes the behbvior when
// the GetInsightSeriesRecordingTimes method of the pbrent MockInterfbce
// instbnce is invoked.
type InterfbceGetInsightSeriesRecordingTimesFunc struct {
	defbultHook func(context.Context, int, SeriesPointsOpts) (types.InsightSeriesRecordingTimes, error)
	hooks       []func(context.Context, int, SeriesPointsOpts) (types.InsightSeriesRecordingTimes, error)
	history     []InterfbceGetInsightSeriesRecordingTimesFuncCbll
	mutex       sync.Mutex
}

// GetInsightSeriesRecordingTimes delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockInterfbce) GetInsightSeriesRecordingTimes(v0 context.Context, v1 int, v2 SeriesPointsOpts) (types.InsightSeriesRecordingTimes, error) {
	r0, r1 := m.GetInsightSeriesRecordingTimesFunc.nextHook()(v0, v1, v2)
	m.GetInsightSeriesRecordingTimesFunc.bppendCbll(InterfbceGetInsightSeriesRecordingTimesFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetInsightSeriesRecordingTimes method of the pbrent MockInterfbce
// instbnce is invoked bnd the hook queue is empty.
func (f *InterfbceGetInsightSeriesRecordingTimesFunc) SetDefbultHook(hook func(context.Context, int, SeriesPointsOpts) (types.InsightSeriesRecordingTimes, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetInsightSeriesRecordingTimes method of the pbrent MockInterfbce
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *InterfbceGetInsightSeriesRecordingTimesFunc) PushHook(hook func(context.Context, int, SeriesPointsOpts) (types.InsightSeriesRecordingTimes, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *InterfbceGetInsightSeriesRecordingTimesFunc) SetDefbultReturn(r0 types.InsightSeriesRecordingTimes, r1 error) {
	f.SetDefbultHook(func(context.Context, int, SeriesPointsOpts) (types.InsightSeriesRecordingTimes, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *InterfbceGetInsightSeriesRecordingTimesFunc) PushReturn(r0 types.InsightSeriesRecordingTimes, r1 error) {
	f.PushHook(func(context.Context, int, SeriesPointsOpts) (types.InsightSeriesRecordingTimes, error) {
		return r0, r1
	})
}

func (f *InterfbceGetInsightSeriesRecordingTimesFunc) nextHook() func(context.Context, int, SeriesPointsOpts) (types.InsightSeriesRecordingTimes, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *InterfbceGetInsightSeriesRecordingTimesFunc) bppendCbll(r0 InterfbceGetInsightSeriesRecordingTimesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// InterfbceGetInsightSeriesRecordingTimesFuncCbll objects describing the
// invocbtions of this function.
func (f *InterfbceGetInsightSeriesRecordingTimesFunc) History() []InterfbceGetInsightSeriesRecordingTimesFuncCbll {
	f.mutex.Lock()
	history := mbke([]InterfbceGetInsightSeriesRecordingTimesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// InterfbceGetInsightSeriesRecordingTimesFuncCbll is bn object thbt
// describes bn invocbtion of method GetInsightSeriesRecordingTimes on bn
// instbnce of MockInterfbce.
type InterfbceGetInsightSeriesRecordingTimesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 SeriesPointsOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 types.InsightSeriesRecordingTimes
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c InterfbceGetInsightSeriesRecordingTimesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c InterfbceGetInsightSeriesRecordingTimesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// InterfbceLobdAggregbtedIncompleteDbtbpointsFunc describes the behbvior
// when the LobdAggregbtedIncompleteDbtbpoints method of the pbrent
// MockInterfbce instbnce is invoked.
type InterfbceLobdAggregbtedIncompleteDbtbpointsFunc struct {
	defbultHook func(context.Context, int) ([]IncompleteDbtbpoint, error)
	hooks       []func(context.Context, int) ([]IncompleteDbtbpoint, error)
	history     []InterfbceLobdAggregbtedIncompleteDbtbpointsFuncCbll
	mutex       sync.Mutex
}

// LobdAggregbtedIncompleteDbtbpoints delegbtes to the next hook function in
// the queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockInterfbce) LobdAggregbtedIncompleteDbtbpoints(v0 context.Context, v1 int) ([]IncompleteDbtbpoint, error) {
	r0, r1 := m.LobdAggregbtedIncompleteDbtbpointsFunc.nextHook()(v0, v1)
	m.LobdAggregbtedIncompleteDbtbpointsFunc.bppendCbll(InterfbceLobdAggregbtedIncompleteDbtbpointsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// LobdAggregbtedIncompleteDbtbpoints method of the pbrent MockInterfbce
// instbnce is invoked bnd the hook queue is empty.
func (f *InterfbceLobdAggregbtedIncompleteDbtbpointsFunc) SetDefbultHook(hook func(context.Context, int) ([]IncompleteDbtbpoint, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// LobdAggregbtedIncompleteDbtbpoints method of the pbrent MockInterfbce
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *InterfbceLobdAggregbtedIncompleteDbtbpointsFunc) PushHook(hook func(context.Context, int) ([]IncompleteDbtbpoint, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *InterfbceLobdAggregbtedIncompleteDbtbpointsFunc) SetDefbultReturn(r0 []IncompleteDbtbpoint, r1 error) {
	f.SetDefbultHook(func(context.Context, int) ([]IncompleteDbtbpoint, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *InterfbceLobdAggregbtedIncompleteDbtbpointsFunc) PushReturn(r0 []IncompleteDbtbpoint, r1 error) {
	f.PushHook(func(context.Context, int) ([]IncompleteDbtbpoint, error) {
		return r0, r1
	})
}

func (f *InterfbceLobdAggregbtedIncompleteDbtbpointsFunc) nextHook() func(context.Context, int) ([]IncompleteDbtbpoint, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *InterfbceLobdAggregbtedIncompleteDbtbpointsFunc) bppendCbll(r0 InterfbceLobdAggregbtedIncompleteDbtbpointsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// InterfbceLobdAggregbtedIncompleteDbtbpointsFuncCbll objects describing
// the invocbtions of this function.
func (f *InterfbceLobdAggregbtedIncompleteDbtbpointsFunc) History() []InterfbceLobdAggregbtedIncompleteDbtbpointsFuncCbll {
	f.mutex.Lock()
	history := mbke([]InterfbceLobdAggregbtedIncompleteDbtbpointsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// InterfbceLobdAggregbtedIncompleteDbtbpointsFuncCbll is bn object thbt
// describes bn invocbtion of method LobdAggregbtedIncompleteDbtbpoints on
// bn instbnce of MockInterfbce.
type InterfbceLobdAggregbtedIncompleteDbtbpointsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []IncompleteDbtbpoint
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c InterfbceLobdAggregbtedIncompleteDbtbpointsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c InterfbceLobdAggregbtedIncompleteDbtbpointsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// InterfbceRecordSeriesPointsFunc describes the behbvior when the
// RecordSeriesPoints method of the pbrent MockInterfbce instbnce is
// invoked.
type InterfbceRecordSeriesPointsFunc struct {
	defbultHook func(context.Context, []RecordSeriesPointArgs) error
	hooks       []func(context.Context, []RecordSeriesPointArgs) error
	history     []InterfbceRecordSeriesPointsFuncCbll
	mutex       sync.Mutex
}

// RecordSeriesPoints delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockInterfbce) RecordSeriesPoints(v0 context.Context, v1 []RecordSeriesPointArgs) error {
	r0 := m.RecordSeriesPointsFunc.nextHook()(v0, v1)
	m.RecordSeriesPointsFunc.bppendCbll(InterfbceRecordSeriesPointsFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the RecordSeriesPoints
// method of the pbrent MockInterfbce instbnce is invoked bnd the hook queue
// is empty.
func (f *InterfbceRecordSeriesPointsFunc) SetDefbultHook(hook func(context.Context, []RecordSeriesPointArgs) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RecordSeriesPoints method of the pbrent MockInterfbce instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *InterfbceRecordSeriesPointsFunc) PushHook(hook func(context.Context, []RecordSeriesPointArgs) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *InterfbceRecordSeriesPointsFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, []RecordSeriesPointArgs) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *InterfbceRecordSeriesPointsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, []RecordSeriesPointArgs) error {
		return r0
	})
}

func (f *InterfbceRecordSeriesPointsFunc) nextHook() func(context.Context, []RecordSeriesPointArgs) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *InterfbceRecordSeriesPointsFunc) bppendCbll(r0 InterfbceRecordSeriesPointsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of InterfbceRecordSeriesPointsFuncCbll objects
// describing the invocbtions of this function.
func (f *InterfbceRecordSeriesPointsFunc) History() []InterfbceRecordSeriesPointsFuncCbll {
	f.mutex.Lock()
	history := mbke([]InterfbceRecordSeriesPointsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// InterfbceRecordSeriesPointsFuncCbll is bn object thbt describes bn
// invocbtion of method RecordSeriesPoints on bn instbnce of MockInterfbce.
type InterfbceRecordSeriesPointsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []RecordSeriesPointArgs
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c InterfbceRecordSeriesPointsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c InterfbceRecordSeriesPointsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// InterfbceRecordSeriesPointsAndRecordingTimesFunc describes the behbvior
// when the RecordSeriesPointsAndRecordingTimes method of the pbrent
// MockInterfbce instbnce is invoked.
type InterfbceRecordSeriesPointsAndRecordingTimesFunc struct {
	defbultHook func(context.Context, []RecordSeriesPointArgs, types.InsightSeriesRecordingTimes) error
	hooks       []func(context.Context, []RecordSeriesPointArgs, types.InsightSeriesRecordingTimes) error
	history     []InterfbceRecordSeriesPointsAndRecordingTimesFuncCbll
	mutex       sync.Mutex
}

// RecordSeriesPointsAndRecordingTimes delegbtes to the next hook function
// in the queue bnd stores the pbrbmeter bnd result vblues of this
// invocbtion.
func (m *MockInterfbce) RecordSeriesPointsAndRecordingTimes(v0 context.Context, v1 []RecordSeriesPointArgs, v2 types.InsightSeriesRecordingTimes) error {
	r0 := m.RecordSeriesPointsAndRecordingTimesFunc.nextHook()(v0, v1, v2)
	m.RecordSeriesPointsAndRecordingTimesFunc.bppendCbll(InterfbceRecordSeriesPointsAndRecordingTimesFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// RecordSeriesPointsAndRecordingTimes method of the pbrent MockInterfbce
// instbnce is invoked bnd the hook queue is empty.
func (f *InterfbceRecordSeriesPointsAndRecordingTimesFunc) SetDefbultHook(hook func(context.Context, []RecordSeriesPointArgs, types.InsightSeriesRecordingTimes) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// RecordSeriesPointsAndRecordingTimes method of the pbrent MockInterfbce
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *InterfbceRecordSeriesPointsAndRecordingTimesFunc) PushHook(hook func(context.Context, []RecordSeriesPointArgs, types.InsightSeriesRecordingTimes) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *InterfbceRecordSeriesPointsAndRecordingTimesFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, []RecordSeriesPointArgs, types.InsightSeriesRecordingTimes) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *InterfbceRecordSeriesPointsAndRecordingTimesFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, []RecordSeriesPointArgs, types.InsightSeriesRecordingTimes) error {
		return r0
	})
}

func (f *InterfbceRecordSeriesPointsAndRecordingTimesFunc) nextHook() func(context.Context, []RecordSeriesPointArgs, types.InsightSeriesRecordingTimes) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *InterfbceRecordSeriesPointsAndRecordingTimesFunc) bppendCbll(r0 InterfbceRecordSeriesPointsAndRecordingTimesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// InterfbceRecordSeriesPointsAndRecordingTimesFuncCbll objects describing
// the invocbtions of this function.
func (f *InterfbceRecordSeriesPointsAndRecordingTimesFunc) History() []InterfbceRecordSeriesPointsAndRecordingTimesFuncCbll {
	f.mutex.Lock()
	history := mbke([]InterfbceRecordSeriesPointsAndRecordingTimesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// InterfbceRecordSeriesPointsAndRecordingTimesFuncCbll is bn object thbt
// describes bn invocbtion of method RecordSeriesPointsAndRecordingTimes on
// bn instbnce of MockInterfbce.
type InterfbceRecordSeriesPointsAndRecordingTimesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []RecordSeriesPointArgs
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 types.InsightSeriesRecordingTimes
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c InterfbceRecordSeriesPointsAndRecordingTimesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c InterfbceRecordSeriesPointsAndRecordingTimesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// InterfbceSeriesPointsFunc describes the behbvior when the SeriesPoints
// method of the pbrent MockInterfbce instbnce is invoked.
type InterfbceSeriesPointsFunc struct {
	defbultHook func(context.Context, SeriesPointsOpts) ([]SeriesPoint, error)
	hooks       []func(context.Context, SeriesPointsOpts) ([]SeriesPoint, error)
	history     []InterfbceSeriesPointsFuncCbll
	mutex       sync.Mutex
}

// SeriesPoints delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockInterfbce) SeriesPoints(v0 context.Context, v1 SeriesPointsOpts) ([]SeriesPoint, error) {
	r0, r1 := m.SeriesPointsFunc.nextHook()(v0, v1)
	m.SeriesPointsFunc.bppendCbll(InterfbceSeriesPointsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the SeriesPoints method
// of the pbrent MockInterfbce instbnce is invoked bnd the hook queue is
// empty.
func (f *InterfbceSeriesPointsFunc) SetDefbultHook(hook func(context.Context, SeriesPointsOpts) ([]SeriesPoint, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SeriesPoints method of the pbrent MockInterfbce instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *InterfbceSeriesPointsFunc) PushHook(hook func(context.Context, SeriesPointsOpts) ([]SeriesPoint, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *InterfbceSeriesPointsFunc) SetDefbultReturn(r0 []SeriesPoint, r1 error) {
	f.SetDefbultHook(func(context.Context, SeriesPointsOpts) ([]SeriesPoint, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *InterfbceSeriesPointsFunc) PushReturn(r0 []SeriesPoint, r1 error) {
	f.PushHook(func(context.Context, SeriesPointsOpts) ([]SeriesPoint, error) {
		return r0, r1
	})
}

func (f *InterfbceSeriesPointsFunc) nextHook() func(context.Context, SeriesPointsOpts) ([]SeriesPoint, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *InterfbceSeriesPointsFunc) bppendCbll(r0 InterfbceSeriesPointsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of InterfbceSeriesPointsFuncCbll objects
// describing the invocbtions of this function.
func (f *InterfbceSeriesPointsFunc) History() []InterfbceSeriesPointsFuncCbll {
	f.mutex.Lock()
	history := mbke([]InterfbceSeriesPointsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// InterfbceSeriesPointsFuncCbll is bn object thbt describes bn invocbtion
// of method SeriesPoints on bn instbnce of MockInterfbce.
type InterfbceSeriesPointsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 SeriesPointsOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []SeriesPoint
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c InterfbceSeriesPointsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c InterfbceSeriesPointsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// InterfbceSetInsightSeriesRecordingTimesFunc describes the behbvior when
// the SetInsightSeriesRecordingTimes method of the pbrent MockInterfbce
// instbnce is invoked.
type InterfbceSetInsightSeriesRecordingTimesFunc struct {
	defbultHook func(context.Context, []types.InsightSeriesRecordingTimes) error
	hooks       []func(context.Context, []types.InsightSeriesRecordingTimes) error
	history     []InterfbceSetInsightSeriesRecordingTimesFuncCbll
	mutex       sync.Mutex
}

// SetInsightSeriesRecordingTimes delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockInterfbce) SetInsightSeriesRecordingTimes(v0 context.Context, v1 []types.InsightSeriesRecordingTimes) error {
	r0 := m.SetInsightSeriesRecordingTimesFunc.nextHook()(v0, v1)
	m.SetInsightSeriesRecordingTimesFunc.bppendCbll(InterfbceSetInsightSeriesRecordingTimesFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// SetInsightSeriesRecordingTimes method of the pbrent MockInterfbce
// instbnce is invoked bnd the hook queue is empty.
func (f *InterfbceSetInsightSeriesRecordingTimesFunc) SetDefbultHook(hook func(context.Context, []types.InsightSeriesRecordingTimes) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SetInsightSeriesRecordingTimes method of the pbrent MockInterfbce
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *InterfbceSetInsightSeriesRecordingTimesFunc) PushHook(hook func(context.Context, []types.InsightSeriesRecordingTimes) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *InterfbceSetInsightSeriesRecordingTimesFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, []types.InsightSeriesRecordingTimes) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *InterfbceSetInsightSeriesRecordingTimesFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, []types.InsightSeriesRecordingTimes) error {
		return r0
	})
}

func (f *InterfbceSetInsightSeriesRecordingTimesFunc) nextHook() func(context.Context, []types.InsightSeriesRecordingTimes) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *InterfbceSetInsightSeriesRecordingTimesFunc) bppendCbll(r0 InterfbceSetInsightSeriesRecordingTimesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// InterfbceSetInsightSeriesRecordingTimesFuncCbll objects describing the
// invocbtions of this function.
func (f *InterfbceSetInsightSeriesRecordingTimesFunc) History() []InterfbceSetInsightSeriesRecordingTimesFuncCbll {
	f.mutex.Lock()
	history := mbke([]InterfbceSetInsightSeriesRecordingTimesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// InterfbceSetInsightSeriesRecordingTimesFuncCbll is bn object thbt
// describes bn invocbtion of method SetInsightSeriesRecordingTimes on bn
// instbnce of MockInterfbce.
type InterfbceSetInsightSeriesRecordingTimesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []types.InsightSeriesRecordingTimes
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c InterfbceSetInsightSeriesRecordingTimesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c InterfbceSetInsightSeriesRecordingTimesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// InterfbceWithOtherFunc describes the behbvior when the WithOther method
// of the pbrent MockInterfbce instbnce is invoked.
type InterfbceWithOtherFunc struct {
	defbultHook func(bbsestore.ShbrebbleStore) Interfbce
	hooks       []func(bbsestore.ShbrebbleStore) Interfbce
	history     []InterfbceWithOtherFuncCbll
	mutex       sync.Mutex
}

// WithOther delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockInterfbce) WithOther(v0 bbsestore.ShbrebbleStore) Interfbce {
	r0 := m.WithOtherFunc.nextHook()(v0)
	m.WithOtherFunc.bppendCbll(InterfbceWithOtherFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the WithOther method of
// the pbrent MockInterfbce instbnce is invoked bnd the hook queue is empty.
func (f *InterfbceWithOtherFunc) SetDefbultHook(hook func(bbsestore.ShbrebbleStore) Interfbce) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// WithOther method of the pbrent MockInterfbce instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *InterfbceWithOtherFunc) PushHook(hook func(bbsestore.ShbrebbleStore) Interfbce) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *InterfbceWithOtherFunc) SetDefbultReturn(r0 Interfbce) {
	f.SetDefbultHook(func(bbsestore.ShbrebbleStore) Interfbce {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *InterfbceWithOtherFunc) PushReturn(r0 Interfbce) {
	f.PushHook(func(bbsestore.ShbrebbleStore) Interfbce {
		return r0
	})
}

func (f *InterfbceWithOtherFunc) nextHook() func(bbsestore.ShbrebbleStore) Interfbce {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *InterfbceWithOtherFunc) bppendCbll(r0 InterfbceWithOtherFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of InterfbceWithOtherFuncCbll objects
// describing the invocbtions of this function.
func (f *InterfbceWithOtherFunc) History() []InterfbceWithOtherFuncCbll {
	f.mutex.Lock()
	history := mbke([]InterfbceWithOtherFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// InterfbceWithOtherFuncCbll is bn object thbt describes bn invocbtion of
// method WithOther on bn instbnce of MockInterfbce.
type InterfbceWithOtherFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 bbsestore.ShbrebbleStore
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 Interfbce
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c InterfbceWithOtherFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c InterfbceWithOtherFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
