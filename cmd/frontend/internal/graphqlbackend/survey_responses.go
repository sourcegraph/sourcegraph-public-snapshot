package graphqlbackend

import (
	"context"
	"errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
)

type surveyResponseConnectionResolver struct {
	opt db.SurveyResponseListOptions
}

func (r *schemaResolver) SurveyResponses(args *struct {
	connectionArgs
}) *surveyResponseConnectionResolver {
	var opt db.SurveyResponseListOptions
	args.connectionArgs.set(&opt.LimitOffset)
	return &surveyResponseConnectionResolver{opt: opt}
}

func (r *surveyResponseConnectionResolver) Nodes(ctx context.Context) ([]*surveyResponseResolver, error) {
	// 🚨 SECURITY: Survey responses can only be viewed by site admins.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	if !conf.HostSurveysLocallyEnabled() {
		return nil, errors.New("Local user survey management is not enabled.")
	}

	responses, err := db.SurveyResponses.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var surveyResponses []*surveyResponseResolver
	for _, resp := range responses {
		surveyResponses = append(surveyResponses, &surveyResponseResolver{surveyResponse: resp})
	}

	return surveyResponses, nil
}

func (r *surveyResponseConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	// 🚨 SECURITY: Only site admins can count survey responses.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return 0, err
	}
	if !conf.HostSurveysLocallyEnabled() {
		return 0, errors.New("Local user survey management is not enabled.")
	}

	count, err := db.SurveyResponses.Count(ctx)
	return int32(count), err
}
