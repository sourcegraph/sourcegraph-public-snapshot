package events

import (
	"code.google.com/p/rog-go/parallel"
	"errors"
	"log"
	"reflect"
	"sync"
)

const maxParallelCallbacks = 8

type EventID string

type Event struct {
	EventID
	Payload interface{}
}

type eventServer struct {
	*sync.Mutex
	parallel  *parallel.Run
	callbacks map[EventID][]interface{}
}

// dispatch asynchronously executes each of the callbacks that are subscribed to a given event's ID.
func (s *eventServer) dispatch(e Event) {
	s.Lock()
	defer s.Unlock()

	pv := reflect.ValueOf(e.Payload)

	for _, callback := range s.callbacks[e.EventID] {
		cv := reflect.ValueOf(callback)
		// Enforce basic type safety by ensuring the first argument of the
		// callback matches the type of the event's payload.
		if cv.Type().In(0) != pv.Type() {
			log.Printf("warning: event dispatcher type mismatch for callback type '%s', payload type '%s'", cv.Type(), pv.Type())
			continue
		}

		args := []reflect.Value{pv}
		go s.parallel.Do(func() error {
			cv.Call(args)
			return nil
		})
	}
}

func (s *eventServer) publish(e Event) {
	s.dispatch(e)
}

func (s *eventServer) subscribe(id EventID, callback interface{}) error {
	s.Lock()
	defer s.Unlock()

	t := reflect.TypeOf(callback)
	if t.Kind() != reflect.Func {
		return errors.New("event callback must be a func")
	}
	if t.NumIn() != 1 {
		return errors.New("event callback must have 1 argument")
	}

	if _, ok := s.callbacks[id]; !ok {
		s.callbacks[id] = make([]interface{}, 0)
	}

	s.callbacks[id] = append(s.callbacks[id], callback)
	return nil
}

func newEventServer() *eventServer {
	return &eventServer{
		Mutex:     &sync.Mutex{},
		parallel:  parallel.NewRun(maxParallelCallbacks),
		callbacks: make(map[EventID][]interface{}),
	}
}

// server is a global eventServer that is shared by all subscribers and
// publishers throughout the application. They can access the server's
// functionality via the API, which is defined as the exported methods of this
// package.
var server *eventServer

func init() {
	server = newEventServer()
}

// Publish globally broadcasts the payload of an event to all of its subscribers.
func Publish(e Event) {
	server.publish(e)
}

// Subscribe stores a supplied callback that will be dispatched each time a
// given EventID is published. The callback must be a func with a single
// argument of any type that corresponds to an event's payload.
func Subscribe(id EventID, callback interface{}) error {
	return server.subscribe(id, callback)
}
