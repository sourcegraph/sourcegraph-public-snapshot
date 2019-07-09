import { ProxyResult, ProxyValue, proxyValueSymbol, proxyValue, ProxyInput } from '@sourcegraph/comlink'
import { from, Unsubscribable, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { ClientDiagnosticsAPI } from '../../client/api/diagnostics'
import { DiagnosticCollection } from '../../types/diagnosticCollection'
import { toDiagnosticData, fromDiagnosticData, DiagnosticData } from '../../types/diagnostic'
import { syncSubscription } from '../../util'
import { toProxyableSubscribable } from './common'

/** @internal */
export const createExtDiagnostics = (
    proxy: ProxyResult<ClientDiagnosticsAPI>
): Pick<typeof sourcegraph.workspace, 'registerDiagnosticProvider'> => {
    return {
        registerDiagnosticProvider: (
            name: string,
            provider: sourcegraph.DiagnosticProvider
        ): sourcegraph.Unsubscribable => {
            const providerFunction: ProxyInput<
                Parameters<ClientDiagnosticsAPI['$registerDiagnosticProvider']>[1]
            > = proxyValue(async scope => toProxyableSubscribable(provider.provideDiagnostics(), toLocations))
            return syncSubscription(proxy.$registerDiagnosticProvider(selector, providerFunction))
        },
    }
}
