import { ProxyInput, ProxyResult, proxyValue } from '@sourcegraph/comlink'
import * as sourcegraph from 'sourcegraph'
import { ClientChecksAPI } from '../../client/api/checks'
import { syncSubscription } from '../../util'
import { toProxyableSubscribable } from './common'

export function createExtChecks(
    proxy: ProxyResult<ClientChecksAPI>
): Pick<typeof sourcegraph.checks, 'registerCheckProvider'> {
    return {
        registerCheckProvider: (type, provider) => {
            const providerFunction: ProxyInput<Parameters<ClientChecksAPI['$registerCheckProvider']>[1]> = proxyValue(
                async (...args: Parameters<sourcegraph.CheckProvider['provideCheck']>) =>
                    toProxyableSubscribable(provider.provideCheck(...args), item =>
                        item ? toTransferableCheck(item) : null
                    )
            )
            return syncSubscription(proxy.$registerCheckProvider(type, proxyValue(providerFunction)))
        },
    }
}
