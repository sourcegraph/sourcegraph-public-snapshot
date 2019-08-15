import { ProxyResult, proxyValue, ProxyInput } from '@sourcegraph/comlink'
import * as sourcegraph from 'sourcegraph'
import { ClientDiagnosticsAPI } from '../../client/api/diagnostics'
import { fromDiagnostic } from '../../types/diagnostic'
import { syncSubscription } from '../../util'
import { toProxyableSubscribable } from './common'

/** @internal */
export const createExtDiagnostics = (
    proxy: ProxyResult<ClientDiagnosticsAPI>
): Pick<typeof sourcegraph.workspace, 'registerDiagnosticProvider'> => {
    return {
        registerDiagnosticProvider: (
            type: string,
            provider: sourcegraph.DiagnosticProvider
        ): sourcegraph.Unsubscribable => {
            const providerFunction: ProxyInput<
                Parameters<ClientDiagnosticsAPI['$registerDiagnosticProvider']>[1]
            > = proxyValue(async (scope, context) =>
                toProxyableSubscribable(provider.provideDiagnostics(scope, context), diagnostics =>
                    diagnostics ? diagnostics.map(fromDiagnostic) : []
                )
            )
            return syncSubscription(proxy.$registerDiagnosticProvider(type, providerFunction))
        },
    }
}
