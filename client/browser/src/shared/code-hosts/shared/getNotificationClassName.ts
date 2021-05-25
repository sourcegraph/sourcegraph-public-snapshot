import classNames from 'classnames'

import { NotificationType } from '@sourcegraph/shared/src/api/extension/extensionHostApi'

type NotificationClassNames = Record<number, string>

type NotificationKind = 'error' | 'warning' | 'info'

export const createNotificationClassNameGetter = (
    notificationClassNames: NotificationClassNames,
    extraClassName?: string
) => (notificationKind: NotificationKind): string => {
    function getNotificationClassName(): string {
        switch (notificationKind) {
            case 'error':
                return notificationClassNames[NotificationType.Error]
            case 'warning':
                return notificationClassNames[NotificationType.Warning]
            default:
                return notificationClassNames[NotificationType.Info]
        }
    }

    return classNames(getNotificationClassName(), extraClassName)
}
