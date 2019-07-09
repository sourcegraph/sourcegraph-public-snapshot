import { useEffect, useState } from 'react'
import { combineLatest, NEVER, of, Subscription } from 'rxjs'
import { catchError, delay, startWith, switchMap } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { WrappedStatus } from '../../../../../shared/src/api/client/services/statusService'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'

const LOADING: 'loading' = 'loading'

/**
 * Wait this long for the status provider to be registered (to allow time for the extension to
 * activate) before showing "not found".
 */
const STATUS_PROVIDER_REGISTRATION_DELAY = 5000 // ms

/**
 * A React hook that observes a single status (looked up by name) for a particular scope.
 *
 * @param name The status name.
 * @param scope The scope in which to compute the status.
 */
export const useCheckByTypeForScope = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    name: string,
    scope: sourcegraph.StatusScope | sourcegraph.WorkspaceRoot
): typeof LOADING | WrappedStatus | null | ErrorLike => {
    const [checkOrError, setStatusOrError] = useState<typeof LOADING | WrappedStatus | null | ErrorLike>(LOADING)
    useEffect(() => {
        const subscriptions = new Subscription()
        subscriptions.add(
            combineLatest([
                extensionsController.services.status.observeStatus(name, scope).pipe(startWith(LOADING)),
                of(true).pipe(
                    delay(STATUS_PROVIDER_REGISTRATION_DELAY),
                    startWith(false)
                ),
            ])
                .pipe(
                    switchMap(([status, isDelayElapsed]) => (status || isDelayElapsed ? of(status) : NEVER)),
                    catchError(err => [asError(err)])
                )
                .subscribe(setStatusOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [extensionsController.services.status, scope, name])
    return checkOrError
}
