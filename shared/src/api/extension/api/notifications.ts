import { ProxyInput, ProxyResult, proxyValue } from '@sourcegraph/comlink'
import * as sourcegraph from 'sourcegraph'
import { syncSubscription } from '../../util'
import { toProxyableSubscribable } from './common'
import { ClientNotificationsAPI } from '../../client/api/notifications'

export function createExtNotifications(
    proxy: ProxyResult<ClientNotificationsAPI>
): Pick<typeof sourcegraph.notifications, 'registerNotificationProvider'> {
    return {
        registerNotificationProvider: (type, provider) => {
            const providerFunction: ProxyInput<
                Parameters<ClientNotificationsAPI['$registerNotificationProvider']>[1]
            > = proxyValue(async (...args: Parameters<sourcegraph.NotificationProvider['provideNotifications']>) =>
                toProxyableSubscribable(provider.provideNotifications(...args), items => items || [])
            )
            return syncSubscription(proxy.$registerNotificationProvider(type, proxyValue(providerFunction)))
        },
    }
}
