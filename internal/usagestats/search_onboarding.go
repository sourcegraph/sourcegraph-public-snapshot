package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/types"
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
		totalOnboardingTourViews   *int32
		viewedLangStep             *int32
		viewedFilterRepoStep       *int32
		viewedAddQueryTermStep     *int32
		viewedSubmitSearchStep     *int32
		viewedSearchReferenceStep  *int32
		closeOnboardingTourClicked *int32
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
		TotalOnboardingTourViews:   totalOnboardingTourViews,
		ViewedLangStep:             viewedLangStep,
		ViewedFilterRepoStep:       viewedFilterRepoStep,
		ViewedAddQueryTermStep:     viewedAddQueryTermStep,
		ViewedSubmitSearchStep:     viewedSubmitSearchStep,
		ViewedSearchReferenceStep:  viewedSearchReferenceStep,
		CloseOnboardingTourClicked: closeOnboardingTourClicked,
	}

	return s, nil
}
