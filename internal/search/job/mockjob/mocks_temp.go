// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge mockjob

import (
	"context"
	"sync"

	sebrch "github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	job "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	strebming "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	bttribute "go.opentelemetry.io/otel/bttribute"
)

// MockJob is b mock implementbtion of the Job interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job) used for unit
// testing.
type MockJob struct {
	// AttributesFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method Attributes.
	AttributesFunc *JobAttributesFunc
	// ChildrenFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Children.
	ChildrenFunc *JobChildrenFunc
	// MbpChildrenFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method MbpChildren.
	MbpChildrenFunc *JobMbpChildrenFunc
	// NbmeFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Nbme.
	NbmeFunc *JobNbmeFunc
	// RunFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Run.
	RunFunc *JobRunFunc
}

// NewMockJob crebtes b new mock of the Job interfbce. All methods return
// zero vblues for bll results, unless overwritten.
func NewMockJob() *MockJob {
	return &MockJob{
		AttributesFunc: &JobAttributesFunc{
			defbultHook: func(job.Verbosity) (r0 []bttribute.KeyVblue) {
				return
			},
		},
		ChildrenFunc: &JobChildrenFunc{
			defbultHook: func() (r0 []job.Describer) {
				return
			},
		},
		MbpChildrenFunc: &JobMbpChildrenFunc{
			defbultHook: func(job.MbpFunc) (r0 job.Job) {
				return
			},
		},
		NbmeFunc: &JobNbmeFunc{
			defbultHook: func() (r0 string) {
				return
			},
		},
		RunFunc: &JobRunFunc{
			defbultHook: func(context.Context, job.RuntimeClients, strebming.Sender) (r0 *sebrch.Alert, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockJob crebtes b new mock of the Job interfbce. All methods
// pbnic on invocbtion, unless overwritten.
func NewStrictMockJob() *MockJob {
	return &MockJob{
		AttributesFunc: &JobAttributesFunc{
			defbultHook: func(job.Verbosity) []bttribute.KeyVblue {
				pbnic("unexpected invocbtion of MockJob.Attributes")
			},
		},
		ChildrenFunc: &JobChildrenFunc{
			defbultHook: func() []job.Describer {
				pbnic("unexpected invocbtion of MockJob.Children")
			},
		},
		MbpChildrenFunc: &JobMbpChildrenFunc{
			defbultHook: func(job.MbpFunc) job.Job {
				pbnic("unexpected invocbtion of MockJob.MbpChildren")
			},
		},
		NbmeFunc: &JobNbmeFunc{
			defbultHook: func() string {
				pbnic("unexpected invocbtion of MockJob.Nbme")
			},
		},
		RunFunc: &JobRunFunc{
			defbultHook: func(context.Context, job.RuntimeClients, strebming.Sender) (*sebrch.Alert, error) {
				pbnic("unexpected invocbtion of MockJob.Run")
			},
		},
	}
}

// NewMockJobFrom crebtes b new mock of the MockJob interfbce. All methods
// delegbte to the given implementbtion, unless overwritten.
func NewMockJobFrom(i job.Job) *MockJob {
	return &MockJob{
		AttributesFunc: &JobAttributesFunc{
			defbultHook: i.Attributes,
		},
		ChildrenFunc: &JobChildrenFunc{
			defbultHook: i.Children,
		},
		MbpChildrenFunc: &JobMbpChildrenFunc{
			defbultHook: i.MbpChildren,
		},
		NbmeFunc: &JobNbmeFunc{
			defbultHook: i.Nbme,
		},
		RunFunc: &JobRunFunc{
			defbultHook: i.Run,
		},
	}
}

// JobAttributesFunc describes the behbvior when the Attributes method of
// the pbrent MockJob instbnce is invoked.
type JobAttributesFunc struct {
	defbultHook func(job.Verbosity) []bttribute.KeyVblue
	hooks       []func(job.Verbosity) []bttribute.KeyVblue
	history     []JobAttributesFuncCbll
	mutex       sync.Mutex
}

// Attributes delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockJob) Attributes(v0 job.Verbosity) []bttribute.KeyVblue {
	r0 := m.AttributesFunc.nextHook()(v0)
	m.AttributesFunc.bppendCbll(JobAttributesFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Attributes method of
// the pbrent MockJob instbnce is invoked bnd the hook queue is empty.
func (f *JobAttributesFunc) SetDefbultHook(hook func(job.Verbosity) []bttribute.KeyVblue) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Attributes method of the pbrent MockJob instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *JobAttributesFunc) PushHook(hook func(job.Verbosity) []bttribute.KeyVblue) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *JobAttributesFunc) SetDefbultReturn(r0 []bttribute.KeyVblue) {
	f.SetDefbultHook(func(job.Verbosity) []bttribute.KeyVblue {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *JobAttributesFunc) PushReturn(r0 []bttribute.KeyVblue) {
	f.PushHook(func(job.Verbosity) []bttribute.KeyVblue {
		return r0
	})
}

func (f *JobAttributesFunc) nextHook() func(job.Verbosity) []bttribute.KeyVblue {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *JobAttributesFunc) bppendCbll(r0 JobAttributesFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of JobAttributesFuncCbll objects describing
// the invocbtions of this function.
func (f *JobAttributesFunc) History() []JobAttributesFuncCbll {
	f.mutex.Lock()
	history := mbke([]JobAttributesFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// JobAttributesFuncCbll is bn object thbt describes bn invocbtion of method
// Attributes on bn instbnce of MockJob.
type JobAttributesFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 job.Verbosity
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []bttribute.KeyVblue
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c JobAttributesFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c JobAttributesFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// JobChildrenFunc describes the behbvior when the Children method of the
// pbrent MockJob instbnce is invoked.
type JobChildrenFunc struct {
	defbultHook func() []job.Describer
	hooks       []func() []job.Describer
	history     []JobChildrenFuncCbll
	mutex       sync.Mutex
}

// Children delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockJob) Children() []job.Describer {
	r0 := m.ChildrenFunc.nextHook()()
	m.ChildrenFunc.bppendCbll(JobChildrenFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Children method of
// the pbrent MockJob instbnce is invoked bnd the hook queue is empty.
func (f *JobChildrenFunc) SetDefbultHook(hook func() []job.Describer) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Children method of the pbrent MockJob instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *JobChildrenFunc) PushHook(hook func() []job.Describer) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *JobChildrenFunc) SetDefbultReturn(r0 []job.Describer) {
	f.SetDefbultHook(func() []job.Describer {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *JobChildrenFunc) PushReturn(r0 []job.Describer) {
	f.PushHook(func() []job.Describer {
		return r0
	})
}

func (f *JobChildrenFunc) nextHook() func() []job.Describer {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *JobChildrenFunc) bppendCbll(r0 JobChildrenFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of JobChildrenFuncCbll objects describing the
// invocbtions of this function.
func (f *JobChildrenFunc) History() []JobChildrenFuncCbll {
	f.mutex.Lock()
	history := mbke([]JobChildrenFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// JobChildrenFuncCbll is bn object thbt describes bn invocbtion of method
// Children on bn instbnce of MockJob.
type JobChildrenFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 []job.Describer
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c JobChildrenFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c JobChildrenFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// JobMbpChildrenFunc describes the behbvior when the MbpChildren method of
// the pbrent MockJob instbnce is invoked.
type JobMbpChildrenFunc struct {
	defbultHook func(job.MbpFunc) job.Job
	hooks       []func(job.MbpFunc) job.Job
	history     []JobMbpChildrenFuncCbll
	mutex       sync.Mutex
}

// MbpChildren delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockJob) MbpChildren(v0 job.MbpFunc) job.Job {
	r0 := m.MbpChildrenFunc.nextHook()(v0)
	m.MbpChildrenFunc.bppendCbll(JobMbpChildrenFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the MbpChildren method
// of the pbrent MockJob instbnce is invoked bnd the hook queue is empty.
func (f *JobMbpChildrenFunc) SetDefbultHook(hook func(job.MbpFunc) job.Job) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// MbpChildren method of the pbrent MockJob instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *JobMbpChildrenFunc) PushHook(hook func(job.MbpFunc) job.Job) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *JobMbpChildrenFunc) SetDefbultReturn(r0 job.Job) {
	f.SetDefbultHook(func(job.MbpFunc) job.Job {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *JobMbpChildrenFunc) PushReturn(r0 job.Job) {
	f.PushHook(func(job.MbpFunc) job.Job {
		return r0
	})
}

func (f *JobMbpChildrenFunc) nextHook() func(job.MbpFunc) job.Job {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *JobMbpChildrenFunc) bppendCbll(r0 JobMbpChildrenFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of JobMbpChildrenFuncCbll objects describing
// the invocbtions of this function.
func (f *JobMbpChildrenFunc) History() []JobMbpChildrenFuncCbll {
	f.mutex.Lock()
	history := mbke([]JobMbpChildrenFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// JobMbpChildrenFuncCbll is bn object thbt describes bn invocbtion of
// method MbpChildren on bn instbnce of MockJob.
type JobMbpChildrenFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 job.MbpFunc
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 job.Job
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c JobMbpChildrenFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c JobMbpChildrenFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// JobNbmeFunc describes the behbvior when the Nbme method of the pbrent
// MockJob instbnce is invoked.
type JobNbmeFunc struct {
	defbultHook func() string
	hooks       []func() string
	history     []JobNbmeFuncCbll
	mutex       sync.Mutex
}

// Nbme delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockJob) Nbme() string {
	r0 := m.NbmeFunc.nextHook()()
	m.NbmeFunc.bppendCbll(JobNbmeFuncCbll{r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Nbme method of the
// pbrent MockJob instbnce is invoked bnd the hook queue is empty.
func (f *JobNbmeFunc) SetDefbultHook(hook func() string) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Nbme method of the pbrent MockJob instbnce invokes the hook bt the front
// of the queue bnd discbrds it. After the queue is empty, the defbult hook
// function is invoked for bny future bction.
func (f *JobNbmeFunc) PushHook(hook func() string) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *JobNbmeFunc) SetDefbultReturn(r0 string) {
	f.SetDefbultHook(func() string {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *JobNbmeFunc) PushReturn(r0 string) {
	f.PushHook(func() string {
		return r0
	})
}

func (f *JobNbmeFunc) nextHook() func() string {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *JobNbmeFunc) bppendCbll(r0 JobNbmeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of JobNbmeFuncCbll objects describing the
// invocbtions of this function.
func (f *JobNbmeFunc) History() []JobNbmeFuncCbll {
	f.mutex.Lock()
	history := mbke([]JobNbmeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// JobNbmeFuncCbll is bn object thbt describes bn invocbtion of method Nbme
// on bn instbnce of MockJob.
type JobNbmeFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 string
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c JobNbmeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c JobNbmeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// JobRunFunc describes the behbvior when the Run method of the pbrent
// MockJob instbnce is invoked.
type JobRunFunc struct {
	defbultHook func(context.Context, job.RuntimeClients, strebming.Sender) (*sebrch.Alert, error)
	hooks       []func(context.Context, job.RuntimeClients, strebming.Sender) (*sebrch.Alert, error)
	history     []JobRunFuncCbll
	mutex       sync.Mutex
}

// Run delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockJob) Run(v0 context.Context, v1 job.RuntimeClients, v2 strebming.Sender) (*sebrch.Alert, error) {
	r0, r1 := m.RunFunc.nextHook()(v0, v1, v2)
	m.RunFunc.bppendCbll(JobRunFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Run method of the
// pbrent MockJob instbnce is invoked bnd the hook queue is empty.
func (f *JobRunFunc) SetDefbultHook(hook func(context.Context, job.RuntimeClients, strebming.Sender) (*sebrch.Alert, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Run method of the pbrent MockJob instbnce invokes the hook bt the front
// of the queue bnd discbrds it. After the queue is empty, the defbult hook
// function is invoked for bny future bction.
func (f *JobRunFunc) PushHook(hook func(context.Context, job.RuntimeClients, strebming.Sender) (*sebrch.Alert, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *JobRunFunc) SetDefbultReturn(r0 *sebrch.Alert, r1 error) {
	f.SetDefbultHook(func(context.Context, job.RuntimeClients, strebming.Sender) (*sebrch.Alert, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *JobRunFunc) PushReturn(r0 *sebrch.Alert, r1 error) {
	f.PushHook(func(context.Context, job.RuntimeClients, strebming.Sender) (*sebrch.Alert, error) {
		return r0, r1
	})
}

func (f *JobRunFunc) nextHook() func(context.Context, job.RuntimeClients, strebming.Sender) (*sebrch.Alert, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *JobRunFunc) bppendCbll(r0 JobRunFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of JobRunFuncCbll objects describing the
// invocbtions of this function.
func (f *JobRunFunc) History() []JobRunFuncCbll {
	f.mutex.Lock()
	history := mbke([]JobRunFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// JobRunFuncCbll is bn object thbt describes bn invocbtion of method Run on
// bn instbnce of MockJob.
type JobRunFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 job.RuntimeClients
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 strebming.Sender
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *sebrch.Alert
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c JobRunFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c JobRunFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}
