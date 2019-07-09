import { ProxyInput, ProxyResult, proxyValue } from '@sourcegraph/comlink'
import * as sourcegraph from 'sourcegraph'
import { ClientChecksAPI } from '../../client/api/checks'
import { syncSubscription } from '../../util'
import { proxySubscribable, ProxySubscribable } from './common'

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
                return {
                    information: (proxySubscribable(provider.information) as any) as ProxyResult<
                        ProxySubscribable<sourcegraph.CheckInformation>
                    >,
                }
            })
            return syncSubscription(proxy.$registerCheckProvider(type, proxiedProviderFactory))
        },
    }
}
