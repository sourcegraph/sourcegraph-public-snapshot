package squirrel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type getDefStackDepthKeyType struct{}

var getDefStackDepthKey = getDefStackDepthKeyType{}

const maxGetDefStackDepth = 10_000

func maxGetDefStackDepthReached(ctx context.Context) error {
	v, ok := ctx.Value(getDefStackDepthKey).(int)
	if !ok {
		return nil
	}
	if v >= maxGetDefStackDepth {
		return errors.New("max get-def stack depth exceeded")
	}
	return nil
}

func withGetDefStackDepth(ctx context.Context, depth int) context.Context {
	return context.WithValue(ctx, getDefStackDepthKey, depth)
}

func incGetDefStackDepth(ctx context.Context) context.Context {
	v, ok := ctx.Value(getDefStackDepthKey).(int)
	if !ok {
		return withGetDefStackDepth(ctx, 0)
	}

	return withGetDefStackDepth(ctx, v+1)
}
