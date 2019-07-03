import { ProxyResult, ProxyValue, proxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import * as sourcegraph from 'sourcegraph'
import { ProxySubscribable } from '../../extension/api/common'
import { wrapRemoteObservable } from './common'
import { Unsubscribable } from 'rxjs'
import { NotificationService } from '../services/notificationService'
import { Notification } from '../../../notifications/notification'
import { map } from 'rxjs/operators'
import { toNotification } from '../../types/notification'

export interface ClientNotificationsAPI extends ProxyValue {
    $registerNotificationProvider(
        type: Parameters<typeof sourcegraph.notifications.registerNotificationProvider>[0],
        providerFunction: ProxyResult<
            ((
                ...args: Parameters<sourcegraph.NotificationProvider['provideNotifications']>
            ) => ProxySubscribable<Notification[] | null | undefined>) &
                ProxyValue
        >
    ): Unsubscribable & ProxyValue
}

export function createClientNotifications(statusService: NotificationService): ClientNotificationsAPI {
    return {
        $registerNotificationProvider: (type, providerFunction) => {
            return proxyValue(
                statusService.registerNotificationProvider(type, {
                    provideNotifications: (...args) =>
                        wrapRemoteObservable(providerFunction(...args)).pipe(
                            map(notifications => (notifications ? notifications.map(toNotification) : notifications))
                        ),
                })
            )
        },
        [proxyValueSymbol]: true,
    }
}
