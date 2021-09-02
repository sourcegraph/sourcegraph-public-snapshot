// For search-related extension API features, such as query transformers

import { Remote } from 'comlink'
import { from, Observable, of, TimeoutError } from 'rxjs'
import { catchError, filter, first, switchMap, timeout } from 'rxjs/operators'

import { FlatExtensionHostAPI } from '../contract'

import { wrapRemoteObservable } from './api/common'

const TRANSFORM_QUERY_TIMEOUT = 3000

/**
 * TODO
 */
export function observeTransformedSearchQuery({
    query,
    extensionHostAPIPromise,
}: {
    query: string
    extensionHostAPIPromise: Promise<Remote<FlatExtensionHostAPI>>
}): Observable<string> {
    return from(extensionHostAPIPromise).pipe(
        switchMap(extensionHostAPI =>
            wrapRemoteObservable(extensionHostAPI.haveInitialExtensionsLoaded()).pipe(
                filter(haveLoaded => haveLoaded),
                first(),
                switchMap(() =>
                    wrapRemoteObservable(extensionHostAPI.transformSearchQuery(query)).pipe(
                        // TODO: explain why
                        first()
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
}
