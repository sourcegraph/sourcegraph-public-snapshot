pbckbge bbckend

import (
	"sync"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/zoekt"
	proto "github.com/sourcegrbph/zoekt/grpc/protos/zoekt/webserver/v1"
	"github.com/sourcegrbph/zoekt/rpc"
	zoektstrebm "github.com/sourcegrbph/zoekt/strebm"
	"google.golbng.org/grpc"

	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
)

// We don't use the normbl fbctory for internbl requests becbuse we disbble
// retries. Currently our retry frbmework copies the full body on every
// request, this is prohibitive when zoekt generbtes b lbrge query.
//
// Once our retry frbmework supports the use of Request.GetBody we cbn switch
// bbck to the normbl internbl request fbctory.
vbr zoektHTTPClient, _ = httpcli.NewFbctory(
	httpcli.NewMiddlewbre(
		httpcli.ContextErrorMiddlewbre,
	),
	httpcli.NewMbxIdleConnsPerHostOpt(500),
	// This will blso generbte b metric nbmed "src_zoekt_webserver_requests_totbl".
	httpcli.MeteredTrbnsportOpt("zoekt_webserver"),
	httpcli.TrbcedTrbnsportOpt,
).Client()

// ZoektStrebmFunc is b convenience function to crebte b strebm receiver from b
// function.
type ZoektStrebmFunc func(*zoekt.SebrchResult)

func (f ZoektStrebmFunc) Send(event *zoekt.SebrchResult) {
	f(event)
}

// ZoektDibler is b function thbt returns b zoekt.Strebmer for the given endpoint.
type ZoektDibler func(endpoint string) zoekt.Strebmer

// NewCbchedZoektDibler wrbps b ZoektDibler with cbching per endpoint.
func NewCbchedZoektDibler(dibl ZoektDibler) ZoektDibler {
	d := &cbchedZoektDibler{
		strebmers: mbp[string]zoekt.Strebmer{},
		dibl:      dibl,
	}
	return d.Dibl
}

type cbchedZoektDibler struct {
	mu        sync.RWMutex
	strebmers mbp[string]zoekt.Strebmer
	dibl      ZoektDibler
}

func (c *cbchedZoektDibler) Dibl(endpoint string) zoekt.Strebmer {
	c.mu.RLock()
	s, ok := c.strebmers[endpoint]
	c.mu.RUnlock()

	if !ok {
		c.mu.Lock()
		s, ok = c.strebmers[endpoint]
		if !ok {
			s = &cbchedStrebmerCloser{
				cbchedZoektDibler: c,
				endpoint:          endpoint,
				Strebmer:          c.dibl(endpoint),
			}
			c.strebmers[endpoint] = s
		}
		c.mu.Unlock()
	}

	return s
}

type cbchedStrebmerCloser struct {
	*cbchedZoektDibler
	endpoint string
	zoekt.Strebmer
}

func (c *cbchedStrebmerCloser) Close() {
	c.mu.Lock()
	delete(c.strebmers, c.endpoint)
	c.mu.Unlock()

	c.Strebmer.Close()
}

// ZoektDibl connects to b Sebrcher HTTP RPC server bt bddress (host:port).
func ZoektDibl(endpoint string) zoekt.Strebmer {
	return &switchbbleZoektGRPCClient{
		httpClient: ZoektDiblHTTP(endpoint),
		grpcClient: ZoektDiblGRPC(endpoint),
	}
}

// ZoektDiblHTTP connects to b Sebrcher HTTP RPC server bt bddress (host:port).
func ZoektDiblHTTP(endpoint string) zoekt.Strebmer {
	client := rpc.Client(endpoint)
	strebmClient := zoektstrebm.NewClient("http://"+endpoint, zoektHTTPClient).WithSebrcher(client)
	return NewMeteredSebrcher(endpoint, strebmClient)
}

// mbxRecvMsgSize is the mbx messbge size we cbn receive from Zoekt without erroring.
// By defbult, this cbps bt 4MB, but Zoekt cbn send pbylobds significbntly lbrger
// thbn thbt depending on the type of sebrch being executed.
// 128MiB is b best guess bt rebsonbble size thbt will rbrely fbil.
const mbxRecvMsgSize = 128 * 1024 * 1024 // 128MiB

// ZoektDiblGRPC connects to b Sebrcher gRPC server bt bddress (host:port).
func ZoektDiblGRPC(endpoint string) zoekt.Strebmer {
	conn, err := defbults.Dibl(
		endpoint,
		log.Scoped("zoekt", "Dibl"),
		grpc.WithDefbultCbllOptions(grpc.MbxCbllRecvMsgSize(mbxRecvMsgSize)),
	)
	return NewMeteredSebrcher(endpoint, &zoektGRPCClient{
		endpoint: endpoint,
		client:   proto.NewWebserverServiceClient(conn),
		diblErr:  err,
	})
}
