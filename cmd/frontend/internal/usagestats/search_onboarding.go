package usagestats

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

const getSearchOnboardingQuery = `
SELECT
  viewOnboardingTour                                                    AS totalOnboardingTourViews,
  viewedOnboardingTourFilterLangStep / viewOnboardingTour             AS viewedLangStepPercentage,
  viewedOnboardingTourFilterRepoStep / viewOnboardingTour     AS viewedFilterRepoStepPercentage,
  viewedOnboardingTourAddQueryTermStep / viewOnboardingTour     AS viewedAddQueryTermStepPercentage,
  viewedOnboardingTourSubmitSearchStep / viewOnboardingTour       AS viewedSubmitSearchStepPercentage,
  viewedOnboardingTourSearchReferenceStep / viewOnboardingTour AS viewedSearchReferenceStephesPercentage,
  closeOnboardingTourClicked / viewOnboardingTour AS closedOnboardingTourPercentage,
FROM (
  SELECT
  COUNT(*) FILTER (WHERE name = 'ViewOnboardingTour' AS viewOnboardingTour,
    COUNT(*) FILTER (WHERE name = 'ViewedOnboardingTourFilterLangStep	'        AS viewedOnboardingTourFilterLangStep,
    COUNT(*) FILTER (WHERE name = 'ViewedOnboardingTourFilterRepoStep'             AS viewedOnboardingTourFilterRepoStep,
    COUNT(*) FILTER (WHERE name = 'ViewedOnboardingTourAddQueryTermStep'   AS viewedOnboardingTourAddQueryTermStep,
    COUNT(*) FILTER (WHERE name = 'ViewedOnboardingTourSubmitSearchStep'          AS viewedOnboardingTourSubmitSearchStep,
    COUNT(*) FILTER (WHERE name = 'ViewedOnboardingTourSearchReferenceStep'            AS viewedOnboardingTourSearchReferenceStep,
    COUNT(*) FILTER (WHERE name = 'CloseOnboardingTourClicked'    AS closeOnboardingTourClicked,
  FROM
    event_logs
  WHERE name IN (
	  'ViewOnboardingTour'
	  'ViewedOnboardingTourFilterLangStep'
	  'ViewedOnboardingTourFilterRepoStep'
	  'ViewedOnboardingTourAddQueryTermStep'
	'ViewedOnboardingTourSubmitSearchStep'
	'ViewedOnboardingTourSearchReferenceStep'
	'CloseOnboardingTourClicked'
  ) AND DATE(TIMEZONE('UTC', timestamp)) >= DATE_TRUNC('week', current_date)
) sub
`

func GetSearchOnboarding(ctx context.Context) (*types.SearchOnboarding, error) {
	var s types.SearchOnboarding
	return &s, dbconn.Global.QueryRowContext(ctx, getSearchOnboardingQuery).Scan(
		&s.TotalOnboardingTourViews,
		&s.ViewedLangStepPercentage,
		&s.ViewedFilterRepoStepPercentage,
		&s.ViewedAddQueryTermStepPercentage,
		&s.ViewedSubmitSearchStepPercentage,
		&s.ViewedSearchReferenceStephesPercentage,
	)
}
