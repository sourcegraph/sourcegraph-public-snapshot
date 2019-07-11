import { NotificationScope } from '@sourcegraph/extension-api-classes'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { CheckAreaContext } from '../CheckArea'
import { CheckNotification } from './CheckNotification'
import { useNotifications } from './useNotifications'

interface Props extends Pick<CheckAreaContext, 'checkID' | 'checkInfo'>, ExtensionsControllerProps {
    className?: string
    itemClassName?: string
    history: H.History
    location: H.Location
}

const LOADING: 'loading' = 'loading'

/**
 * The check notifications page.
 */
export const CheckNotificationsPage: React.FunctionComponent<Props> = ({
    checkID,
    checkInfo,
    className = '',
    ...props
}) => {
    const notificationsOrError = useNotifications(
        props.extensionsController,
        NotificationScope.Global,
        checkID.type /* TODO!(sqs) assumes that notif type == check provider type as convention */
    )
    return (
        <div className={`check-notifications-page ${className}`}>
            {isErrorLike(notificationsOrError) ? (
                <div className="alert alert-danger mt-2">{notificationsOrError.message}</div>
            ) : notificationsOrError === LOADING ? (
                <LoadingSpinner className="mt-3" />
            ) : notificationsOrError.length === 0 ? (
                <p className="p-2 mb-0 text-muted">No notifications found.</p>
            ) : (
                <ul className="list-unstyled mb-0">
                    {notificationsOrError.map((notification, i) => (
                        <li key={i}>
                            <CheckNotification
                                {...props}
                                notification={notification}
                                className="card my-5"
                                contentClassName="card-body"
                            />
                        </li>
                    ))}
                </ul>
            )}
        </div>
    )
}
