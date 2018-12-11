import * as React from 'react'
import { Subscription } from 'rxjs'
import { ExtensionsControllerProps } from '../extensions/controller'
import { Notification } from './notification'
import { NotificationItem } from './NotificationItem'

interface Props extends ExtensionsControllerProps {}

interface State {
    notifications: Notification[]
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
            this.props.extensionsController.notifications.subscribe(notification => {
                this.setState(prevState => ({
                    notifications: [notification, ...prevState.notifications.slice(0, Notifications.MAX_RETAIN - 1)],
                }))
            })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="sourcegraph-notifications">
                {this.state.notifications
                    .slice(0, Notifications.MAX_RETAIN)
                    .map((notification, i) => (
                        <NotificationItem
                            key={i}
                            notification={notification}
                            onDismiss={this.onDismiss}
                            className="sourcegraph-notifications__notification rounded-0 m-2"
                        />
                    ))}
            </div>
        )
    }

    private onDismiss = (notification: Notification) => {
        this.setState(prevState => ({ notifications: prevState.notifications.filter(n => n !== notification) }))
    }
}
