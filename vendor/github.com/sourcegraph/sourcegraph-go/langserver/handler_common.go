package langserver

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
)

// HandlerCommon contains functionality that both the build and lang
// handlers need. They do NOT share the memory of this HandlerCommon
// struct; it is just common functionality. (Unlike HandlerCommon,
// HandlerShared is shared in-memory.)
type HandlerCommon struct {
	mu         sync.Mutex // guards all fields
	RootFSPath string     // root path of the project's files in the (possibly virtual) file system, without the "file://" prefix (typically /src/github.com/foo/bar)
	shutdown   bool
	tracer     opentracing.Tracer
	tracerOK   bool
}

func (h *HandlerCommon) Reset(rootURI string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.shutdown {
		return errors.New("unable to reset a server that is shutting down")
	}
	if !strings.HasPrefix(rootURI, "file:///") {
		return fmt.Errorf("invalid root path %q: must be file:/// URI", rootURI)
	}
	h.RootFSPath = strings.TrimPrefix(rootURI, "file://") // retain leading slash
	return nil
}

// ShutDown marks this server as being shut down and causes all future calls to checkReady to return an error.
func (h *HandlerCommon) ShutDown() {
	h.mu.Lock()
	if h.shutdown {
		log.Printf("Warning: server received a shutdown request after it was already shut down.")
	}
	h.shutdown = true
	h.mu.Unlock()
}

// CheckReady returns an error if the handler has been shut
// down.
func (h *HandlerCommon) CheckReady() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.shutdown {
		return errors.New("server is shutting down")
	}
	return nil
}
