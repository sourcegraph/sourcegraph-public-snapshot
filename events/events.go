package events

import (
	"errors"
	"log"
	"reflect"
	"sync"
)

type EventID string

type Event struct {
	EventID
	Payload interface{}
}

type eventServer struct {
	*sync.Mutex
	callbacks map[EventID][]interface{}
}

func (s *eventServer) dispatch(e Event) {
	s.Lock()
	defer s.Unlock()

	pv := reflect.ValueOf(e.Payload)

	for _, callback := range s.callbacks[e.EventID] {
		cv := reflect.ValueOf(callback)
		if cv.Type().In(0) != pv.Type() {
			log.Printf("warning: event dispatcher type mismatch for callback type '%s', payload type '%s'", cv.Type(), pv.Type())
			continue
		}
		args := []reflect.Value{pv}
		go cv.Call(args)
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
		callbacks: make(map[EventID][]interface{}),
	}
}

var server *eventServer

func init() {
	server = newEventServer()
}

func Publish(e Event) {
	server.publish(e)
}

func Subscribe(id EventID, callback interface{}) error {
	return server.subscribe(id, callback)
}
