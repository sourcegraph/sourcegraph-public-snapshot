import { useMemo } from 'react'
import { Observable, combineLatest, of } from 'rxjs'
import { Contributions, ViewContribution, Evaluated } from '../../../../shared/src/api/protocol'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { map, delay, startWith } from 'rxjs/operators'
import { ViewService, View } from '../../../../shared/src/api/client/services/viewService'

/**
 * A React hook that returns the view and its associated contents, `undefined` if loading, and
 * `null` if not found. The view must be both registered as a contribution and at runtime.
 */
export const useView = (
    viewID: string,
    viewContainer: ViewContribution['where'],
    params: { [key: string]: string },
    contributions: Observable<Pick<Evaluated<Contributions>, 'views'>>,
    viewService: Pick<ViewService, 'get'>
): View | undefined | null =>
    useObservable(
        useMemo(
            () =>
                combineLatest([
                    contributions.pipe(
                        map(contributions =>
                            contributions.views?.some(({ id, where }) => id === viewID && where === viewContainer)
                        )
                    ),
                    viewService.get(viewID, params),

                    // Wait for extensions to load for up to 5 seconds (grace period) before showing
                    // "not found", to avoid showing an error for a brief period during initial
                    // load.
                    of(false).pipe(delay(5000), startWith(true)),
                ]).pipe(
                    map(
                        ([isContributed, view, isInitialLoadGracePeriod]) =>
                            (isContributed && view) || (isInitialLoadGracePeriod ? undefined : null)
                    )
                ),
            [contributions, viewService, viewID, viewContainer, params]
        )
    )
