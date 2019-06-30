import React from 'react'
import { Link } from 'react-router-dom'
import { StatusStateIcon } from '../../components/StatusStateIcon'
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
                <Link to={statusURL} className="d-inline-flex align-items-center">
                    {status.name} <StatusStateIcon status={status.status} className="icon-inline ml-2" />
                </Link>
            </li>
        </ol>
    </nav>
)
