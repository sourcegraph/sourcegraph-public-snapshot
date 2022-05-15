package resolver

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type resolver struct {
	dependenciesSvc *dependencies.Service
}

var _ goroutine.Handler = &resolver{}
var _ goroutine.ErrorHandler = &resolver{}

func (r *resolver) Handle(ctx context.Context) error {
	// TODO
	return nil
}

func (r *resolver) HandleError(err error) {
}
