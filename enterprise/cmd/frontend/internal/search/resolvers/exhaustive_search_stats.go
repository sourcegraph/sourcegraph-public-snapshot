package resolvers

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

var _ graphqlbackend.ExhaustiveSearchStatsResolver = &exhaustiveSearchStatsResolver{}

type exhaustiveSearchStatsResolver struct {
}

func (e *exhaustiveSearchStatsResolver) Total() int32 {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchStatsResolver) Completed() int32 {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchStatsResolver) Failed() int32 {
	//TODO implement me
	panic("implement me")
}

func (e *exhaustiveSearchStatsResolver) InProgress() int32 {
	//TODO implement me
	panic("implement me")
}
