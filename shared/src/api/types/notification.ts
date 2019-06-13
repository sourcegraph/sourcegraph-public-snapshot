import * as sourcegraph from 'sourcegraph'
import { Notification } from '../../notifications/notification'
import { fromAction, toAction } from './action'
import { toDiagnosticData, fromDiagnosticData } from './diagnostic'

export const fromNotification = (notification: sourcegraph.Notification): Notification => {
    return {
        ...notification,
        actions: notification.actions && notification.actions.map(fromAction),
    }
}

export const toNotification = (notification: Notification): sourcegraph.Notification => {
    return {
        ...notification,
        actions: notification.actions && notification.actions.map(toAction),
    }
}
