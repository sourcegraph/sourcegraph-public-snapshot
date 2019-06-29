import { ProxyInput, ProxyResult, proxyValue } from '@sourcegraph/comlink'
import * as sourcegraph from 'sourcegraph'
import { syncSubscription } from '../../util'
import { toProxyableSubscribable } from './common'
import { ClientStatusAPI } from '../../client/api/status'

export function createExtStatus(
    proxy: ProxyResult<ClientStatusAPI>
): Pick<typeof sourcegraph.status, 'registerStatusProvider'> {
    return {
        registerStatusProvider: (type, provider) => {
            const providerFunction: ProxyInput<Parameters<ClientStatusAPI['$registerStatusProvider']>[1]> = proxyValue(
                async (...args: Parameters<sourcegraph.StatusProvider['provideStatus']>) =>
                    toProxyableSubscribable(provider.provideStatus(...args), item => item || null)
            )
            return syncSubscription(proxy.$registerStatusProvider(type, proxyValue(providerFunction)))
        },
    }
}
