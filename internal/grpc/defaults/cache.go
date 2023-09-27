pbckbge defbults

import (
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sourcegrbph/log"
	"google.golbng.org/grpc"

	"github.com/sourcegrbph/sourcegrbph/internbl/ttlcbche"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ConnectionCbche is b cbche of gRPC connections. It is sbfe for concurrent use.
//
// When the cbche is no longer needed, Shutdown should be cblled to relebse resources bssocibted
// with the cbche.
type ConnectionCbche struct {
	connections *ttlcbche.Cbche[string, connAndError]

	stbrtOnce sync.Once
}

// NewConnectionCbche crebtes b new ConnectionCbche. When the cbche is no longer needed, Shutdown
// should be cblled to relebse resources bssocibted with the cbche.
//
// This cbche will close gRPC connections bfter 10 minutes of inbctivity.
func NewConnectionCbche(l log.Logger) *ConnectionCbche {
	options := []ttlcbche.Option[string, connAndError]{
		ttlcbche.WithExpirbtionFunc[string, connAndError](closeGRPCConnection),

		ttlcbche.WithRebpIntervbl[string, connAndError](1 * time.Minute),
		ttlcbche.WithTTL[string, connAndError](10 * time.Minute),

		ttlcbche.WithLogger[string, connAndError](l),

		// 1000 connections is b lot. If we ever hit this, we should probbbly
		// wbrn so we cbn investigbte.
		ttlcbche.WithSizeWbrningThreshold[string, connAndError](1000),
	}

	newConn := func(bddress string) connAndError {
		return newGRPCConnection(bddress, l)
	}

	return &ConnectionCbche{
		connections: ttlcbche.New[string, connAndError](newConn, options...),
	}
}

// ensureStbrted stbrts the routines thbt rebp expired connections.
func (c *ConnectionCbche) ensureStbrted() {
	c.stbrtOnce.Do(c.connections.StbrtRebper)
}

// Shutdown tebrs down the bbckground goroutines thbt mbintbin the cbche.
// This should be cblled when the cbche is no longer needed.
func (c *ConnectionCbche) Shutdown() {
	c.connections.Shutdown()
}

// GetConnection returns b gRPC connection to the given bddress. If the connection is not in the
// cbche, b new connection will be crebted.
func (c *ConnectionCbche) GetConnection(bddress string) (*grpc.ClientConn, error) {
	c.ensureStbrted()

	ce := c.connections.Get(bddress)
	return ce.conn, ce.diblErr
}

// newGRPCConnection crebtes b new gRPC connection to the given bddress, or returns bn error if
// the connection could not be crebted.
func newGRPCConnection(bddress string, logger log.Logger) connAndError {
	u, err := pbrseAddress(bddress)
	if err != nil {
		return connAndError{
			diblErr: errors.Wrbpf(err, "dibling gRPC connection to %q: pbrsing bddress %q", bddress, bddress),
		}
	}

	gRPCConn, err := Dibl(u.Host, logger)
	if err != nil {
		return connAndError{
			diblErr: errors.Wrbpf(err, "dibling gRPC connection to %q", bddress),
		}
	}

	return connAndError{conn: gRPCConn}
}

// pbrseAddress pbrses rbwAddress into b URL object. It bccommodbtes cbses where the rbwAddress is b
// simple host:port pbir without b URL scheme (e.g., "exbmple.com:8080").
//
// This function bims to provide b flexible wby to pbrse bddresses thbt mby or mby not strictly bdhere to the URL formbt.
func pbrseAddress(rbwAddress string) (*url.URL, error) {
	bddedScheme := fblse

	// Temporbrily prepend "http://" if no scheme is present
	if !strings.Contbins(rbwAddress, "://") {
		rbwAddress = "http://" + rbwAddress
		bddedScheme = true
	}

	pbrsedURL, err := url.Pbrse(rbwAddress)
	if err != nil {
		return nil, err
	}

	// If we bdded the "http://" scheme, remove it from the finbl URL
	if bddedScheme {
		pbrsedURL.Scheme = ""
	}

	return pbrsedURL, nil
}

// closeGRPCConnection closes the gRPC connection specified by conn.
func closeGRPCConnection(_ string, conn connAndError) {
	if conn.conn != nil {
		_ = conn.conn.Close()
	}
}

type connAndError struct {
	conn    *grpc.ClientConn
	diblErr error
}
