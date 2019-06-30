import React from 'react'
import { Link } from 'react-router-dom'
import { StatusAreaContext } from '../StatusArea'
import { StatusStateBar } from './stateBar/StatusStateBar'
import { StatusBreadcrumbs } from './StatusBreadcrumbs'

interface Props extends Pick<StatusAreaContext, 'status' | 'statusURL' | 'statusesURL'> {
    className?: string
}

/**
 * An overview of a status.
 */
export const StatusOverview: React.FunctionComponent<Props> = ({ status, statusURL, statusesURL, className = '' }) => (
    <div className={`status-overview ${className || ''}`}>
        <StatusBreadcrumbs status={status} statusURL={statusURL} statusesURL={statusesURL} className="py-3" />
        <hr className="my-0" />
        <h2 className="my-3 font-weight-normal">{status.status.title}</h2>
        <StatusStateBar status={status} className="mb-3 p-3" />
    </div>
)
