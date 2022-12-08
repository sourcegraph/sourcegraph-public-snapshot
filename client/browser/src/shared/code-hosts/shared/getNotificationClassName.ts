import classNames from 'classnames'
import { NotificationType } from 'sourcegraph'

export const createNotificationClassNameGetter =
    (notificationClassNames: Record<NotificationType, string>, extraClassName?: string) =>
    (notificationType: NotificationType): string =>
        classNames(notificationClassNames[notificationType], extraClassName)
