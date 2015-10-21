package events

import (
	"errors"
	"log"
	"reflect"
)

type EventID string

type Event struct {
	EventID
	Payload interface{}
}

type subscription struct {
	ID       EventID
	Callback interface{}
	Errc     error
}

type eventServer struct {
	callbacks map[EventID][]interface{}
	publish   chan Event
	register  chan EventID
	subscribe chan subscription
	closing   chan bool
}

func (s *eventServer) dispatch(e Event) {
	pv := reflect.ValueOf(e.Payload)

	for _, callback := range s.callbacks[e.EventID] {
		cv := reflect.ValueOf(callback)
		if cv.Type().In(0) != pv.Type() {
			panic("event payload-callback type mismatch")
		}
		args := []reflect.Value{pv}
		go cv.Call(args)
	}
}

func (s *eventServer) start() {
	for {
		select {
		case id := <-s.register:
			if _, ok := s.callbacks[id]; !ok {
				s.callbacks[id] = make([]interface{}, 0)
			}
		case sub := <-s.subscribe:
			s.callbacks[sub.ID] = append(s.callbacks[sub.ID], sub.Callback)
		case e := <-s.publish:
			s.dispatch(e)
		case <-s.closing:
			log.Println("CLOSING")
			return
		}
	}
}

func (s *eventServer) stop() {
	log.Println("STOP")
	s.closing <- true
}

func newEventServer() *eventServer {
	return &eventServer{
		callbacks: make(map[EventID][]interface{}),
		publish:   make(chan Event),
		register:  make(chan EventID),
		subscribe: make(chan subscription),
		closing:   make(chan bool, 1),
	}
}

var server *eventServer

func init() {
	log.Println("Init")
	server = newEventServer()
	go server.start()
}

func Register(id EventID) {
	server.register <- id
}

func Publish(e Event) {
	server.publish <- e
}

func Subscribe(id EventID, callback interface{}) error {
	t := reflect.TypeOf(callback)
	if t.Kind() != reflect.Func {
		return errors.New("event callback must be a func")
	}
	if t.NumIn() != 1 {
		return errors.New("event callback must have 1 argument")
	}

	server.subscribe <- subscription{ID: id, Callback: callback}
	return nil
}
