package globalstatedb

import "context"

var Mock = struct {
	Get func(ctx context.Context) (*State, error)
}{}
