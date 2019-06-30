import React from 'react'
import { Link } from 'react-router-dom'
import { StatusAreaContext } from '../StatusArea'

interface Props extends Pick<StatusAreaContext, 'status' | 'statusURL' | 'statusesURL'> {
    className?: string
}

/**
 * The breadcrumbs for a status.
 */
export const StatusBreadcrumbs: React.FunctionComponent<Props> = ({
    status,
    statusURL,
    statusesURL,
    className = '',
}) => (
    <nav className={`d-flex align-items-center ${className}`} aria-label="breadcrumb">
        <ol className="breadcrumb">
            <li className="breadcrumb-item">
                <Link to={statusesURL}>Status</Link>
            </li>
            <li className="breadcrumb-item active font-weight-bold">
                <Link to={statusURL}>{status.name}</Link>
            </li>
        </ol>
    </nav>
)
