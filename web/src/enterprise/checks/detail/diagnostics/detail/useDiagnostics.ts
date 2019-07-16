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
 * @param type Only observe diagnostics matching the type.
 */
export const useDiagnostics = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    type?: sourcegraph.DiagnosticQuery['type']
): typeof LOADING | DiagnosticInfo[] | ErrorLike => {
    const [diagnosticsOrError, setDiagnosticsOrError] = useState<typeof LOADING | DiagnosticInfo[] | ErrorLike>(LOADING)
    useEffect(() => {
        const subscriptions = new Subscription()
        subscriptions.add(
            getDiagnosticInfos(extensionsController, type)
                .pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
                .subscribe(setDiagnosticsOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [extensionsController, type])
    return diagnosticsOrError
}
