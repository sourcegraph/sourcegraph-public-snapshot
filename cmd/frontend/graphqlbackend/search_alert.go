package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
)

type searchAlert struct {
	alert *search.Alert
}

func NewSearchAlertResolver(alert *search.Alert) *searchAlert {
	if alert == nil {
		return nil
	}
	return &searchAlert{alert: alert}
}

func (a searchAlert) Title() string { return a.alert.Title }

func (a searchAlert) Description() *string {
	if a.alert.Description == "" {
		return nil
	}
	return &a.alert.Description
}

func (a searchAlert) PrometheusType() string {
	return a.alert.PrometheusType
}

func (a searchAlert) ProposedQueries() *[]*searchQueryDescription {
	if len(a.alert.ProposedQueries) == 0 {
		return nil
	}
	var proposedQueries []*searchQueryDescription
	for _, q := range a.alert.ProposedQueries {
		proposedQueries = append(proposedQueries, &searchQueryDescription{query: q})
	}
	return &proposedQueries
}

func (a searchAlert) wrapResults() *SearchResults {
	return &SearchResults{Alert: &a}
}

func (a searchAlert) wrapSearchImplementer(db database.DB) *alertSearchImplementer {
	return &alertSearchImplementer{
		db:    db,
		alert: a,
	}
}

// alertSearchImplementer is a light wrapper type around an alert that implements
// SearchImplementer. This helps avoid needing to have a db on the searchAlert type
type alertSearchImplementer struct {
	db    database.DB
	alert searchAlert
}

func (a alertSearchImplementer) Results(context.Context) (*SearchResultsResolver, error) {
	return &SearchResultsResolver{db: a.db, SearchResults: a.alert.wrapResults()}, nil
}

func (alertSearchImplementer) Stats(context.Context) (*searchResultsStats, error) { return nil, nil }
func (alertSearchImplementer) Inputs() run.SearchInputs {
	return run.SearchInputs{}
}
