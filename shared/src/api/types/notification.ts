import * as sourcegraph from 'sourcegraph'
import { Notification } from '../../notifications/notification'
import { fromCodeAction, toCodeAction } from './action'
import { toDiagnosticData, fromDiagnosticData } from './diagnostic'

export const fromNotification = (notification: sourcegraph.Notification): Notification => {
    return {
        ...notification,
        diagnostics: notification.diagnostics && toDiagnosticData(notification.diagnostics),
        actions: notification.actions && notification.actions.map(fromCodeAction),
    }
}

export const toNotification = (notification: Notification): sourcegraph.Notification => {
    return {
        ...notification,
        diagnostics: notification.diagnostics && fromDiagnosticData(notification.diagnostics),
        actions: notification.actions && notification.actions.map(toCodeAction),
    }
}
