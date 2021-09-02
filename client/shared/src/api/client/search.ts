// For search-related extension API features, such as query transformers

import { from, Observable, of, TimeoutError } from 'rxjs'
import { catchError, filter, first, switchMap, timeout } from 'rxjs/operators'

import { Controller as ExtensionsController } from '../../extensions/controller'

import { wrapRemoteObservable } from './api/common'

const TRANSFORM_QUERY_TIMEOUT = 3000

/**
 * TODO
 */
export function observeTransformedSearchQuery({
    query,
    extensionsController,
}: {
    query: string
    extensionsController: Pick<ExtensionsController, 'extHostAPI'>
}): Observable<string> {
    return from(extensionsController.extHostAPI).pipe(
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
