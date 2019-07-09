import { useEffect, useState } from 'react'
import { from, Subscription } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import {
    CheckInformationOrError,
    observeChecksInformation,
} from '../../../../../shared/src/api/client/services/checkService'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes all checks for a particular scope.
 *
 * @param scope The scope in which to observe the checks.
 */
export const useChecksForScope = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    scope: sourcegraph.CheckScope | sourcegraph.WorkspaceRoot
): typeof LOADING | CheckInformationOrError[] | ErrorLike => {
    const [checksOrError, setChecksOrError] = useState<typeof LOADING | CheckInformationOrError[] | ErrorLike>(LOADING)
    useEffect(() => {
        const subscriptions = new Subscription()
        subscriptions.add(
            from(observeChecksInformation(extensionsController.services.checks, scope))
                .pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
                .subscribe(setChecksOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [extensionsController.services.checks, scope])
    return checksOrError
}
