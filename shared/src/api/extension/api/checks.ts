import { ProxyInput, ProxyResult, proxyValue, ProxyValue } from '@sourcegraph/comlink'
import * as sourcegraph from 'sourcegraph'
import { ClientChecksAPI, ProxiedCheckProvider } from '../../client/api/checks'
import { syncSubscription } from '../../util'
import { toProxyableSubscribable, proxySubscribable } from './common'

export function createExtChecks(
    proxy: ProxyResult<ClientChecksAPI>
): Pick<typeof sourcegraph.checks, 'registerCheckProvider'> {
    return {
        registerCheckProvider: (type, providerFactory) => {
            const proxiedProviderFactory = proxyValue(async (context: sourcegraph.CheckContext<any>) => {
                const provider = providerFactory(context)
                return { information: proxySubscribable(provider.information) }
            })
            return syncSubscription(proxy.$registerCheckProvider(type, proxiedProviderFactory))
        },
    }
}
