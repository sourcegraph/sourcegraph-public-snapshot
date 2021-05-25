import classNames from 'classnames'

import { NotificationType } from '@sourcegraph/shared/src/api/extension/extensionHostApi'

type NotificationClassNames = Record<number, string>

type NotificationKind = 'error' | 'warning' | 'info'

export const createNotificationClassNameGetter = (
    notificationClassNames: NotificationClassNames,
    extraClassName?: string
) => (notificationKind: NotificationKind): string => {
    switch (notificationKind) {
        case 'error':
            return classNames(notificationClassNames[NotificationType.Error], extraClassName)
        case 'warning':
            return classNames(notificationClassNames[NotificationType.Warning], extraClassName)
        default:
            return classNames(notificationClassNames[NotificationType.Info], extraClassName)
    }
}
