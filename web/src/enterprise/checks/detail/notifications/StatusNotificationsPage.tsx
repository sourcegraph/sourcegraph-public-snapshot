import { NotificationScope } from '@sourcegraph/extension-api-classes'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { CheckAreaContext } from '../CheckArea'
import { StatusNotification } from './StatusNotification'
import { useNotifications } from './useNotifications'

interface Props extends Pick<CheckAreaContext, 'name' | 'status'>, ExtensionsControllerProps {
    className?: string
    itemClassName?: string
    history: H.History
    location: H.Location
}

const LOADING: 'loading' = 'loading'

/**
 * The status notifications page.
 */
export const CheckNotificationsPage: React.FunctionComponent<Props> = ({
    checkID,
    check,
    className = '',
    itemClassName = '',
    ...props
}) => {
    const notificationsOrError = useNotifications(props.extensionsController, NotificationScope.Global, checkID)
    return (
        <div className={`status-notifications-page ${className}`}>
            {isErrorLike(notificationsOrError) ? (
                <div className={itemClassName}>
                    <div className="alert alert-danger mt-2">{notificationsOrError.message}</div>
                </div>
            ) : notificationsOrError === LOADING ? (
                <div className={itemClassName}>
                    <LoadingSpinner className="mt-3" />
                </div>
            ) : notificationsOrError.length === 0 ? (
                <div className={itemClassName}>
                    <p className="p-2 mb-0 text-muted">No notifications found.</p>
                </div>
            ) : (
                <ul className="list-unstyled mb-0">
                    {notificationsOrError.map((notification, i) => (
                        <li key={i}>
                            <StatusNotification
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
