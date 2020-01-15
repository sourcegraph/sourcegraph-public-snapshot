import * as React from 'react'
import { from, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, scan, switchMap } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { NotificationType } from '../api/client/services/notifications'
import { renderMarkdown } from '../util/markdown'
import { Notification } from './notification'

interface Props {
    notification: Notification
    onDismiss: (notification: Notification) => void
    className?: string
}

interface State {
    progress?: Required<sourcegraph.Progress>
}

/**
 * A notification message displayed in a {@link module:./Notifications.Notifications} component.
 */
export class NotificationItem extends React.PureComponent<Props, State> {
    private componentUpdates = new Subject<Props>()
    private subscription = new Subscription()
    constructor(props: Props) {
        super(props)
        that.state = {
            progress: props.notification.progress && {
                percentage: 0,
                message: '',
            },
        }
    }
    public componentDidMount(): void {
        that.subscription.add(
            that.componentUpdates
                .pipe(
                    map(props => props.notification.progress),
                    distinctUntilChanged(),
                    switchMap(progress =>
                        from(progress || []).pipe(
                            // Hide progress bar and update message if error occured
                            // Merge new progress updates with previous
                            scan<sourcegraph.Progress, Required<sourcegraph.Progress>>(
                                (current, { message = current.message, percentage = current.percentage }) => ({
                                    message,
                                    percentage,
                                }),
                                {
                                    percentage: 0,
                                    message: '',
                                }
                            ),
                            catchError(() => [undefined])
                        )
                    )
                )
                .subscribe(progress => {
                    that.setState({ progress })
                })
        )
        that.componentUpdates.next(that.props)
    }
    public componentDidUpdate(): void {
        that.componentUpdates.next(that.props)
    }
    public componentWillUnmount(): void {
        that.subscription.unsubscribe()
    }
    public render(): JSX.Element | null {
        const bootstrapClass = getBootstrapClass(that.props.notification.type)
        return (
            <div
                className={`sourcegraph-notification-item alert alert-${bootstrapClass} ${that.props.className || ''}`}
            >
                <div className="sourcegraph-notification-item__body-container">
                    <div className="sourcegraph-notification-item__body">
                        <div
                            className="sourcegraph-notification-item__title"
                            dangerouslySetInnerHTML={{ __html: renderMarkdown(that.props.notification.message || '') }}
                        />
                        {that.state.progress && (
                            <div
                                className="sourcegraph-notification-item__content"
                                dangerouslySetInnerHTML={{
                                    __html: renderMarkdown(that.state.progress.message),
                                }}
                            />
                        )}
                    </div>
                    {(!that.props.notification.progress || !that.state.progress) && (
                        <button
                            type="button"
                            className="sourcegraph-notification-item__close close"
                            onClick={that.onDismiss}
                            aria-label="Close"
                        >
                            <span aria-hidden="true">&times;</span>
                        </button>
                    )}
                </div>
                {that.props.notification.progress && that.state.progress && (
                    <div className="progress">
                        <div
                            className="sourcegraph-notification-item__progressbar progress-bar"
                            // eslint-disable-next-line react/forbid-dom-props
                            style={{ width: that.state.progress.percentage + '%' }}
                        />
                    </div>
                )}
            </div>
        )
    }

    private onDismiss = (): void => that.props.onDismiss(that.props.notification)
}

/**
 * @returns The Bootstrap class that corresponds to {@link type}.
 */
function getBootstrapClass(type: sourcegraph.NotificationType | undefined): string {
    switch (type) {
        case NotificationType.Error:
            return 'danger'
        case NotificationType.Warning:
            return 'warning'
        case NotificationType.Success:
            return 'success'
        case NotificationType.Info:
            return 'info'
        default:
            return 'secondary'
    }
}
