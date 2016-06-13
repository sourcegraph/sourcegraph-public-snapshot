package chanrpc

import (
	"encoding/gob"
	"errors"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"
)

type action int

const actionSend action = 1
const actionClose action = 2

type chanEntry struct {
	ID      int
	Path    string
	Dir     reflect.ChanDir
	Cap     int
	channel reflect.Value
}

type Server struct {
	Addr        string
	RequestChan interface{}
}

func ListenAndServe(addr string, requestChan interface{}) error {
	srv := &Server{Addr: addr, RequestChan: requestChan}
	return srv.ListenAndServe()
}

func (srv *Server) ListenAndServe() error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}

func (srv *Server) Serve(l net.Listener) error {
	defer l.Close()
	for {
		rw, err := l.Accept()
		if err != nil {
			return err
		}

		go func() {
			c := NewConn(rw, rw)
			c.chanMap[0] = reflect.ValueOf(srv.RequestChan)
			c.receiveValues()
		}()
	}
}

type Conn struct {
	dec            *gob.Decoder
	enc            *gob.Encoder
	encMutex       sync.Mutex
	closer         io.Closer
	chanMap        map[int]reflect.Value
	nextSendChanID int
	nextRecvChanID int
	chanMutex      sync.RWMutex
}

func NewConn(r io.Reader, w io.WriteCloser) *Conn {
	return &Conn{
		dec:            gob.NewDecoder(r),
		enc:            gob.NewEncoder(w),
		closer:         w,
		chanMap:        make(map[int]reflect.Value),
		nextSendChanID: 1,
		nextRecvChanID: -1,
	}
}

func DialAndDeliver(addr string, requestChan interface{}) error {
	rw, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	c := NewConn(rw, rw)
	c.Deliver(requestChan)
	return nil
}

func (c *Conn) Deliver(requestChan interface{}) {
	go c.receiveValues()
	c.deliverChannel(0, reflect.ValueOf(requestChan))
}

func (c *Conn) deliverChannel(chanID int, channel reflect.Value) {
	for {
		x, ok := channel.Recv()
		if !ok {
			break
		}
		if err := c.transmitSend(chanID, x); err != nil {
			c.handleError(err)
			break
		}
	}

	if err := c.transmitClose(chanID); err != nil {
		c.handleError(err)
	}
}

func (c *Conn) transmitSend(chanID int, x reflect.Value) error {
	var chanList []*chanEntry
	extractChannels(x, "", &chanList)

	c.encMutex.Lock()
	defer c.encMutex.Unlock()

	// registerChannels needs to be protected by encMutex because else a value
	// on a receive channel might be sent before the channel itself has
	// reached the other side.
	if err := c.registerChannels(chanList); err != nil {
		return err
	}

	if err := c.enc.Encode(actionSend); err != nil {
		return err
	}
	if err := c.enc.Encode(chanID); err != nil {
		return err
	}
	if err := c.enc.EncodeValue(x); err != nil {
		return err
	}
	if err := c.enc.Encode(chanList); err != nil {
		return err
	}
	return nil
}

func (c *Conn) transmitClose(chanID int) error {
	c.encMutex.Lock()
	defer c.encMutex.Unlock()

	if err := c.enc.Encode(actionClose); err != nil {
		return err
	}
	if err := c.enc.Encode(chanID); err != nil {
		return err
	}
	return nil
}

func extractChannels(x reflect.Value, path string, list *[]*chanEntry) {
	switch x.Kind() {
	case reflect.Ptr:
		extractChannels(x.Elem(), path+".*", list)
	case reflect.Struct:
		for i := 0; i < x.NumField(); i++ {
			extractChannels(x.Field(i), path+"."+x.Type().Field(i).Name, list)
		}
	case reflect.Chan:
		if x.IsNil() {
			return
		}
		*list = append(*list, &chanEntry{
			Path:    path,
			Dir:     x.Type().ChanDir(),
			Cap:     x.Cap(),
			channel: x,
		})
	}
}

func (c *Conn) registerChannels(chanList []*chanEntry) error {
	c.chanMutex.Lock()
	defer c.chanMutex.Unlock()

	for _, entry := range chanList {
		if entry.Cap == 0 {
			return errors.New("chanrpc: type error, channel must be buffered")
		}
		switch entry.Dir {
		case reflect.SendDir:
			if c.chanMap == nil { // connection closed
				entry.channel.Close()
				break
			}
			entry.ID = c.nextSendChanID
			c.nextSendChanID++
			c.chanMap[entry.ID] = entry.channel
		case reflect.RecvDir:
			entry.ID = c.nextRecvChanID
			c.nextRecvChanID--
			go c.deliverChannel(entry.ID, entry.channel)
		case reflect.BothDir:
			return errors.New("chanrpc: type error, channel must have a direction")
		}
	}

	return nil
}

func (c *Conn) receiveValues() {
receiveLoop:
	for {
		var act action
		if err := c.dec.Decode(&act); err != nil {
			c.handleError(err)
			break receiveLoop
		}

		var chanID int
		if err := c.dec.Decode(&chanID); err != nil {
			c.handleError(err)
			break receiveLoop
		}
		c.chanMutex.RLock()
		channel, ok := c.chanMap[chanID]
		c.chanMutex.RUnlock()
		if !ok {
			c.handleError(errors.New("chanrpc: protocol error, unknown channel id"))
			break receiveLoop
		}

		switch act {
		case actionSend:
			x := reflect.New(channel.Type().Elem())
			if err := c.dec.DecodeValue(x); err != nil {
				c.handleError(err)
				break receiveLoop
			}

			var chanList []*chanEntry
			if err := c.dec.Decode(&chanList); err != nil {
				c.handleError(err)
				break receiveLoop
			}
			for _, entry := range chanList {
				target := x.Elem()
				for _, field := range strings.Split(entry.Path, ".")[1:] {
					if field == "*" {
						target = target.Elem()
						continue
					}
					target = target.FieldByName(field)
				}
				if target.Kind() != reflect.Chan {
					c.handleError(errors.New("chanrpc: type error, channel expected"))
					break receiveLoop
				}
				if target.Type().ChanDir() != entry.Dir {
					c.handleError(errors.New("chanrpc: type error, wrong channel direction"))
					break receiveLoop
				}
				newChannel := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, target.Type().Elem()), entry.Cap)
				target.Set(newChannel)
				switch entry.Dir {
				case reflect.SendDir:
					go c.deliverChannel(entry.ID, newChannel)
				case reflect.RecvDir:
					c.chanMutex.Lock()
					c.chanMap[entry.ID] = newChannel
					c.chanMutex.Unlock()
				}
			}

			channel.Send(x.Elem())

		case actionClose:
			channel.Close()
			c.chanMutex.Lock()
			delete(c.chanMap, chanID)
			c.chanMutex.Unlock()

		default:
			c.handleError(errors.New("chanrpc: protocol error, invalid action"))
			break receiveLoop
		}
	}

	c.chanMutex.Lock()
	for id, channel := range c.chanMap {
		if id != 0 {
			channel.Close()
		}
	}
	c.chanMap = nil
	c.chanMutex.Unlock()
}

func (c *Conn) handleError(err error) {
	c.closer.Close()
	if err != io.EOF {
		log.Print(err) // more flexible logging
	}
}
