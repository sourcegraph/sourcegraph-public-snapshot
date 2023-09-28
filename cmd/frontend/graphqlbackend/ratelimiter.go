package graphqlbackend

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
)

type rateLimiterStateResolver struct {
	state ratelimit.GlobalLimiterInfo
}

func (rl *rateLimiterStateResolver) Burst() int32 {
	return int32(rl.state.Burst)
}

func (rl *rateLimiterStateResolver) CurrentCapacity() int32 {
	return int32(rl.state.CurrentCapacity)
}

func (rl *rateLimiterStateResolver) Infinite() bool {
	return rl.state.Infinite
}

func (rl *rateLimiterStateResolver) Interval() int32 {
	return int32(rl.state.Interval / time.Second)
}

func (rl *rateLimiterStateResolver) LastReplenishment() gqlutil.DateTime {
	return gqlutil.DateTime{Time: rl.state.LastReplenishment}
}

func (rl *rateLimiterStateResolver) Limit() int32 {
	return int32(rl.state.Limit)
}
