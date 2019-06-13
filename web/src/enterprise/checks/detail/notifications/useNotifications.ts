import { useEffect, useState } from 'react'
import { Subscription } from 'rxjs'
import { catchError, startWith } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'

const LOADING: 'loading' = 'loading'

/**
 * A React hook that observes notifications.
 *
 * @param type Only observe notifications from the provider registered with this type. If undefined,
 * notifications from all providers are observed.
 */
export const useNotifications = (
    extensionsController: ExtensionsControllerProps['extensionsController'],
    scope: Parameters<sourcegraph.NotificationProvider['provideNotifications']>[0],
    type?: Parameters<typeof sourcegraph.notifications.registerNotificationProvider>[0]
): typeof LOADING | sourcegraph.Notification[] | ErrorLike => {
    const [notificationsOrError, setNotificationsOrError] = useState<
        typeof LOADING | sourcegraph.Notification[] | ErrorLike
    >(LOADING)
    useEffect(() => {
        const subscriptions = new Subscription()
        subscriptions.add(
            extensionsController.services.notifications2
                .observeNotifications(scope, type)
                .pipe(
                    catchError(err => [asError(err)]),
                    startWith(LOADING)
                )
                .subscribe(setNotificationsOrError)
        )
        return () => subscriptions.unsubscribe()
    }, [extensionsController, scope, type])
    return notificationsOrError
}
