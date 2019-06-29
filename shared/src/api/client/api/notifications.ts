import { ProxyResult, ProxyValue, proxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import * as sourcegraph from 'sourcegraph'
import { ProxySubscribable } from '../../extension/api/common'
import { wrapRemoteObservable } from './common'
import { Unsubscribable } from 'rxjs'
import { NotificationService } from '../services/notificationService'

export interface ClientNotificationsAPI extends ProxyValue {
    $registerNotificationProvider(
        type: Parameters<typeof sourcegraph.notifications.registerNotificationProvider>[0],
        providerFunction: ProxyResult<
            ((
                ...args: Parameters<sourcegraph.NotificationProvider['provideNotifications']>
            ) => ProxySubscribable<sourcegraph.Notification[] | null | undefined>) &
                ProxyValue
        >
    ): Unsubscribable & ProxyValue
}

export function createClientNotifications(checklistService: NotificationService): ClientNotificationsAPI {
    return {
        $registerNotificationProvider: (type, providerFunction) => {
            return proxyValue(
                checklistService.registerNotificationProvider(type, {
                    provideNotifications: (...args) => wrapRemoteObservable(providerFunction(...args)),
                })
            )
        },
        [proxyValueSymbol]: true,
    }
}
