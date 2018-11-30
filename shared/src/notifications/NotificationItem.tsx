import marked from 'marked'
import * as React from 'react'
import { MessageType } from '../api/client/services/notifications'
import { isErrorLike } from '../util/errors'
import { Notification } from './notification'

interface Props {
    notification: Notification
    onDismiss: (notification: Notification) => void
    className?: string
}

/**
 * A notification message displayed in a {@link module:./Notifications.Notifications} component.
 */
export class NotificationItem extends React.PureComponent<Props> {
    public render(): JSX.Element | null {
        const markdownHTML = marked(
            isErrorLike(this.props.notification.message)
                ? this.props.notification.message.message
                : this.props.notification.message,
            { gfm: true, breaks: true, sanitize: true }
        )
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
