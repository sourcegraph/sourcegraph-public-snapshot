import * as React from 'react'

import { uniqueId } from 'lodash'
import { from, merge, Subscription } from 'rxjs'
import { delay, map, mergeMap, switchMap, takeWhile } from 'rxjs/operators'
import { tabbable } from 'tabbable'

import { asError, logger } from '@sourcegraph/common'

import { wrapRemoteObservable } from '../api/client/api/common'
import { NotificationType } from '../api/extension/extensionHostApi'
import { syncRemoteSubscription } from '../api/util'
import { RequiredExtensionsControllerProps } from '../extensions/controller'

import { Notification } from './notification'
import { NotificationItem, NotificationItemProps } from './NotificationItem'

import styles from './Notifications.module.scss'

export interface NotificationsProps
    extends RequiredExtensionsControllerProps,
        Pick<NotificationItemProps, 'notificationItemStyleProps'> {}

interface NotificationsState {
    // TODO(tj): use remote progress observable type
    notifications: (Notification & { id: string })[]
}

const HAS_NOTIFICATIONS_CONTEXT_KEY = 'hasNotifications'

/**
 * A notifications center that displays global, non-modal messages.
 */
export class Notifications extends React.PureComponent<NotificationsProps, NotificationsState> {
    /**
     * The maximum number of notifications at a time. Older notifications are truncated when the length exceeds
     * this number.
     */
    private static MAX_RETAIN = 7

    public state: NotificationsState = {
        notifications: [],
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
            .catch(error => logger.error('Error updating context for notifications', error))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className={styles.sourcegraphNotifications} ref={this.notificationsReference}>
                {this.state.notifications.slice(0, Notifications.MAX_RETAIN).map(notification => (
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

    private onDismiss = (dismissedNotification: Notification): void => {
        this.setState(previousState => ({
            notifications: previousState.notifications.filter(notification => notification !== dismissedNotification),
        }))
    }
}
