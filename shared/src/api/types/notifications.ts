import * as sourcegraph from 'sourcegraph'
import { Notification } from '../../notifications/notification'
import { fromCodeAction, toCodeAction } from '../extension/api/types'

export const fromNotification = (notification: sourcegraph.Notification): Notification => {
    return { ...notification, actions: notification.actions && notification.actions.map(fromCodeAction) }
}

export const toNotification = (notification: Notification): sourcegraph.Notification => {
    return { ...notification, actions: notification.actions && notification.actions.map(toCodeAction) }
}
