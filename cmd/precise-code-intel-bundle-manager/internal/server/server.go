package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

const Port = 3187

type Server struct {
	bundleDir          string
	databaseCache      *database.DatabaseCache
	documentCache      *database.DocumentCache
	resultChunkCache   *database.ResultChunkCache
	observationContext *observation.Context
	server             *http.Server
	once               sync.Once
}

func New(
	bundleDir string,
	databaseCache *database.DatabaseCache,
	documentCache *database.DocumentCache,
	resultChunkCache *database.ResultChunkCache,
	observationContext *observation.Context,
) *Server {
	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}

	s := &Server{
		bundleDir:          bundleDir,
		databaseCache:      databaseCache,
		documentCache:      documentCache,
		resultChunkCache:   resultChunkCache,
		observationContext: observationContext,
	}

	s.server = &http.Server{
		Addr:    net.JoinHostPort(host, strconv.FormatInt(int64(Port), 10)),
		Handler: ot.Middleware(s.handler()),
	}

	return s
}

func (s *Server) Start() {
	if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
		log15.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}

func (s *Server) Stop() {
	s.once.Do(func() {
		if err := s.server.Shutdown(context.Background()); err != nil {
			log15.Error("Failed to shutdown server", "error", err)
		}
	})
}
