package resolvers

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

var _ graphqlbackend.InsightBackfillDebugResolver = &insightsBackfillDebugResolver{}
var _ graphqlbackend.InsightBackfillPlanNodeResolver = &insightsBackfillPlanNodeResolver{}

type insightsBackfillDebugResolver struct {
}

func newInsightsBackfillDebugResolver() *insightsBackfillDebugResolver {
	return &insightsBackfillDebugResolver{}
}

func (i insightsBackfillDebugResolver) Plan(ctx context.Context, args graphqlbackend.InsightBackfillDebugArgs) ([]graphqlbackend.InsightBackfillPlanNodeResolver, error) {

	return nil, nil
}

type insightsBackfillPlanNodeResolver struct {
	query            string
	time             time.Time
	childTimes       []time.Time
	generationMethod types.GenerationMethod
}

func (i *insightsBackfillPlanNodeResolver) Query() string {
	// TODO implement me
	panic("implement me")
}

func (i *insightsBackfillPlanNodeResolver) Time() gqlutil.DateTime {
	// TODO implement me
	panic("implement me")
}

func (i *insightsBackfillPlanNodeResolver) ChildTimes() []gqlutil.DateTime {
	// TODO implement me
	panic("implement me")
}

func (i *insightsBackfillPlanNodeResolver) GenerationMethod() string {
	// TODO implement me
	panic("implement me")
}
