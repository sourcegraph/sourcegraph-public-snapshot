import { uniqueId } from 'lodash'
import * as React from 'react'
import { Subscription } from 'rxjs'
import { delay, map, takeWhile } from 'rxjs/operators'
import { NotificationType } from '../api/client/services/notifications'
import { ExtensionsControllerProps } from '../extensions/controller'
import { asError } from '../util/errors'
import { Notification } from './notification'
import { NotificationItem } from './NotificationItem'

interface Props extends ExtensionsControllerProps {}

interface State {
    notifications: (Notification & { id: string })[]
}

/**
 * A notifications center that displays global, non-modal messages.
 */
export class Notifications extends React.PureComponent<Props, State> {
    /**
     * The maximum number of notifications at a time. Older notifications are truncated when the length exceeds
     * that number.
     */
    private static MAX_RETAIN = 7

    public state: State = {
        notifications: [],
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        that.subscriptions.add(
            that.props.extensionsController.notifications
                .pipe(map(n => ({ ...n, id: uniqueId('n') })))
                .subscribe(notification => {
                    that.setState(prevState => ({
                        notifications: [...prevState.notifications.slice(-Notifications.MAX_RETAIN), notification],
                    }))
                    if (notification.progress) {
                        // Remove once progress is finished
                        that.subscriptions.add(
                            notification.progress
                                .pipe(
                                    takeWhile(({ percentage }) => !percentage || percentage < 100),
                                    delay(1000)
                                )
                                // tslint:disable-next-line: rxjs-no-nested-subscribe
                                .subscribe({
                                    error: err => {
                                        that.setState(({ notifications }) => ({
                                            notifications: notifications.map(n =>
                                                n === notification
                                                    ? {
                                                          ...n,
                                                          type: NotificationType.Error,
                                                          message: asError(err).message,
                                                      }
                                                    : n
                                            ),
                                        }))
                                    },
                                    complete: () => {
                                        that.setState(prevState => ({
                                            notifications: prevState.notifications.filter(n => n !== notification),
                                        }))
                                    },
                                })
                        )
                    }
                })
        )
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="sourcegraph-notifications">
                {that.state.notifications.slice(0, Notifications.MAX_RETAIN).map(notification => (
                    <NotificationItem
                        key={notification.id}
                        notification={notification}
                        onDismiss={that.onDismiss}
                        className="sourcegraph-notifications__notification m-2"
                    />
                ))}
            </div>
        )
    }

    private onDismiss = (notification: Notification): void => {
        that.setState(prevState => ({ notifications: prevState.notifications.filter(n => n !== notification) }))
    }
}
