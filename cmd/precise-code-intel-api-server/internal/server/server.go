package server

import (
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/api"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

type Server struct {
	Host                string
	Port                int
	DB                  db.DB
	BundleManagerClient bundles.BundleManagerClient
	CodeIntelAPI        api.CodeIntelAPI
}

func (s *Server) Start() {
	addr := net.JoinHostPort(s.Host, strconv.FormatInt(int64(s.Port), 10))
	handler := ot.Middleware(s.handler())
	server := &http.Server{Addr: addr, Handler: handler}

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log15.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
