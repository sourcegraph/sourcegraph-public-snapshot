package repodb

import (
	"context"

	"github.com/go-kit/kit/log"
)

// Backend is how the RepoDB opens and creates repositories.
type Backend interface {
	// CanAccess is called to determine if the client can access the
	// given repo (and all of its refs).
	CanAccess(ctx context.Context, logger log.Logger, repo string) (bool, error)
}
