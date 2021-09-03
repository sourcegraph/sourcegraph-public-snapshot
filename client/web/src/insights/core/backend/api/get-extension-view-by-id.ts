import { Remote } from 'comlink'
import { from, Observable } from 'rxjs'
import { switchMap } from 'rxjs/operators'

import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { ViewProviderResult } from '@sourcegraph/shared/src/api/extension/extensionHostApi'

/**
 * Returns view provider result data resolved by id.
 *
 * @param id - view id.
 * @param extensionApi - flat extension host API.
 */
export function getExtensionViewById(
    id: string,
    extensionApi: Promise<Remote<FlatExtensionHostAPI>>
): Observable<ViewProviderResult> {
    return from(extensionApi).pipe(
        switchMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getInsightViewById(id, {})))
    )
}
