import { ProxyInput, ProxyResult, proxyValue } from '@sourcegraph/comlink'
import * as sourcegraph from 'sourcegraph'
import { ClientChecksAPI, ProxiedCheckProvider } from '../../client/api/checks'
import { syncSubscription } from '../../util'
import { proxySubscribable, ProxySubscribable, toProxyableSubscribable } from './common'

export function createExtChecks(
    proxy: ProxyResult<ClientChecksAPI>
): Pick<typeof sourcegraph.checks, 'registerCheckProvider'> {
    return {
        registerCheckProvider: (type, providerFactory) => {
            const proxiedProviderFactory: ProxyInput<
                Parameters<ClientChecksAPI['$registerCheckProvider']>[1]
            > = proxyValue(async context => {
                const provider = providerFactory(context)
                // TODO!(sqs): fix type error, remove casts below
                return proxyValue({
                    information: (proxyValue(proxySubscribable(provider.information)) as any) as ProxyResult<
                        ProxySubscribable<sourcegraph.CheckInformation>
                    >,
                    provideDiagnosticGroups: proxyValue(async () =>
                        proxySubscribable(provider.provideDiagnosticGroups())
                    ),
                    provideDiagnosticBatchActions: proxyValue(async (query: sourcegraph.DiagnosticQuery) =>
                        proxySubscribable(provider.provideDiagnosticBatchActions(query))
                    ),
                })
            })
            return syncSubscription(proxy.$registerCheckProvider(type, proxiedProviderFactory))
        },
    }
}
