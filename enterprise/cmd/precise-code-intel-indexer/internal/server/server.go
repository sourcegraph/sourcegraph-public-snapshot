package server

import (
	indexmanager "github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer/internal/index_manager"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
)

const addr = ":3189"

type Server struct {
	indexManager indexmanager.Manager
}

func New(indexManager indexmanager.Manager) (goroutine.BackgroundRoutine, error) {
	server := &Server{
		indexManager: indexManager,
	}

	return httpserver.NewFromAddr(addr, httpserver.NewHandler(server.setupRoutes), httpserver.Options{})
}
