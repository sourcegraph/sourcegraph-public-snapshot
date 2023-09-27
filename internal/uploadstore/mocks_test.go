// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge uplobdstore

import (
	"context"
	"io"
	"sync"

	storbge "cloud.google.com/go/storbge"
	s3 "github.com/bws/bws-sdk-go-v2/service/s3"
)

// MockGcsAPI is b mock implementbtion of the gcsAPI interfbce (from the
// pbckbge github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore) used for
// unit testing.
type MockGcsAPI struct {
	// BucketFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Bucket.
	BucketFunc *GcsAPIBucketFunc
}

// NewMockGcsAPI crebtes b new mock of the gcsAPI interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockGcsAPI() *MockGcsAPI {
	return &MockGcsAPI{
		BucketFunc: &GcsAPIBucketFunc{
			defbultHook: func(string) (r0 gcsBucketHbndle) {
				return
			},
		},
	}
}

// NewStrictMockGcsAPI crebtes b new mock of the gcsAPI interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockGcsAPI() *MockGcsAPI {
	return &MockGcsAPI{
		BucketFunc: &GcsAPIBucketFunc{
			defbultHook: func(string) gcsBucketHbndle {
				pbnic("unexpected invocbtion of MockGcsAPI.Bucket")
			},
		},
	}
}

// surrogbteMockGcsAPI is b copy of the gcsAPI interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore). It is redefined
// here bs it is unexported in the source pbckbge.
type surrogbteMockGcsAPI interfbce {
	Bucket(string) gcsBucketHbndle
}

// NewMockGcsAPIFrom crebtes b new mock of the MockGcsAPI interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockGcsAPIFrom(i surrogbteMockGcsAPI) *MockGcsAPI {
	return &MockGcsAPI{
		BucketFunc: &GcsAPIBucketFunc{
			defbultHook: i.Bucket,
		},
	}
}

// GcsAPIBucketFunc describes the behbvior when the Bucket method of the
// pbrent MockGcsAPI instbnce is invoked.
type GcsAPIBucketFunc struct {
	defbultHook func(string) gcsBucketHbndle
	hooks       []func(string) gcsBucketHbndle
	history     []GcsAPIBucketFuncCbll
	mutex       sync.Mutex
}

// Bucket delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGcsAPI) Bucket(v0 string) gcsBucketHbndle {
	r0 := m.BucketFunc.nextHook()(v0)
	m.BucketFunc.bppendCbll(GcsAPIBucketFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Bucket method of the
// pbrent MockGcsAPI instbnce is invoked bnd the hook queue is empty.
func (f *GcsAPIBucketFunc) SetDefbultHook(hook func(string) gcsBucketHbndle) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Bucket method of the pbrent MockGcsAPI instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *GcsAPIBucketFunc) PushHook(hook func(string) gcsBucketHbndle) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GcsAPIBucketFunc) SetDefbultReturn(r0 gcsBucketHbndle) {
	f.SetDefbultHook(func(string) gcsBucketHbndle {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GcsAPIBucketFunc) PushReturn(r0 gcsBucketHbndle) {
	f.PushHook(func(string) gcsBucketHbndle {
		return r0
	})
}

func (f *GcsAPIBucketFunc) nextHook() func(string) gcsBucketHbndle {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GcsAPIBucketFunc) bppendCbll(r0 GcsAPIBucketFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GcsAPIBucketFuncCbll objects describing the
// invocbtions of this function.
func (f *GcsAPIBucketFunc) History() []GcsAPIBucketFuncCbll {
	f.mutex.Lock()
	history := mbke([]GcsAPIBucketFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GcsAPIBucketFuncCbll is bn object thbt describes bn invocbtion of method
// Bucket on bn instbnce of MockGcsAPI.
type GcsAPIBucketFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 gcsBucketHbndle
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GcsAPIBucketFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GcsAPIBucketFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockGcsBucketHbndle is b mock implementbtion of the gcsBucketHbndle
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore) used for unit
// testing.
type MockGcsBucketHbndle struct {
	// AttrsFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Attrs.
	AttrsFunc *GcsBucketHbndleAttrsFunc
	// CrebteFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Crebte.
	CrebteFunc *GcsBucketHbndleCrebteFunc
	// ObjectFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Object.
	ObjectFunc *GcsBucketHbndleObjectFunc
	// ObjectsFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Objects.
	ObjectsFunc *GcsBucketHbndleObjectsFunc
}

// NewMockGcsBucketHbndle crebtes b new mock of the gcsBucketHbndle
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockGcsBucketHbndle() *MockGcsBucketHbndle {
	return &MockGcsBucketHbndle{
		AttrsFunc: &GcsBucketHbndleAttrsFunc{
			defbultHook: func(context.Context) (r0 *storbge.BucketAttrs, r1 error) {
				return
			},
		},
		CrebteFunc: &GcsBucketHbndleCrebteFunc{
			defbultHook: func(context.Context, string, *storbge.BucketAttrs) (r0 error) {
				return
			},
		},
		ObjectFunc: &GcsBucketHbndleObjectFunc{
			defbultHook: func(string) (r0 gcsObjectHbndle) {
				return
			},
		},
		ObjectsFunc: &GcsBucketHbndleObjectsFunc{
			defbultHook: func(context.Context, *storbge.Query) (r0 gcsObjectIterbtor) {
				return
			},
		},
	}
}

// NewStrictMockGcsBucketHbndle crebtes b new mock of the gcsBucketHbndle
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockGcsBucketHbndle() *MockGcsBucketHbndle {
	return &MockGcsBucketHbndle{
		AttrsFunc: &GcsBucketHbndleAttrsFunc{
			defbultHook: func(context.Context) (*storbge.BucketAttrs, error) {
				pbnic("unexpected invocbtion of MockGcsBucketHbndle.Attrs")
			},
		},
		CrebteFunc: &GcsBucketHbndleCrebteFunc{
			defbultHook: func(context.Context, string, *storbge.BucketAttrs) error {
				pbnic("unexpected invocbtion of MockGcsBucketHbndle.Crebte")
			},
		},
		ObjectFunc: &GcsBucketHbndleObjectFunc{
			defbultHook: func(string) gcsObjectHbndle {
				pbnic("unexpected invocbtion of MockGcsBucketHbndle.Object")
			},
		},
		ObjectsFunc: &GcsBucketHbndleObjectsFunc{
			defbultHook: func(context.Context, *storbge.Query) gcsObjectIterbtor {
				pbnic("unexpected invocbtion of MockGcsBucketHbndle.Objects")
			},
		},
	}
}

// surrogbteMockGcsBucketHbndle is b copy of the gcsBucketHbndle interfbce
// (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore). It is redefined
// here bs it is unexported in the source pbckbge.
type surrogbteMockGcsBucketHbndle interfbce {
	Attrs(context.Context) (*storbge.BucketAttrs, error)
	Crebte(context.Context, string, *storbge.BucketAttrs) error
	Object(string) gcsObjectHbndle
	Objects(context.Context, *storbge.Query) gcsObjectIterbtor
}

// NewMockGcsBucketHbndleFrom crebtes b new mock of the MockGcsBucketHbndle
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockGcsBucketHbndleFrom(i surrogbteMockGcsBucketHbndle) *MockGcsBucketHbndle {
	return &MockGcsBucketHbndle{
		AttrsFunc: &GcsBucketHbndleAttrsFunc{
			defbultHook: i.Attrs,
		},
		CrebteFunc: &GcsBucketHbndleCrebteFunc{
			defbultHook: i.Crebte,
		},
		ObjectFunc: &GcsBucketHbndleObjectFunc{
			defbultHook: i.Object,
		},
		ObjectsFunc: &GcsBucketHbndleObjectsFunc{
			defbultHook: i.Objects,
		},
	}
}

// GcsBucketHbndleAttrsFunc describes the behbvior when the Attrs method of
// the pbrent MockGcsBucketHbndle instbnce is invoked.
type GcsBucketHbndleAttrsFunc struct {
	defbultHook func(context.Context) (*storbge.BucketAttrs, error)
	hooks       []func(context.Context) (*storbge.BucketAttrs, error)
	history     []GcsBucketHbndleAttrsFuncCbll
	mutex       sync.Mutex
}

// Attrs delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGcsBucketHbndle) Attrs(v0 context.Context) (*storbge.BucketAttrs, error) {
	r0, r1 := m.AttrsFunc.nextHook()(v0)
	m.AttrsFunc.bppendCbll(GcsBucketHbndleAttrsFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Attrs method of the
// pbrent MockGcsBucketHbndle instbnce is invoked bnd the hook queue is
// empty.
func (f *GcsBucketHbndleAttrsFunc) SetDefbultHook(hook func(context.Context) (*storbge.BucketAttrs, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Attrs method of the pbrent MockGcsBucketHbndle instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GcsBucketHbndleAttrsFunc) PushHook(hook func(context.Context) (*storbge.BucketAttrs, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GcsBucketHbndleAttrsFunc) SetDefbultReturn(r0 *storbge.BucketAttrs, r1 error) {
	f.SetDefbultHook(func(context.Context) (*storbge.BucketAttrs, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GcsBucketHbndleAttrsFunc) PushReturn(r0 *storbge.BucketAttrs, r1 error) {
	f.PushHook(func(context.Context) (*storbge.BucketAttrs, error) {
		return r0, r1
	})
}

func (f *GcsBucketHbndleAttrsFunc) nextHook() func(context.Context) (*storbge.BucketAttrs, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GcsBucketHbndleAttrsFunc) bppendCbll(r0 GcsBucketHbndleAttrsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GcsBucketHbndleAttrsFuncCbll objects
// describing the invocbtions of this function.
func (f *GcsBucketHbndleAttrsFunc) History() []GcsBucketHbndleAttrsFuncCbll {
	f.mutex.Lock()
	history := mbke([]GcsBucketHbndleAttrsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GcsBucketHbndleAttrsFuncCbll is bn object thbt describes bn invocbtion of
// method Attrs on bn instbnce of MockGcsBucketHbndle.
type GcsBucketHbndleAttrsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *storbge.BucketAttrs
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GcsBucketHbndleAttrsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GcsBucketHbndleAttrsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GcsBucketHbndleCrebteFunc describes the behbvior when the Crebte method
// of the pbrent MockGcsBucketHbndle instbnce is invoked.
type GcsBucketHbndleCrebteFunc struct {
	defbultHook func(context.Context, string, *storbge.BucketAttrs) error
	hooks       []func(context.Context, string, *storbge.BucketAttrs) error
	history     []GcsBucketHbndleCrebteFuncCbll
	mutex       sync.Mutex
}

// Crebte delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGcsBucketHbndle) Crebte(v0 context.Context, v1 string, v2 *storbge.BucketAttrs) error {
	r0 := m.CrebteFunc.nextHook()(v0, v1, v2)
	m.CrebteFunc.bppendCbll(GcsBucketHbndleCrebteFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Crebte method of the
// pbrent MockGcsBucketHbndle instbnce is invoked bnd the hook queue is
// empty.
func (f *GcsBucketHbndleCrebteFunc) SetDefbultHook(hook func(context.Context, string, *storbge.BucketAttrs) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Crebte method of the pbrent MockGcsBucketHbndle instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GcsBucketHbndleCrebteFunc) PushHook(hook func(context.Context, string, *storbge.BucketAttrs) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GcsBucketHbndleCrebteFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, string, *storbge.BucketAttrs) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GcsBucketHbndleCrebteFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, string, *storbge.BucketAttrs) error {
		return r0
	})
}

func (f *GcsBucketHbndleCrebteFunc) nextHook() func(context.Context, string, *storbge.BucketAttrs) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GcsBucketHbndleCrebteFunc) bppendCbll(r0 GcsBucketHbndleCrebteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GcsBucketHbndleCrebteFuncCbll objects
// describing the invocbtions of this function.
func (f *GcsBucketHbndleCrebteFunc) History() []GcsBucketHbndleCrebteFuncCbll {
	f.mutex.Lock()
	history := mbke([]GcsBucketHbndleCrebteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GcsBucketHbndleCrebteFuncCbll is bn object thbt describes bn invocbtion
// of method Crebte on bn instbnce of MockGcsBucketHbndle.
type GcsBucketHbndleCrebteFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 *storbge.BucketAttrs
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GcsBucketHbndleCrebteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GcsBucketHbndleCrebteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GcsBucketHbndleObjectFunc describes the behbvior when the Object method
// of the pbrent MockGcsBucketHbndle instbnce is invoked.
type GcsBucketHbndleObjectFunc struct {
	defbultHook func(string) gcsObjectHbndle
	hooks       []func(string) gcsObjectHbndle
	history     []GcsBucketHbndleObjectFuncCbll
	mutex       sync.Mutex
}

// Object delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGcsBucketHbndle) Object(v0 string) gcsObjectHbndle {
	r0 := m.ObjectFunc.nextHook()(v0)
	m.ObjectFunc.bppendCbll(GcsBucketHbndleObjectFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Object method of the
// pbrent MockGcsBucketHbndle instbnce is invoked bnd the hook queue is
// empty.
func (f *GcsBucketHbndleObjectFunc) SetDefbultHook(hook func(string) gcsObjectHbndle) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Object method of the pbrent MockGcsBucketHbndle instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GcsBucketHbndleObjectFunc) PushHook(hook func(string) gcsObjectHbndle) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GcsBucketHbndleObjectFunc) SetDefbultReturn(r0 gcsObjectHbndle) {
	f.SetDefbultHook(func(string) gcsObjectHbndle {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GcsBucketHbndleObjectFunc) PushReturn(r0 gcsObjectHbndle) {
	f.PushHook(func(string) gcsObjectHbndle {
		return r0
	})
}

func (f *GcsBucketHbndleObjectFunc) nextHook() func(string) gcsObjectHbndle {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GcsBucketHbndleObjectFunc) bppendCbll(r0 GcsBucketHbndleObjectFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GcsBucketHbndleObjectFuncCbll objects
// describing the invocbtions of this function.
func (f *GcsBucketHbndleObjectFunc) History() []GcsBucketHbndleObjectFuncCbll {
	f.mutex.Lock()
	history := mbke([]GcsBucketHbndleObjectFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GcsBucketHbndleObjectFuncCbll is bn object thbt describes bn invocbtion
// of method Object on bn instbnce of MockGcsBucketHbndle.
type GcsBucketHbndleObjectFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 gcsObjectHbndle
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GcsBucketHbndleObjectFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GcsBucketHbndleObjectFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GcsBucketHbndleObjectsFunc describes the behbvior when the Objects method
// of the pbrent MockGcsBucketHbndle instbnce is invoked.
type GcsBucketHbndleObjectsFunc struct {
	defbultHook func(context.Context, *storbge.Query) gcsObjectIterbtor
	hooks       []func(context.Context, *storbge.Query) gcsObjectIterbtor
	history     []GcsBucketHbndleObjectsFuncCbll
	mutex       sync.Mutex
}

// Objects delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGcsBucketHbndle) Objects(v0 context.Context, v1 *storbge.Query) gcsObjectIterbtor {
	r0 := m.ObjectsFunc.nextHook()(v0, v1)
	m.ObjectsFunc.bppendCbll(GcsBucketHbndleObjectsFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Objects method of
// the pbrent MockGcsBucketHbndle instbnce is invoked bnd the hook queue is
// empty.
func (f *GcsBucketHbndleObjectsFunc) SetDefbultHook(hook func(context.Context, *storbge.Query) gcsObjectIterbtor) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Objects method of the pbrent MockGcsBucketHbndle instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GcsBucketHbndleObjectsFunc) PushHook(hook func(context.Context, *storbge.Query) gcsObjectIterbtor) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GcsBucketHbndleObjectsFunc) SetDefbultReturn(r0 gcsObjectIterbtor) {
	f.SetDefbultHook(func(context.Context, *storbge.Query) gcsObjectIterbtor {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GcsBucketHbndleObjectsFunc) PushReturn(r0 gcsObjectIterbtor) {
	f.PushHook(func(context.Context, *storbge.Query) gcsObjectIterbtor {
		return r0
	})
}

func (f *GcsBucketHbndleObjectsFunc) nextHook() func(context.Context, *storbge.Query) gcsObjectIterbtor {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GcsBucketHbndleObjectsFunc) bppendCbll(r0 GcsBucketHbndleObjectsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GcsBucketHbndleObjectsFuncCbll objects
// describing the invocbtions of this function.
func (f *GcsBucketHbndleObjectsFunc) History() []GcsBucketHbndleObjectsFuncCbll {
	f.mutex.Lock()
	history := mbke([]GcsBucketHbndleObjectsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GcsBucketHbndleObjectsFuncCbll is bn object thbt describes bn invocbtion
// of method Objects on bn instbnce of MockGcsBucketHbndle.
type GcsBucketHbndleObjectsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *storbge.Query
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 gcsObjectIterbtor
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GcsBucketHbndleObjectsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GcsBucketHbndleObjectsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockGcsComposer is b mock implementbtion of the gcsComposer interfbce
// (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore) used for unit
// testing.
type MockGcsComposer struct {
	// RunFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Run.
	RunFunc *GcsComposerRunFunc
}

// NewMockGcsComposer crebtes b new mock of the gcsComposer interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockGcsComposer() *MockGcsComposer {
	return &MockGcsComposer{
		RunFunc: &GcsComposerRunFunc{
			defbultHook: func(context.Context) (r0 *storbge.ObjectAttrs, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockGcsComposer crebtes b new mock of the gcsComposer interfbce.
// All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockGcsComposer() *MockGcsComposer {
	return &MockGcsComposer{
		RunFunc: &GcsComposerRunFunc{
			defbultHook: func(context.Context) (*storbge.ObjectAttrs, error) {
				pbnic("unexpected invocbtion of MockGcsComposer.Run")
			},
		},
	}
}

// surrogbteMockGcsComposer is b copy of the gcsComposer interfbce (from the
// pbckbge github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore). It is
// redefined here bs it is unexported in the source pbckbge.
type surrogbteMockGcsComposer interfbce {
	Run(context.Context) (*storbge.ObjectAttrs, error)
}

// NewMockGcsComposerFrom crebtes b new mock of the MockGcsComposer
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockGcsComposerFrom(i surrogbteMockGcsComposer) *MockGcsComposer {
	return &MockGcsComposer{
		RunFunc: &GcsComposerRunFunc{
			defbultHook: i.Run,
		},
	}
}

// GcsComposerRunFunc describes the behbvior when the Run method of the
// pbrent MockGcsComposer instbnce is invoked.
type GcsComposerRunFunc struct {
	defbultHook func(context.Context) (*storbge.ObjectAttrs, error)
	hooks       []func(context.Context) (*storbge.ObjectAttrs, error)
	history     []GcsComposerRunFuncCbll
	mutex       sync.Mutex
}

// Run delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGcsComposer) Run(v0 context.Context) (*storbge.ObjectAttrs, error) {
	r0, r1 := m.RunFunc.nextHook()(v0)
	m.RunFunc.bppendCbll(GcsComposerRunFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Run method of the
// pbrent MockGcsComposer instbnce is invoked bnd the hook queue is empty.
func (f *GcsComposerRunFunc) SetDefbultHook(hook func(context.Context) (*storbge.ObjectAttrs, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Run method of the pbrent MockGcsComposer instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *GcsComposerRunFunc) PushHook(hook func(context.Context) (*storbge.ObjectAttrs, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GcsComposerRunFunc) SetDefbultReturn(r0 *storbge.ObjectAttrs, r1 error) {
	f.SetDefbultHook(func(context.Context) (*storbge.ObjectAttrs, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GcsComposerRunFunc) PushReturn(r0 *storbge.ObjectAttrs, r1 error) {
	f.PushHook(func(context.Context) (*storbge.ObjectAttrs, error) {
		return r0, r1
	})
}

func (f *GcsComposerRunFunc) nextHook() func(context.Context) (*storbge.ObjectAttrs, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GcsComposerRunFunc) bppendCbll(r0 GcsComposerRunFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GcsComposerRunFuncCbll objects describing
// the invocbtions of this function.
func (f *GcsComposerRunFunc) History() []GcsComposerRunFuncCbll {
	f.mutex.Lock()
	history := mbke([]GcsComposerRunFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GcsComposerRunFuncCbll is bn object thbt describes bn invocbtion of
// method Run on bn instbnce of MockGcsComposer.
type GcsComposerRunFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *storbge.ObjectAttrs
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GcsComposerRunFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GcsComposerRunFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockGcsObjectHbndle is b mock implementbtion of the gcsObjectHbndle
// interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore) used for unit
// testing.
type MockGcsObjectHbndle struct {
	// ComposerFromFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method ComposerFrom.
	ComposerFromFunc *GcsObjectHbndleComposerFromFunc
	// DeleteFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Delete.
	DeleteFunc *GcsObjectHbndleDeleteFunc
	// NewRbngeRebderFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method NewRbngeRebder.
	NewRbngeRebderFunc *GcsObjectHbndleNewRbngeRebderFunc
	// NewWriterFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method NewWriter.
	NewWriterFunc *GcsObjectHbndleNewWriterFunc
}

// NewMockGcsObjectHbndle crebtes b new mock of the gcsObjectHbndle
// interfbce. All methods return zero vblues for bll results, unless
// overwritten.
func NewMockGcsObjectHbndle() *MockGcsObjectHbndle {
	return &MockGcsObjectHbndle{
		ComposerFromFunc: &GcsObjectHbndleComposerFromFunc{
			defbultHook: func(...gcsObjectHbndle) (r0 gcsComposer) {
				return
			},
		},
		DeleteFunc: &GcsObjectHbndleDeleteFunc{
			defbultHook: func(context.Context) (r0 error) {
				return
			},
		},
		NewRbngeRebderFunc: &GcsObjectHbndleNewRbngeRebderFunc{
			defbultHook: func(context.Context, int64, int64) (r0 io.RebdCloser, r1 error) {
				return
			},
		},
		NewWriterFunc: &GcsObjectHbndleNewWriterFunc{
			defbultHook: func(context.Context) (r0 io.WriteCloser) {
				return
			},
		},
	}
}

// NewStrictMockGcsObjectHbndle crebtes b new mock of the gcsObjectHbndle
// interfbce. All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockGcsObjectHbndle() *MockGcsObjectHbndle {
	return &MockGcsObjectHbndle{
		ComposerFromFunc: &GcsObjectHbndleComposerFromFunc{
			defbultHook: func(...gcsObjectHbndle) gcsComposer {
				pbnic("unexpected invocbtion of MockGcsObjectHbndle.ComposerFrom")
			},
		},
		DeleteFunc: &GcsObjectHbndleDeleteFunc{
			defbultHook: func(context.Context) error {
				pbnic("unexpected invocbtion of MockGcsObjectHbndle.Delete")
			},
		},
		NewRbngeRebderFunc: &GcsObjectHbndleNewRbngeRebderFunc{
			defbultHook: func(context.Context, int64, int64) (io.RebdCloser, error) {
				pbnic("unexpected invocbtion of MockGcsObjectHbndle.NewRbngeRebder")
			},
		},
		NewWriterFunc: &GcsObjectHbndleNewWriterFunc{
			defbultHook: func(context.Context) io.WriteCloser {
				pbnic("unexpected invocbtion of MockGcsObjectHbndle.NewWriter")
			},
		},
	}
}

// surrogbteMockGcsObjectHbndle is b copy of the gcsObjectHbndle interfbce
// (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore). It is redefined
// here bs it is unexported in the source pbckbge.
type surrogbteMockGcsObjectHbndle interfbce {
	ComposerFrom(...gcsObjectHbndle) gcsComposer
	Delete(context.Context) error
	NewRbngeRebder(context.Context, int64, int64) (io.RebdCloser, error)
	NewWriter(context.Context) io.WriteCloser
}

// NewMockGcsObjectHbndleFrom crebtes b new mock of the MockGcsObjectHbndle
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockGcsObjectHbndleFrom(i surrogbteMockGcsObjectHbndle) *MockGcsObjectHbndle {
	return &MockGcsObjectHbndle{
		ComposerFromFunc: &GcsObjectHbndleComposerFromFunc{
			defbultHook: i.ComposerFrom,
		},
		DeleteFunc: &GcsObjectHbndleDeleteFunc{
			defbultHook: i.Delete,
		},
		NewRbngeRebderFunc: &GcsObjectHbndleNewRbngeRebderFunc{
			defbultHook: i.NewRbngeRebder,
		},
		NewWriterFunc: &GcsObjectHbndleNewWriterFunc{
			defbultHook: i.NewWriter,
		},
	}
}

// GcsObjectHbndleComposerFromFunc describes the behbvior when the
// ComposerFrom method of the pbrent MockGcsObjectHbndle instbnce is
// invoked.
type GcsObjectHbndleComposerFromFunc struct {
	defbultHook func(...gcsObjectHbndle) gcsComposer
	hooks       []func(...gcsObjectHbndle) gcsComposer
	history     []GcsObjectHbndleComposerFromFuncCbll
	mutex       sync.Mutex
}

// ComposerFrom delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGcsObjectHbndle) ComposerFrom(v0 ...gcsObjectHbndle) gcsComposer {
	r0 := m.ComposerFromFunc.nextHook()(v0...)
	m.ComposerFromFunc.bppendCbll(GcsObjectHbndleComposerFromFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the ComposerFrom method
// of the pbrent MockGcsObjectHbndle instbnce is invoked bnd the hook queue
// is empty.
func (f *GcsObjectHbndleComposerFromFunc) SetDefbultHook(hook func(...gcsObjectHbndle) gcsComposer) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// ComposerFrom method of the pbrent MockGcsObjectHbndle instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GcsObjectHbndleComposerFromFunc) PushHook(hook func(...gcsObjectHbndle) gcsComposer) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GcsObjectHbndleComposerFromFunc) SetDefbultReturn(r0 gcsComposer) {
	f.SetDefbultHook(func(...gcsObjectHbndle) gcsComposer {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GcsObjectHbndleComposerFromFunc) PushReturn(r0 gcsComposer) {
	f.PushHook(func(...gcsObjectHbndle) gcsComposer {
		return r0
	})
}

func (f *GcsObjectHbndleComposerFromFunc) nextHook() func(...gcsObjectHbndle) gcsComposer {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GcsObjectHbndleComposerFromFunc) bppendCbll(r0 GcsObjectHbndleComposerFromFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GcsObjectHbndleComposerFromFuncCbll objects
// describing the invocbtions of this function.
func (f *GcsObjectHbndleComposerFromFunc) History() []GcsObjectHbndleComposerFromFuncCbll {
	f.mutex.Lock()
	history := mbke([]GcsObjectHbndleComposerFromFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GcsObjectHbndleComposerFromFuncCbll is bn object thbt describes bn
// invocbtion of method ComposerFrom on bn instbnce of MockGcsObjectHbndle.
type GcsObjectHbndleComposerFromFuncCbll struct {
	// Arg0 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg0 []gcsObjectHbndle
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 gcsComposer
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c GcsObjectHbndleComposerFromFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg0 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GcsObjectHbndleComposerFromFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GcsObjectHbndleDeleteFunc describes the behbvior when the Delete method
// of the pbrent MockGcsObjectHbndle instbnce is invoked.
type GcsObjectHbndleDeleteFunc struct {
	defbultHook func(context.Context) error
	hooks       []func(context.Context) error
	history     []GcsObjectHbndleDeleteFuncCbll
	mutex       sync.Mutex
}

// Delete delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGcsObjectHbndle) Delete(v0 context.Context) error {
	r0 := m.DeleteFunc.nextHook()(v0)
	m.DeleteFunc.bppendCbll(GcsObjectHbndleDeleteFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Delete method of the
// pbrent MockGcsObjectHbndle instbnce is invoked bnd the hook queue is
// empty.
func (f *GcsObjectHbndleDeleteFunc) SetDefbultHook(hook func(context.Context) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Delete method of the pbrent MockGcsObjectHbndle instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *GcsObjectHbndleDeleteFunc) PushHook(hook func(context.Context) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GcsObjectHbndleDeleteFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GcsObjectHbndleDeleteFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context) error {
		return r0
	})
}

func (f *GcsObjectHbndleDeleteFunc) nextHook() func(context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GcsObjectHbndleDeleteFunc) bppendCbll(r0 GcsObjectHbndleDeleteFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GcsObjectHbndleDeleteFuncCbll objects
// describing the invocbtions of this function.
func (f *GcsObjectHbndleDeleteFunc) History() []GcsObjectHbndleDeleteFuncCbll {
	f.mutex.Lock()
	history := mbke([]GcsObjectHbndleDeleteFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GcsObjectHbndleDeleteFuncCbll is bn object thbt describes bn invocbtion
// of method Delete on bn instbnce of MockGcsObjectHbndle.
type GcsObjectHbndleDeleteFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GcsObjectHbndleDeleteFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GcsObjectHbndleDeleteFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// GcsObjectHbndleNewRbngeRebderFunc describes the behbvior when the
// NewRbngeRebder method of the pbrent MockGcsObjectHbndle instbnce is
// invoked.
type GcsObjectHbndleNewRbngeRebderFunc struct {
	defbultHook func(context.Context, int64, int64) (io.RebdCloser, error)
	hooks       []func(context.Context, int64, int64) (io.RebdCloser, error)
	history     []GcsObjectHbndleNewRbngeRebderFuncCbll
	mutex       sync.Mutex
}

// NewRbngeRebder delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGcsObjectHbndle) NewRbngeRebder(v0 context.Context, v1 int64, v2 int64) (io.RebdCloser, error) {
	r0, r1 := m.NewRbngeRebderFunc.nextHook()(v0, v1, v2)
	m.NewRbngeRebderFunc.bppendCbll(GcsObjectHbndleNewRbngeRebderFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the NewRbngeRebder
// method of the pbrent MockGcsObjectHbndle instbnce is invoked bnd the hook
// queue is empty.
func (f *GcsObjectHbndleNewRbngeRebderFunc) SetDefbultHook(hook func(context.Context, int64, int64) (io.RebdCloser, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// NewRbngeRebder method of the pbrent MockGcsObjectHbndle instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *GcsObjectHbndleNewRbngeRebderFunc) PushHook(hook func(context.Context, int64, int64) (io.RebdCloser, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GcsObjectHbndleNewRbngeRebderFunc) SetDefbultReturn(r0 io.RebdCloser, r1 error) {
	f.SetDefbultHook(func(context.Context, int64, int64) (io.RebdCloser, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GcsObjectHbndleNewRbngeRebderFunc) PushReturn(r0 io.RebdCloser, r1 error) {
	f.PushHook(func(context.Context, int64, int64) (io.RebdCloser, error) {
		return r0, r1
	})
}

func (f *GcsObjectHbndleNewRbngeRebderFunc) nextHook() func(context.Context, int64, int64) (io.RebdCloser, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GcsObjectHbndleNewRbngeRebderFunc) bppendCbll(r0 GcsObjectHbndleNewRbngeRebderFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GcsObjectHbndleNewRbngeRebderFuncCbll
// objects describing the invocbtions of this function.
func (f *GcsObjectHbndleNewRbngeRebderFunc) History() []GcsObjectHbndleNewRbngeRebderFuncCbll {
	f.mutex.Lock()
	history := mbke([]GcsObjectHbndleNewRbngeRebderFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GcsObjectHbndleNewRbngeRebderFuncCbll is bn object thbt describes bn
// invocbtion of method NewRbngeRebder on bn instbnce of
// MockGcsObjectHbndle.
type GcsObjectHbndleNewRbngeRebderFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int64
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int64
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 io.RebdCloser
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GcsObjectHbndleNewRbngeRebderFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GcsObjectHbndleNewRbngeRebderFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// GcsObjectHbndleNewWriterFunc describes the behbvior when the NewWriter
// method of the pbrent MockGcsObjectHbndle instbnce is invoked.
type GcsObjectHbndleNewWriterFunc struct {
	defbultHook func(context.Context) io.WriteCloser
	hooks       []func(context.Context) io.WriteCloser
	history     []GcsObjectHbndleNewWriterFuncCbll
	mutex       sync.Mutex
}

// NewWriter delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockGcsObjectHbndle) NewWriter(v0 context.Context) io.WriteCloser {
	r0 := m.NewWriterFunc.nextHook()(v0)
	m.NewWriterFunc.bppendCbll(GcsObjectHbndleNewWriterFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the NewWriter method of
// the pbrent MockGcsObjectHbndle instbnce is invoked bnd the hook queue is
// empty.
func (f *GcsObjectHbndleNewWriterFunc) SetDefbultHook(hook func(context.Context) io.WriteCloser) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// NewWriter method of the pbrent MockGcsObjectHbndle instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *GcsObjectHbndleNewWriterFunc) PushHook(hook func(context.Context) io.WriteCloser) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *GcsObjectHbndleNewWriterFunc) SetDefbultReturn(r0 io.WriteCloser) {
	f.SetDefbultHook(func(context.Context) io.WriteCloser {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *GcsObjectHbndleNewWriterFunc) PushReturn(r0 io.WriteCloser) {
	f.PushHook(func(context.Context) io.WriteCloser {
		return r0
	})
}

func (f *GcsObjectHbndleNewWriterFunc) nextHook() func(context.Context) io.WriteCloser {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *GcsObjectHbndleNewWriterFunc) bppendCbll(r0 GcsObjectHbndleNewWriterFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of GcsObjectHbndleNewWriterFuncCbll objects
// describing the invocbtions of this function.
func (f *GcsObjectHbndleNewWriterFunc) History() []GcsObjectHbndleNewWriterFuncCbll {
	f.mutex.Lock()
	history := mbke([]GcsObjectHbndleNewWriterFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// GcsObjectHbndleNewWriterFuncCbll is bn object thbt describes bn
// invocbtion of method NewWriter on bn instbnce of MockGcsObjectHbndle.
type GcsObjectHbndleNewWriterFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 io.WriteCloser
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c GcsObjectHbndleNewWriterFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c GcsObjectHbndleNewWriterFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// MockS3API is b mock implementbtion of the s3API interfbce (from the
// pbckbge github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore) used for
// unit testing.
type MockS3API struct {
	// AbortMultipbrtUplobdFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method AbortMultipbrtUplobd.
	AbortMultipbrtUplobdFunc *S3APIAbortMultipbrtUplobdFunc
	// CompleteMultipbrtUplobdFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CompleteMultipbrtUplobd.
	CompleteMultipbrtUplobdFunc *S3APICompleteMultipbrtUplobdFunc
	// CrebteBucketFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method CrebteBucket.
	CrebteBucketFunc *S3APICrebteBucketFunc
	// CrebteMultipbrtUplobdFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method CrebteMultipbrtUplobd.
	CrebteMultipbrtUplobdFunc *S3APICrebteMultipbrtUplobdFunc
	// DeleteObjectFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method DeleteObject.
	DeleteObjectFunc *S3APIDeleteObjectFunc
	// DeleteObjectsFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method DeleteObjects.
	DeleteObjectsFunc *S3APIDeleteObjectsFunc
	// GetObjectFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method GetObject.
	GetObjectFunc *S3APIGetObjectFunc
	// HebdObjectFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method HebdObject.
	HebdObjectFunc *S3APIHebdObjectFunc
	// NewListObjectsV2PbginbtorFunc is bn instbnce of b mock function
	// object controlling the behbvior of the method
	// NewListObjectsV2Pbginbtor.
	NewListObjectsV2PbginbtorFunc *S3APINewListObjectsV2PbginbtorFunc
	// UplobdPbrtCopyFunc is bn instbnce of b mock function object
	// controlling the behbvior of the method UplobdPbrtCopy.
	UplobdPbrtCopyFunc *S3APIUplobdPbrtCopyFunc
}

// NewMockS3API crebtes b new mock of the s3API interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockS3API() *MockS3API {
	return &MockS3API{
		AbortMultipbrtUplobdFunc: &S3APIAbortMultipbrtUplobdFunc{
			defbultHook: func(context.Context, *s3.AbortMultipbrtUplobdInput) (r0 *s3.AbortMultipbrtUplobdOutput, r1 error) {
				return
			},
		},
		CompleteMultipbrtUplobdFunc: &S3APICompleteMultipbrtUplobdFunc{
			defbultHook: func(context.Context, *s3.CompleteMultipbrtUplobdInput) (r0 *s3.CompleteMultipbrtUplobdOutput, r1 error) {
				return
			},
		},
		CrebteBucketFunc: &S3APICrebteBucketFunc{
			defbultHook: func(context.Context, *s3.CrebteBucketInput) (r0 *s3.CrebteBucketOutput, r1 error) {
				return
			},
		},
		CrebteMultipbrtUplobdFunc: &S3APICrebteMultipbrtUplobdFunc{
			defbultHook: func(context.Context, *s3.CrebteMultipbrtUplobdInput) (r0 *s3.CrebteMultipbrtUplobdOutput, r1 error) {
				return
			},
		},
		DeleteObjectFunc: &S3APIDeleteObjectFunc{
			defbultHook: func(context.Context, *s3.DeleteObjectInput) (r0 *s3.DeleteObjectOutput, r1 error) {
				return
			},
		},
		DeleteObjectsFunc: &S3APIDeleteObjectsFunc{
			defbultHook: func(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (r0 *s3.DeleteObjectsOutput, r1 error) {
				return
			},
		},
		GetObjectFunc: &S3APIGetObjectFunc{
			defbultHook: func(context.Context, *s3.GetObjectInput) (r0 *s3.GetObjectOutput, r1 error) {
				return
			},
		},
		HebdObjectFunc: &S3APIHebdObjectFunc{
			defbultHook: func(context.Context, *s3.HebdObjectInput) (r0 *s3.HebdObjectOutput, r1 error) {
				return
			},
		},
		NewListObjectsV2PbginbtorFunc: &S3APINewListObjectsV2PbginbtorFunc{
			defbultHook: func(*s3.ListObjectsV2Input) (r0 *s3.ListObjectsV2Pbginbtor) {
				return
			},
		},
		UplobdPbrtCopyFunc: &S3APIUplobdPbrtCopyFunc{
			defbultHook: func(context.Context, *s3.UplobdPbrtCopyInput) (r0 *s3.UplobdPbrtCopyOutput, r1 error) {
				return
			},
		},
	}
}

// NewStrictMockS3API crebtes b new mock of the s3API interfbce. All methods
// pbnic on invocbtion, unless overwritten.
func NewStrictMockS3API() *MockS3API {
	return &MockS3API{
		AbortMultipbrtUplobdFunc: &S3APIAbortMultipbrtUplobdFunc{
			defbultHook: func(context.Context, *s3.AbortMultipbrtUplobdInput) (*s3.AbortMultipbrtUplobdOutput, error) {
				pbnic("unexpected invocbtion of MockS3API.AbortMultipbrtUplobd")
			},
		},
		CompleteMultipbrtUplobdFunc: &S3APICompleteMultipbrtUplobdFunc{
			defbultHook: func(context.Context, *s3.CompleteMultipbrtUplobdInput) (*s3.CompleteMultipbrtUplobdOutput, error) {
				pbnic("unexpected invocbtion of MockS3API.CompleteMultipbrtUplobd")
			},
		},
		CrebteBucketFunc: &S3APICrebteBucketFunc{
			defbultHook: func(context.Context, *s3.CrebteBucketInput) (*s3.CrebteBucketOutput, error) {
				pbnic("unexpected invocbtion of MockS3API.CrebteBucket")
			},
		},
		CrebteMultipbrtUplobdFunc: &S3APICrebteMultipbrtUplobdFunc{
			defbultHook: func(context.Context, *s3.CrebteMultipbrtUplobdInput) (*s3.CrebteMultipbrtUplobdOutput, error) {
				pbnic("unexpected invocbtion of MockS3API.CrebteMultipbrtUplobd")
			},
		},
		DeleteObjectFunc: &S3APIDeleteObjectFunc{
			defbultHook: func(context.Context, *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
				pbnic("unexpected invocbtion of MockS3API.DeleteObject")
			},
		},
		DeleteObjectsFunc: &S3APIDeleteObjectsFunc{
			defbultHook: func(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
				pbnic("unexpected invocbtion of MockS3API.DeleteObjects")
			},
		},
		GetObjectFunc: &S3APIGetObjectFunc{
			defbultHook: func(context.Context, *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
				pbnic("unexpected invocbtion of MockS3API.GetObject")
			},
		},
		HebdObjectFunc: &S3APIHebdObjectFunc{
			defbultHook: func(context.Context, *s3.HebdObjectInput) (*s3.HebdObjectOutput, error) {
				pbnic("unexpected invocbtion of MockS3API.HebdObject")
			},
		},
		NewListObjectsV2PbginbtorFunc: &S3APINewListObjectsV2PbginbtorFunc{
			defbultHook: func(*s3.ListObjectsV2Input) *s3.ListObjectsV2Pbginbtor {
				pbnic("unexpected invocbtion of MockS3API.NewListObjectsV2Pbginbtor")
			},
		},
		UplobdPbrtCopyFunc: &S3APIUplobdPbrtCopyFunc{
			defbultHook: func(context.Context, *s3.UplobdPbrtCopyInput) (*s3.UplobdPbrtCopyOutput, error) {
				pbnic("unexpected invocbtion of MockS3API.UplobdPbrtCopy")
			},
		},
	}
}

// surrogbteMockS3API is b copy of the s3API interfbce (from the pbckbge
// github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore). It is redefined
// here bs it is unexported in the source pbckbge.
type surrogbteMockS3API interfbce {
	AbortMultipbrtUplobd(context.Context, *s3.AbortMultipbrtUplobdInput) (*s3.AbortMultipbrtUplobdOutput, error)
	CompleteMultipbrtUplobd(context.Context, *s3.CompleteMultipbrtUplobdInput) (*s3.CompleteMultipbrtUplobdOutput, error)
	CrebteBucket(context.Context, *s3.CrebteBucketInput) (*s3.CrebteBucketOutput, error)
	CrebteMultipbrtUplobd(context.Context, *s3.CrebteMultipbrtUplobdInput) (*s3.CrebteMultipbrtUplobdOutput, error)
	DeleteObject(context.Context, *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
	DeleteObjects(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
	GetObject(context.Context, *s3.GetObjectInput) (*s3.GetObjectOutput, error)
	HebdObject(context.Context, *s3.HebdObjectInput) (*s3.HebdObjectOutput, error)
	NewListObjectsV2Pbginbtor(*s3.ListObjectsV2Input) *s3.ListObjectsV2Pbginbtor
	UplobdPbrtCopy(context.Context, *s3.UplobdPbrtCopyInput) (*s3.UplobdPbrtCopyOutput, error)
}

// NewMockS3APIFrom crebtes b new mock of the MockS3API interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockS3APIFrom(i surrogbteMockS3API) *MockS3API {
	return &MockS3API{
		AbortMultipbrtUplobdFunc: &S3APIAbortMultipbrtUplobdFunc{
			defbultHook: i.AbortMultipbrtUplobd,
		},
		CompleteMultipbrtUplobdFunc: &S3APICompleteMultipbrtUplobdFunc{
			defbultHook: i.CompleteMultipbrtUplobd,
		},
		CrebteBucketFunc: &S3APICrebteBucketFunc{
			defbultHook: i.CrebteBucket,
		},
		CrebteMultipbrtUplobdFunc: &S3APICrebteMultipbrtUplobdFunc{
			defbultHook: i.CrebteMultipbrtUplobd,
		},
		DeleteObjectFunc: &S3APIDeleteObjectFunc{
			defbultHook: i.DeleteObject,
		},
		DeleteObjectsFunc: &S3APIDeleteObjectsFunc{
			defbultHook: i.DeleteObjects,
		},
		GetObjectFunc: &S3APIGetObjectFunc{
			defbultHook: i.GetObject,
		},
		HebdObjectFunc: &S3APIHebdObjectFunc{
			defbultHook: i.HebdObject,
		},
		NewListObjectsV2PbginbtorFunc: &S3APINewListObjectsV2PbginbtorFunc{
			defbultHook: i.NewListObjectsV2Pbginbtor,
		},
		UplobdPbrtCopyFunc: &S3APIUplobdPbrtCopyFunc{
			defbultHook: i.UplobdPbrtCopy,
		},
	}
}

// S3APIAbortMultipbrtUplobdFunc describes the behbvior when the
// AbortMultipbrtUplobd method of the pbrent MockS3API instbnce is invoked.
type S3APIAbortMultipbrtUplobdFunc struct {
	defbultHook func(context.Context, *s3.AbortMultipbrtUplobdInput) (*s3.AbortMultipbrtUplobdOutput, error)
	hooks       []func(context.Context, *s3.AbortMultipbrtUplobdInput) (*s3.AbortMultipbrtUplobdOutput, error)
	history     []S3APIAbortMultipbrtUplobdFuncCbll
	mutex       sync.Mutex
}

// AbortMultipbrtUplobd delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockS3API) AbortMultipbrtUplobd(v0 context.Context, v1 *s3.AbortMultipbrtUplobdInput) (*s3.AbortMultipbrtUplobdOutput, error) {
	r0, r1 := m.AbortMultipbrtUplobdFunc.nextHook()(v0, v1)
	m.AbortMultipbrtUplobdFunc.bppendCbll(S3APIAbortMultipbrtUplobdFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the AbortMultipbrtUplobd
// method of the pbrent MockS3API instbnce is invoked bnd the hook queue is
// empty.
func (f *S3APIAbortMultipbrtUplobdFunc) SetDefbultHook(hook func(context.Context, *s3.AbortMultipbrtUplobdInput) (*s3.AbortMultipbrtUplobdOutput, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// AbortMultipbrtUplobd method of the pbrent MockS3API instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *S3APIAbortMultipbrtUplobdFunc) PushHook(hook func(context.Context, *s3.AbortMultipbrtUplobdInput) (*s3.AbortMultipbrtUplobdOutput, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *S3APIAbortMultipbrtUplobdFunc) SetDefbultReturn(r0 *s3.AbortMultipbrtUplobdOutput, r1 error) {
	f.SetDefbultHook(func(context.Context, *s3.AbortMultipbrtUplobdInput) (*s3.AbortMultipbrtUplobdOutput, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *S3APIAbortMultipbrtUplobdFunc) PushReturn(r0 *s3.AbortMultipbrtUplobdOutput, r1 error) {
	f.PushHook(func(context.Context, *s3.AbortMultipbrtUplobdInput) (*s3.AbortMultipbrtUplobdOutput, error) {
		return r0, r1
	})
}

func (f *S3APIAbortMultipbrtUplobdFunc) nextHook() func(context.Context, *s3.AbortMultipbrtUplobdInput) (*s3.AbortMultipbrtUplobdOutput, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *S3APIAbortMultipbrtUplobdFunc) bppendCbll(r0 S3APIAbortMultipbrtUplobdFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of S3APIAbortMultipbrtUplobdFuncCbll objects
// describing the invocbtions of this function.
func (f *S3APIAbortMultipbrtUplobdFunc) History() []S3APIAbortMultipbrtUplobdFuncCbll {
	f.mutex.Lock()
	history := mbke([]S3APIAbortMultipbrtUplobdFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// S3APIAbortMultipbrtUplobdFuncCbll is bn object thbt describes bn
// invocbtion of method AbortMultipbrtUplobd on bn instbnce of MockS3API.
type S3APIAbortMultipbrtUplobdFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *s3.AbortMultipbrtUplobdInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *s3.AbortMultipbrtUplobdOutput
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c S3APIAbortMultipbrtUplobdFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c S3APIAbortMultipbrtUplobdFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// S3APICompleteMultipbrtUplobdFunc describes the behbvior when the
// CompleteMultipbrtUplobd method of the pbrent MockS3API instbnce is
// invoked.
type S3APICompleteMultipbrtUplobdFunc struct {
	defbultHook func(context.Context, *s3.CompleteMultipbrtUplobdInput) (*s3.CompleteMultipbrtUplobdOutput, error)
	hooks       []func(context.Context, *s3.CompleteMultipbrtUplobdInput) (*s3.CompleteMultipbrtUplobdOutput, error)
	history     []S3APICompleteMultipbrtUplobdFuncCbll
	mutex       sync.Mutex
}

// CompleteMultipbrtUplobd delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockS3API) CompleteMultipbrtUplobd(v0 context.Context, v1 *s3.CompleteMultipbrtUplobdInput) (*s3.CompleteMultipbrtUplobdOutput, error) {
	r0, r1 := m.CompleteMultipbrtUplobdFunc.nextHook()(v0, v1)
	m.CompleteMultipbrtUplobdFunc.bppendCbll(S3APICompleteMultipbrtUplobdFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// CompleteMultipbrtUplobd method of the pbrent MockS3API instbnce is
// invoked bnd the hook queue is empty.
func (f *S3APICompleteMultipbrtUplobdFunc) SetDefbultHook(hook func(context.Context, *s3.CompleteMultipbrtUplobdInput) (*s3.CompleteMultipbrtUplobdOutput, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CompleteMultipbrtUplobd method of the pbrent MockS3API instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *S3APICompleteMultipbrtUplobdFunc) PushHook(hook func(context.Context, *s3.CompleteMultipbrtUplobdInput) (*s3.CompleteMultipbrtUplobdOutput, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *S3APICompleteMultipbrtUplobdFunc) SetDefbultReturn(r0 *s3.CompleteMultipbrtUplobdOutput, r1 error) {
	f.SetDefbultHook(func(context.Context, *s3.CompleteMultipbrtUplobdInput) (*s3.CompleteMultipbrtUplobdOutput, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *S3APICompleteMultipbrtUplobdFunc) PushReturn(r0 *s3.CompleteMultipbrtUplobdOutput, r1 error) {
	f.PushHook(func(context.Context, *s3.CompleteMultipbrtUplobdInput) (*s3.CompleteMultipbrtUplobdOutput, error) {
		return r0, r1
	})
}

func (f *S3APICompleteMultipbrtUplobdFunc) nextHook() func(context.Context, *s3.CompleteMultipbrtUplobdInput) (*s3.CompleteMultipbrtUplobdOutput, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *S3APICompleteMultipbrtUplobdFunc) bppendCbll(r0 S3APICompleteMultipbrtUplobdFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of S3APICompleteMultipbrtUplobdFuncCbll
// objects describing the invocbtions of this function.
func (f *S3APICompleteMultipbrtUplobdFunc) History() []S3APICompleteMultipbrtUplobdFuncCbll {
	f.mutex.Lock()
	history := mbke([]S3APICompleteMultipbrtUplobdFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// S3APICompleteMultipbrtUplobdFuncCbll is bn object thbt describes bn
// invocbtion of method CompleteMultipbrtUplobd on bn instbnce of MockS3API.
type S3APICompleteMultipbrtUplobdFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *s3.CompleteMultipbrtUplobdInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *s3.CompleteMultipbrtUplobdOutput
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c S3APICompleteMultipbrtUplobdFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c S3APICompleteMultipbrtUplobdFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// S3APICrebteBucketFunc describes the behbvior when the CrebteBucket method
// of the pbrent MockS3API instbnce is invoked.
type S3APICrebteBucketFunc struct {
	defbultHook func(context.Context, *s3.CrebteBucketInput) (*s3.CrebteBucketOutput, error)
	hooks       []func(context.Context, *s3.CrebteBucketInput) (*s3.CrebteBucketOutput, error)
	history     []S3APICrebteBucketFuncCbll
	mutex       sync.Mutex
}

// CrebteBucket delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockS3API) CrebteBucket(v0 context.Context, v1 *s3.CrebteBucketInput) (*s3.CrebteBucketOutput, error) {
	r0, r1 := m.CrebteBucketFunc.nextHook()(v0, v1)
	m.CrebteBucketFunc.bppendCbll(S3APICrebteBucketFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the CrebteBucket method
// of the pbrent MockS3API instbnce is invoked bnd the hook queue is empty.
func (f *S3APICrebteBucketFunc) SetDefbultHook(hook func(context.Context, *s3.CrebteBucketInput) (*s3.CrebteBucketOutput, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CrebteBucket method of the pbrent MockS3API instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *S3APICrebteBucketFunc) PushHook(hook func(context.Context, *s3.CrebteBucketInput) (*s3.CrebteBucketOutput, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *S3APICrebteBucketFunc) SetDefbultReturn(r0 *s3.CrebteBucketOutput, r1 error) {
	f.SetDefbultHook(func(context.Context, *s3.CrebteBucketInput) (*s3.CrebteBucketOutput, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *S3APICrebteBucketFunc) PushReturn(r0 *s3.CrebteBucketOutput, r1 error) {
	f.PushHook(func(context.Context, *s3.CrebteBucketInput) (*s3.CrebteBucketOutput, error) {
		return r0, r1
	})
}

func (f *S3APICrebteBucketFunc) nextHook() func(context.Context, *s3.CrebteBucketInput) (*s3.CrebteBucketOutput, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *S3APICrebteBucketFunc) bppendCbll(r0 S3APICrebteBucketFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of S3APICrebteBucketFuncCbll objects
// describing the invocbtions of this function.
func (f *S3APICrebteBucketFunc) History() []S3APICrebteBucketFuncCbll {
	f.mutex.Lock()
	history := mbke([]S3APICrebteBucketFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// S3APICrebteBucketFuncCbll is bn object thbt describes bn invocbtion of
// method CrebteBucket on bn instbnce of MockS3API.
type S3APICrebteBucketFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *s3.CrebteBucketInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *s3.CrebteBucketOutput
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c S3APICrebteBucketFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c S3APICrebteBucketFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// S3APICrebteMultipbrtUplobdFunc describes the behbvior when the
// CrebteMultipbrtUplobd method of the pbrent MockS3API instbnce is invoked.
type S3APICrebteMultipbrtUplobdFunc struct {
	defbultHook func(context.Context, *s3.CrebteMultipbrtUplobdInput) (*s3.CrebteMultipbrtUplobdOutput, error)
	hooks       []func(context.Context, *s3.CrebteMultipbrtUplobdInput) (*s3.CrebteMultipbrtUplobdOutput, error)
	history     []S3APICrebteMultipbrtUplobdFuncCbll
	mutex       sync.Mutex
}

// CrebteMultipbrtUplobd delegbtes to the next hook function in the queue
// bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockS3API) CrebteMultipbrtUplobd(v0 context.Context, v1 *s3.CrebteMultipbrtUplobdInput) (*s3.CrebteMultipbrtUplobdOutput, error) {
	r0, r1 := m.CrebteMultipbrtUplobdFunc.nextHook()(v0, v1)
	m.CrebteMultipbrtUplobdFunc.bppendCbll(S3APICrebteMultipbrtUplobdFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the
// CrebteMultipbrtUplobd method of the pbrent MockS3API instbnce is invoked
// bnd the hook queue is empty.
func (f *S3APICrebteMultipbrtUplobdFunc) SetDefbultHook(hook func(context.Context, *s3.CrebteMultipbrtUplobdInput) (*s3.CrebteMultipbrtUplobdOutput, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// CrebteMultipbrtUplobd method of the pbrent MockS3API instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *S3APICrebteMultipbrtUplobdFunc) PushHook(hook func(context.Context, *s3.CrebteMultipbrtUplobdInput) (*s3.CrebteMultipbrtUplobdOutput, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *S3APICrebteMultipbrtUplobdFunc) SetDefbultReturn(r0 *s3.CrebteMultipbrtUplobdOutput, r1 error) {
	f.SetDefbultHook(func(context.Context, *s3.CrebteMultipbrtUplobdInput) (*s3.CrebteMultipbrtUplobdOutput, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *S3APICrebteMultipbrtUplobdFunc) PushReturn(r0 *s3.CrebteMultipbrtUplobdOutput, r1 error) {
	f.PushHook(func(context.Context, *s3.CrebteMultipbrtUplobdInput) (*s3.CrebteMultipbrtUplobdOutput, error) {
		return r0, r1
	})
}

func (f *S3APICrebteMultipbrtUplobdFunc) nextHook() func(context.Context, *s3.CrebteMultipbrtUplobdInput) (*s3.CrebteMultipbrtUplobdOutput, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *S3APICrebteMultipbrtUplobdFunc) bppendCbll(r0 S3APICrebteMultipbrtUplobdFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of S3APICrebteMultipbrtUplobdFuncCbll objects
// describing the invocbtions of this function.
func (f *S3APICrebteMultipbrtUplobdFunc) History() []S3APICrebteMultipbrtUplobdFuncCbll {
	f.mutex.Lock()
	history := mbke([]S3APICrebteMultipbrtUplobdFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// S3APICrebteMultipbrtUplobdFuncCbll is bn object thbt describes bn
// invocbtion of method CrebteMultipbrtUplobd on bn instbnce of MockS3API.
type S3APICrebteMultipbrtUplobdFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *s3.CrebteMultipbrtUplobdInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *s3.CrebteMultipbrtUplobdOutput
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c S3APICrebteMultipbrtUplobdFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c S3APICrebteMultipbrtUplobdFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// S3APIDeleteObjectFunc describes the behbvior when the DeleteObject method
// of the pbrent MockS3API instbnce is invoked.
type S3APIDeleteObjectFunc struct {
	defbultHook func(context.Context, *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
	hooks       []func(context.Context, *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)
	history     []S3APIDeleteObjectFuncCbll
	mutex       sync.Mutex
}

// DeleteObject delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockS3API) DeleteObject(v0 context.Context, v1 *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	r0, r1 := m.DeleteObjectFunc.nextHook()(v0, v1)
	m.DeleteObjectFunc.bppendCbll(S3APIDeleteObjectFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the DeleteObject method
// of the pbrent MockS3API instbnce is invoked bnd the hook queue is empty.
func (f *S3APIDeleteObjectFunc) SetDefbultHook(hook func(context.Context, *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DeleteObject method of the pbrent MockS3API instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *S3APIDeleteObjectFunc) PushHook(hook func(context.Context, *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *S3APIDeleteObjectFunc) SetDefbultReturn(r0 *s3.DeleteObjectOutput, r1 error) {
	f.SetDefbultHook(func(context.Context, *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *S3APIDeleteObjectFunc) PushReturn(r0 *s3.DeleteObjectOutput, r1 error) {
	f.PushHook(func(context.Context, *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
		return r0, r1
	})
}

func (f *S3APIDeleteObjectFunc) nextHook() func(context.Context, *s3.DeleteObjectInput) (*s3.DeleteObjectOutput, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *S3APIDeleteObjectFunc) bppendCbll(r0 S3APIDeleteObjectFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of S3APIDeleteObjectFuncCbll objects
// describing the invocbtions of this function.
func (f *S3APIDeleteObjectFunc) History() []S3APIDeleteObjectFuncCbll {
	f.mutex.Lock()
	history := mbke([]S3APIDeleteObjectFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// S3APIDeleteObjectFuncCbll is bn object thbt describes bn invocbtion of
// method DeleteObject on bn instbnce of MockS3API.
type S3APIDeleteObjectFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *s3.DeleteObjectInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *s3.DeleteObjectOutput
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c S3APIDeleteObjectFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c S3APIDeleteObjectFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// S3APIDeleteObjectsFunc describes the behbvior when the DeleteObjects
// method of the pbrent MockS3API instbnce is invoked.
type S3APIDeleteObjectsFunc struct {
	defbultHook func(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
	hooks       []func(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
	history     []S3APIDeleteObjectsFuncCbll
	mutex       sync.Mutex
}

// DeleteObjects delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockS3API) DeleteObjects(v0 context.Context, v1 *s3.DeleteObjectsInput, v2 ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	r0, r1 := m.DeleteObjectsFunc.nextHook()(v0, v1, v2...)
	m.DeleteObjectsFunc.bppendCbll(S3APIDeleteObjectsFuncCbll{v0, v1, v2, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the DeleteObjects method
// of the pbrent MockS3API instbnce is invoked bnd the hook queue is empty.
func (f *S3APIDeleteObjectsFunc) SetDefbultHook(hook func(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// DeleteObjects method of the pbrent MockS3API instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *S3APIDeleteObjectsFunc) PushHook(hook func(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *S3APIDeleteObjectsFunc) SetDefbultReturn(r0 *s3.DeleteObjectsOutput, r1 error) {
	f.SetDefbultHook(func(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *S3APIDeleteObjectsFunc) PushReturn(r0 *s3.DeleteObjectsOutput, r1 error) {
	f.PushHook(func(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
		return r0, r1
	})
}

func (f *S3APIDeleteObjectsFunc) nextHook() func(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *S3APIDeleteObjectsFunc) bppendCbll(r0 S3APIDeleteObjectsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of S3APIDeleteObjectsFuncCbll objects
// describing the invocbtions of this function.
func (f *S3APIDeleteObjectsFunc) History() []S3APIDeleteObjectsFuncCbll {
	f.mutex.Lock()
	history := mbke([]S3APIDeleteObjectsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// S3APIDeleteObjectsFuncCbll is bn object thbt describes bn invocbtion of
// method DeleteObjects on bn instbnce of MockS3API.
type S3APIDeleteObjectsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *s3.DeleteObjectsInput
	// Arg2 is b slice contbining the vblues of the vbribdic brguments
	// pbssed to this method invocbtion.
	Arg2 []func(*s3.Options)
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *s3.DeleteObjectsOutput
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion. The vbribdic slice brgument is flbttened in this brrby such
// thbt one positionbl brgument bnd three vbribdic brguments would result in
// b slice of four, not two.
func (c S3APIDeleteObjectsFuncCbll) Args() []interfbce{} {
	trbiling := []interfbce{}{}
	for _, vbl := rbnge c.Arg2 {
		trbiling = bppend(trbiling, vbl)
	}

	return bppend([]interfbce{}{c.Arg0, c.Arg1}, trbiling...)
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c S3APIDeleteObjectsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// S3APIGetObjectFunc describes the behbvior when the GetObject method of
// the pbrent MockS3API instbnce is invoked.
type S3APIGetObjectFunc struct {
	defbultHook func(context.Context, *s3.GetObjectInput) (*s3.GetObjectOutput, error)
	hooks       []func(context.Context, *s3.GetObjectInput) (*s3.GetObjectOutput, error)
	history     []S3APIGetObjectFuncCbll
	mutex       sync.Mutex
}

// GetObject delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockS3API) GetObject(v0 context.Context, v1 *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	r0, r1 := m.GetObjectFunc.nextHook()(v0, v1)
	m.GetObjectFunc.bppendCbll(S3APIGetObjectFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the GetObject method of
// the pbrent MockS3API instbnce is invoked bnd the hook queue is empty.
func (f *S3APIGetObjectFunc) SetDefbultHook(hook func(context.Context, *s3.GetObjectInput) (*s3.GetObjectOutput, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetObject method of the pbrent MockS3API instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *S3APIGetObjectFunc) PushHook(hook func(context.Context, *s3.GetObjectInput) (*s3.GetObjectOutput, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *S3APIGetObjectFunc) SetDefbultReturn(r0 *s3.GetObjectOutput, r1 error) {
	f.SetDefbultHook(func(context.Context, *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *S3APIGetObjectFunc) PushReturn(r0 *s3.GetObjectOutput, r1 error) {
	f.PushHook(func(context.Context, *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
		return r0, r1
	})
}

func (f *S3APIGetObjectFunc) nextHook() func(context.Context, *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *S3APIGetObjectFunc) bppendCbll(r0 S3APIGetObjectFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of S3APIGetObjectFuncCbll objects describing
// the invocbtions of this function.
func (f *S3APIGetObjectFunc) History() []S3APIGetObjectFuncCbll {
	f.mutex.Lock()
	history := mbke([]S3APIGetObjectFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// S3APIGetObjectFuncCbll is bn object thbt describes bn invocbtion of
// method GetObject on bn instbnce of MockS3API.
type S3APIGetObjectFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *s3.GetObjectInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *s3.GetObjectOutput
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c S3APIGetObjectFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c S3APIGetObjectFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// S3APIHebdObjectFunc describes the behbvior when the HebdObject method of
// the pbrent MockS3API instbnce is invoked.
type S3APIHebdObjectFunc struct {
	defbultHook func(context.Context, *s3.HebdObjectInput) (*s3.HebdObjectOutput, error)
	hooks       []func(context.Context, *s3.HebdObjectInput) (*s3.HebdObjectOutput, error)
	history     []S3APIHebdObjectFuncCbll
	mutex       sync.Mutex
}

// HebdObject delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockS3API) HebdObject(v0 context.Context, v1 *s3.HebdObjectInput) (*s3.HebdObjectOutput, error) {
	r0, r1 := m.HebdObjectFunc.nextHook()(v0, v1)
	m.HebdObjectFunc.bppendCbll(S3APIHebdObjectFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the HebdObject method of
// the pbrent MockS3API instbnce is invoked bnd the hook queue is empty.
func (f *S3APIHebdObjectFunc) SetDefbultHook(hook func(context.Context, *s3.HebdObjectInput) (*s3.HebdObjectOutput, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// HebdObject method of the pbrent MockS3API instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *S3APIHebdObjectFunc) PushHook(hook func(context.Context, *s3.HebdObjectInput) (*s3.HebdObjectOutput, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *S3APIHebdObjectFunc) SetDefbultReturn(r0 *s3.HebdObjectOutput, r1 error) {
	f.SetDefbultHook(func(context.Context, *s3.HebdObjectInput) (*s3.HebdObjectOutput, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *S3APIHebdObjectFunc) PushReturn(r0 *s3.HebdObjectOutput, r1 error) {
	f.PushHook(func(context.Context, *s3.HebdObjectInput) (*s3.HebdObjectOutput, error) {
		return r0, r1
	})
}

func (f *S3APIHebdObjectFunc) nextHook() func(context.Context, *s3.HebdObjectInput) (*s3.HebdObjectOutput, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *S3APIHebdObjectFunc) bppendCbll(r0 S3APIHebdObjectFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of S3APIHebdObjectFuncCbll objects describing
// the invocbtions of this function.
func (f *S3APIHebdObjectFunc) History() []S3APIHebdObjectFuncCbll {
	f.mutex.Lock()
	history := mbke([]S3APIHebdObjectFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// S3APIHebdObjectFuncCbll is bn object thbt describes bn invocbtion of
// method HebdObject on bn instbnce of MockS3API.
type S3APIHebdObjectFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *s3.HebdObjectInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *s3.HebdObjectOutput
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c S3APIHebdObjectFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c S3APIHebdObjectFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// S3APINewListObjectsV2PbginbtorFunc describes the behbvior when the
// NewListObjectsV2Pbginbtor method of the pbrent MockS3API instbnce is
// invoked.
type S3APINewListObjectsV2PbginbtorFunc struct {
	defbultHook func(*s3.ListObjectsV2Input) *s3.ListObjectsV2Pbginbtor
	hooks       []func(*s3.ListObjectsV2Input) *s3.ListObjectsV2Pbginbtor
	history     []S3APINewListObjectsV2PbginbtorFuncCbll
	mutex       sync.Mutex
}

// NewListObjectsV2Pbginbtor delegbtes to the next hook function in the
// queue bnd stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockS3API) NewListObjectsV2Pbginbtor(v0 *s3.ListObjectsV2Input) *s3.ListObjectsV2Pbginbtor {
	r0 := m.NewListObjectsV2PbginbtorFunc.nextHook()(v0)
	m.NewListObjectsV2PbginbtorFunc.bppendCbll(S3APINewListObjectsV2PbginbtorFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the
// NewListObjectsV2Pbginbtor method of the pbrent MockS3API instbnce is
// invoked bnd the hook queue is empty.
func (f *S3APINewListObjectsV2PbginbtorFunc) SetDefbultHook(hook func(*s3.ListObjectsV2Input) *s3.ListObjectsV2Pbginbtor) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// NewListObjectsV2Pbginbtor method of the pbrent MockS3API instbnce invokes
// the hook bt the front of the queue bnd discbrds it. After the queue is
// empty, the defbult hook function is invoked for bny future bction.
func (f *S3APINewListObjectsV2PbginbtorFunc) PushHook(hook func(*s3.ListObjectsV2Input) *s3.ListObjectsV2Pbginbtor) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *S3APINewListObjectsV2PbginbtorFunc) SetDefbultReturn(r0 *s3.ListObjectsV2Pbginbtor) {
	f.SetDefbultHook(func(*s3.ListObjectsV2Input) *s3.ListObjectsV2Pbginbtor {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *S3APINewListObjectsV2PbginbtorFunc) PushReturn(r0 *s3.ListObjectsV2Pbginbtor) {
	f.PushHook(func(*s3.ListObjectsV2Input) *s3.ListObjectsV2Pbginbtor {
		return r0
	})
}

func (f *S3APINewListObjectsV2PbginbtorFunc) nextHook() func(*s3.ListObjectsV2Input) *s3.ListObjectsV2Pbginbtor {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *S3APINewListObjectsV2PbginbtorFunc) bppendCbll(r0 S3APINewListObjectsV2PbginbtorFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of S3APINewListObjectsV2PbginbtorFuncCbll
// objects describing the invocbtions of this function.
func (f *S3APINewListObjectsV2PbginbtorFunc) History() []S3APINewListObjectsV2PbginbtorFuncCbll {
	f.mutex.Lock()
	history := mbke([]S3APINewListObjectsV2PbginbtorFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// S3APINewListObjectsV2PbginbtorFuncCbll is bn object thbt describes bn
// invocbtion of method NewListObjectsV2Pbginbtor on bn instbnce of
// MockS3API.
type S3APINewListObjectsV2PbginbtorFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 *s3.ListObjectsV2Input
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *s3.ListObjectsV2Pbginbtor
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c S3APINewListObjectsV2PbginbtorFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c S3APINewListObjectsV2PbginbtorFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// S3APIUplobdPbrtCopyFunc describes the behbvior when the UplobdPbrtCopy
// method of the pbrent MockS3API instbnce is invoked.
type S3APIUplobdPbrtCopyFunc struct {
	defbultHook func(context.Context, *s3.UplobdPbrtCopyInput) (*s3.UplobdPbrtCopyOutput, error)
	hooks       []func(context.Context, *s3.UplobdPbrtCopyInput) (*s3.UplobdPbrtCopyOutput, error)
	history     []S3APIUplobdPbrtCopyFuncCbll
	mutex       sync.Mutex
}

// UplobdPbrtCopy delegbtes to the next hook function in the queue bnd
// stores the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockS3API) UplobdPbrtCopy(v0 context.Context, v1 *s3.UplobdPbrtCopyInput) (*s3.UplobdPbrtCopyOutput, error) {
	r0, r1 := m.UplobdPbrtCopyFunc.nextHook()(v0, v1)
	m.UplobdPbrtCopyFunc.bppendCbll(S3APIUplobdPbrtCopyFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the UplobdPbrtCopy
// method of the pbrent MockS3API instbnce is invoked bnd the hook queue is
// empty.
func (f *S3APIUplobdPbrtCopyFunc) SetDefbultHook(hook func(context.Context, *s3.UplobdPbrtCopyInput) (*s3.UplobdPbrtCopyOutput, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// UplobdPbrtCopy method of the pbrent MockS3API instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *S3APIUplobdPbrtCopyFunc) PushHook(hook func(context.Context, *s3.UplobdPbrtCopyInput) (*s3.UplobdPbrtCopyOutput, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *S3APIUplobdPbrtCopyFunc) SetDefbultReturn(r0 *s3.UplobdPbrtCopyOutput, r1 error) {
	f.SetDefbultHook(func(context.Context, *s3.UplobdPbrtCopyInput) (*s3.UplobdPbrtCopyOutput, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *S3APIUplobdPbrtCopyFunc) PushReturn(r0 *s3.UplobdPbrtCopyOutput, r1 error) {
	f.PushHook(func(context.Context, *s3.UplobdPbrtCopyInput) (*s3.UplobdPbrtCopyOutput, error) {
		return r0, r1
	})
}

func (f *S3APIUplobdPbrtCopyFunc) nextHook() func(context.Context, *s3.UplobdPbrtCopyInput) (*s3.UplobdPbrtCopyOutput, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *S3APIUplobdPbrtCopyFunc) bppendCbll(r0 S3APIUplobdPbrtCopyFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of S3APIUplobdPbrtCopyFuncCbll objects
// describing the invocbtions of this function.
func (f *S3APIUplobdPbrtCopyFunc) History() []S3APIUplobdPbrtCopyFuncCbll {
	f.mutex.Lock()
	history := mbke([]S3APIUplobdPbrtCopyFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// S3APIUplobdPbrtCopyFuncCbll is bn object thbt describes bn invocbtion of
// method UplobdPbrtCopy on bn instbnce of MockS3API.
type S3APIUplobdPbrtCopyFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *s3.UplobdPbrtCopyInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *s3.UplobdPbrtCopyOutput
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c S3APIUplobdPbrtCopyFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c S3APIUplobdPbrtCopyFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// MockS3Uplobder is b mock implementbtion of the s3Uplobder interfbce (from
// the pbckbge github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore) used
// for unit testing.
type MockS3Uplobder struct {
	// UplobdFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Uplobd.
	UplobdFunc *S3UplobderUplobdFunc
}

// NewMockS3Uplobder crebtes b new mock of the s3Uplobder interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockS3Uplobder() *MockS3Uplobder {
	return &MockS3Uplobder{
		UplobdFunc: &S3UplobderUplobdFunc{
			defbultHook: func(context.Context, *s3.PutObjectInput) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockS3Uplobder crebtes b new mock of the s3Uplobder interfbce.
// All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockS3Uplobder() *MockS3Uplobder {
	return &MockS3Uplobder{
		UplobdFunc: &S3UplobderUplobdFunc{
			defbultHook: func(context.Context, *s3.PutObjectInput) error {
				pbnic("unexpected invocbtion of MockS3Uplobder.Uplobd")
			},
		},
	}
}

// surrogbteMockS3Uplobder is b copy of the s3Uplobder interfbce (from the
// pbckbge github.com/sourcegrbph/sourcegrbph/internbl/uplobdstore). It is
// redefined here bs it is unexported in the source pbckbge.
type surrogbteMockS3Uplobder interfbce {
	Uplobd(context.Context, *s3.PutObjectInput) error
}

// NewMockS3UplobderFrom crebtes b new mock of the MockS3Uplobder interfbce.
// All methods delegbte to the given implementbtion, unless overwritten.
func NewMockS3UplobderFrom(i surrogbteMockS3Uplobder) *MockS3Uplobder {
	return &MockS3Uplobder{
		UplobdFunc: &S3UplobderUplobdFunc{
			defbultHook: i.Uplobd,
		},
	}
}

// S3UplobderUplobdFunc describes the behbvior when the Uplobd method of the
// pbrent MockS3Uplobder instbnce is invoked.
type S3UplobderUplobdFunc struct {
	defbultHook func(context.Context, *s3.PutObjectInput) error
	hooks       []func(context.Context, *s3.PutObjectInput) error
	history     []S3UplobderUplobdFuncCbll
	mutex       sync.Mutex
}

// Uplobd delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockS3Uplobder) Uplobd(v0 context.Context, v1 *s3.PutObjectInput) error {
	r0 := m.UplobdFunc.nextHook()(v0, v1)
	m.UplobdFunc.bppendCbll(S3UplobderUplobdFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Uplobd method of the
// pbrent MockS3Uplobder instbnce is invoked bnd the hook queue is empty.
func (f *S3UplobderUplobdFunc) SetDefbultHook(hook func(context.Context, *s3.PutObjectInput) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Uplobd method of the pbrent MockS3Uplobder instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *S3UplobderUplobdFunc) PushHook(hook func(context.Context, *s3.PutObjectInput) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *S3UplobderUplobdFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, *s3.PutObjectInput) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *S3UplobderUplobdFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, *s3.PutObjectInput) error {
		return r0
	})
}

func (f *S3UplobderUplobdFunc) nextHook() func(context.Context, *s3.PutObjectInput) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *S3UplobderUplobdFunc) bppendCbll(r0 S3UplobderUplobdFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of S3UplobderUplobdFuncCbll objects describing
// the invocbtions of this function.
func (f *S3UplobderUplobdFunc) History() []S3UplobderUplobdFuncCbll {
	f.mutex.Lock()
	history := mbke([]S3UplobderUplobdFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// S3UplobderUplobdFuncCbll is bn object thbt describes bn invocbtion of
// method Uplobd on bn instbnce of MockS3Uplobder.
type S3UplobderUplobdFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 *s3.PutObjectInput
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c S3UplobderUplobdFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c S3UplobderUplobdFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
