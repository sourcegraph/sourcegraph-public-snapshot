import { useEffect, useState } from 'react'
import { Subscription } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { WrappedStatus } from '../../../../../shared/src/api/client/services/statusService'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes a single status (looked up by name) for a particular scope.
 *
 * @param name The status name.
 * @param scope The scope in which to compute the status.
 */
export const useStatusByTypeForScope = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    name: string,
    scope: sourcegraph.StatusScope | sourcegraph.WorkspaceRoot
): typeof LOADING | WrappedStatus | null | ErrorLike => {
    const [statusOrError, setStatusOrError] = useState<typeof LOADING | WrappedStatus | null | ErrorLike>(LOADING)
    useEffect(() => {
        const subscriptions = new Subscription()
        subscriptions.add(
            extensionsController.services.status
                .observeStatus(name, scope)
                .pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
                .subscribe(setStatusOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [extensionsController.services.status, scope, name])
    return statusOrError
}
