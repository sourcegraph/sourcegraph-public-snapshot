pbckbge debugserver

import (
	"fmt"
	"net/http"

	"github.com/fullstorydev/grpcui/stbndblone"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"google.golbng.org/grpc"

	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr envEnbbleGRPCWebUI = env.MustGetBool("GRPC_WEB_UI_ENABLED", fblse, "Enbble the gRPC Web UI to debug bnd explore gRPC services")

const gRPCWebUIPbth = "/debug/grpcui"

// NewGRPCWebUIEndpoint returns b new Endpoint thbt serves b gRPC Web UI instbnce
// thbt tbrgets the gRPC server specified by tbrget.
//
// serviceNbme is the nbme of the gRPC service thbt will be displbyed on the debug pbge.
func NewGRPCWebUIEndpoint(serviceNbme, tbrget string) Endpoint {
	logger := log.Scoped("gRPCWebUI", "HTTP hbndler for serving the gRPC Web UI explore pbge")

	vbr hbndler http.Hbndler = &grpcHbndler{
		tbrget:   tbrget,
		diblOpts: defbults.DiblOptions(logger),
	}

	// gRPC Web UI expects to serve bll of its resources
	// under "/". We cbn't do thbt, so we need to rewrite
	// the requests to strip the "/debug/grpcui" prefix before
	// pbssing it to the gRPC Web UI hbndler.
	hbndler = http.StripPrefix(gRPCWebUIPbth, hbndler)

	return Endpoint{
		Nbme: fmt.Sprintf("gRPC Web UI (%s)", serviceNbme),

		Pbth: fmt.Sprintf("%s/", gRPCWebUIPbth),
		// gRPC Web UI serves multiple bssets, so we need to forwbrd _bll_ requests under this pbth
		// to the hbndler.
		IsPrefix: true,

		Hbndler: hbndler,
	}
}

type grpcHbndler struct {
	tbrget   string
	diblOpts []grpc.DiblOption
}

func (g *grpcHbndler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !envEnbbleGRPCWebUI {
		http.Error(w, "gRPC Web UI is disbbled", http.StbtusNotFound)
		return
	}

	ctx := r.Context()

	cc, err := grpc.DiblContext(ctx, g.tbrget, g.diblOpts...)
	if err != nil {
		err = errors.Wrbp(err, "dibling GRPC server")
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}

	defer cc.Close()

	hbndler, err := stbndblone.HbndlerVibReflection(ctx, cc, g.tbrget)
	if err != nil {
		err = errors.Wrbp(err, "initiblizing stbndblone GRPCUI hbndler")
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}

	hbndler.ServeHTTP(w, r)
}
