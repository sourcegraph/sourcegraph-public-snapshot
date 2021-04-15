import { concat, from, Observable, of } from 'rxjs'
import { catchError, debounceTime, defaultIfEmpty, map, mergeMap, scan, switchMap } from 'rxjs/operators'
import sourcegraph from 'sourcegraph'

import { asError, ErrorLike } from '../../../util/errors'
import { allOf, isDefined, isExactly, isNot, property } from '../../../util/types'
import { ContributableViewContainer } from '../../protocol'
import { RegisteredViewProvider, ViewContexts, ViewProviderResult } from '../extensionHostApi'

import { providerResultToObservable } from './common'

const DEFAULT_MAX_PARALLEL_QUERIES = 2

/** Load view providers in parallel with parallel queries limit. */
export function callViewProvidersInParallel<W extends ContributableViewContainer>(
    context: ViewContexts[W],
    providers: Observable<readonly RegisteredViewProvider<W>[]>,
    maxParallelQueries = DEFAULT_MAX_PARALLEL_QUERIES
): Observable<ViewProviderResult[]> {
    return providers.pipe(
        debounceTime(0),
        switchMap(providers =>
            // Add first synthetic observable with null withing to trigger
            // all operators chain immediately in first time
            concat(of(null), from(providers)).pipe(
                mergeMap(
                    (provider, index) =>
                        provider
                            ? // Just because we have this first nullable synthetic event we have to avoid
                              // calling provideView on null value
                              providerResultToObservable(provider.viewProvider.provideView(context)).pipe(
                                  defaultIfEmpty<sourcegraph.View | null | undefined>(null),
                                  catchError((error): [ErrorLike] => {
                                      console.error('View provider errored:', error)
                                      return [asError(error)]
                                  }),
                                  // Add index to view to put response in right position of result views array below in scan operator
                                  map(view => ({ id: provider.id, view, index }))
                              )
                            : of(provider),
                    maxParallelQueries
                ),

                // Collect all responses to one result array
                scan(
                    (accumulator, current) => {
                        // Skip null step
                        if (current === null) {
                            return accumulator
                        }

                        const { index, ...payload } = current

                        accumulator[index] = payload

                        return accumulator
                    },
                    [null, ...providers.map(provider => ({ id: provider.id, view: undefined }))] as [
                        null,
                        undefined | { id: string; view: sourcegraph.View | null | undefined | ErrorLike }
                    ]
                )
            )
        ),
        // Filter all inappropriate values (nullish value and value with view: null)
        map(views => views.filter(allOf(isDefined, property('view', isNot(isExactly(null))))))
    )
}
