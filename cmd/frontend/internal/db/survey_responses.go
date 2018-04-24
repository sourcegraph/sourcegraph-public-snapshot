package db

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
)

// SurveyResponseListOptions specifies the options for listing survey responses.
type SurveyResponseListOptions struct {
	*LimitOffset
}

type surveyResponses struct{}

// Create creates a survey response.
func (s *surveyResponses) Create(ctx context.Context, userID *int32, email *string, score int, reason *string, better *string) (id int64, err error) {
	err = globalDB.QueryRowContext(ctx,
		"INSERT INTO survey_responses(user_id, email, score, reason, better) VALUES($1, $2, $3, $4, $5) RETURNING id",
		userID, email, score, reason, better,
	).Scan(&id)
	return id, err
}

func (*surveyResponses) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*types.SurveyResponse, error) {
	rows, err := globalDB.QueryContext(ctx, "SELECT id, user_id, email, score, reason, better, created_at FROM survey_responses "+query, args...)
	if err != nil {
		return nil, err
	}
	responses := []*types.SurveyResponse{}
	defer rows.Close()
	for rows.Next() {
		r := types.SurveyResponse{}
		err := rows.Scan(&r.ID, &r.UserID, &r.Email, &r.Score, &r.Reason, &r.Better, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		responses = append(responses, &r)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return responses, nil
}

// GetAll gets all survey responses.
func (s *surveyResponses) GetAll(ctx context.Context) ([]*types.SurveyResponse, error) {
	return s.getBySQL(ctx, "ORDER BY created_at DESC")
}

// Count returns the count of all survey responses.
func (s *surveyResponses) Count(ctx context.Context) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) FROM survey_responses")

	var count int
	err := globalDB.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count)
	return count, err
}
