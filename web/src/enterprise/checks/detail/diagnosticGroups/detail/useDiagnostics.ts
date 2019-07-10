import { useEffect, useState } from 'react'
import { Subscription } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { asError, ErrorLike } from '../../../../../../../shared/src/util/errors'
import { DiagnosticInfo, getDiagnosticInfos } from '../../../../threads/detail/backend'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes diagnostics.
 *
 * @param query Only observe diagnostics matching the {@link sourcegraph.DiagnosticQuery}.
 */
export const useDiagnostics = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    query?: sourcegraph.DiagnosticQuery
): typeof LOADING | DiagnosticInfo[] | ErrorLike => {
    const [diagnosticsOrError, setDiagnosticsOrError] = useState<typeof LOADING | DiagnosticInfo[] | ErrorLike>(LOADING)
    useEffect(() => {
        const subscriptions = new Subscription()
        subscriptions.add(
            getDiagnosticInfos(extensionsController, query)
                .pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
                .subscribe(setDiagnosticsOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [extensionsController, query])
    return diagnosticsOrError
}
