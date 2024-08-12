package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type surveyResponseConnectionResolver struct {
	db  database.DB
	opt database.SurveyResponseListOptions
}

func (r *schemaResolver) SurveyResponses(args *struct {
	gqlutil.ConnectionArgs
}) *surveyResponseConnectionResolver {
	var opt database.SurveyResponseListOptions
	args.ConnectionArgs.Set(&opt.LimitOffset)
	return &surveyResponseConnectionResolver{db: r.db, opt: opt}
}

func (r *surveyResponseConnectionResolver) Nodes(ctx context.Context) ([]*surveyResponseResolver, error) {
	// 🚨 SECURITY: Survey responses can only be viewed by site admins.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	responses, err := database.SurveyResponses(r.db).GetAll(ctx)
	if err != nil {
		return nil, err
	}

	var surveyResponses []*surveyResponseResolver
	for _, resp := range responses {
		surveyResponses = append(surveyResponses, &surveyResponseResolver{db: r.db, surveyResponse: resp})
	}

	return surveyResponses, nil
}

func (r *surveyResponseConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	// 🚨 SECURITY: Only site admins can count survey responses.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return 0, err
	}

	count, err := database.SurveyResponses(r.db).Count(ctx)
	return int32(count), err
}

func (r *surveyResponseConnectionResolver) AverageScore(ctx context.Context) (float64, error) {
	// 🚨 SECURITY: Only site admins can see average scores.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return 0, err
	}
	return database.SurveyResponses(r.db).Last30DaysAverageScore(ctx)
}

func (r *surveyResponseConnectionResolver) NetPromoterScore(ctx context.Context) (int32, error) {
	// 🚨 SECURITY: Only site admins can see net promoter scores.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return 0, err
	}
	nps, err := database.SurveyResponses(r.db).Last30DaysNetPromoterScore(ctx)
	return int32(nps), err
}

func (r *surveyResponseConnectionResolver) Last30DaysCount(ctx context.Context) (int32, error) {
	// 🚨 SECURITY: Only site admins can count survey responses.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return 0, err
	}
	count, err := database.SurveyResponses(r.db).Last30DaysCount(ctx)
	return int32(count), err
}
