import marked from 'marked'
import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { Progress } from 'sourcegraph'
import { MessageType } from '../api/client/services/notifications'
import { isErrorLike } from '../util/errors'
import { Notification } from './notification'

interface Props {
    notification: Notification
    onDismiss: (notification: Notification) => void
    className?: string
}

interface State {
    progress?: Progress
}

/**
 * A notification message displayed in a {@link module:./Notifications.Notifications} component.
 */
export class NotificationItem extends React.PureComponent<Props, State> {
    public state: State = {}
    private componentUpdates = new Subject<Props>()
    private subscription = new Subscription()
    public componentDidMount(): void {
        this.subscription.add(
            this.componentUpdates
                .pipe(switchMap(props => props.notification.progress || []))
                .subscribe(progress => this.setState({ progress }))
        )
    }
    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }
    public componentWillUnmount(): void {
        this.subscription.unsubscribe()
    }
    public render(): JSX.Element | null {
        let message = isErrorLike(this.props.notification.message)
            ? this.props.notification.message.message
            : this.props.notification.message
        if (this.state.progress) {
            message += '  \n' + this.state.progress.message
        }
        const markdownHTML = marked(message, { gfm: true, breaks: true, sanitize: true })
        return (
            <div
                className={`sourcegraph-notification-item alert alert-${alertClass(
                    this.props.notification.type
                )} p-0 ${this.props.className || ''}`}
            >
                <div
                    className="sourcegraph-notification-item__content py-2 pl-2 pr-0"
                    dangerouslySetInnerHTML={{ __html: markdownHTML }}
                />
                <button
                    type="button"
                    className="sourcegraph-notification-item__close p-2"
                    onClick={this.onDismiss}
                    aria-label="Close"
                >
                    <span aria-hidden="true">&times;</span>
                </button>
                {this.state.progress && (
                    <div className="w-100">
                        <div
                            className={`p-1 bg-${this.props.notification.type}`}
                            // tslint:disable-next-line:jsx-ban-props
                            style={{ width: this.state.progress.percentage + '%' }}
                        />
                    </div>
                )}
            </div>
        )
    }

    private onDismiss = () => this.props.onDismiss(this.props.notification)
}

/**
 * @return The Bootstrap class that corresponds to {@link type}.
 */
function alertClass(type: MessageType | undefined): string {
    switch (type) {
        case MessageType.Error:
            return 'danger'
        case MessageType.Warning:
            return 'warning'
        default:
            return 'info'
    }
}
