import { ProxyInput, ProxyResult, proxyValue } from '@sourcegraph/comlink'
import * as sourcegraph from 'sourcegraph'
import { syncSubscription } from '../../util'
import { toProxyableSubscribable } from './common'
import { ClientChecklistAPI } from '../../client/api/checklist'

export function createExtChecklist(
    proxy: ProxyResult<ClientChecklistAPI>
): Pick<typeof sourcegraph.checklist, 'registerChecklistProvider'> {
    return {
        registerChecklistProvider: (type, provider) => {
            const providerFunction: ProxyInput<
                Parameters<ClientChecklistAPI['$registerChecklistProvider']>[1]
            > = proxyValue(async (...args: Parameters<sourcegraph.ChecklistProvider['provideChecklistItems']>) =>
                toProxyableSubscribable(provider.provideChecklistItems(...args), items => items || [])
            )
            return syncSubscription(proxy.$registerChecklistProvider(type, proxyValue(providerFunction)))
        },
    }
}
