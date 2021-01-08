import { uniqueId } from 'lodash'
import * as React from 'react'
import { from, Subscription } from 'rxjs'
import { delay, map, mergeMap, takeWhile } from 'rxjs/operators'
import { wrapRemoteObservable } from '../api/client/api/common'
import { NotificationType } from '../api/contract'
import { ExtensionsControllerProps } from '../extensions/controller'
import { asError } from '../util/errors'
import { Notification } from './notification'
import { NotificationItem, NotificationClassNameProps } from './NotificationItem'

interface Props extends ExtensionsControllerProps, NotificationClassNameProps {}

interface State {
    // TODO(tj): use remote progress observable type
    notifications: (Notification & { id: string })[]
}

/**
 * A notifications center that displays global, non-modal messages.
 */
export class Notifications extends React.PureComponent<Props, State> {
    /**
     * The maximum number of notifications at a time. Older notifications are truncated when the length exceeds
     * this number.
     */
    private static MAX_RETAIN = 7

    public state: State = {
        notifications: [],
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Subscribe to plain notifications
        this.subscriptions.add(
            from(this.props.extensionsController.extHostAPI)
                .pipe(
                    mergeMap(extensionHostAPI => wrapRemoteObservable(extensionHostAPI.getPlainNotifications())),
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
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="sourcegraph-notifications">
                {this.state.notifications.slice(0, Notifications.MAX_RETAIN).map(notification => (
                    <NotificationItem
                        key={notification.id}
                        notification={notification}
                        onDismiss={this.onDismiss}
                        className="sourcegraph-notifications__notification m-2"
                        notificationClassNames={this.props.notificationClassNames}
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
