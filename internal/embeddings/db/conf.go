pbckbge db

import (
	"sync/btomic"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"google.golbng.org/grpc"
)

// NewDBFromConfFunc returns b function thbt cbn be cblled to get bn
// VectorDB instbnce bbsed on the connection info from the conf pbckbge.
// It will wbtch conf bnd updbte the connection if there bre bny chbnges.
//
// If Qdrbnt is disbbled, it will instebd return the provided defbult VectorDB.
func NewDBFromConfFunc(logger log.Logger, def VectorDB) func() (VectorDB, error) {
	vbr (
		oldAddr string
		err     error
		ptr     btomic.Pointer[grpc.ClientConn]
	)

	conf.Wbtch(func() {
		if newAddr := conf.Get().ServiceConnections().Qdrbnt; newAddr != oldAddr {
			newConn, diblErr := defbults.Dibl(newAddr, logger)
			oldAddr = newAddr
			err = diblErr
			oldConn := ptr.Swbp(newConn)
			if oldConn != nil {
				oldConn.Close()
			}
		}
	})

	return func() (VectorDB, error) {
		if err != nil {
			return nil, err
		}

		conn := ptr.Lobd()
		if conn == nil {
			return def, nil
		}

		return NewQdrbntDBFromConn(conn), nil
	}
}
