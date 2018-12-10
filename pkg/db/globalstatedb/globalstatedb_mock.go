package globalstatedb

import (
	"context"
)

var Mock = struct {
	Get                           func(ctx context.Context) (*State, error)
	AuthenticateManagementConsole func(ctx context.Context, password string) error
}{}
