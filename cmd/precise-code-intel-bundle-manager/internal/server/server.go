package server

import (
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

type Server struct {
	Host                 string
	Port                 int
	BundleDir            string
	DatabaseCache        *database.DatabaseCache
	DocumentDataCache    *database.DocumentDataCache
	ResultChunkDataCache *database.ResultChunkDataCache
	ObservationContext   *observation.Context
}

func (s *Server) Start() {
	addr := net.JoinHostPort(s.Host, strconv.FormatInt(int64(s.Port), 10))
	handler := ot.Middleware(s.handler())
	server := &http.Server{Addr: addr, Handler: handler}

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log15.Error("Failed to start server", "err", err)
		os.Exit(1)
	}
}
