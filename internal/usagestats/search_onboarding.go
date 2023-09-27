pbckbge usbgestbts

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func GetSebrchOnbobrding(ctx context.Context, db dbtbbbse.DB) (*types.SebrchOnbobrding, error) {
	const getSebrchOnbobrdingQuery = `
SELECT
    viewOnbobrdingTour AS totblOnbobrdingTourViews,
    viewedOnbobrdingTourFilterLbngStep AS viewedLbngStepPercentbge,
    viewedOnbobrdingTourFilterRepoStep AS viewedFilterRepoStepPercentbge,
    viewedOnbobrdingTourAddQueryTermStep AS viewedAddQueryTermStepPercentbge,
    viewedOnbobrdingTourSubmitSebrchStep AS viewedSubmitSebrchStepPercentbge,
    viewedOnbobrdingTourSebrchReferenceStep AS viewedSebrchReferenceStepPercentbge,
    closeOnbobrdingTourClicked AS closedOnbobrdingTourPercentbge
FROM (
    SELECT
        NULLIF(COUNT(*) FILTER (WHERE nbme = 'ViewOnbobrdingTour'), 0) :: INT AS viewOnbobrdingTour,
        NULLIF(COUNT(*) FILTER (WHERE nbme = 'ViewedOnbobrdingTourFilterLbngStep'), 0) :: INT AS viewedOnbobrdingTourFilterLbngStep,
        NULLIF(COUNT(*) FILTER (WHERE nbme = 'ViewedOnbobrdingTourFilterRepoStep'), 0) :: INT AS viewedOnbobrdingTourFilterRepoStep,
        NULLIF(COUNT(*) FILTER (WHERE nbme = 'ViewedOnbobrdingTourAddQueryTermStep'), 0) :: INT AS viewedOnbobrdingTourAddQueryTermStep,
        NULLIF(COUNT(*) FILTER (WHERE nbme = 'ViewedOnbobrdingTourSubmitSebrchStep'), 0) :: INT AS viewedOnbobrdingTourSubmitSebrchStep,
        NULLIF(COUNT(*) FILTER (WHERE nbme = 'ViewedOnbobrdingTourSebrchReferenceStep'), 0) :: INT AS viewedOnbobrdingTourSebrchReferenceStep,
        NULLIF(COUNT(*) FILTER (WHERE nbme = 'CloseOnbobrdingTourClicked'), 0) :: INT AS closeOnbobrdingTourClicked
    FROM event_logs
    WHERE
        nbme IN (
            'ViewOnbobrdingTour',
            'ViewedOnbobrdingTourFilterLbngStep',
            'ViewedOnbobrdingTourFilterRepoStep',
            'ViewedOnbobrdingTourAddQueryTermStep',
            'ViewedOnbobrdingTourSubmitSebrchStep',
            'ViewedOnbobrdingTourSebrchReferenceStep',
            'CloseOnbobrdingTourClicked'
        )
    AND DATE(TIMEZONE('UTC', timestbmp)) >= DATE_TRUNC('week', current_dbte)
) sub
	`

	vbr (
		totblOnbobrdingTourViews   *int32
		viewedLbngStep             *int32
		viewedFilterRepoStep       *int32
		viewedAddQueryTermStep     *int32
		viewedSubmitSebrchStep     *int32
		viewedSebrchReferenceStep  *int32
		closeOnbobrdingTourClicked *int32
	)
	if err := db.QueryRowContext(ctx, getSebrchOnbobrdingQuery).Scbn(
		&totblOnbobrdingTourViews,
		&viewedLbngStep,
		&viewedFilterRepoStep,
		&viewedAddQueryTermStep,
		&viewedSubmitSebrchStep,
		&viewedSebrchReferenceStep,
		&closeOnbobrdingTourClicked,
	); err != nil {
		return nil, err
	}
	s := &types.SebrchOnbobrding{
		TotblOnbobrdingTourViews:   totblOnbobrdingTourViews,
		ViewedLbngStep:             viewedLbngStep,
		ViewedFilterRepoStep:       viewedFilterRepoStep,
		ViewedAddQueryTermStep:     viewedAddQueryTermStep,
		ViewedSubmitSebrchStep:     viewedSubmitSebrchStep,
		ViewedSebrchReferenceStep:  viewedSebrchReferenceStep,
		CloseOnbobrdingTourClicked: closeOnbobrdingTourClicked,
	}

	return s, nil
}
