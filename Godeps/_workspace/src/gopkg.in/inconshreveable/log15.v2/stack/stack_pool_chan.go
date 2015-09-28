// +build !go1.3 appengine

package stack

const (
	stackPoolSize = 64
)

var (
	pcStackPool = make(chan []uintptr, stackPoolSize)
)

func poolBuf() []uintptr {
	select {
	case p := <-pcStackPool:
		return p
	default:
		return make([]uintptr, 1000)
	}
}

func putPoolBuf(p []uintptr) {
	select {
	case pcStackPool <- p:
	default:
	}
}
