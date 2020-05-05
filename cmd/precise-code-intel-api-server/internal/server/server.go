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
	host                string
	port                int
	db                  db.DB
	bundleManagerClient bundles.BundleManagerClient
	api                 api.CodeIntelAPI
}

type ServerOpts struct {
	Host                string
	Port                int
	DB                  db.DB
	BundleManagerClient bundles.BundleManagerClient
}

func New(opts ServerOpts) *Server {
	return &Server{
		host:                opts.Host,
		port:                opts.Port,
		db:                  opts.DB,
		bundleManagerClient: opts.BundleManagerClient,
		api:                 api.New(opts.DB, opts.BundleManagerClient),
	}
}

func (s *Server) Start() {
	addr := net.JoinHostPort(s.host, strconv.FormatInt(int64(s.port), 10))
	handler := ot.Middleware(s.handler())
	server := &http.Server{Addr: addr, Handler: handler}

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log15.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
