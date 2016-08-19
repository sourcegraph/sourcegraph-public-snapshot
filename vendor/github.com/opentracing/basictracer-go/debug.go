package basictracer

import (
	"bytes"
	"fmt"
	"runtime"
	"strconv"
	"sync"
)

const debugGoroutineIDTag = "_initial_goroutine"

type errAssertionFailed struct {
	msg string
}

func (s *spanImpl) Lock() {
	s.Mutex.Lock()
	s.maybeAssertSanityLocked()
}

func (s *spanImpl) maybeAssertSanityLocked() {
	if s.tracer == nil {
		s.Mutex.Unlock()
		panic(&errAssertionFailed{
			msg: fmt.Sprintf("span used after Finish()"),
		})
	}
	if s.tracer.options.DebugAssertSingleGoroutine {
		startID := curGoroutineID()
		curID, ok := s.raw.Tags[debugGoroutineIDTag].(uint64)
		if !ok {
			// This is likely invoked in the context of the SetTag which sets
			// debugGoroutineTag.
			return
		}
		if startID != curID {
			s.Mutex.Unlock()
			panic(&errAssertionFailed{
				msg: fmt.Sprintf("span started on goroutine %d, but now running on %d", startID, curID),
			})
		}
	}
}

var goroutineSpace = []byte("goroutine ")
var littleBuf = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 64)
		return &buf
	},
}

// Credit to @bradfitz:
// https://github.com/golang/net/blob/master/http2/gotrack.go#L51
func curGoroutineID() uint64 {
	bp := littleBuf.Get().(*[]byte)
	defer littleBuf.Put(bp)
	b := *bp
	b = b[:runtime.Stack(b, false)]
	// Parse the 4707 out of "goroutine 4707 ["
	b = bytes.TrimPrefix(b, goroutineSpace)
	i := bytes.IndexByte(b, ' ')
	if i < 0 {
		panic(fmt.Sprintf("No space found in %q", b))
	}
	b = b[:i]
	n, err := strconv.ParseUint(string(b), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse goroutine ID out of %q: %v", b, err))
	}
	return n
}
