package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type surveyResponseConnectionResolver struct {
	opt db.SurveyResponseListOptions
}

func (r *schemaResolver) SurveyResponses(args *struct {
	graphqlutil.ConnectionArgs
}) *surveyResponseConnectionResolver {
	var opt db.SurveyResponseListOptions
	args.ConnectionArgs.Set(&opt.LimitOffset)
	return &surveyResponseConnectionResolver{opt: opt}
}

func (r *surveyResponseConnectionResolver) Nodes(ctx context.Context) ([]*surveyResponseResolver, error) {
	// 🚨 SECURITY: Survey responses can only be viewed by site admins.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
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

	count, err := db.SurveyResponses.Count(ctx)
	return int32(count), err
}

func (r *surveyResponseConnectionResolver) AverageScore(ctx context.Context) (float64, error) {
	// 🚨 SECURITY: Only site admins can see average scores.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return 0, err
	}
	return db.SurveyResponses.Last30DaysAverageScore(ctx)
}

func (r *surveyResponseConnectionResolver) NetPromoterScore(ctx context.Context) (int32, error) {
	// 🚨 SECURITY: Only site admins can see net promoter scores.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return 0, err
	}
	nps, err := db.SurveyResponses.Last30DaysNetPromoterScore(ctx)
	return int32(nps), err
}

func (r *surveyResponseConnectionResolver) Last30DaysCount(ctx context.Context) (int32, error) {
	// 🚨 SECURITY: Only site admins can count survey responses.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return 0, err
	}
	count, err := db.SurveyResponses.Last30DaysCount(ctx)
	return int32(count), err
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_227(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
