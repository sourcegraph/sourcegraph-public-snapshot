// For search-related extension API features, such as query transformers

import { Remote } from 'comlink'
import { from, of, TimeoutError } from 'rxjs'
import { catchError, filter, first, switchMap, timeout } from 'rxjs/operators'

import { FlatExtensionHostAPI } from '../contract'

import { wrapRemoteObservable } from './api/common'

const TRANSFORM_QUERY_TIMEOUT = 3000

/**
 * Executes search query transformers contributed by Sourcegraph extensions.
 */
export function transformSearchQuery({
    query,
    extensionHostAPIPromise,
}: {
    query: string
    extensionHostAPIPromise: Promise<Remote<FlatExtensionHostAPI>>
}): Promise<string> {
    return from(extensionHostAPIPromise)
        .pipe(
            switchMap(extensionHostAPI =>
                // Since we won't re-compute on subsequent extension activation, ensure that
                // at least the initial set of extensions, which should include always-activated
                // query-transforming extensions, have been loaded to ensure that the initial
                // search query is transformed
                wrapRemoteObservable(extensionHostAPI.haveInitialExtensionsLoaded()).pipe(
                    filter(haveLoaded => haveLoaded),
                    first(), // Ensure that it only emits once
                    switchMap(() =>
                        wrapRemoteObservable(extensionHostAPI.transformSearchQuery(query)).pipe(
                            first() // Ensure that it only emits once
                        )
                    )
                )
            ),
            // Timeout: if this is hanging due to any sort of extension bug, it may not result in a thrown error,
            // but will degrade search UX.
            // Wait up to 5 seconds and log to console for users to debug slow query transformer extensions
            timeout(TRANSFORM_QUERY_TIMEOUT),
            catchError(error => {
                if (error instanceof TimeoutError) {
                    console.error(`Extension query transformers took more than ${TRANSFORM_QUERY_TIMEOUT}ms`)
                }
                return of(query)
            })
        )
        .toPromise()
}
