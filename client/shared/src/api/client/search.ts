// For search-related extension API features, such as query transformers

import { from, Observable } from 'rxjs'
import { filter, first, switchMap } from 'rxjs/operators'

import { Controller as ExtensionsController } from '../../extensions/controller'

import { wrapRemoteObservable } from './api/common'

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
        )
    )
}
