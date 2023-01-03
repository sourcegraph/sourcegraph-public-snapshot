import classNames from 'classnames'

import type { NotificationType } from '@sourcegraph/shared/src/codeintel/legacy-extensions/api'

export const createNotificationClassNameGetter =
    (notificationClassNames: Record<NotificationType, string>, extraClassName?: string) =>
    (notificationType: NotificationType): string =>
        classNames(notificationClassNames[notificationType], extraClassName)
