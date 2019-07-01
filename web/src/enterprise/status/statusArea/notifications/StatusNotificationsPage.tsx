import H from 'history'
import React from 'react'
import { StatusAreaContext } from '../StatusArea'

interface Props extends Pick<StatusAreaContext, 'status'> {
    className?: string
    history: H.History
    location: H.Location
}

/**
 * The status notifications page.
 */
export const StatusNotificationsPage: React.FunctionComponent<Props> = ({ status, className = '', ...props }) => (
    <div className={`status-notifications-page ${className}`}>
        {status.status.notifications &&
            status.status.notifications.map((n, i) => (
                <div key={i} className="card mb-5">
                    <section className="card-body">
                        <h4 className="card-title mb-0 font-weight-normal d-flex align-items-center">{n.title}</h4>
                        <div className="d-flex align-items-center flex-wrap">
                            <button className="btn btn-sm btn-success mt-3 mr-3">Preview fix</button>
                            <button className="btn btn-sm btn-secondary mt-3 mr-3">Show details</button>
                            <button className="btn btn-sm btn-secondary mt-3 mr-3">Ignore this problem</button>
                        </div>
                    </section>
                </div>
            ))}
    </div>
)
