import { useMemo, useState } from 'react'
import { merge, of, timer, throwError, NEVER, Observable } from 'rxjs'
import { catchError, concatMap, switchMap, takeWhile, tap } from 'rxjs/operators'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { isDefined } from '../../../../../../shared/src/util/types'
import { useObservable } from '../../../../util/useObservable'
import { fetchCampaignPlanById, previewCampaignPlan } from '../backend'

/**
 * Query and poll a campaign plan until it is no longer processing. If `specOrID` is a spec, it
 * creates a campaign plan from the spec; otherwise it fetches an existing campaign plan by ID. In
 * both cases, it polls.
 */
const queryAndPollCampaignPlan = (
    specOrID: GQL.ICampaignPlanSpecification | GQL.ID
): Observable<GQL.ICampaignPlan | ErrorLike> => {
    const initialResult =
        typeof specOrID === 'string'
            ? fetchCampaignPlanById(specOrID).pipe(
                  switchMap(plan => (plan !== null ? of(plan) : throwError(new Error('Campaign plan not found'))))
              )
            : previewCampaignPlan(specOrID, false)

    return initialResult.pipe(
        switchMap(previewPlan =>
            merge(
                of(previewPlan),
                timer(0, 2000).pipe(
                    concatMap(() => fetchCampaignPlanById(previewPlan.id)),
                    takeWhile(isDefined),
                    takeWhile(plan => plan.status.state === GQL.BackgroundProcessState.PROCESSING, true)
                )
            )
        ),
        catchError(error => of<ErrorLike>(asError(error)))
    )
}

/**
 * A React hook that observes a campaign plan.
 *
 * @param specOrID A campaign plan specification (to preview), an existing campaign plan ID (to
 * fetch), or `undefined` if not known yet.
 * @param queryCampaignPlan For testing only.
 * @returns [campaignPlanOrError, isLoading]
 */
export const useCampaignPlan = (
    specOrID: GQL.ICampaignPlanSpecification | GQL.ID | undefined,
    queryCampaignPlan = queryAndPollCampaignPlan
): [GQL.ICampaignPlan | ErrorLike | undefined, boolean] => {
    const [plan, setPlan] = useState<GQL.ICampaignPlan | ErrorLike | undefined>()
    const [isLoading, setIsLoading] = useState(false)

    useObservable(
        useMemo(() => {
            if (specOrID === undefined) {
                setPlan(undefined)
                setIsLoading(false)
                return NEVER
            }
            return queryCampaignPlan(specOrID).pipe(
                tap(plan => {
                    setPlan(plan)
                    setIsLoading(
                        Boolean(
                            plan === null ||
                                (!isErrorLike(plan) && plan.status.state === GQL.BackgroundProcessState.PROCESSING)
                        )
                    )
                })
            )
        }, [queryCampaignPlan, specOrID])
    )

    return [plan, isLoading]
}
