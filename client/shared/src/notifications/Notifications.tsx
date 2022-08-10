import * as React from 'react'

import { uniqueId } from 'lodash'
import { from, merge, Subscription } from 'rxjs'
import { delay, map, mergeMap, switchMap, takeWhile } from 'rxjs/operators'
import { tabbable } from 'tabbable'

import { asError } from '@sourcegraph/common'

import { wrapRemoteObservable } from '../api/client/api/common'
import { NotificationType } from '../api/extension/extensionHostApi'
import { syncRemoteSubscription } from '../api/util'
import { ExtensionsControllerProps } from '../extensions/controller'

import { Notification, NotificationWithId, WebAppNotification } from './notification'
import { NotificationItem, NotificationItemProps } from './NotificationItem'

import styles from './Notifications.module.scss'

export interface NotificationsProps
    extends ExtensionsControllerProps,
        Pick<NotificationItemProps, 'notificationItemStyleProps'> {}

interface NotificationsState {
    // TODO(tj): use remote progress observable type
    notifications: NotificationWithId[]
}

const HAS_NOTIFICATIONS_CONTEXT_KEY = 'hasNotifications'

// Context to manage notifications contributed by Sourcegraph web app parts.
export const NotificationContext = React.createContext<{
    notifications: WebAppNotification[]
    addNotification: (notification: Omit<WebAppNotification, 'id'>) => void
    removeNotification: (notification: WebAppNotification) => void
}>({
    notifications: [],
    addNotification: () => {},
    removeNotification: () => {},
})

export const NotificationContextProvider: React.FC<React.PropsWithChildren<{}>> = ({ children }) => {
    const [notifications, setNotifications] = React.useState<(Notification & { id: string })[]>([])
    const addNotification = React.useCallback(
        (notification: WebAppNotification) =>
            setNotifications(current => [...current, { ...notification, id: uniqueId('n') }]),
        [setNotifications]
    )
    const removeNotification = React.useCallback(
        (notification: WebAppNotification) =>
            setNotifications(current => current.filter(item => item.id !== notification.id)),
        [setNotifications]
    )

    return (
        <NotificationContext.Provider value={{ notifications, addNotification, removeNotification }}>
            {children}
        </NotificationContext.Provider>
    )
}

/**
 * A notifications center that displays global, non-modal messages from Sourcegraph web app and extensions.
 */
export class Notifications extends React.PureComponent<NotificationsProps, NotificationsState> {
    public static contextType = NotificationContext
    public context!: React.ContextType<typeof NotificationContext> // web app notifications

    /**
     * The maximum number of notifications at a time. Older notifications are truncated when the length exceeds
     * this number.
     */
    private static MAX_RETAIN = 7

    public state: NotificationsState = {
        notifications: [], // extensions notifications
    }

    private subscriptions = new Subscription()

    private notificationsReference = React.createRef<HTMLDivElement>()

    public componentDidMount(): void {
        // Subscribe to plain notifications
        this.subscriptions.add(
            from(this.props.extensionsController.extHostAPI)
                .pipe(
                    switchMap(extensionHostAPI =>
                        merge(
                            wrapRemoteObservable(extensionHostAPI.getPlainNotifications()),
                            // Subscribe to command error notifications (also plain)
                            this.props.extensionsController.commandErrors
                        )
                    ),
                    map(notification => ({ ...notification, id: uniqueId('n') }))
                )
                .subscribe(notification => {
                    this.setState(previousState => ({
                        notifications: [...previousState.notifications.slice(-Notifications.MAX_RETAIN), notification],
                    }))
                })
        )

        // Subscribe to progress notifications. This is more complex to handle than
        // plain notifications because the emissions of the progress notification observable
        // have to be proxied as well
        this.subscriptions.add(
            from(this.props.extensionsController.extHostAPI)
                .pipe(mergeMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getProgressNotifications())))
                .subscribe(progressNotification => {
                    // Progress notifications are remote, so property access is asynchronous
                    progressNotification.baseNotification
                        .then(baseNotification => {
                            // Turn ExtensionNotification type into client Notification type
                            // for NotificationItem to render (and subscribe to progressObservable)
                            const progressObservable = wrapRemoteObservable(progressNotification.progress)
                            const notification: Notification & { id: string } = {
                                ...baseNotification,
                                progress: progressObservable,
                                id: uniqueId('n'),
                            }

                            this.setState(previousState => ({
                                notifications: [
                                    ...previousState.notifications.slice(-Notifications.MAX_RETAIN),
                                    notification,
                                ],
                            }))

                            // Remove this notification once progress is finished
                            this.subscriptions.add(
                                progressObservable
                                    .pipe(
                                        takeWhile(({ percentage }) => !percentage || percentage < 100),
                                        delay(1000)
                                    )
                                    // eslint-disable-next-line rxjs/no-nested-subscribe
                                    .subscribe({
                                        error: error => {
                                            const erroredNotification = notification
                                            this.setState(({ notifications }) => ({
                                                notifications: notifications.map(notification =>
                                                    notification === erroredNotification
                                                        ? {
                                                              ...notification,
                                                              type: NotificationType.Error,
                                                              message: asError(error).message,
                                                          }
                                                        : notification
                                                ),
                                            }))
                                        },
                                        complete: () => {
                                            const completedNotification = notification
                                            this.setState(previousState => ({
                                                notifications: previousState.notifications.filter(
                                                    notification => notification !== completedNotification
                                                ),
                                            }))
                                        },
                                    })
                            )
                        })
                        .catch(() => {
                            // noop. there's no meaningful information to log if accessing
                            // baseNotification somehow failed
                        })
                })
        )

        // Register command to focus notifications.
        this.subscriptions.add(
            this.props.extensionsController.registerCommand({
                command: 'focusNotifications',
                run: () => {
                    const notificationsElement = this.notificationsReference.current
                    if (notificationsElement) {
                        tabbable(notificationsElement)[0]?.focus()
                    }
                    return Promise.resolve()
                },
            })
        )
        this.subscriptions.add(
            syncRemoteSubscription(
                this.props.extensionsController.extHostAPI.then(extensionHostAPI =>
                    extensionHostAPI.registerContributions({
                        menus: {
                            commandPalette: [
                                {
                                    action: 'focusNotifications',
                                    when: HAS_NOTIFICATIONS_CONTEXT_KEY,
                                },
                            ],
                        },
                        actions: [
                            {
                                id: 'focusNotifications',
                                title: 'Focus notifications',
                                command: 'focusNotifications',
                            },
                        ],
                    })
                )
            )
        )
    }

    public componentDidUpdate(): void {
        // Update context to show/hide "Focus notifications" command.
        this.props.extensionsController.extHostAPI
            .then(extensionHostAPI =>
                extensionHostAPI.updateContext({
                    [HAS_NOTIFICATIONS_CONTEXT_KEY]: this.state.notifications.length > 0,
                })
            )
            .catch(error => console.error('Error updating context for notifications', error))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const notifications = [
            ...this.context.notifications, // from Sourcegraph web app
            ...this.state.notifications, // from extension host state
        ]

        return (
            <div className={styles.sourcegraphNotifications} ref={this.notificationsReference}>
                {notifications.slice(0, Notifications.MAX_RETAIN).map(notification => (
                    <NotificationItem
                        key={notification.id}
                        notification={notification}
                        onDismiss={this.onDismiss}
                        className="sourcegraph-notifications__notification m-2"
                        notificationItemStyleProps={this.props.notificationItemStyleProps}
                    />
                ))}
            </div>
        )
    }

    private isWebAppNotification = (
        notification: NotificationWithId | WebAppNotification
    ): notification is WebAppNotification => this.context.notifications.some(item => item.id === notification.id)

    private onDismiss = (dismissedNotification: NotificationWithId | WebAppNotification): void => {
        if (this.isWebAppNotification(dismissedNotification)) {
            this.context.removeNotification(dismissedNotification)
            dismissedNotification.onDismiss?.()
        } else {
            this.setState(previousState => ({
                notifications: previousState.notifications.filter(
                    notification => notification !== dismissedNotification
                ),
            }))
        }
    }
}
