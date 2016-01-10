package sourcegraph

import (
	"sync"

	"google.golang.org/grpc"
)

var (
	connTargetMusMu sync.Mutex
	connTargetMus   map[string]*sync.Mutex

	connsMu sync.Mutex
	conns   map[string]*grpc.ClientConn // keyed on GRPC target (i.e., addr)
)

// lockTargetMutex creates the mutex for target (if it doesn't already
// exist) and obtains the lock. It is used to implement per-target
// locks for pooledGRPCDial.
func lockTargetMutex(target string) *sync.Mutex {
	connTargetMusMu.Lock()
	if connTargetMus == nil {
		connTargetMus = map[string]*sync.Mutex{}
	}
	if _, ok := connTargetMus[target]; !ok {
		connTargetMus[target] = new(sync.Mutex)
	}
	mu := connTargetMus[target]
	mu.Lock()
	connTargetMusMu.Unlock()
	return mu
}

// pooledGRPCDial is a global connection pool for grpc.Dial.
func pooledGRPCDial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// Make sure we are the only goroutine dealing with this target.
	targetMu := lockTargetMutex(target)
	defer targetMu.Unlock()

	connsMu.Lock()
	if conns == nil {
		conns = map[string]*grpc.ClientConn{}
	}
	if conn := conns[target]; conn != nil {
		st, err := conn.State()
		if err == nil && st != grpc.Shutdown {
			connsMu.Unlock()
			return conn, nil
		}
	}
	connsMu.Unlock()

	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err
	}

	connsMu.Lock()
	conns[target] = conn
	connsMu.Unlock()

	return conn, nil
}

// removeConnFromPool deletes the ClientConnection to the specified target
// from the connection pool.
func removeConnFromPool(target string) {
	// Make sure we are the only goroutine dealing with this target.
	targetMu := lockTargetMutex(target)
	defer targetMu.Unlock()

	connsMu.Lock()
	defer connsMu.Unlock()

	if conns != nil {
		conns[target] = nil
	}
}
