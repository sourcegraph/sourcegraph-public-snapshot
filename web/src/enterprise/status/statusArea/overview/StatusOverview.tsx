import React from 'react'
import { StatusAreaContext } from '../StatusArea'
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
    </div>
)
