import { useEffect, useState } from 'react'
import { Subscription } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { Status } from '../status'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes a single status (looked up by type) for a particular scope.
 *
 * @param type The status type.
 * @param scope The scope in which to compute the status.
 */
export const useStatusByTypeForScope = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    type: string,
    scope: sourcegraph.StatusScope | sourcegraph.WorkspaceRoot
): typeof LOADING | Status | null | ErrorLike => {
    const [statusOrError, setStatusOrError] = useState<typeof LOADING | Status | null | ErrorLike>(LOADING)
    useEffect(() => {
        const subscriptions = new Subscription()
        subscriptions.add(
            extensionsController.services.status
                .observeStatus(type, scope)
                .pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
                .subscribe(setStatusOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [extensionsController.services.status, scope, type])
    return statusOrError
}
