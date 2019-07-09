import { useEffect, useState } from 'react'
import { Subscription } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { CheckWithType } from '../../../../../shared/src/api/client/services/checkService'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes a combined status for a particular scope.
 *
 * @param scope The scope in which to compute the status.
 */
export const useCombinedStatusForScope = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    scope: sourcegraph.CheckScope | sourcegraph.WorkspaceRoot
): typeof LOADING | CheckWithType[] | ErrorLike => {
    const [combinedStatusOrError, setCombinedStatusOrError] = useState<typeof LOADING | CheckWithType[] | ErrorLike>(
        LOADING
    )
    useEffect(() => {
        const subscriptions = new Subscription()
        subscriptions.add(
            extensionsController.services.status
                .observeChecks(scope)
                .pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
                .subscribe(setCombinedStatusOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [extensionsController.services.status, scope])
    return combinedStatusOrError
}
