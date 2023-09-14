package resolvers

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

var _ graphqlbackend.SearchJobStatsResolver = &searchJobStatsResolver{}

type searchJobStatsResolver struct {
	total      int32
	completed  int32
	failed     int32
	inProgress int32
}

func (e *searchJobStatsResolver) Total() int32 {
	return e.total
}

func (e *searchJobStatsResolver) Completed() int32 {
	return e.completed
}

func (e *searchJobStatsResolver) Failed() int32 {
	return e.failed
}

func (e *searchJobStatsResolver) InProgress() int32 {
	return e.inProgress
}
