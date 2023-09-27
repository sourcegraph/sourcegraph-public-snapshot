// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge repo

import (
	"context"
	"sync"

	sqlf "github.com/keegbncsmith/sqlf"
	bpi "github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	bbsestore "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
)

// MockRepoEmbeddingJobsStore is b mock implementbtion of the
// RepoEmbeddingJobsStore interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo)
// used for unit testing.
type MockRepoEmbeddingJobsStore struct {
	// CbncelRepoEmbeddingJobFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CbncelRepoEmbeddingJob.
	CbncelRepoEmbeddingJobFunc *RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFunc
	// CountRepoEmbeddingJobsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CountRepoEmbeddingJobs.
	CountRepoEmbeddingJobsFunc *RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFunc
	// CountRepoEmbeddingsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CountRepoEmbeddings.
	CountRepoEmbeddingsFunc *RepoEmbeddingJobsStoreCountRepoEmbeddingsFunc
	// CrebteRepoEmbeddingJobFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CrebteRepoEmbeddingJob.
	CrebteRepoEmbeddingJobFunc *RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFunc
	// DoneFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Done.
	DoneFunc *RepoEmbeddingJobsStoreDoneFunc
	// ExecFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Exec.
	ExecFunc *RepoEmbeddingJobsStoreExecFunc
	// GetEmbeddbbleReposFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetEmbeddbbleRepos.
	GetEmbeddbbleReposFunc *RepoEmbeddingJobsStoreGetEmbeddbbleReposFunc
	// GetLbstCompletedRepoEmbeddingJobFunc is bn instbnce of b mock
	// function object controlling the behbvior of the method
	// GetLbstCompletedRepoEmbeddingJob.
	GetLbstCompletedRepoEmbeddingJobFunc *RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFunc
	// GetLbstRepoEmbeddingJobForRevisionFunc is bn instbnce of b mock
	// function object controlling the behbvior of the method
	// GetLbstRepoEmbeddingJobForRevision.
	GetLbstRepoEmbeddingJobForRevisionFunc *RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFunc
	// GetRepoEmbeddingJobStbtsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method GetRepoEmbeddingJobStbts.
	GetRepoEmbeddingJobStbtsFunc *RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFunc
	// HbndleFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Hbndle.
	HbndleFunc *RepoEmbeddingJobsStoreHbndleFunc
	// ListRepoEmbeddingJobsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method ListRepoEmbeddingJobs.
	ListRepoEmbeddingJobsFunc *RepoEmbeddingJobsStoreListRepoEmbeddingJobsFunc
	// TrbnsbctFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Trbnsbct.
	TrbnsbctFunc *RepoEmbeddingJobsStoreTrbnsbctFunc
	// UpdbteRepoEmbeddingJobStbtsFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// UpdbteRepoEmbeddingJobStbts.
	UpdbteRepoEmbeddingJobStbtsFunc *RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFunc
}

// NewMockRepoEmbeddingJobsStore crebtes b new mock of the
// RepoEmbeddingJobsStore interfbce. All methods return zero vblues for bll
// results, unless overwritten.
func NewMockRepoEmbeddingJobsStore() *MockRepoEmbeddingJobsStore {
	return &MockRepoEmbeddingJobsStore{
		CbncelRepoEmbeddingJobFunc: &RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFunc{
			defbultHook: func(context.Context, int) (r0 error) {
				return
			},
		},
		CountRepoEmbeddingJobsFunc: &RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFunc{
			defbultHook: func(context.Context, ListOpts) (r0 int, r1 error) {
				return
			},
		},
		CountRepoEmbeddingsFunc: &RepoEmbeddingJobsStoreCountRepoEmbeddingsFunc{
			defbultHook: func(context.Context) (r0 int, r1 error) {
				return
			},
		},
		CrebteRepoEmbeddingJobFunc: &RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFunc{
			defbultHook: func(context.Context, bpi.RepoID, bpi.CommitID) (r0 int, r1 error) {
				return
			},
		},
		DoneFunc: &RepoEmbeddingJobsStoreDoneFunc{
			defbultHook: func(error) (r0 error) {
				return
			},
		},
		ExecFunc: &RepoEmbeddingJobsStoreExecFunc{
			defbultHook: func(context.Context, *sqlf.Query) (r0 error) {
				return
			},
		},
		GetEmbeddbbleReposFunc: &RepoEmbeddingJobsStoreGetEmbeddbbleReposFunc{
			defbultHook: func(context.Context, EmbeddbbleRepoOpts) (r0 []EmbeddbbleRepo, r1 error) {
				return
			},
		},
		GetLbstCompletedRepoEmbeddingJobFunc: &RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFunc{
			defbultHook: func(context.Context, bpi.RepoID) (r0 *RepoEmbeddingJob, r1 error) {
				return
			},
		},
		GetLbstRepoEmbeddingJobForRevisionFunc: &RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFunc{
			defbultHook: func(context.Context, bpi.RepoID, bpi.CommitID) (r0 *RepoEmbeddingJob, r1 error) {
				return
			},
		},
		GetRepoEmbeddingJobStbtsFunc: &RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFunc{
			defbultHook: func(context.Context, int) (r0 EmbedRepoStbts, r1 error) {
				return
			},
		},
		HbndleFunc: &RepoEmbeddingJobsStoreHbndleFunc{
			defbultHook: func() (r0 bbsestore.TrbnsbctbbleHbndle) {
				return
			},
		},
		ListRepoEmbeddingJobsFunc: &RepoEmbeddingJobsStoreListRepoEmbeddingJobsFunc{
			defbultHook: func(context.Context, ListOpts) (r0 []*RepoEmbeddingJob, r1 error) {
				return
			},
		},
		TrbnsbctFunc: &RepoEmbeddingJobsStoreTrbnsbctFunc{
			defbultHook: func(context.Context) (r0 RepoEmbeddingJobsStore, r1 error) {
				return
			},
		},
		UpdbteRepoEmbeddingJobStbtsFunc: &RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFunc{
			defbultHook: func(context.Context, int, *EmbedRepoStbts) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockRepoEmbeddingJobsStore crebtes b new mock of the
// RepoEmbeddingJobsStore interfbce. All methods pbnic on invocbtion, unless
// overwritten.
func NewStrictMockRepoEmbeddingJobsStore() *MockRepoEmbeddingJobsStore {
	return &MockRepoEmbeddingJobsStore{
		CbncelRepoEmbeddingJobFunc: &RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFunc{
			defbultHook: func(context.Context, int) error {
				pbnic("unexpected invocbtion of MockRepoEmbeddingJobsStore.CbncelRepoEmbeddingJob")
			},
		},
		CountRepoEmbeddingJobsFunc: &RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFunc{
			defbultHook: func(context.Context, ListOpts) (int, error) {
				pbnic("unexpected invocbtion of MockRepoEmbeddingJobsStore.CountRepoEmbeddingJobs")
			},
		},
		CountRepoEmbeddingsFunc: &RepoEmbeddingJobsStoreCountRepoEmbeddingsFunc{
			defbultHook: func(context.Context) (int, error) {
				pbnic("unexpected invocbtion of MockRepoEmbeddingJobsStore.CountRepoEmbeddings")
			},
		},
		CrebteRepoEmbeddingJobFunc: &RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFunc{
			defbultHook: func(context.Context, bpi.RepoID, bpi.CommitID) (int, error) {
				pbnic("unexpected invocbtion of MockRepoEmbeddingJobsStore.CrebteRepoEmbeddingJob")
			},
		},
		DoneFunc: &RepoEmbeddingJobsStoreDoneFunc{
			defbultHook: func(error) error {
				pbnic("unexpected invocbtion of MockRepoEmbeddingJobsStore.Done")
			},
		},
		ExecFunc: &RepoEmbeddingJobsStoreExecFunc{
			defbultHook: func(context.Context, *sqlf.Query) error {
				pbnic("unexpected invocbtion of MockRepoEmbeddingJobsStore.Exec")
			},
		},
		GetEmbeddbbleReposFunc: &RepoEmbeddingJobsStoreGetEmbeddbbleReposFunc{
			defbultHook: func(context.Context, EmbeddbbleRepoOpts) ([]EmbeddbbleRepo, error) {
				pbnic("unexpected invocbtion of MockRepoEmbeddingJobsStore.GetEmbeddbbleRepos")
			},
		},
		GetLbstCompletedRepoEmbeddingJobFunc: &RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFunc{
			defbultHook: func(context.Context, bpi.RepoID) (*RepoEmbeddingJob, error) {
				pbnic("unexpected invocbtion of MockRepoEmbeddingJobsStore.GetLbstCompletedRepoEmbeddingJob")
			},
		},
		GetLbstRepoEmbeddingJobForRevisionFunc: &RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFunc{
			defbultHook: func(context.Context, bpi.RepoID, bpi.CommitID) (*RepoEmbeddingJob, error) {
				pbnic("unexpected invocbtion of MockRepoEmbeddingJobsStore.GetLbstRepoEmbeddingJobForRevision")
			},
		},
		GetRepoEmbeddingJobStbtsFunc: &RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFunc{
			defbultHook: func(context.Context, int) (EmbedRepoStbts, error) {
				pbnic("unexpected invocbtion of MockRepoEmbeddingJobsStore.GetRepoEmbeddingJobStbts")
			},
		},
		HbndleFunc: &RepoEmbeddingJobsStoreHbndleFunc{
			defbultHook: func() bbsestore.TrbnsbctbbleHbndle {
				pbnic("unexpected invocbtion of MockRepoEmbeddingJobsStore.Hbndle")
			},
		},
		ListRepoEmbeddingJobsFunc: &RepoEmbeddingJobsStoreListRepoEmbeddingJobsFunc{
			defbultHook: func(context.Context, ListOpts) ([]*RepoEmbeddingJob, error) {
				pbnic("unexpected invocbtion of MockRepoEmbeddingJobsStore.ListRepoEmbeddingJobs")
			},
		},
		TrbnsbctFunc: &RepoEmbeddingJobsStoreTrbnsbctFunc{
			defbultHook: func(context.Context) (RepoEmbeddingJobsStore, error) {
				pbnic("unexpected invocbtion of MockRepoEmbeddingJobsStore.Trbnsbct")
			},
		},
		UpdbteRepoEmbeddingJobStbtsFunc: &RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFunc{
			defbultHook: func(context.Context, int, *EmbedRepoStbts) error {
				pbnic("unexpected invocbtion of MockRepoEmbeddingJobsStore.UpdbteRepoEmbeddingJobStbts")
			},
		},
	}
}

// NewMockRepoEmbeddingJobsStoreFrom crebtes b new mock of the
// MockRepoEmbeddingJobsStore interfbce. All methods delegbte to the given
// implementbtion, unless overwritten.
func NewMockRepoEmbeddingJobsStoreFrom(i RepoEmbeddingJobsStore) *MockRepoEmbeddingJobsStore {
	return &MockRepoEmbeddingJobsStore{
		CbncelRepoEmbeddingJobFunc: &RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFunc{
			defbultHook: i.CbncelRepoEmbeddingJob,
		},
		CountRepoEmbeddingJobsFunc: &RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFunc{
			defbultHook: i.CountRepoEmbeddingJobs,
		},
		CountRepoEmbeddingsFunc: &RepoEmbeddingJobsStoreCountRepoEmbeddingsFunc{
			defbultHook: i.CountRepoEmbeddings,
		},
		CrebteRepoEmbeddingJobFunc: &RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFunc{
			defbultHook: i.CrebteRepoEmbeddingJob,
		},
		DoneFunc: &RepoEmbeddingJobsStoreDoneFunc{
			defbultHook: i.Done,
		},
		ExecFunc: &RepoEmbeddingJobsStoreExecFunc{
			defbultHook: i.Exec,
		},
		GetEmbeddbbleReposFunc: &RepoEmbeddingJobsStoreGetEmbeddbbleReposFunc{
			defbultHook: i.GetEmbeddbbleRepos,
		},
		GetLbstCompletedRepoEmbeddingJobFunc: &RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFunc{
			defbultHook: i.GetLbstCompletedRepoEmbeddingJob,
		},
		GetLbstRepoEmbeddingJobForRevisionFunc: &RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFunc{
			defbultHook: i.GetLbstRepoEmbeddingJobForRevision,
		},
		GetRepoEmbeddingJobStbtsFunc: &RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFunc{
			defbultHook: i.GetRepoEmbeddingJobStbts,
		},
		HbndleFunc: &RepoEmbeddingJobsStoreHbndleFunc{
			defbultHook: i.Hbndle,
		},
		ListRepoEmbeddingJobsFunc: &RepoEmbeddingJobsStoreListRepoEmbeddingJobsFunc{
			defbultHook: i.ListRepoEmbeddingJobs,
		},
		TrbnsbctFunc: &RepoEmbeddingJobsStoreTrbnsbctFunc{
			defbultHook: i.Trbnsbct,
		},
		UpdbteRepoEmbeddingJobStbtsFunc: &RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFunc{
			defbultHook: i.UpdbteRepoEmbeddingJobStbts,
		},
	}
}

// RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFunc describes the behbvior
// when the CbncelRepoEmbeddingJob method of the pbrent
// MockRepoEmbeddingJobsStore instbnce is invoked.
type RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFunc struct {
	defbultHook func(context.Context, int) error
	hooks       []func(context.Context, int) error
	history     []RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFuncCbll
	mutex       sync.Mutex
}

// CbncelRepoEmbeddingJob delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRepoEmbeddingJobsStore) CbncelRepoEmbeddingJob(v0 context.Context, v1 int) error {
	r0 := m.CbncelRepoEmbeddingJobFunc.nextHook()(v0, v1)
	m.CbncelRepoEmbeddingJobFunc.bppendCbll(RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// CbncelRepoEmbeddingJob method of the pbrent MockRepoEmbeddingJobsStore
// instbnce is invoked bnd the hook queue is empty.
func (f *RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFunc) SetDefbultHook(hook func(context.Context, int) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CbncelRepoEmbeddingJob method of the pbrent MockRepoEmbeddingJobsStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFunc) PushHook(hook func(context.Context, int) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int) error {
		return r0
	})
}

func (f *RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFunc) nextHook() func(context.Context, int) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFunc) bppendCbll(r0 RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFuncCbll objects describing
// the invocbtions of this function.
func (f *RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFunc) History() []RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFuncCbll {
	f.mutex.Lock()
	history := mbke([]RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFuncCbll is bn object thbt
// describes bn invocbtion of method CbncelRepoEmbeddingJob on bn instbnce
// of MockRepoEmbeddingJobsStore.
type RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFuncCbll struct {
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
func (c RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RepoEmbeddingJobsStoreCbncelRepoEmbeddingJobFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFunc describes the behbvior
// when the CountRepoEmbeddingJobs method of the pbrent
// MockRepoEmbeddingJobsStore instbnce is invoked.
type RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFunc struct {
	defbultHook func(context.Context, ListOpts) (int, error)
	hooks       []func(context.Context, ListOpts) (int, error)
	history     []RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFuncCbll
	mutex       sync.Mutex
}

// CountRepoEmbeddingJobs delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRepoEmbeddingJobsStore) CountRepoEmbeddingJobs(v0 context.Context, v1 ListOpts) (int, error) {
	r0, r1 := m.CountRepoEmbeddingJobsFunc.nextHook()(v0, v1)
	m.CountRepoEmbeddingJobsFunc.bppendCbll(RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// CountRepoEmbeddingJobs method of the pbrent MockRepoEmbeddingJobsStore
// instbnce is invoked bnd the hook queue is empty.
func (f *RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFunc) SetDefbultHook(hook func(context.Context, ListOpts) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CountRepoEmbeddingJobs method of the pbrent MockRepoEmbeddingJobsStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFunc) PushHook(hook func(context.Context, ListOpts) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context, ListOpts) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, ListOpts) (int, error) {
		return r0, r1
	})
}

func (f *RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFunc) nextHook() func(context.Context, ListOpts) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFunc) bppendCbll(r0 RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFuncCbll objects describing
// the invocbtions of this function.
func (f *RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFunc) History() []RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFuncCbll {
	f.mutex.Lock()
	history := mbke([]RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFuncCbll is bn object thbt
// describes bn invocbtion of method CountRepoEmbeddingJobs on bn instbnce
// of MockRepoEmbeddingJobsStore.
type RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 ListOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RepoEmbeddingJobsStoreCountRepoEmbeddingJobsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// RepoEmbeddingJobsStoreCountRepoEmbeddingsFunc describes the behbvior when
// the CountRepoEmbeddings method of the pbrent MockRepoEmbeddingJobsStore
// instbnce is invoked.
type RepoEmbeddingJobsStoreCountRepoEmbeddingsFunc struct {
	defbultHook func(context.Context) (int, error)
	hooks       []func(context.Context) (int, error)
	history     []RepoEmbeddingJobsStoreCountRepoEmbeddingsFuncCbll
	mutex       sync.Mutex
}

// CountRepoEmbeddings delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRepoEmbeddingJobsStore) CountRepoEmbeddings(v0 context.Context) (int, error) {
	r0, r1 := m.CountRepoEmbeddingsFunc.nextHook()(v0)
	m.CountRepoEmbeddingsFunc.bppendCbll(RepoEmbeddingJobsStoreCountRepoEmbeddingsFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CountRepoEmbeddings
// method of the pbrent MockRepoEmbeddingJobsStore instbnce is invoked bnd
// the hook queue is empty.
func (f *RepoEmbeddingJobsStoreCountRepoEmbeddingsFunc) SetDefbultHook(hook func(context.Context) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CountRepoEmbeddings method of the pbrent MockRepoEmbeddingJobsStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *RepoEmbeddingJobsStoreCountRepoEmbeddingsFunc) PushHook(hook func(context.Context) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RepoEmbeddingJobsStoreCountRepoEmbeddingsFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RepoEmbeddingJobsStoreCountRepoEmbeddingsFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context) (int, error) {
		return r0, r1
	})
}

func (f *RepoEmbeddingJobsStoreCountRepoEmbeddingsFunc) nextHook() func(context.Context) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RepoEmbeddingJobsStoreCountRepoEmbeddingsFunc) bppendCbll(r0 RepoEmbeddingJobsStoreCountRepoEmbeddingsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// RepoEmbeddingJobsStoreCountRepoEmbeddingsFuncCbll objects describing the
// invocbtions of this function.
func (f *RepoEmbeddingJobsStoreCountRepoEmbeddingsFunc) History() []RepoEmbeddingJobsStoreCountRepoEmbeddingsFuncCbll {
	f.mutex.Lock()
	history := mbke([]RepoEmbeddingJobsStoreCountRepoEmbeddingsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RepoEmbeddingJobsStoreCountRepoEmbeddingsFuncCbll is bn object thbt
// describes bn invocbtion of method CountRepoEmbeddings on bn instbnce of
// MockRepoEmbeddingJobsStore.
type RepoEmbeddingJobsStoreCountRepoEmbeddingsFuncCbll struct {
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
func (c RepoEmbeddingJobsStoreCountRepoEmbeddingsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RepoEmbeddingJobsStoreCountRepoEmbeddingsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFunc describes the behbvior
// when the CrebteRepoEmbeddingJob method of the pbrent
// MockRepoEmbeddingJobsStore instbnce is invoked.
type RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFunc struct {
	defbultHook func(context.Context, bpi.RepoID, bpi.CommitID) (int, error)
	hooks       []func(context.Context, bpi.RepoID, bpi.CommitID) (int, error)
	history     []RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFuncCbll
	mutex       sync.Mutex
}

// CrebteRepoEmbeddingJob delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRepoEmbeddingJobsStore) CrebteRepoEmbeddingJob(v0 context.Context, v1 bpi.RepoID, v2 bpi.CommitID) (int, error) {
	r0, r1 := m.CrebteRepoEmbeddingJobFunc.nextHook()(v0, v1, v2)
	m.CrebteRepoEmbeddingJobFunc.bppendCbll(RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// CrebteRepoEmbeddingJob method of the pbrent MockRepoEmbeddingJobsStore
// instbnce is invoked bnd the hook queue is empty.
func (f *RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFunc) SetDefbultHook(hook func(context.Context, bpi.RepoID, bpi.CommitID) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CrebteRepoEmbeddingJob method of the pbrent MockRepoEmbeddingJobsStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFunc) PushHook(hook func(context.Context, bpi.RepoID, bpi.CommitID) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoID, bpi.CommitID) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoID, bpi.CommitID) (int, error) {
		return r0, r1
	})
}

func (f *RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFunc) nextHook() func(context.Context, bpi.RepoID, bpi.CommitID) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFunc) bppendCbll(r0 RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFuncCbll objects describing
// the invocbtions of this function.
func (f *RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFunc) History() []RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFuncCbll {
	f.mutex.Lock()
	history := mbke([]RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFuncCbll is bn object thbt
// describes bn invocbtion of method CrebteRepoEmbeddingJob on bn instbnce
// of MockRepoEmbeddingJobsStore.
type RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoID
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.CommitID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RepoEmbeddingJobsStoreCrebteRepoEmbeddingJobFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// RepoEmbeddingJobsStoreDoneFunc describes the behbvior when the Done
// method of the pbrent MockRepoEmbeddingJobsStore instbnce is invoked.
type RepoEmbeddingJobsStoreDoneFunc struct {
	defbultHook func(error) error
	hooks       []func(error) error
	history     []RepoEmbeddingJobsStoreDoneFuncCbll
	mutex       sync.Mutex
}

// Done delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRepoEmbeddingJobsStore) Done(v0 error) error {
	r0 := m.DoneFunc.nextHook()(v0)
	m.DoneFunc.bppendCbll(RepoEmbeddingJobsStoreDoneFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Done method of the
// pbrent MockRepoEmbeddingJobsStore instbnce is invoked bnd the hook queue
// is empty.
func (f *RepoEmbeddingJobsStoreDoneFunc) SetDefbultHook(hook func(error) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Done method of the pbrent MockRepoEmbeddingJobsStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *RepoEmbeddingJobsStoreDoneFunc) PushHook(hook func(error) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RepoEmbeddingJobsStoreDoneFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(error) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RepoEmbeddingJobsStoreDoneFunc) PushReturn(r0 error) {
	f.PushHook(func(error) error {
		return r0
	})
}

func (f *RepoEmbeddingJobsStoreDoneFunc) nextHook() func(error) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RepoEmbeddingJobsStoreDoneFunc) bppendCbll(r0 RepoEmbeddingJobsStoreDoneFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of RepoEmbeddingJobsStoreDoneFuncCbll objects
// describing the invocbtions of this function.
func (f *RepoEmbeddingJobsStoreDoneFunc) History() []RepoEmbeddingJobsStoreDoneFuncCbll {
	f.mutex.Lock()
	history := mbke([]RepoEmbeddingJobsStoreDoneFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RepoEmbeddingJobsStoreDoneFuncCbll is bn object thbt describes bn
// invocbtion of method Done on bn instbnce of MockRepoEmbeddingJobsStore.
type RepoEmbeddingJobsStoreDoneFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 error
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RepoEmbeddingJobsStoreDoneFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RepoEmbeddingJobsStoreDoneFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// RepoEmbeddingJobsStoreExecFunc describes the behbvior when the Exec
// method of the pbrent MockRepoEmbeddingJobsStore instbnce is invoked.
type RepoEmbeddingJobsStoreExecFunc struct {
	defbultHook func(context.Context, *sqlf.Query) error
	hooks       []func(context.Context, *sqlf.Query) error
	history     []RepoEmbeddingJobsStoreExecFuncCbll
	mutex       sync.Mutex
}

// Exec delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRepoEmbeddingJobsStore) Exec(v0 context.Context, v1 *sqlf.Query) error {
	r0 := m.ExecFunc.nextHook()(v0, v1)
	m.ExecFunc.bppendCbll(RepoEmbeddingJobsStoreExecFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Exec method of the
// pbrent MockRepoEmbeddingJobsStore instbnce is invoked bnd the hook queue
// is empty.
func (f *RepoEmbeddingJobsStoreExecFunc) SetDefbultHook(hook func(context.Context, *sqlf.Query) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Exec method of the pbrent MockRepoEmbeddingJobsStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *RepoEmbeddingJobsStoreExecFunc) PushHook(hook func(context.Context, *sqlf.Query) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RepoEmbeddingJobsStoreExecFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *sqlf.Query) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RepoEmbeddingJobsStoreExecFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *sqlf.Query) error {
		return r0
	})
}

func (f *RepoEmbeddingJobsStoreExecFunc) nextHook() func(context.Context, *sqlf.Query) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RepoEmbeddingJobsStoreExecFunc) bppendCbll(r0 RepoEmbeddingJobsStoreExecFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of RepoEmbeddingJobsStoreExecFuncCbll objects
// describing the invocbtions of this function.
func (f *RepoEmbeddingJobsStoreExecFunc) History() []RepoEmbeddingJobsStoreExecFuncCbll {
	f.mutex.Lock()
	history := mbke([]RepoEmbeddingJobsStoreExecFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RepoEmbeddingJobsStoreExecFuncCbll is bn object thbt describes bn
// invocbtion of method Exec on bn instbnce of MockRepoEmbeddingJobsStore.
type RepoEmbeddingJobsStoreExecFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *sqlf.Query
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RepoEmbeddingJobsStoreExecFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RepoEmbeddingJobsStoreExecFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// RepoEmbeddingJobsStoreGetEmbeddbbleReposFunc describes the behbvior when
// the GetEmbeddbbleRepos method of the pbrent MockRepoEmbeddingJobsStore
// instbnce is invoked.
type RepoEmbeddingJobsStoreGetEmbeddbbleReposFunc struct {
	defbultHook func(context.Context, EmbeddbbleRepoOpts) ([]EmbeddbbleRepo, error)
	hooks       []func(context.Context, EmbeddbbleRepoOpts) ([]EmbeddbbleRepo, error)
	history     []RepoEmbeddingJobsStoreGetEmbeddbbleReposFuncCbll
	mutex       sync.Mutex
}

// GetEmbeddbbleRepos delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRepoEmbeddingJobsStore) GetEmbeddbbleRepos(v0 context.Context, v1 EmbeddbbleRepoOpts) ([]EmbeddbbleRepo, error) {
	r0, r1 := m.GetEmbeddbbleReposFunc.nextHook()(v0, v1)
	m.GetEmbeddbbleReposFunc.bppendCbll(RepoEmbeddingJobsStoreGetEmbeddbbleReposFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetEmbeddbbleRepos
// method of the pbrent MockRepoEmbeddingJobsStore instbnce is invoked bnd
// the hook queue is empty.
func (f *RepoEmbeddingJobsStoreGetEmbeddbbleReposFunc) SetDefbultHook(hook func(context.Context, EmbeddbbleRepoOpts) ([]EmbeddbbleRepo, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetEmbeddbbleRepos method of the pbrent MockRepoEmbeddingJobsStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *RepoEmbeddingJobsStoreGetEmbeddbbleReposFunc) PushHook(hook func(context.Context, EmbeddbbleRepoOpts) ([]EmbeddbbleRepo, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RepoEmbeddingJobsStoreGetEmbeddbbleReposFunc) SetDefbultReturn(r0 []EmbeddbbleRepo, r1 error) {
	f.SetDefbultHook(func(context.Context, EmbeddbbleRepoOpts) ([]EmbeddbbleRepo, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RepoEmbeddingJobsStoreGetEmbeddbbleReposFunc) PushReturn(r0 []EmbeddbbleRepo, r1 error) {
	f.PushHook(func(context.Context, EmbeddbbleRepoOpts) ([]EmbeddbbleRepo, error) {
		return r0, r1
	})
}

func (f *RepoEmbeddingJobsStoreGetEmbeddbbleReposFunc) nextHook() func(context.Context, EmbeddbbleRepoOpts) ([]EmbeddbbleRepo, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RepoEmbeddingJobsStoreGetEmbeddbbleReposFunc) bppendCbll(r0 RepoEmbeddingJobsStoreGetEmbeddbbleReposFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// RepoEmbeddingJobsStoreGetEmbeddbbleReposFuncCbll objects describing the
// invocbtions of this function.
func (f *RepoEmbeddingJobsStoreGetEmbeddbbleReposFunc) History() []RepoEmbeddingJobsStoreGetEmbeddbbleReposFuncCbll {
	f.mutex.Lock()
	history := mbke([]RepoEmbeddingJobsStoreGetEmbeddbbleReposFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RepoEmbeddingJobsStoreGetEmbeddbbleReposFuncCbll is bn object thbt
// describes bn invocbtion of method GetEmbeddbbleRepos on bn instbnce of
// MockRepoEmbeddingJobsStore.
type RepoEmbeddingJobsStoreGetEmbeddbbleReposFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 EmbeddbbleRepoOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []EmbeddbbleRepo
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RepoEmbeddingJobsStoreGetEmbeddbbleReposFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RepoEmbeddingJobsStoreGetEmbeddbbleReposFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFunc describes the
// behbvior when the GetLbstCompletedRepoEmbeddingJob method of the pbrent
// MockRepoEmbeddingJobsStore instbnce is invoked.
type RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFunc struct {
	defbultHook func(context.Context, bpi.RepoID) (*RepoEmbeddingJob, error)
	hooks       []func(context.Context, bpi.RepoID) (*RepoEmbeddingJob, error)
	history     []RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFuncCbll
	mutex       sync.Mutex
}

// GetLbstCompletedRepoEmbeddingJob delegbtes to the next hook function in
// the queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRepoEmbeddingJobsStore) GetLbstCompletedRepoEmbeddingJob(v0 context.Context, v1 bpi.RepoID) (*RepoEmbeddingJob, error) {
	r0, r1 := m.GetLbstCompletedRepoEmbeddingJobFunc.nextHook()(v0, v1)
	m.GetLbstCompletedRepoEmbeddingJobFunc.bppendCbll(RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetLbstCompletedRepoEmbeddingJob method of the pbrent
// MockRepoEmbeddingJobsStore instbnce is invoked bnd the hook queue is
// empty.
func (f *RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFunc) SetDefbultHook(hook func(context.Context, bpi.RepoID) (*RepoEmbeddingJob, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetLbstCompletedRepoEmbeddingJob method of the pbrent
// MockRepoEmbeddingJobsStore instbnce invokes the hook bt the front of the
// queue bnd discbrds it. After the queue is empty, the defbult hook
// function is invoked for bny future bction.
func (f *RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFunc) PushHook(hook func(context.Context, bpi.RepoID) (*RepoEmbeddingJob, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFunc) SetDefbultReturn(r0 *RepoEmbeddingJob, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoID) (*RepoEmbeddingJob, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFunc) PushReturn(r0 *RepoEmbeddingJob, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoID) (*RepoEmbeddingJob, error) {
		return r0, r1
	})
}

func (f *RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFunc) nextHook() func(context.Context, bpi.RepoID) (*RepoEmbeddingJob, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFunc) bppendCbll(r0 RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFuncCbll objects
// describing the invocbtions of this function.
func (f *RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFunc) History() []RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFuncCbll {
	f.mutex.Lock()
	history := mbke([]RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFuncCbll is bn
// object thbt describes bn invocbtion of method
// GetLbstCompletedRepoEmbeddingJob on bn instbnce of
// MockRepoEmbeddingJobsStore.
type RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *RepoEmbeddingJob
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RepoEmbeddingJobsStoreGetLbstCompletedRepoEmbeddingJobFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFunc describes
// the behbvior when the GetLbstRepoEmbeddingJobForRevision method of the
// pbrent MockRepoEmbeddingJobsStore instbnce is invoked.
type RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFunc struct {
	defbultHook func(context.Context, bpi.RepoID, bpi.CommitID) (*RepoEmbeddingJob, error)
	hooks       []func(context.Context, bpi.RepoID, bpi.CommitID) (*RepoEmbeddingJob, error)
	history     []RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFuncCbll
	mutex       sync.Mutex
}

// GetLbstRepoEmbeddingJobForRevision delegbtes to the next hook function in
// the queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRepoEmbeddingJobsStore) GetLbstRepoEmbeddingJobForRevision(v0 context.Context, v1 bpi.RepoID, v2 bpi.CommitID) (*RepoEmbeddingJob, error) {
	r0, r1 := m.GetLbstRepoEmbeddingJobForRevisionFunc.nextHook()(v0, v1, v2)
	m.GetLbstRepoEmbeddingJobForRevisionFunc.bppendCbll(RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetLbstRepoEmbeddingJobForRevision method of the pbrent
// MockRepoEmbeddingJobsStore instbnce is invoked bnd the hook queue is
// empty.
func (f *RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFunc) SetDefbultHook(hook func(context.Context, bpi.RepoID, bpi.CommitID) (*RepoEmbeddingJob, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetLbstRepoEmbeddingJobForRevision method of the pbrent
// MockRepoEmbeddingJobsStore instbnce invokes the hook bt the front of the
// queue bnd discbrds it. After the queue is empty, the defbult hook
// function is invoked for bny future bction.
func (f *RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFunc) PushHook(hook func(context.Context, bpi.RepoID, bpi.CommitID) (*RepoEmbeddingJob, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFunc) SetDefbultReturn(r0 *RepoEmbeddingJob, r1 error) {
	f.SetDefbultHook(func(context.Context, bpi.RepoID, bpi.CommitID) (*RepoEmbeddingJob, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFunc) PushReturn(r0 *RepoEmbeddingJob, r1 error) {
	f.PushHook(func(context.Context, bpi.RepoID, bpi.CommitID) (*RepoEmbeddingJob, error) {
		return r0, r1
	})
}

func (f *RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFunc) nextHook() func(context.Context, bpi.RepoID, bpi.CommitID) (*RepoEmbeddingJob, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFunc) bppendCbll(r0 RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFuncCbll objects
// describing the invocbtions of this function.
func (f *RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFunc) History() []RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFuncCbll {
	f.mutex.Lock()
	history := mbke([]RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFuncCbll is bn
// object thbt describes bn invocbtion of method
// GetLbstRepoEmbeddingJobForRevision on bn instbnce of
// MockRepoEmbeddingJobsStore.
type RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 bpi.RepoID
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 bpi.CommitID
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *RepoEmbeddingJob
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RepoEmbeddingJobsStoreGetLbstRepoEmbeddingJobForRevisionFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFunc describes the behbvior
// when the GetRepoEmbeddingJobStbts method of the pbrent
// MockRepoEmbeddingJobsStore instbnce is invoked.
type RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFunc struct {
	defbultHook func(context.Context, int) (EmbedRepoStbts, error)
	hooks       []func(context.Context, int) (EmbedRepoStbts, error)
	history     []RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFuncCbll
	mutex       sync.Mutex
}

// GetRepoEmbeddingJobStbts delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRepoEmbeddingJobsStore) GetRepoEmbeddingJobStbts(v0 context.Context, v1 int) (EmbedRepoStbts, error) {
	r0, r1 := m.GetRepoEmbeddingJobStbtsFunc.nextHook()(v0, v1)
	m.GetRepoEmbeddingJobStbtsFunc.bppendCbll(RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// GetRepoEmbeddingJobStbts method of the pbrent MockRepoEmbeddingJobsStore
// instbnce is invoked bnd the hook queue is empty.
func (f *RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFunc) SetDefbultHook(hook func(context.Context, int) (EmbedRepoStbts, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetRepoEmbeddingJobStbts method of the pbrent MockRepoEmbeddingJobsStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFunc) PushHook(hook func(context.Context, int) (EmbedRepoStbts, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFunc) SetDefbultReturn(r0 EmbedRepoStbts, r1 error) {
	f.SetDefbultHook(func(context.Context, int) (EmbedRepoStbts, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFunc) PushReturn(r0 EmbedRepoStbts, r1 error) {
	f.PushHook(func(context.Context, int) (EmbedRepoStbts, error) {
		return r0, r1
	})
}

func (f *RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFunc) nextHook() func(context.Context, int) (EmbedRepoStbts, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFunc) bppendCbll(r0 RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFuncCbll objects describing
// the invocbtions of this function.
func (f *RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFunc) History() []RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFuncCbll {
	f.mutex.Lock()
	history := mbke([]RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFuncCbll is bn object thbt
// describes bn invocbtion of method GetRepoEmbeddingJobStbts on bn instbnce
// of MockRepoEmbeddingJobsStore.
type RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 EmbedRepoStbts
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RepoEmbeddingJobsStoreGetRepoEmbeddingJobStbtsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// RepoEmbeddingJobsStoreHbndleFunc describes the behbvior when the Hbndle
// method of the pbrent MockRepoEmbeddingJobsStore instbnce is invoked.
type RepoEmbeddingJobsStoreHbndleFunc struct {
	defbultHook func() bbsestore.TrbnsbctbbleHbndle
	hooks       []func() bbsestore.TrbnsbctbbleHbndle
	history     []RepoEmbeddingJobsStoreHbndleFuncCbll
	mutex       sync.Mutex
}

// Hbndle delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRepoEmbeddingJobsStore) Hbndle() bbsestore.TrbnsbctbbleHbndle {
	r0 := m.HbndleFunc.nextHook()()
	m.HbndleFunc.bppendCbll(RepoEmbeddingJobsStoreHbndleFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Hbndle method of the
// pbrent MockRepoEmbeddingJobsStore instbnce is invoked bnd the hook queue
// is empty.
func (f *RepoEmbeddingJobsStoreHbndleFunc) SetDefbultHook(hook func() bbsestore.TrbnsbctbbleHbndle) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Hbndle method of the pbrent MockRepoEmbeddingJobsStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *RepoEmbeddingJobsStoreHbndleFunc) PushHook(hook func() bbsestore.TrbnsbctbbleHbndle) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RepoEmbeddingJobsStoreHbndleFunc) SetDefbultReturn(r0 bbsestore.TrbnsbctbbleHbndle) {
	f.SetDefbultHook(func() bbsestore.TrbnsbctbbleHbndle {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RepoEmbeddingJobsStoreHbndleFunc) PushReturn(r0 bbsestore.TrbnsbctbbleHbndle) {
	f.PushHook(func() bbsestore.TrbnsbctbbleHbndle {
		return r0
	})
}

func (f *RepoEmbeddingJobsStoreHbndleFunc) nextHook() func() bbsestore.TrbnsbctbbleHbndle {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RepoEmbeddingJobsStoreHbndleFunc) bppendCbll(r0 RepoEmbeddingJobsStoreHbndleFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of RepoEmbeddingJobsStoreHbndleFuncCbll
// objects describing the invocbtions of this function.
func (f *RepoEmbeddingJobsStoreHbndleFunc) History() []RepoEmbeddingJobsStoreHbndleFuncCbll {
	f.mutex.Lock()
	history := mbke([]RepoEmbeddingJobsStoreHbndleFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RepoEmbeddingJobsStoreHbndleFuncCbll is bn object thbt describes bn
// invocbtion of method Hbndle on bn instbnce of MockRepoEmbeddingJobsStore.
type RepoEmbeddingJobsStoreHbndleFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bbsestore.TrbnsbctbbleHbndle
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RepoEmbeddingJobsStoreHbndleFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RepoEmbeddingJobsStoreHbndleFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// RepoEmbeddingJobsStoreListRepoEmbeddingJobsFunc describes the behbvior
// when the ListRepoEmbeddingJobs method of the pbrent
// MockRepoEmbeddingJobsStore instbnce is invoked.
type RepoEmbeddingJobsStoreListRepoEmbeddingJobsFunc struct {
	defbultHook func(context.Context, ListOpts) ([]*RepoEmbeddingJob, error)
	hooks       []func(context.Context, ListOpts) ([]*RepoEmbeddingJob, error)
	history     []RepoEmbeddingJobsStoreListRepoEmbeddingJobsFuncCbll
	mutex       sync.Mutex
}

// ListRepoEmbeddingJobs delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRepoEmbeddingJobsStore) ListRepoEmbeddingJobs(v0 context.Context, v1 ListOpts) ([]*RepoEmbeddingJob, error) {
	r0, r1 := m.ListRepoEmbeddingJobsFunc.nextHook()(v0, v1)
	m.ListRepoEmbeddingJobsFunc.bppendCbll(RepoEmbeddingJobsStoreListRepoEmbeddingJobsFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// ListRepoEmbeddingJobs method of the pbrent MockRepoEmbeddingJobsStore
// instbnce is invoked bnd the hook queue is empty.
func (f *RepoEmbeddingJobsStoreListRepoEmbeddingJobsFunc) SetDefbultHook(hook func(context.Context, ListOpts) ([]*RepoEmbeddingJob, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ListRepoEmbeddingJobs method of the pbrent MockRepoEmbeddingJobsStore
// instbnce invokes the hook bt the front of the queue bnd discbrds it.
// After the queue is empty, the defbult hook function is invoked for bny
// future bction.
func (f *RepoEmbeddingJobsStoreListRepoEmbeddingJobsFunc) PushHook(hook func(context.Context, ListOpts) ([]*RepoEmbeddingJob, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RepoEmbeddingJobsStoreListRepoEmbeddingJobsFunc) SetDefbultReturn(r0 []*RepoEmbeddingJob, r1 error) {
	f.SetDefbultHook(func(context.Context, ListOpts) ([]*RepoEmbeddingJob, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RepoEmbeddingJobsStoreListRepoEmbeddingJobsFunc) PushReturn(r0 []*RepoEmbeddingJob, r1 error) {
	f.PushHook(func(context.Context, ListOpts) ([]*RepoEmbeddingJob, error) {
		return r0, r1
	})
}

func (f *RepoEmbeddingJobsStoreListRepoEmbeddingJobsFunc) nextHook() func(context.Context, ListOpts) ([]*RepoEmbeddingJob, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RepoEmbeddingJobsStoreListRepoEmbeddingJobsFunc) bppendCbll(r0 RepoEmbeddingJobsStoreListRepoEmbeddingJobsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// RepoEmbeddingJobsStoreListRepoEmbeddingJobsFuncCbll objects describing
// the invocbtions of this function.
func (f *RepoEmbeddingJobsStoreListRepoEmbeddingJobsFunc) History() []RepoEmbeddingJobsStoreListRepoEmbeddingJobsFuncCbll {
	f.mutex.Lock()
	history := mbke([]RepoEmbeddingJobsStoreListRepoEmbeddingJobsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RepoEmbeddingJobsStoreListRepoEmbeddingJobsFuncCbll is bn object thbt
// describes bn invocbtion of method ListRepoEmbeddingJobs on bn instbnce of
// MockRepoEmbeddingJobsStore.
type RepoEmbeddingJobsStoreListRepoEmbeddingJobsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 ListOpts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []*RepoEmbeddingJob
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RepoEmbeddingJobsStoreListRepoEmbeddingJobsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RepoEmbeddingJobsStoreListRepoEmbeddingJobsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// RepoEmbeddingJobsStoreTrbnsbctFunc describes the behbvior when the
// Trbnsbct method of the pbrent MockRepoEmbeddingJobsStore instbnce is
// invoked.
type RepoEmbeddingJobsStoreTrbnsbctFunc struct {
	defbultHook func(context.Context) (RepoEmbeddingJobsStore, error)
	hooks       []func(context.Context) (RepoEmbeddingJobsStore, error)
	history     []RepoEmbeddingJobsStoreTrbnsbctFuncCbll
	mutex       sync.Mutex
}

// Trbnsbct delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRepoEmbeddingJobsStore) Trbnsbct(v0 context.Context) (RepoEmbeddingJobsStore, error) {
	r0, r1 := m.TrbnsbctFunc.nextHook()(v0)
	m.TrbnsbctFunc.bppendCbll(RepoEmbeddingJobsStoreTrbnsbctFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Trbnsbct method of
// the pbrent MockRepoEmbeddingJobsStore instbnce is invoked bnd the hook
// queue is empty.
func (f *RepoEmbeddingJobsStoreTrbnsbctFunc) SetDefbultHook(hook func(context.Context) (RepoEmbeddingJobsStore, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Trbnsbct method of the pbrent MockRepoEmbeddingJobsStore instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *RepoEmbeddingJobsStoreTrbnsbctFunc) PushHook(hook func(context.Context) (RepoEmbeddingJobsStore, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RepoEmbeddingJobsStoreTrbnsbctFunc) SetDefbultReturn(r0 RepoEmbeddingJobsStore, r1 error) {
	f.SetDefbultHook(func(context.Context) (RepoEmbeddingJobsStore, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RepoEmbeddingJobsStoreTrbnsbctFunc) PushReturn(r0 RepoEmbeddingJobsStore, r1 error) {
	f.PushHook(func(context.Context) (RepoEmbeddingJobsStore, error) {
		return r0, r1
	})
}

func (f *RepoEmbeddingJobsStoreTrbnsbctFunc) nextHook() func(context.Context) (RepoEmbeddingJobsStore, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RepoEmbeddingJobsStoreTrbnsbctFunc) bppendCbll(r0 RepoEmbeddingJobsStoreTrbnsbctFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of RepoEmbeddingJobsStoreTrbnsbctFuncCbll
// objects describing the invocbtions of this function.
func (f *RepoEmbeddingJobsStoreTrbnsbctFunc) History() []RepoEmbeddingJobsStoreTrbnsbctFuncCbll {
	f.mutex.Lock()
	history := mbke([]RepoEmbeddingJobsStoreTrbnsbctFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RepoEmbeddingJobsStoreTrbnsbctFuncCbll is bn object thbt describes bn
// invocbtion of method Trbnsbct on bn instbnce of
// MockRepoEmbeddingJobsStore.
type RepoEmbeddingJobsStoreTrbnsbctFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 RepoEmbeddingJobsStore
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RepoEmbeddingJobsStoreTrbnsbctFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RepoEmbeddingJobsStoreTrbnsbctFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFunc describes the
// behbvior when the UpdbteRepoEmbeddingJobStbts method of the pbrent
// MockRepoEmbeddingJobsStore instbnce is invoked.
type RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFunc struct {
	defbultHook func(context.Context, int, *EmbedRepoStbts) error
	hooks       []func(context.Context, int, *EmbedRepoStbts) error
	history     []RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFuncCbll
	mutex       sync.Mutex
}

// UpdbteRepoEmbeddingJobStbts delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockRepoEmbeddingJobsStore) UpdbteRepoEmbeddingJobStbts(v0 context.Context, v1 int, v2 *EmbedRepoStbts) error {
	r0 := m.UpdbteRepoEmbeddingJobStbtsFunc.nextHook()(v0, v1, v2)
	m.UpdbteRepoEmbeddingJobStbtsFunc.bppendCbll(RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// UpdbteRepoEmbeddingJobStbts method of the pbrent
// MockRepoEmbeddingJobsStore instbnce is invoked bnd the hook queue is
// empty.
func (f *RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFunc) SetDefbultHook(hook func(context.Context, int, *EmbedRepoStbts) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UpdbteRepoEmbeddingJobStbts method of the pbrent
// MockRepoEmbeddingJobsStore instbnce invokes the hook bt the front of the
// queue bnd discbrds it. After the queue is empty, the defbult hook
// function is invoked for bny future bction.
func (f *RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFunc) PushHook(hook func(context.Context, int, *EmbedRepoStbts) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, int, *EmbedRepoStbts) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, int, *EmbedRepoStbts) error {
		return r0
	})
}

func (f *RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFunc) nextHook() func(context.Context, int, *EmbedRepoStbts) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFunc) bppendCbll(r0 RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of
// RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFuncCbll objects
// describing the invocbtions of this function.
func (f *RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFunc) History() []RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFuncCbll {
	f.mutex.Lock()
	history := mbke([]RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFuncCbll is bn object
// thbt describes bn invocbtion of method UpdbteRepoEmbeddingJobStbts on bn
// instbnce of MockRepoEmbeddingJobsStore.
type RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 *EmbedRepoStbts
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c RepoEmbeddingJobsStoreUpdbteRepoEmbeddingJobStbtsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
