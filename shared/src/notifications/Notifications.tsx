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
     * this number.
     */
    private static MAX_RETAIN = 7

    public state: State = {
        notifications: [],
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.props.extensionsController.notifications
                .pipe(map(n => ({ ...n, id: uniqueId('n') })))
                .subscribe(notification => {
                    this.setState(prevState => ({
                        notifications: [...prevState.notifications.slice(-Notifications.MAX_RETAIN), notification],
                    }))
                    if (notification.progress) {
                        // Remove once progress is finished
                        this.subscriptions.add(
                            notification.progress
                                .pipe(
                                    takeWhile(({ percentage }) => !percentage || percentage < 100),
                                    delay(1000)
                                )
                                // tslint:disable-next-line: rxjs-no-nested-subscribe
                                .subscribe({
                                    error: err => {
                                        this.setState(({ notifications }) => ({
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
                                        this.setState(prevState => ({
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
                    />
                ))}
            </div>
        )
    }

    private onDismiss = (notification: Notification): void => {
        this.setState(prevState => ({ notifications: prevState.notifications.filter(n => n !== notification) }))
    }
}
