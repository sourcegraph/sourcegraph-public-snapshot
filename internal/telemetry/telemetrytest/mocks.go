// Code generbted by go-mockgen 1.3.7; DO NOT EDIT.
//
// This file wbs generbted by running `sg generbte` (or `go-mockgen`) bt the root of
// this repository. To bdd bdditionbl mocks to this or bnother pbckbge, bdd b new entry
// to the mockgen.ybml file in the root of this repository.

pbckbge telemetrytest

import (
	"context"
	"sync"

	telemetry "github.com/sourcegrbph/sourcegrbph/internbl/telemetry"
	v1 "github.com/sourcegrbph/sourcegrbph/internbl/telemetrygbtewby/v1"
)

// MockEventsStore is b mock implementbtion of the EventsStore interfbce
// (from the pbckbge github.com/sourcegrbph/sourcegrbph/internbl/telemetry)
// used for unit testing.
type MockEventsStore struct {
	// StoreEventsFunc is bn instbnce of b mock function object controlling
	// the behbvior of the method StoreEvents.
	StoreEventsFunc *EventsStoreStoreEventsFunc
}

// NewMockEventsStore crebtes b new mock of the EventsStore interfbce. All
// methods return zero vblues for bll results, unless overwritten.
func NewMockEventsStore() *MockEventsStore {
	return &MockEventsStore{
		StoreEventsFunc: &EventsStoreStoreEventsFunc{
			defbultHook: func(context.Context, []*v1.Event) (r0 error) {
				return
			},
		},
	}
}

// NewStrictMockEventsStore crebtes b new mock of the EventsStore interfbce.
// All methods pbnic on invocbtion, unless overwritten.
func NewStrictMockEventsStore() *MockEventsStore {
	return &MockEventsStore{
		StoreEventsFunc: &EventsStoreStoreEventsFunc{
			defbultHook: func(context.Context, []*v1.Event) error {
				pbnic("unexpected invocbtion of MockEventsStore.StoreEvents")
			},
		},
	}
}

// NewMockEventsStoreFrom crebtes b new mock of the MockEventsStore
// interfbce. All methods delegbte to the given implementbtion, unless
// overwritten.
func NewMockEventsStoreFrom(i telemetry.EventsStore) *MockEventsStore {
	return &MockEventsStore{
		StoreEventsFunc: &EventsStoreStoreEventsFunc{
			defbultHook: i.StoreEvents,
		},
	}
}

// EventsStoreStoreEventsFunc describes the behbvior when the StoreEvents
// method of the pbrent MockEventsStore instbnce is invoked.
type EventsStoreStoreEventsFunc struct {
	defbultHook func(context.Context, []*v1.Event) error
	hooks       []func(context.Context, []*v1.Event) error
	history     []EventsStoreStoreEventsFuncCbll
	mutex       sync.Mutex
}

// StoreEvents delegbtes to the next hook function in the queue bnd stores
// the pbrbmeter bnd result vblues of this invocbtion.
func (m *MockEventsStore) StoreEvents(v0 context.Context, v1 []*v1.Event) error {
	r0 := m.StoreEventsFunc.nextHook()(v0, v1)
	m.StoreEventsFunc.bppendCbll(EventsStoreStoreEventsFuncCbll{v0, v1, r0})
	return r0
}

// SetDefbultHook sets function thbt is cblled when the StoreEvents method
// of the pbrent MockEventsStore instbnce is invoked bnd the hook queue is
// empty.
func (f *EventsStoreStoreEventsFunc) SetDefbultHook(hook func(context.Context, []*v1.Event) error) {
	f.defbultHook = hook
}

// PushHook bdds b function to the end of hook queue. Ebch invocbtion of the
// StoreEvents method of the pbrent MockEventsStore instbnce invokes the
// hook bt the front of the queue bnd discbrds it. After the queue is empty,
// the defbult hook function is invoked for bny future bction.
func (f *EventsStoreStoreEventsFunc) PushHook(hook func(context.Context, []*v1.Event) error) {
	f.mutex.Lock()
	f.hooks = bppend(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefbultReturn cblls SetDefbultHook with b function thbt returns the
// given vblues.
func (f *EventsStoreStoreEventsFunc) SetDefbultReturn(r0 error) {
	f.SetDefbultHook(func(context.Context, []*v1.Event) error {
		return r0
	})
}

// PushReturn cblls PushHook with b function thbt returns the given vblues.
func (f *EventsStoreStoreEventsFunc) PushReturn(r0 error) {
	f.PushHook(func(context.Context, []*v1.Event) error {
		return r0
	})
}

func (f *EventsStoreStoreEventsFunc) nextHook() func(context.Context, []*v1.Event) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defbultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *EventsStoreStoreEventsFunc) bppendCbll(r0 EventsStoreStoreEventsFuncCbll) {
	f.mutex.Lock()
	f.history = bppend(f.history, r0)
	f.mutex.Unlock()
}

// History returns b sequence of EventsStoreStoreEventsFuncCbll objects
// describing the invocbtions of this function.
func (f *EventsStoreStoreEventsFunc) History() []EventsStoreStoreEventsFuncCbll {
	f.mutex.Lock()
	history := mbke([]EventsStoreStoreEventsFuncCbll, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// EventsStoreStoreEventsFuncCbll is bn object thbt describes bn invocbtion
// of method StoreEvents on bn instbnce of MockEventsStore.
type EventsStoreStoreEventsFuncCbll struct {
	// Arg0 is the vblue of the 1st brgument pbssed to this method
	// invocbtion.
	Arg0 context.Context
	// Arg1 is the vblue of the 2nd brgument pbssed to this method
	// invocbtion.
	Arg1 []*v1.Event
	// Result0 is the vblue of the 1st result returned from this method
	// invocbtion.
	Result0 error
}

// Args returns bn interfbce slice contbining the brguments of this
// invocbtion.
func (c EventsStoreStoreEventsFuncCbll) Args() []interfbce{} {
	return []interfbce{}{c.Arg0, c.Arg1}
}

// Results returns bn interfbce slice contbining the results of this
// invocbtion.
func (c EventsStoreStoreEventsFuncCbll) Results() []interfbce{} {
	return []interfbce{}{c.Result0}
}
