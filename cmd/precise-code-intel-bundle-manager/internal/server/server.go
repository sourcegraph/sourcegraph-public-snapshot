package server

import (
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

type Server struct {
	host                 string
	port                 int
	bundleDir            string
	databaseCache        *database.DatabaseCache
	documentDataCache    *database.DocumentDataCache
	resultChunkDataCache *database.ResultChunkDataCache
}

type ServerOpts struct {
	Host                     string
	Port                     int
	BundleDir                string
	DatabaseCacheSize        int64
	DocumentDataCacheSize    int64
	ResultChunkDataCacheSize int64
}

func New(opts ServerOpts) (*Server, error) {
	databaseCache, err := database.NewDatabaseCache(opts.DatabaseCacheSize)
	if err != nil {
		return nil, err
	}

	documentDataCache, err := database.NewDocumentDataCache(opts.DocumentDataCacheSize)
	if err != nil {
		return nil, err
	}

	resultChunkDataCache, err := database.NewResultChunkDataCache(opts.ResultChunkDataCacheSize)
	if err != nil {
		return nil, err
	}

	return &Server{
		host:                 opts.Host,
		port:                 opts.Port,
		bundleDir:            opts.BundleDir,
		databaseCache:        databaseCache,
		documentDataCache:    documentDataCache,
		resultChunkDataCache: resultChunkDataCache,
	}, nil
}

func (s *Server) Start() {
	addr := net.JoinHostPort(s.host, strconv.FormatInt(int64(s.port), 10))
	handler := ot.Middleware(s.handler())
	server := &http.Server{Addr: addr, Handler: handler}

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log15.Error("Failed to start server", "err", err)
		os.Exit(1)
	}
}
