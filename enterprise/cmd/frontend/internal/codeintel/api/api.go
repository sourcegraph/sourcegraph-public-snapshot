package api

import (
	"errors"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type CodeIntelAPI struct {
	dbStore         DBStore
	lsifStore       LSIFStore
	gitserverClient GitserverClient
	operations      *operations
}

var ErrMissingDump = errors.New("missing dump")

func New(dbStore DBStore, lsifStore LSIFStore, gitserverClient GitserverClient, observationContext *observation.Context) *CodeIntelAPI {
	return &CodeIntelAPI{
		dbStore:         dbStore,
		lsifStore:       lsifStore,
		gitserverClient: gitserverClient,
		operations:      makeOperations(observationContext),
	}
}
