import { useEffect, useState } from 'react'
import { from, Subscription } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes diagnostic groups for a check.
 */
export const useCheckDiagnosticGroups = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    checkProvider: Pick<sourcegraph.CheckProvider, 'provideDiagnosticGroups'>
): typeof LOADING | sourcegraph.DiagnosticGroup[] | ErrorLike => {
    const [diagnosticGroupsOrError, setDiagnosticGroupsOrError] = useState<
        typeof LOADING | sourcegraph.DiagnosticGroup[] | ErrorLike
    >(LOADING)
    useEffect(() => {
        const subscriptions = new Subscription()
        subscriptions.add(
            from(checkProvider.provideDiagnosticGroups())
                .pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
                .subscribe(setDiagnosticGroupsOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [checkProvider, extensionsController])
    return diagnosticGroupsOrError
}
