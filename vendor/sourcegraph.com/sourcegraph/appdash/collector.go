package appdash

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	pio "github.com/gogo/protobuf/io"
	"sourcegraph.com/sourcegraph/appdash/internal/wire"
)

// maxMessageSize is the maximum buffer size for delimited protobuf messages.
// Effectively, the client may request the server to allocate a buffer of up to
// maxMessageSize -- so choose carefully.
//
// We use 1MB here.
const maxMessageSize = 1 * 1024 * 1024

// A Collector collects events that occur in spans.
type Collector interface {
	Collect(SpanID, ...Annotation) error
}

// NewLocalCollector returns a Collector that writes directly to a
// Store.
func NewLocalCollector(s Store) Collector {
	return s
}

// newCollectPacket returns an initialized *wire.CollectPacket given a span and
// set of annotations.
func newCollectPacket(s SpanID, as Annotations) *wire.CollectPacket {
	return &wire.CollectPacket{
		Spanid:     s.wire(),
		Annotation: as.wire(),
	}
}

// A ChunkedCollector groups annotations together that have the same
// span and calls its underlying collector's Collect method with the
// chunked data periodically (instead of immediately).
type ChunkedCollector struct {
	// Collector is the underlying collector that spans are sent to.
	Collector

	// MinInterval is the minimum time period between calls to the
	// underlying collector's Collect method.
	MinInterval time.Duration

	// OnFlush, if non-nil, will be directly invoked at the start of each Flush
	// operation that is performed by this collector.
	//
	// It is primarily used for debugging purposes.
	OnFlush func()

	// The last error from the underlying Collector's Collect method,
	// if any. It will be returned to the next caller of Collect and
	// this field will be set to nil.
	lastErr error

	started, stopped bool
	stopChan         chan struct{}

	pendingBySpanID map[SpanID]Annotations

	// mu protects pendingBySpanID, lastErr, started, stopped, and stopChan.
	mu sync.Mutex
}

// Collect adds the span and annotations to a local buffer until the
// next call to Flush (or when MinInterval elapses), at which point
// they are sent (grouped by span) to the underlying collector.
func (cc *ChunkedCollector) Collect(span SpanID, anns ...Annotation) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.stopped {
		return errors.New("ChunkedCollector is stopped")
	}
	if !cc.started {
		cc.start()
	}

	if cc.pendingBySpanID == nil {
		cc.pendingBySpanID = make(map[SpanID]Annotations)
	}

	if p, present := cc.pendingBySpanID[span]; present {
		if len(anns) > 0 {
			cc.pendingBySpanID[span] = append(p, anns...)
		}
	} else {
		cc.pendingBySpanID[span] = anns
	}

	if err := cc.lastErr; err != nil {
		cc.lastErr = nil
		return err
	}
	return nil
}

// Flush immediately sends all pending spans to the underlying
// collector.
func (cc *ChunkedCollector) Flush() error {
	if cc.OnFlush != nil {
		cc.OnFlush()
	}

	cc.mu.Lock()
	pendingBySpanID := cc.pendingBySpanID
	cc.pendingBySpanID = nil
	cc.mu.Unlock()

	var errs []error
	for spanID, p := range pendingBySpanID {
		if err := cc.Collector.Collect(spanID, p...); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == 1 {
		return errs[0]
	} else if len(errs) > 1 {
		return fmt.Errorf("ChunkedCollector: multiple errors: %v", errs)
	}
	return nil
}

func (cc *ChunkedCollector) start() {
	cc.stopChan = make(chan struct{})
	cc.started = true
	go func() {
		for {
			t := time.After(cc.MinInterval)
			select {
			case <-t:
				if err := cc.Flush(); err != nil {
					cc.mu.Lock()
					cc.lastErr = err
					cc.mu.Unlock()
				}
			case <-cc.stopChan:
				return // stop
			}
		}
	}()
}

// Stop stops the collector. After stopping, no more data will be sent
// to the underlying collector and calls to Collect will fail.
func (cc *ChunkedCollector) Stop() {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	close(cc.stopChan)
	cc.stopped = true
}

// NewRemoteCollector creates a collector that sends data to a
// collector server (created with NewServer). It sends data
// immediately when Collect is called. To send data in chunks, use a
// ChunkedCollector.
func NewRemoteCollector(addr string) *RemoteCollector {
	return &RemoteCollector{
		addr: addr,
		dial: func() (net.Conn, error) {
			return net.Dial("tcp", addr)
		},
	}
}

// NewTLSRemoteCollector creates a RemoteCollector that uses TLS.
func NewTLSRemoteCollector(addr string, tlsConfig *tls.Config) *RemoteCollector {
	return &RemoteCollector{
		addr: addr,
		dial: func() (net.Conn, error) {
			return tls.Dial("tcp", addr, tlsConfig)
		},
	}
}

// A RemoteCollector sends data to a collector server (created with
// NewServer).
type RemoteCollector struct {
	addr string

	dial func() (net.Conn, error)

	mu    sync.Mutex      // guards pconn
	pconn pio.WriteCloser // delimited-protobuf remote connection

	// Log is the logger to use for errors and warnings. If nil, a new
	// logger is created.
	Log   *log.Logger
	logMu sync.Mutex

	// Debug is whether to log debug messages.
	Debug bool
}

// Collect implements the Collector interface by sending the events that
// occured in the span to the remote collector server (see CollectorServer).
func (rc *RemoteCollector) Collect(span SpanID, anns ...Annotation) error {
	return rc.collectAndRetry(newCollectPacket(span, anns))
}

// connect makes a connection to the collector server. It must be
// called with rc.mu held.
func (rc *RemoteCollector) connect() error {
	if rc.pconn != nil {
		rc.pconn.Close()
		rc.pconn = nil
	}

	c, err := rc.dial()
	if err == nil {
		// Create a protobuf delimited writer wrapping the connection. When the
		// writer is closed, it also closes the underlying connection (see
		// source code for details).
		rc.pconn = pio.NewDelimitedWriter(c)
	}
	return err
}

// Close closes the connection to the server.
func (rc *RemoteCollector) Close() error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.pconn != nil {
		err := rc.pconn.Close()
		rc.pconn = nil
		return err
	}
	return nil
}

func (rc *RemoteCollector) collectAndRetry(p *wire.CollectPacket) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	if rc.pconn != nil {
		if err := rc.collect(p); err == nil {
			return nil
		}
		if rc.Debug {
			rc.log().Printf("Reconnecting to send %v", spanIDFromWire(p.Spanid))
		}
	}
	if err := rc.connect(); err != nil {
		return err
	}
	return rc.collect(p)
}

func (rc *RemoteCollector) collect(p *wire.CollectPacket) error {
	if rc.Debug {
		rc.log().Printf("Sending %v", spanIDFromWire(p.Spanid))
	}

	// Send our message, close writer.
	if err := rc.pconn.WriteMsg(p); err != nil {
		return err
	}

	if rc.Debug {
		rc.log().Printf("Sent %v", spanIDFromWire(p.Spanid))
	}
	return nil
}

func (rc *RemoteCollector) log() *log.Logger {
	rc.logMu.Lock()
	defer rc.logMu.Unlock()
	if rc.Log == nil {
		rc.Log = log.New(os.Stderr, fmt.Sprintf("RemoteCollector[%s]: ", rc.addr), log.LstdFlags|log.Lmicroseconds)
	}
	return rc.Log
}

// NewServer creates and starts a new server that listens for
// spans and annotations on l and adds them to the collector c.
//
// Call the CollectorServer's Start method to start listening and
// serving.
func NewServer(l net.Listener, c Collector) *CollectorServer {
	cs := &CollectorServer{c: c, l: l}
	return cs
}

// A CollectorServer listens for spans and annotations and adds them
// to a local collector.
type CollectorServer struct {
	c Collector
	l net.Listener

	// Log is the logger to use for errors and warnings. If nil, a new
	// logger is created.
	Log   *log.Logger
	logMu sync.Mutex

	// Debug is whether to log debug messages.
	Debug bool

	// Trace is whether to log all data that is received.
	Trace bool
}

// Start starts the server.
func (cs *CollectorServer) Start() {
	for {
		conn, err := cs.l.Accept()
		if err != nil {
			cs.log().Printf("Accept: %s", err)
			continue
		}

		if cs.Debug {
			cs.log().Printf("Client %s connected", conn.RemoteAddr())
		}

		go cs.handleConn(conn)
	}
}

func (cs *CollectorServer) handleConn(conn net.Conn) (err error) {
	defer func() {
		if err != nil {
			cs.log().Printf("Client %s: %s", conn.RemoteAddr(), err)
		}
	}()
	defer conn.Close()

	rdr := pio.NewDelimitedReader(conn, maxMessageSize)
	defer rdr.Close()
	for {
		p := &wire.CollectPacket{}
		if err = rdr.ReadMsg(p); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("ReadMsg: %s", err)
		}

		spanID := spanIDFromWire(p.Spanid)
		if cs.Debug || cs.Trace {
			cs.log().Printf("Client %s: received span %v with %d annotations", conn.RemoteAddr(), spanID, len(p.Annotation))
		}
		if cs.Trace {
			for i, ann := range p.Annotation {
				cs.log().Printf("Client %s: span %v: annotation %d: %s=%q", conn.RemoteAddr(), p.Spanid.Span, i, *ann.Key, ann.Value)
			}
		}

		if err = cs.c.Collect(spanID, annotationsFromWire(p.Annotation)...); err != nil {
			return fmt.Errorf("Collect %v: %s", spanID, err)
		}
	}
}

func (cs *CollectorServer) log() *log.Logger {
	cs.logMu.Lock()
	defer cs.logMu.Unlock()
	if cs.Log == nil {
		cs.Log = log.New(os.Stderr, fmt.Sprintf("CollectorServer[%s]: ", cs.l.Addr()), log.LstdFlags|log.Lmicroseconds)
	}
	return cs.Log
}
