package golang

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	opentracing "github.com/opentracing/opentracing-go"
)

// handlerCommon contains functionality that both the build and lang
// handlers need. They do NOT share the memory of this handlerCommon
// struct; it is just common functionality. (Unlike handlerCommon,
// handlerShared is shared in-memory.)
type handlerCommon struct {
	mu         sync.Mutex // guards all fields
	rootFSPath string     // root path of the project's files in the file system, without the "file://" prefix (typically /src/github.com/foo/bar)
	shutdown   bool
	tracer     opentracing.Tracer
	tracerOK   bool
}

func (h *handlerCommon) reset(rootURI string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.shutdown {
		return errors.New("unable to reset a server that is shutting down")
	}
	if !strings.HasPrefix(rootURI, "file:///") {
		return fmt.Errorf("invalid root path %q: must be file:///", rootURI)
	}
	h.rootFSPath = strings.TrimPrefix(rootURI, "file://") // retain leading slash
	return nil
}

// shutDown marks this server as being shut down and causes all future calls to checkReady to return an error.
func (h *handlerCommon) shutDown() {
	h.mu.Lock()
	if h.shutdown {
		log.Printf("Warning: server received a shutdown request after it was already shut down.")
	}
	h.shutdown = true
	h.mu.Unlock()
}

// checkReady returns an error if the handler has been shut
// down.
func (h *handlerCommon) checkReady() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.shutdown {
		return errors.New("server is shutting down")
	}
	return nil
}
