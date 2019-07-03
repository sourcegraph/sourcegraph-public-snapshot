import { useEffect, useState } from 'react'
import { Subscription } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { asError, ErrorLike } from '../../../../../shared/src/util/errors'
import { DiagnosticInfo, getDiagnosticInfos } from '../../threads/detail/backend'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes diagnostics.
 *
 * @param diagnosticCollectionName Only observe diagnostics from the named {@link sourcegraph.DiagnosticCollection}.
 */
export const useDiagnostics = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    diagnosticCollectionName?: string
): typeof LOADING | DiagnosticInfo[] | ErrorLike => {
    const [diagnosticsOrError, setDiagnosticsOrError] = useState<typeof LOADING | DiagnosticInfo[] | ErrorLike>(LOADING)
    useEffect(() => {
        const subscriptions = new Subscription()
        subscriptions.add(
            getDiagnosticInfos(extensionsController, diagnosticCollectionName)
                .pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
                .subscribe(setDiagnosticsOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [diagnosticCollectionName, extensionsController])
    return diagnosticsOrError
}
