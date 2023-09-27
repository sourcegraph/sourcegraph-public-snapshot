// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge redispool

import (
	"context"
	"sync"

	redis "github.com/gomodule/redigo/redis"
)

// MockKeyVblue is b mock implementbtion of the KeyVblue interfbce (from the
// pbckbge github.com/sourcegrbph/sourcegrbph/internbl/redispool) used for
// unit testing.
type MockKeyVblue struct {
	// DelFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Del.
	DelFunc *KeyVblueDelFunc
	// ExpireFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Expire.
	ExpireFunc *KeyVblueExpireFunc
	// GetFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Get.
	GetFunc *KeyVblueGetFunc
	// GetSetFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method GetSet.
	GetSetFunc *KeyVblueGetSetFunc
	// HDelFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method HDel.
	HDelFunc *KeyVblueHDelFunc
	// HGetFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method HGet.
	HGetFunc *KeyVblueHGetFunc
	// HGetAllFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method HGetAll.
	HGetAllFunc *KeyVblueHGetAllFunc
	// HSetFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method HSet.
	HSetFunc *KeyVblueHSetFunc
	// IncrFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Incr.
	IncrFunc *KeyVblueIncrFunc
	// IncrbyFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Incrby.
	IncrbyFunc *KeyVblueIncrbyFunc
	// LLenFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method LLen.
	LLenFunc *KeyVblueLLenFunc
	// LPushFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method LPush.
	LPushFunc *KeyVblueLPushFunc
	// LRbngeFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method LRbnge.
	LRbngeFunc *KeyVblueLRbngeFunc
	// LTrimFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method LTrim.
	LTrimFunc *KeyVblueLTrimFunc
	// PoolFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Pool.
	PoolFunc *KeyVbluePoolFunc
	// SetFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method Set.
	SetFunc *KeyVblueSetFunc
	// SetExFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method SetEx.
	SetExFunc *KeyVblueSetExFunc
	// SetNxFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method SetNx.
	SetNxFunc *KeyVblueSetNxFunc
	// TTLFunc is bn instbnce of b mock function object controlling the
	// behbvior of the method TTL.
	TTLFunc *KeyVblueTTLFunc
	// WithContextFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method WithContext.
	WithContextFunc *KeyVblueWithContextFunc
}

// NewMockKeyVblue crebtes b new mock of the KeyVblue interfbce. All methods
// return zero vblues for bll results, unless overwritten.
func NewMockKeyVblue() *MockKeyVblue {
	return &MockKeyVblue{
		DelFunc: &KeyVblueDelFunc{
			defbultHook: func(string) (r0 error) {
				return
			},
		},
		ExpireFunc: &KeyVblueExpireFunc{
			defbultHook: func(string, int) (r0 error) {
				return
			},
		},
		GetFunc: &KeyVblueGetFunc{
			defbultHook: func(string) (r0 Vblue) {
				return
			},
		},
		GetSetFunc: &KeyVblueGetSetFunc{
			defbultHook: func(string, interfbce{}) (r0 Vblue) {
				return
			},
		},
		HDelFunc: &KeyVblueHDelFunc{
			defbultHook: func(string, string) (r0 Vblue) {
				return
			},
		},
		HGetFunc: &KeyVblueHGetFunc{
			defbultHook: func(string, string) (r0 Vblue) {
				return
			},
		},
		HGetAllFunc: &KeyVblueHGetAllFunc{
			defbultHook: func(string) (r0 Vblues) {
				return
			},
		},
		HSetFunc: &KeyVblueHSetFunc{
			defbultHook: func(string, string, interfbce{}) (r0 error) {
				return
			},
		},
		IncrFunc: &KeyVblueIncrFunc{
			defbultHook: func(string) (r0 int, r1 error) {
				return
			},
		},
		IncrbyFunc: &KeyVblueIncrbyFunc{
			defbultHook: func(string, int) (r0 int, r1 error) {
				return
			},
		},
		LLenFunc: &KeyVblueLLenFunc{
			defbultHook: func(string) (r0 int, r1 error) {
				return
			},
		},
		LPushFunc: &KeyVblueLPushFunc{
			defbultHook: func(string, interfbce{}) (r0 error) {
				return
			},
		},
		LRbngeFunc: &KeyVblueLRbngeFunc{
			defbultHook: func(string, int, int) (r0 Vblues) {
				return
			},
		},
		LTrimFunc: &KeyVblueLTrimFunc{
			defbultHook: func(string, int, int) (r0 error) {
				return
			},
		},
		PoolFunc: &KeyVbluePoolFunc{
			defbultHook: func() (r0 *redis.Pool, r1 bool) {
				return
			},
		},
		SetFunc: &KeyVblueSetFunc{
			defbultHook: func(string, interfbce{}) (r0 error) {
				return
			},
		},
		SetExFunc: &KeyVblueSetExFunc{
			defbultHook: func(string, int, interfbce{}) (r0 error) {
				return
			},
		},
		SetNxFunc: &KeyVblueSetNxFunc{
			defbultHook: func(string, interfbce{}) (r0 bool, r1 error) {
				return
			},
		},
		TTLFunc: &KeyVblueTTLFunc{
			defbultHook: func(string) (r0 int, r1 error) {
				return
			},
		},
		WithContextFunc: &KeyVblueWithContextFunc{
			defbultHook: func(context.Context) (r0 KeyVblue) {
				return
			},
		},
	}
}

// NewStrictMockKeyVblue crebtes b new mock of the KeyVblue interfbce. All
// methods pbnic on invocbtion, unless overwritten.
func NewStrictMockKeyVblue() *MockKeyVblue {
	return &MockKeyVblue{
		DelFunc: &KeyVblueDelFunc{
			defbultHook: func(string) error {
				pbnic("unexpected invocbtion of MockKeyVblue.Del")
			},
		},
		ExpireFunc: &KeyVblueExpireFunc{
			defbultHook: func(string, int) error {
				pbnic("unexpected invocbtion of MockKeyVblue.Expire")
			},
		},
		GetFunc: &KeyVblueGetFunc{
			defbultHook: func(string) Vblue {
				pbnic("unexpected invocbtion of MockKeyVblue.Get")
			},
		},
		GetSetFunc: &KeyVblueGetSetFunc{
			defbultHook: func(string, interfbce{}) Vblue {
				pbnic("unexpected invocbtion of MockKeyVblue.GetSet")
			},
		},
		HDelFunc: &KeyVblueHDelFunc{
			defbultHook: func(string, string) Vblue {
				pbnic("unexpected invocbtion of MockKeyVblue.HDel")
			},
		},
		HGetFunc: &KeyVblueHGetFunc{
			defbultHook: func(string, string) Vblue {
				pbnic("unexpected invocbtion of MockKeyVblue.HGet")
			},
		},
		HGetAllFunc: &KeyVblueHGetAllFunc{
			defbultHook: func(string) Vblues {
				pbnic("unexpected invocbtion of MockKeyVblue.HGetAll")
			},
		},
		HSetFunc: &KeyVblueHSetFunc{
			defbultHook: func(string, string, interfbce{}) error {
				pbnic("unexpected invocbtion of MockKeyVblue.HSet")
			},
		},
		IncrFunc: &KeyVblueIncrFunc{
			defbultHook: func(string) (int, error) {
				pbnic("unexpected invocbtion of MockKeyVblue.Incr")
			},
		},
		IncrbyFunc: &KeyVblueIncrbyFunc{
			defbultHook: func(string, int) (int, error) {
				pbnic("unexpected invocbtion of MockKeyVblue.Incrby")
			},
		},
		LLenFunc: &KeyVblueLLenFunc{
			defbultHook: func(string) (int, error) {
				pbnic("unexpected invocbtion of MockKeyVblue.LLen")
			},
		},
		LPushFunc: &KeyVblueLPushFunc{
			defbultHook: func(string, interfbce{}) error {
				pbnic("unexpected invocbtion of MockKeyVblue.LPush")
			},
		},
		LRbngeFunc: &KeyVblueLRbngeFunc{
			defbultHook: func(string, int, int) Vblues {
				pbnic("unexpected invocbtion of MockKeyVblue.LRbnge")
			},
		},
		LTrimFunc: &KeyVblueLTrimFunc{
			defbultHook: func(string, int, int) error {
				pbnic("unexpected invocbtion of MockKeyVblue.LTrim")
			},
		},
		PoolFunc: &KeyVbluePoolFunc{
			defbultHook: func() (*redis.Pool, bool) {
				pbnic("unexpected invocbtion of MockKeyVblue.Pool")
			},
		},
		SetFunc: &KeyVblueSetFunc{
			defbultHook: func(string, interfbce{}) error {
				pbnic("unexpected invocbtion of MockKeyVblue.Set")
			},
		},
		SetExFunc: &KeyVblueSetExFunc{
			defbultHook: func(string, int, interfbce{}) error {
				pbnic("unexpected invocbtion of MockKeyVblue.SetEx")
			},
		},
		SetNxFunc: &KeyVblueSetNxFunc{
			defbultHook: func(string, interfbce{}) (bool, error) {
				pbnic("unexpected invocbtion of MockKeyVblue.SetNx")
			},
		},
		TTLFunc: &KeyVblueTTLFunc{
			defbultHook: func(string) (int, error) {
				pbnic("unexpected invocbtion of MockKeyVblue.TTL")
			},
		},
		WithContextFunc: &KeyVblueWithContextFunc{
			defbultHook: func(context.Context) KeyVblue {
				pbnic("unexpected invocbtion of MockKeyVblue.WithContext")
			},
		},
	}
}

// NewMockKeyVblueFrom crebtes b new mock of the MockKeyVblue interfbce. All
// methods delegbte to the given implementbtion, unless overwritten.
func NewMockKeyVblueFrom(i KeyVblue) *MockKeyVblue {
	return &MockKeyVblue{
		DelFunc: &KeyVblueDelFunc{
			defbultHook: i.Del,
		},
		ExpireFunc: &KeyVblueExpireFunc{
			defbultHook: i.Expire,
		},
		GetFunc: &KeyVblueGetFunc{
			defbultHook: i.Get,
		},
		GetSetFunc: &KeyVblueGetSetFunc{
			defbultHook: i.GetSet,
		},
		HDelFunc: &KeyVblueHDelFunc{
			defbultHook: i.HDel,
		},
		HGetFunc: &KeyVblueHGetFunc{
			defbultHook: i.HGet,
		},
		HGetAllFunc: &KeyVblueHGetAllFunc{
			defbultHook: i.HGetAll,
		},
		HSetFunc: &KeyVblueHSetFunc{
			defbultHook: i.HSet,
		},
		IncrFunc: &KeyVblueIncrFunc{
			defbultHook: i.Incr,
		},
		IncrbyFunc: &KeyVblueIncrbyFunc{
			defbultHook: i.Incrby,
		},
		LLenFunc: &KeyVblueLLenFunc{
			defbultHook: i.LLen,
		},
		LPushFunc: &KeyVblueLPushFunc{
			defbultHook: i.LPush,
		},
		LRbngeFunc: &KeyVblueLRbngeFunc{
			defbultHook: i.LRbnge,
		},
		LTrimFunc: &KeyVblueLTrimFunc{
			defbultHook: i.LTrim,
		},
		PoolFunc: &KeyVbluePoolFunc{
			defbultHook: i.Pool,
		},
		SetFunc: &KeyVblueSetFunc{
			defbultHook: i.Set,
		},
		SetExFunc: &KeyVblueSetExFunc{
			defbultHook: i.SetEx,
		},
		SetNxFunc: &KeyVblueSetNxFunc{
			defbultHook: i.SetNx,
		},
		TTLFunc: &KeyVblueTTLFunc{
			defbultHook: i.TTL,
		},
		WithContextFunc: &KeyVblueWithContextFunc{
			defbultHook: i.WithContext,
		},
	}
}

// KeyVblueDelFunc describes the behbvior when the Del method of the pbrent
// MockKeyVblue instbnce is invoked.
type KeyVblueDelFunc struct {
	defbultHook func(string) error
	hooks       []func(string) error
	history     []KeyVblueDelFuncCbll
	mutex       sync.Mutex
}

// Del delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) Del(v0 string) error {
	r0 := m.DelFunc.nextHook()(v0)
	m.DelFunc.bppendCbll(KeyVblueDelFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Del method of the
// pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVblueDelFunc) SetDefbultHook(hook func(string) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Del method of the pbrent MockKeyVblue instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *KeyVblueDelFunc) PushHook(hook func(string) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueDelFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(string) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueDelFunc) PushReturn(r0 error) {
	f.PushHook(func(string) error {
		return r0
	})
}

func (f *KeyVblueDelFunc) nextHook() func(string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueDelFunc) bppendCbll(r0 KeyVblueDelFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueDelFuncCbll objects describing the
// invocbtions of this function.
func (f *KeyVblueDelFunc) History() []KeyVblueDelFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueDelFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueDelFuncCbll is bn object thbt describes bn invocbtion of method
// Del on bn instbnce of MockKeyVblue.
type KeyVblueDelFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueDelFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueDelFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// KeyVblueExpireFunc describes the behbvior when the Expire method of the
// pbrent MockKeyVblue instbnce is invoked.
type KeyVblueExpireFunc struct {
	defbultHook func(string, int) error
	hooks       []func(string, int) error
	history     []KeyVblueExpireFuncCbll
	mutex       sync.Mutex
}

// Expire delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) Expire(v0 string, v1 int) error {
	r0 := m.ExpireFunc.nextHook()(v0, v1)
	m.ExpireFunc.bppendCbll(KeyVblueExpireFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Expire method of the
// pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVblueExpireFunc) SetDefbultHook(hook func(string, int) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Expire method of the pbrent MockKeyVblue instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *KeyVblueExpireFunc) PushHook(hook func(string, int) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueExpireFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(string, int) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueExpireFunc) PushReturn(r0 error) {
	f.PushHook(func(string, int) error {
		return r0
	})
}

func (f *KeyVblueExpireFunc) nextHook() func(string, int) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueExpireFunc) bppendCbll(r0 KeyVblueExpireFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueExpireFuncCbll objects describing
// the invocbtions of this function.
func (f *KeyVblueExpireFunc) History() []KeyVblueExpireFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueExpireFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueExpireFuncCbll is bn object thbt describes bn invocbtion of
// method Expire on bn instbnce of MockKeyVblue.
type KeyVblueExpireFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueExpireFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueExpireFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// KeyVblueGetFunc describes the behbvior when the Get method of the pbrent
// MockKeyVblue instbnce is invoked.
type KeyVblueGetFunc struct {
	defbultHook func(string) Vblue
	hooks       []func(string) Vblue
	history     []KeyVblueGetFuncCbll
	mutex       sync.Mutex
}

// Get delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) Get(v0 string) Vblue {
	r0 := m.GetFunc.nextHook()(v0)
	m.GetFunc.bppendCbll(KeyVblueGetFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Get method of the
// pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVblueGetFunc) SetDefbultHook(hook func(string) Vblue) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Get method of the pbrent MockKeyVblue instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *KeyVblueGetFunc) PushHook(hook func(string) Vblue) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueGetFunc) SetDefbultReturn(r0 Vblue) {
	f.SetDefbultHook(func(string) Vblue {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueGetFunc) PushReturn(r0 Vblue) {
	f.PushHook(func(string) Vblue {
		return r0
	})
}

func (f *KeyVblueGetFunc) nextHook() func(string) Vblue {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueGetFunc) bppendCbll(r0 KeyVblueGetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueGetFuncCbll objects describing the
// invocbtions of this function.
func (f *KeyVblueGetFunc) History() []KeyVblueGetFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueGetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueGetFuncCbll is bn object thbt describes bn invocbtion of method
// Get on bn instbnce of MockKeyVblue.
type KeyVblueGetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 Vblue
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueGetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueGetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// KeyVblueGetSetFunc describes the behbvior when the GetSet method of the
// pbrent MockKeyVblue instbnce is invoked.
type KeyVblueGetSetFunc struct {
	defbultHook func(string, interfbce{}) Vblue
	hooks       []func(string, interfbce{}) Vblue
	history     []KeyVblueGetSetFuncCbll
	mutex       sync.Mutex
}

// GetSet delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) GetSet(v0 string, v1 interfbce{}) Vblue {
	r0 := m.GetSetFunc.nextHook()(v0, v1)
	m.GetSetFunc.bppendCbll(KeyVblueGetSetFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the GetSet method of the
// pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVblueGetSetFunc) SetDefbultHook(hook func(string, interfbce{}) Vblue) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// GetSet method of the pbrent MockKeyVblue instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *KeyVblueGetSetFunc) PushHook(hook func(string, interfbce{}) Vblue) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueGetSetFunc) SetDefbultReturn(r0 Vblue) {
	f.SetDefbultHook(func(string, interfbce{}) Vblue {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueGetSetFunc) PushReturn(r0 Vblue) {
	f.PushHook(func(string, interfbce{}) Vblue {
		return r0
	})
}

func (f *KeyVblueGetSetFunc) nextHook() func(string, interfbce{}) Vblue {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueGetSetFunc) bppendCbll(r0 KeyVblueGetSetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueGetSetFuncCbll objects describing
// the invocbtions of this function.
func (f *KeyVblueGetSetFunc) History() []KeyVblueGetSetFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueGetSetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueGetSetFuncCbll is bn object thbt describes bn invocbtion of
// method GetSet on bn instbnce of MockKeyVblue.
type KeyVblueGetSetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 interfbce{}
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 Vblue
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueGetSetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueGetSetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// KeyVblueHDelFunc describes the behbvior when the HDel method of the
// pbrent MockKeyVblue instbnce is invoked.
type KeyVblueHDelFunc struct {
	defbultHook func(string, string) Vblue
	hooks       []func(string, string) Vblue
	history     []KeyVblueHDelFuncCbll
	mutex       sync.Mutex
}

// HDel delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) HDel(v0 string, v1 string) Vblue {
	r0 := m.HDelFunc.nextHook()(v0, v1)
	m.HDelFunc.bppendCbll(KeyVblueHDelFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the HDel method of the
// pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVblueHDelFunc) SetDefbultHook(hook func(string, string) Vblue) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// HDel method of the pbrent MockKeyVblue instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *KeyVblueHDelFunc) PushHook(hook func(string, string) Vblue) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueHDelFunc) SetDefbultReturn(r0 Vblue) {
	f.SetDefbultHook(func(string, string) Vblue {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueHDelFunc) PushReturn(r0 Vblue) {
	f.PushHook(func(string, string) Vblue {
		return r0
	})
}

func (f *KeyVblueHDelFunc) nextHook() func(string, string) Vblue {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueHDelFunc) bppendCbll(r0 KeyVblueHDelFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueHDelFuncCbll objects describing the
// invocbtions of this function.
func (f *KeyVblueHDelFunc) History() []KeyVblueHDelFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueHDelFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueHDelFuncCbll is bn object thbt describes bn invocbtion of method
// HDel on bn instbnce of MockKeyVblue.
type KeyVblueHDelFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 Vblue
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueHDelFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueHDelFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// KeyVblueHGetFunc describes the behbvior when the HGet method of the
// pbrent MockKeyVblue instbnce is invoked.
type KeyVblueHGetFunc struct {
	defbultHook func(string, string) Vblue
	hooks       []func(string, string) Vblue
	history     []KeyVblueHGetFuncCbll
	mutex       sync.Mutex
}

// HGet delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) HGet(v0 string, v1 string) Vblue {
	r0 := m.HGetFunc.nextHook()(v0, v1)
	m.HGetFunc.bppendCbll(KeyVblueHGetFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the HGet method of the
// pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVblueHGetFunc) SetDefbultHook(hook func(string, string) Vblue) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// HGet method of the pbrent MockKeyVblue instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *KeyVblueHGetFunc) PushHook(hook func(string, string) Vblue) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueHGetFunc) SetDefbultReturn(r0 Vblue) {
	f.SetDefbultHook(func(string, string) Vblue {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueHGetFunc) PushReturn(r0 Vblue) {
	f.PushHook(func(string, string) Vblue {
		return r0
	})
}

func (f *KeyVblueHGetFunc) nextHook() func(string, string) Vblue {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueHGetFunc) bppendCbll(r0 KeyVblueHGetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueHGetFuncCbll objects describing the
// invocbtions of this function.
func (f *KeyVblueHGetFunc) History() []KeyVblueHGetFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueHGetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueHGetFuncCbll is bn object thbt describes bn invocbtion of method
// HGet on bn instbnce of MockKeyVblue.
type KeyVblueHGetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 Vblue
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueHGetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueHGetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// KeyVblueHGetAllFunc describes the behbvior when the HGetAll method of the
// pbrent MockKeyVblue instbnce is invoked.
type KeyVblueHGetAllFunc struct {
	defbultHook func(string) Vblues
	hooks       []func(string) Vblues
	history     []KeyVblueHGetAllFuncCbll
	mutex       sync.Mutex
}

// HGetAll delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) HGetAll(v0 string) Vblues {
	r0 := m.HGetAllFunc.nextHook()(v0)
	m.HGetAllFunc.bppendCbll(KeyVblueHGetAllFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the HGetAll method of
// the pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVblueHGetAllFunc) SetDefbultHook(hook func(string) Vblues) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// HGetAll method of the pbrent MockKeyVblue instbnce invokes the hook bt
// the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *KeyVblueHGetAllFunc) PushHook(hook func(string) Vblues) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueHGetAllFunc) SetDefbultReturn(r0 Vblues) {
	f.SetDefbultHook(func(string) Vblues {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueHGetAllFunc) PushReturn(r0 Vblues) {
	f.PushHook(func(string) Vblues {
		return r0
	})
}

func (f *KeyVblueHGetAllFunc) nextHook() func(string) Vblues {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueHGetAllFunc) bppendCbll(r0 KeyVblueHGetAllFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueHGetAllFuncCbll objects describing
// the invocbtions of this function.
func (f *KeyVblueHGetAllFunc) History() []KeyVblueHGetAllFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueHGetAllFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueHGetAllFuncCbll is bn object thbt describes bn invocbtion of
// method HGetAll on bn instbnce of MockKeyVblue.
type KeyVblueHGetAllFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 Vblues
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueHGetAllFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueHGetAllFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// KeyVblueHSetFunc describes the behbvior when the HSet method of the
// pbrent MockKeyVblue instbnce is invoked.
type KeyVblueHSetFunc struct {
	defbultHook func(string, string, interfbce{}) error
	hooks       []func(string, string, interfbce{}) error
	history     []KeyVblueHSetFuncCbll
	mutex       sync.Mutex
}

// HSet delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) HSet(v0 string, v1 string, v2 interfbce{}) error {
	r0 := m.HSetFunc.nextHook()(v0, v1, v2)
	m.HSetFunc.bppendCbll(KeyVblueHSetFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the HSet method of the
// pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVblueHSetFunc) SetDefbultHook(hook func(string, string, interfbce{}) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// HSet method of the pbrent MockKeyVblue instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *KeyVblueHSetFunc) PushHook(hook func(string, string, interfbce{}) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueHSetFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(string, string, interfbce{}) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueHSetFunc) PushReturn(r0 error) {
	f.PushHook(func(string, string, interfbce{}) error {
		return r0
	})
}

func (f *KeyVblueHSetFunc) nextHook() func(string, string, interfbce{}) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueHSetFunc) bppendCbll(r0 KeyVblueHSetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueHSetFuncCbll objects describing the
// invocbtions of this function.
func (f *KeyVblueHSetFunc) History() []KeyVblueHSetFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueHSetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueHSetFuncCbll is bn object thbt describes bn invocbtion of method
// HSet on bn instbnce of MockKeyVblue.
type KeyVblueHSetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 string
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 interfbce{}
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueHSetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueHSetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// KeyVblueIncrFunc describes the behbvior when the Incr method of the
// pbrent MockKeyVblue instbnce is invoked.
type KeyVblueIncrFunc struct {
	defbultHook func(string) (int, error)
	hooks       []func(string) (int, error)
	history     []KeyVblueIncrFuncCbll
	mutex       sync.Mutex
}

// Incr delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) Incr(v0 string) (int, error) {
	r0, r1 := m.IncrFunc.nextHook()(v0)
	m.IncrFunc.bppendCbll(KeyVblueIncrFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Incr method of the
// pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVblueIncrFunc) SetDefbultHook(hook func(string) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Incr method of the pbrent MockKeyVblue instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *KeyVblueIncrFunc) PushHook(hook func(string) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueIncrFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(string) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueIncrFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(string) (int, error) {
		return r0, r1
	})
}

func (f *KeyVblueIncrFunc) nextHook() func(string) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueIncrFunc) bppendCbll(r0 KeyVblueIncrFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueIncrFuncCbll objects describing the
// invocbtions of this function.
func (f *KeyVblueIncrFunc) History() []KeyVblueIncrFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueIncrFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueIncrFuncCbll is bn object thbt describes bn invocbtion of method
// Incr on bn instbnce of MockKeyVblue.
type KeyVblueIncrFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueIncrFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueIncrFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// KeyVblueIncrbyFunc describes the behbvior when the Incrby method of the
// pbrent MockKeyVblue instbnce is invoked.
type KeyVblueIncrbyFunc struct {
	defbultHook func(string, int) (int, error)
	hooks       []func(string, int) (int, error)
	history     []KeyVblueIncrbyFuncCbll
	mutex       sync.Mutex
}

// Incrby delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) Incrby(v0 string, v1 int) (int, error) {
	r0, r1 := m.IncrbyFunc.nextHook()(v0, v1)
	m.IncrbyFunc.bppendCbll(KeyVblueIncrbyFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Incrby method of the
// pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVblueIncrbyFunc) SetDefbultHook(hook func(string, int) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Incrby method of the pbrent MockKeyVblue instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *KeyVblueIncrbyFunc) PushHook(hook func(string, int) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueIncrbyFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(string, int) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueIncrbyFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(string, int) (int, error) {
		return r0, r1
	})
}

func (f *KeyVblueIncrbyFunc) nextHook() func(string, int) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueIncrbyFunc) bppendCbll(r0 KeyVblueIncrbyFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueIncrbyFuncCbll objects describing
// the invocbtions of this function.
func (f *KeyVblueIncrbyFunc) History() []KeyVblueIncrbyFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueIncrbyFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueIncrbyFuncCbll is bn object thbt describes bn invocbtion of
// method Incrby on bn instbnce of MockKeyVblue.
type KeyVblueIncrbyFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueIncrbyFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueIncrbyFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// KeyVblueLLenFunc describes the behbvior when the LLen method of the
// pbrent MockKeyVblue instbnce is invoked.
type KeyVblueLLenFunc struct {
	defbultHook func(string) (int, error)
	hooks       []func(string) (int, error)
	history     []KeyVblueLLenFuncCbll
	mutex       sync.Mutex
}

// LLen delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) LLen(v0 string) (int, error) {
	r0, r1 := m.LLenFunc.nextHook()(v0)
	m.LLenFunc.bppendCbll(KeyVblueLLenFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the LLen method of the
// pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVblueLLenFunc) SetDefbultHook(hook func(string) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// LLen method of the pbrent MockKeyVblue instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *KeyVblueLLenFunc) PushHook(hook func(string) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueLLenFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(string) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueLLenFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(string) (int, error) {
		return r0, r1
	})
}

func (f *KeyVblueLLenFunc) nextHook() func(string) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueLLenFunc) bppendCbll(r0 KeyVblueLLenFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueLLenFuncCbll objects describing the
// invocbtions of this function.
func (f *KeyVblueLLenFunc) History() []KeyVblueLLenFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueLLenFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueLLenFuncCbll is bn object thbt describes bn invocbtion of method
// LLen on bn instbnce of MockKeyVblue.
type KeyVblueLLenFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueLLenFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueLLenFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// KeyVblueLPushFunc describes the behbvior when the LPush method of the
// pbrent MockKeyVblue instbnce is invoked.
type KeyVblueLPushFunc struct {
	defbultHook func(string, interfbce{}) error
	hooks       []func(string, interfbce{}) error
	history     []KeyVblueLPushFuncCbll
	mutex       sync.Mutex
}

// LPush delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) LPush(v0 string, v1 interfbce{}) error {
	r0 := m.LPushFunc.nextHook()(v0, v1)
	m.LPushFunc.bppendCbll(KeyVblueLPushFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the LPush method of the
// pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVblueLPushFunc) SetDefbultHook(hook func(string, interfbce{}) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// LPush method of the pbrent MockKeyVblue instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *KeyVblueLPushFunc) PushHook(hook func(string, interfbce{}) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueLPushFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(string, interfbce{}) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueLPushFunc) PushReturn(r0 error) {
	f.PushHook(func(string, interfbce{}) error {
		return r0
	})
}

func (f *KeyVblueLPushFunc) nextHook() func(string, interfbce{}) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueLPushFunc) bppendCbll(r0 KeyVblueLPushFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueLPushFuncCbll objects describing
// the invocbtions of this function.
func (f *KeyVblueLPushFunc) History() []KeyVblueLPushFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueLPushFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueLPushFuncCbll is bn object thbt describes bn invocbtion of method
// LPush on bn instbnce of MockKeyVblue.
type KeyVblueLPushFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 interfbce{}
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueLPushFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueLPushFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// KeyVblueLRbngeFunc describes the behbvior when the LRbnge method of the
// pbrent MockKeyVblue instbnce is invoked.
type KeyVblueLRbngeFunc struct {
	defbultHook func(string, int, int) Vblues
	hooks       []func(string, int, int) Vblues
	history     []KeyVblueLRbngeFuncCbll
	mutex       sync.Mutex
}

// LRbnge delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) LRbnge(v0 string, v1 int, v2 int) Vblues {
	r0 := m.LRbngeFunc.nextHook()(v0, v1, v2)
	m.LRbngeFunc.bppendCbll(KeyVblueLRbngeFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the LRbnge method of the
// pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVblueLRbngeFunc) SetDefbultHook(hook func(string, int, int) Vblues) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// LRbnge method of the pbrent MockKeyVblue instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *KeyVblueLRbngeFunc) PushHook(hook func(string, int, int) Vblues) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueLRbngeFunc) SetDefbultReturn(r0 Vblues) {
	f.SetDefbultHook(func(string, int, int) Vblues {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueLRbngeFunc) PushReturn(r0 Vblues) {
	f.PushHook(func(string, int, int) Vblues {
		return r0
	})
}

func (f *KeyVblueLRbngeFunc) nextHook() func(string, int, int) Vblues {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueLRbngeFunc) bppendCbll(r0 KeyVblueLRbngeFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueLRbngeFuncCbll objects describing
// the invocbtions of this function.
func (f *KeyVblueLRbngeFunc) History() []KeyVblueLRbngeFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueLRbngeFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueLRbngeFuncCbll is bn object thbt describes bn invocbtion of
// method LRbnge on bn instbnce of MockKeyVblue.
type KeyVblueLRbngeFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 Vblues
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueLRbngeFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueLRbngeFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// KeyVblueLTrimFunc describes the behbvior when the LTrim method of the
// pbrent MockKeyVblue instbnce is invoked.
type KeyVblueLTrimFunc struct {
	defbultHook func(string, int, int) error
	hooks       []func(string, int, int) error
	history     []KeyVblueLTrimFuncCbll
	mutex       sync.Mutex
}

// LTrim delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) LTrim(v0 string, v1 int, v2 int) error {
	r0 := m.LTrimFunc.nextHook()(v0, v1, v2)
	m.LTrimFunc.bppendCbll(KeyVblueLTrimFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the LTrim method of the
// pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVblueLTrimFunc) SetDefbultHook(hook func(string, int, int) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// LTrim method of the pbrent MockKeyVblue instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *KeyVblueLTrimFunc) PushHook(hook func(string, int, int) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueLTrimFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(string, int, int) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueLTrimFunc) PushReturn(r0 error) {
	f.PushHook(func(string, int, int) error {
		return r0
	})
}

func (f *KeyVblueLTrimFunc) nextHook() func(string, int, int) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueLTrimFunc) bppendCbll(r0 KeyVblueLTrimFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueLTrimFuncCbll objects describing
// the invocbtions of this function.
func (f *KeyVblueLTrimFunc) History() []KeyVblueLTrimFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueLTrimFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueLTrimFuncCbll is bn object thbt describes bn invocbtion of method
// LTrim on bn instbnce of MockKeyVblue.
type KeyVblueLTrimFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 int
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueLTrimFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueLTrimFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// KeyVbluePoolFunc describes the behbvior when the Pool method of the
// pbrent MockKeyVblue instbnce is invoked.
type KeyVbluePoolFunc struct {
	defbultHook func() (*redis.Pool, bool)
	hooks       []func() (*redis.Pool, bool)
	history     []KeyVbluePoolFuncCbll
	mutex       sync.Mutex
}

// Pool delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) Pool() (*redis.Pool, bool) {
	r0, r1 := m.PoolFunc.nextHook()()
	m.PoolFunc.bppendCbll(KeyVbluePoolFuncCbll{r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the Pool method of the
// pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVbluePoolFunc) SetDefbultHook(hook func() (*redis.Pool, bool)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Pool method of the pbrent MockKeyVblue instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *KeyVbluePoolFunc) PushHook(hook func() (*redis.Pool, bool)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVbluePoolFunc) SetDefbultReturn(r0 *redis.Pool, r1 bool) {
	f.SetDefbultHook(func() (*redis.Pool, bool) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVbluePoolFunc) PushReturn(r0 *redis.Pool, r1 bool) {
	f.PushHook(func() (*redis.Pool, bool) {
		return r0, r1
	})
}

func (f *KeyVbluePoolFunc) nextHook() func() (*redis.Pool, bool) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVbluePoolFunc) bppendCbll(r0 KeyVbluePoolFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVbluePoolFuncCbll objects describing the
// invocbtions of this function.
func (f *KeyVbluePoolFunc) History() []KeyVbluePoolFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVbluePoolFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVbluePoolFuncCbll is bn object thbt describes bn invocbtion of method
// Pool on bn instbnce of MockKeyVblue.
type KeyVbluePoolFuncCbll struct {
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 *redis.Pool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 bool
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVbluePoolFuncCbll) Args() []interfbce{} {
	return []interfbce{}{}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVbluePoolFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// KeyVblueSetFunc describes the behbvior when the Set method of the pbrent
// MockKeyVblue instbnce is invoked.
type KeyVblueSetFunc struct {
	defbultHook func(string, interfbce{}) error
	hooks       []func(string, interfbce{}) error
	history     []KeyVblueSetFuncCbll
	mutex       sync.Mutex
}

// Set delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) Set(v0 string, v1 interfbce{}) error {
	r0 := m.SetFunc.nextHook()(v0, v1)
	m.SetFunc.bppendCbll(KeyVblueSetFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the Set method of the
// pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVblueSetFunc) SetDefbultHook(hook func(string, interfbce{}) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// Set method of the pbrent MockKeyVblue instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *KeyVblueSetFunc) PushHook(hook func(string, interfbce{}) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueSetFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(string, interfbce{}) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueSetFunc) PushReturn(r0 error) {
	f.PushHook(func(string, interfbce{}) error {
		return r0
	})
}

func (f *KeyVblueSetFunc) nextHook() func(string, interfbce{}) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueSetFunc) bppendCbll(r0 KeyVblueSetFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueSetFuncCbll objects describing the
// invocbtions of this function.
func (f *KeyVblueSetFunc) History() []KeyVblueSetFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueSetFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueSetFuncCbll is bn object thbt describes bn invocbtion of method
// Set on bn instbnce of MockKeyVblue.
type KeyVblueSetFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 interfbce{}
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueSetFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueSetFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// KeyVblueSetExFunc describes the behbvior when the SetEx method of the
// pbrent MockKeyVblue instbnce is invoked.
type KeyVblueSetExFunc struct {
	defbultHook func(string, int, interfbce{}) error
	hooks       []func(string, int, interfbce{}) error
	history     []KeyVblueSetExFuncCbll
	mutex       sync.Mutex
}

// SetEx delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) SetEx(v0 string, v1 int, v2 interfbce{}) error {
	r0 := m.SetExFunc.nextHook()(v0, v1, v2)
	m.SetExFunc.bppendCbll(KeyVblueSetExFuncCbll{v0, v1, v2, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the SetEx method of the
// pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVblueSetExFunc) SetDefbultHook(hook func(string, int, interfbce{}) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SetEx method of the pbrent MockKeyVblue instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *KeyVblueSetExFunc) PushHook(hook func(string, int, interfbce{}) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueSetExFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(string, int, interfbce{}) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueSetExFunc) PushReturn(r0 error) {
	f.PushHook(func(string, int, interfbce{}) error {
		return r0
	})
}

func (f *KeyVblueSetExFunc) nextHook() func(string, int, interfbce{}) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueSetExFunc) bppendCbll(r0 KeyVblueSetExFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueSetExFuncCbll objects describing
// the invocbtions of this function.
func (f *KeyVblueSetExFunc) History() []KeyVblueSetExFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueSetExFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueSetExFuncCbll is bn object thbt describes bn invocbtion of method
// SetEx on bn instbnce of MockKeyVblue.
type KeyVblueSetExFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 int
	// Arg2 is the vblue of the 3rd brgument pbssed to this method
	// invocbtion.
	Arg2 interfbce{}
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueSetExFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1, c.Arg2}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueSetExFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}

// KeyVblueSetNxFunc describes the behbvior when the SetNx method of the
// pbrent MockKeyVblue instbnce is invoked.
type KeyVblueSetNxFunc struct {
	defbultHook func(string, interfbce{}) (bool, error)
	hooks       []func(string, interfbce{}) (bool, error)
	history     []KeyVblueSetNxFuncCbll
	mutex       sync.Mutex
}

// SetNx delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) SetNx(v0 string, v1 interfbce{}) (bool, error) {
	r0, r1 := m.SetNxFunc.nextHook()(v0, v1)
	m.SetNxFunc.bppendCbll(KeyVblueSetNxFuncCbll{v0, v1, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the SetNx method of the
// pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVblueSetNxFunc) SetDefbultHook(hook func(string, interfbce{}) (bool, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// SetNx method of the pbrent MockKeyVblue instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *KeyVblueSetNxFunc) PushHook(hook func(string, interfbce{}) (bool, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueSetNxFunc) SetDefbultReturn(r0 bool, r1 error) {
	f.SetDefbultHook(func(string, interfbce{}) (bool, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueSetNxFunc) PushReturn(r0 bool, r1 error) {
	f.PushHook(func(string, interfbce{}) (bool, error) {
		return r0, r1
	})
}

func (f *KeyVblueSetNxFunc) nextHook() func(string, interfbce{}) (bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueSetNxFunc) bppendCbll(r0 KeyVblueSetNxFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueSetNxFuncCbll objects describing
// the invocbtions of this function.
func (f *KeyVblueSetNxFunc) History() []KeyVblueSetNxFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueSetNxFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueSetNxFuncCbll is bn object thbt describes bn invocbtion of method
// SetNx on bn instbnce of MockKeyVblue.
type KeyVblueSetNxFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 interfbce{}
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 bool
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueSetNxFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueSetNxFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// KeyVblueTTLFunc describes the behbvior when the TTL method of the pbrent
// MockKeyVblue instbnce is invoked.
type KeyVblueTTLFunc struct {
	defbultHook func(string) (int, error)
	hooks       []func(string) (int, error)
	history     []KeyVblueTTLFuncCbll
	mutex       sync.Mutex
}

// TTL delegbtes to the next hook function in the queue bnd stores the
// pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) TTL(v0 string) (int, error) {
	r0, r1 := m.TTLFunc.nextHook()(v0)
	m.TTLFunc.bppendCbll(KeyVblueTTLFuncCbll{v0, r0, r1})
	return r0, r1
}

// SetDefbultHook sets function thbt is cblled when the TTL method of the
// pbrent MockKeyVblue instbnce is invoked bnd the hook queue is empty.
func (f *KeyVblueTTLFunc) SetDefbultHook(hook func(string) (int, error)) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// TTL method of the pbrent MockKeyVblue instbnce invokes the hook bt the
// front of the queue bnd discbrds it. After the queue is empty, the defbult
// hook function is invoked for bny future bction.
func (f *KeyVblueTTLFunc) PushHook(hook func(string) (int, error)) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueTTLFunc) SetDefbultReturn(r0 int, r1 error) {
	f.SetDefbultHook(func(string) (int, error) {
		return r0, r1
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueTTLFunc) PushReturn(r0 int, r1 error) {
	f.PushHook(func(string) (int, error) {
		return r0, r1
	})
}

func (f *KeyVblueTTLFunc) nextHook() func(string) (int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueTTLFunc) bppendCbll(r0 KeyVblueTTLFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueTTLFuncCbll objects describing the
// invocbtions of this function.
func (f *KeyVblueTTLFunc) History() []KeyVblueTTLFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueTTLFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueTTLFuncCbll is bn object thbt describes bn invocbtion of method
// TTL on bn instbnce of MockKeyVblue.
type KeyVblueTTLFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 string
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 int
	// Result1 is the vblue of the 2nd result returned from this method
	// invocbtion.
	Result1 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueTTLFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueTTLFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0, c.Result1}
}

// KeyVblueWithContextFunc describes the behbvior when the WithContext
// method of the pbrent MockKeyVblue instbnce is invoked.
type KeyVblueWithContextFunc struct {
	defbultHook func(context.Context) KeyVblue
	hooks       []func(context.Context) KeyVblue
	history     []KeyVblueWithContextFuncCbll
	mutex       sync.Mutex
}

// WithContext delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockKeyVblue) WithContext(v0 context.Context) KeyVblue {
	r0 := m.WithContextFunc.nextHook()(v0)
	m.WithContextFunc.bppendCbll(KeyVblueWithContextFuncCbll{v0, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the WithContext method
// of the pbrent MockKeyVblue instbnce is invoked bnd the hook queue is
// empty.
func (f *KeyVblueWithContextFunc) SetDefbultHook(hook func(context.Context) KeyVblue) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// WithContext method of the pbrent MockKeyVblue instbnce invokes the hook
// bt the front of the queue bnd discbrds it. After the queue is empty, the
// defbult hook function is invoked for bny future bction.
func (f *KeyVblueWithContextFunc) PushHook(hook func(context.Context) KeyVblue) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *KeyVblueWithContextFunc) SetDefbultReturn(r0 KeyVblue) {
	f.SetDefbultHook(func(context.Context) KeyVblue {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *KeyVblueWithContextFunc) PushReturn(r0 KeyVblue) {
	f.PushHook(func(context.Context) KeyVblue {
		return r0
	})
}

func (f *KeyVblueWithContextFunc) nextHook() func(context.Context) KeyVblue {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *KeyVblueWithContextFunc) bppendCbll(r0 KeyVblueWithContextFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of KeyVblueWithContextFuncCbll objects
// describing the invocbtions of this function.
func (f *KeyVblueWithContextFunc) History() []KeyVblueWithContextFuncCbll {
	f.mutex.Lock()
	history := mbke([]KeyVblueWithContextFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// KeyVblueWithContextFuncCbll is bn object thbt describes bn invocbtion of
// method WithContext on bn instbnce of MockKeyVblue.
type KeyVblueWithContextFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 KeyVblue
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c KeyVblueWithContextFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c KeyVblueWithContextFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
