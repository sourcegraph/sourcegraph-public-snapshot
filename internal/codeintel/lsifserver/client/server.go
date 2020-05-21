package client

import (
	"github.com/sourcegraph/sourcegraph/internal/codeintel/api"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
)

type Server struct {
	db                  db.DB
	bundleManagerClient bundles.BundleManagerClient
	codeIntelAPI        api.CodeIntelAPI
}

func NewServer(
	db db.DB,
	bundleManagerClient bundles.BundleManagerClient,
	codeIntelAPI api.CodeIntelAPI,
) *Server {
	return &Server{
		db:                  db,
		bundleManagerClient: bundleManagerClient,
		codeIntelAPI:        codeIntelAPI,
	}
}
