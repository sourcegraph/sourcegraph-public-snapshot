import { NotificationScope } from '@sourcegraph/extension-api-classes'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { StatusAreaContext } from '../StatusArea'
import { useNotifications } from './useNotifications'

interface Props extends Pick<StatusAreaContext, 'name' | 'status'>, ExtensionsControllerProps {
    className?: string
    itemClassName?: string
    history: H.History
    location: H.Location
}

const LOADING: 'loading' = 'loading'

/**
 * The status notifications page.
 */
export const StatusNotificationsPage: React.FunctionComponent<Props> = ({
    name,
    status,
    className = '',
    itemClassName = '',
    ...props
}) => {
    const notificationsOrError = useNotifications(props.extensionsController, NotificationScope.Global, name)
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
                    {notificationsOrError.map((n, i) => (
                        <li key={i}>
                            <div key={i} className="card mb-5">
                                <section className="card-body">
                                    <h4 className="card-title mb-0 font-weight-normal d-flex align-items-center">
                                        {n.title}
                                    </h4>
                                    <div className="d-flex align-items-center flex-wrap">
                                        <button className="btn btn-sm btn-success mt-3 mr-3">Preview fix</button>
                                        <button className="btn btn-sm btn-secondary mt-3 mr-3">Show details</button>
                                        <button className="btn btn-sm btn-secondary mt-3 mr-3">
                                            Ignore this problem
                                        </button>
                                    </div>
                                </section>
                            </div>
                        </li>
                    ))}
                </ul>
            )}
        </div>
    )
}
