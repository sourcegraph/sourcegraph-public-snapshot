import { from, Observable, of } from 'rxjs'
import { catchError, defaultIfEmpty, map, mergeMap, scan, startWith, switchMap } from 'rxjs/operators'
import sourcegraph from 'sourcegraph'

import { ContributableViewContainer } from '@sourcegraph/client-api'
import { allOf, asError, ErrorLike, isDefined, isExactly, isNot, property } from '@sourcegraph/common'

import { RegisteredViewProvider, ViewContexts, ViewProviderResult } from '../extensionHostApi'

import { providerResultToObservable } from './common'

const DEFAULT_MAX_PARALLEL_QUERIES = 2

/** Type of view provider result from extension stream */
interface NullishViewProviderResult extends Omit<ViewProviderResult, 'view'> {
    /**
     * Since in some cases we may have null response from extension stream
     * we have to mark this empty stream with default null value for view field.
     * Because of that we can't use just ViewProviderResult type.
     * */
    view: sourcegraph.View | undefined | ErrorLike | null
}

/**
 * This is kind of synthetic event to trigger first run of pipe operators flow
 * of callViewProvidersInParallel sequence.
 */
const COLD_START = 'ColdCallViewProviderStart' as const

/**
 * Load view providers in parallel with parallel queries limit.
 * With default value of parallel requests (2) in case if we got 3 provider views
 * In first run this observable immediately returns value above
 * [{ id: 1, views: undefined }, { id: 2, views: undefined }, { id: 3, views: undefined }]
 *
 * Right after that we run first two providers view and in case if second was resolved we get
 * [{ id: 1, views: undefined }, { id: 2, views: {...data} }, { id: 3, views: undefined }]
 *
 * Right after that we run third one while first is still in progress. When first will resolved
 * [{ id: 1, views: {...data} }, { id: 2, views: {...data} }, { id: 3, views: undefined }]
 *
 * And finally the when third request will be resolved
 * [{ id: 1, views: {...data} }, { id: 2, views: {...data} }, { id: 3, views: data }]
 */
export function callViewProvidersInParallel<W extends ContributableViewContainer>(
    context: ViewContexts[W],
    providers: Observable<readonly RegisteredViewProvider<W>[]>,
    maxParallelQueries = DEFAULT_MAX_PARALLEL_QUERIES
): Observable<ViewProviderResult[]> {
    return providers.pipe(
        switchMap(providers =>
            // Add first synthetic observable with null withing to trigger
            // all operators chain immediately in first time
            from(providers).pipe(
                startWith(COLD_START),
                mergeMap(
                    (provider, index) =>
                        provider !== COLD_START
                            ? // Just because we have this first nullable synthetic event we have to avoid
                              // calling provideView on null value
                              providerResultToObservable(provider.viewProvider.provideView(context)).pipe(
                                  defaultIfEmpty<sourcegraph.View | null | undefined>(null),
                                  catchError((error: unknown): [ErrorLike] => {
                                      console.error('View provider errored:', error)

                                      // Pass only primitive copied values because Error object is not
                                      // cloneable in Firefox and Safari
                                      const { message, name, stack } = asError(error)
                                      return [{ message, name, stack } as ErrorLike]
                                  }),
                                  // Add index to view to put response in right position of result views array below in scan operator
                                  map(view => ({ id: provider.id, view, index }))
                              )
                            : of(COLD_START),
                    maxParallelQueries
                ),

                // Collect all responses to one result array
                scan(
                    (accumulator, current) => {
                        // Skip cold start step
                        if (current === COLD_START) {
                            return accumulator
                        }

                        const { index, ...payload } = current

                        // Just because we use null value for as a first synthetic event
                        // we skip this event above and put other elements in right position in
                        // result array.
                        accumulator[index - 1] = payload

                        return accumulator
                    },
                    [
                        ...providers.map(provider => ({ id: provider.id, view: undefined })),
                    ] as NullishViewProviderResult[]
                )
            )
        ),
        // Filter all inappropriate values (nullish value and value with view: null)
        map(views => views.filter(allOf(isDefined, property('view', isNot(isExactly(null))))))
    )
}
