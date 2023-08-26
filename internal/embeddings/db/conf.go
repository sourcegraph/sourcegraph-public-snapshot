package db

import (
	"sync/atomic"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"google.golang.org/grpc"
)

// NewDBFromConfFunc returns a function that can be called to get an
// VectorDB instance based on the connection info from the conf package.
// It will watch conf and update the connection if there are any changes.
//
// If Qdrant is disabled, it will instead return the provided default VectorDB.
func NewDBFromConfFunc(logger log.Logger, def VectorDB) func() (VectorDB, error) {
	var (
		oldAddr string
		err     error
		ptr     atomic.Pointer[grpc.ClientConn]
	)

	conf.Watch(func() {
		if newAddr := conf.Get().ServiceConnections().Qdrant; newAddr != oldAddr {
			newConn, dialErr := defaults.Dial(newAddr, logger)
			oldAddr = newAddr
			err = dialErr
			oldConn := ptr.Swap(newConn)
			if oldConn != nil {
				oldConn.Close()
			}
		}
	})

	return func() (VectorDB, error) {
		if err != nil {
			return nil, err
		}

		conn := ptr.Load()
		if conn == nil {
			return def, nil
		}

		return NewQdrantDBFromConn(conn), nil
	}
}
