package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

func GetSearchOnboarding(ctx context.Context) (*types.SearchOnboarding, error) {
	const getSearchOnboardingQuery = `
SELECT
    viewOnboardingTour AS totalOnboardingTourViews,
    viewedOnboardingTourFilterLangStep AS viewedLangStepPercentage,
    viewedOnboardingTourFilterRepoStep AS viewedFilterRepoStepPercentage,
    viewedOnboardingTourAddQueryTermStep AS viewedAddQueryTermStepPercentage,
    viewedOnboardingTourSubmitSearchStep AS viewedSubmitSearchStepPercentage,
    viewedOnboardingTourSearchReferenceStep AS viewedSearchReferenceStepPercentage,
    closeOnboardingTourClicked AS closedOnboardingTourPercentage
FROM (
    SELECT
        NULLIF(COUNT(*) FILTER (WHERE name = 'ViewOnboardingTour'), 0) :: INT AS viewOnboardingTour,
        NULLIF(COUNT(*) FILTER (WHERE name = 'ViewedOnboardingTourFilterLangStep'), 0) :: INT AS viewedOnboardingTourFilterLangStep,
        NULLIF(COUNT(*) FILTER (WHERE name = 'ViewedOnboardingTourFilterRepoStep'), 0) :: INT AS viewedOnboardingTourFilterRepoStep,
        NULLIF(COUNT(*) FILTER (WHERE name = 'ViewedOnboardingTourAddQueryTermStep'), 0) :: INT AS viewedOnboardingTourAddQueryTermStep,
        NULLIF(COUNT(*) FILTER (WHERE name = 'ViewedOnboardingTourSubmitSearchStep'), 0) :: INT AS viewedOnboardingTourSubmitSearchStep,
        NULLIF(COUNT(*) FILTER (WHERE name = 'ViewedOnboardingTourSearchReferenceStep'), 0) :: INT AS viewedOnboardingTourSearchReferenceStep,
        NULLIF(COUNT(*) FILTER (WHERE name = 'CloseOnboardingTourClicked'), 0) :: INT AS closeOnboardingTourClicked
    FROM event_logs
    WHERE
        name IN (
            'ViewOnboardingTour',
            'ViewedOnboardingTourFilterLangStep',
            'ViewedOnboardingTourFilterRepoStep',
            'ViewedOnboardingTourAddQueryTermStep',
            'ViewedOnboardingTourSubmitSearchStep',
            'ViewedOnboardingTourSearchReferenceStep',
            'CloseOnboardingTourClicked'
        )
    AND DATE(TIMEZONE('UTC', timestamp)) >= DATE_TRUNC('week', current_date)
) sub
	`

	var (
		totalOnboardingTourViews   int
		viewedLangStep             int
		viewedFilterRepoStep       int
		viewedAddQueryTermStep     int
		viewedSubmitSearchStep     int
		viewedSearchReferenceStep  int
		closeOnboardingTourClicked int
	)
	if err := dbconn.Global.QueryRowContext(ctx, getSearchOnboardingQuery).Scan(
		&totalOnboardingTourViews,
		&viewedLangStep,
		&viewedFilterRepoStep,
		&viewedAddQueryTermStep,
		&viewedSubmitSearchStep,
		&viewedSearchReferenceStep,
		&closeOnboardingTourClicked,
	); err != nil {
		return nil, err
	}
	s := &types.SearchOnboarding{
		TotalOnboardingTourViews:   int32(totalOnboardingTourViews),
		ViewedLangStep:             int32(viewedLangStep),
		ViewedFilterRepoStep:       int32(viewedFilterRepoStep),
		ViewedAddQueryTermStep:     int32(viewedAddQueryTermStep),
		ViewedSubmitSearchStep:     int32(viewedSubmitSearchStep),
		ViewedSearchReferenceStep:  int32(viewedSearchReferenceStep),
		CloseOnboardingTourClicked: int32(closeOnboardingTourClicked),
	}

	return s, nil
}
