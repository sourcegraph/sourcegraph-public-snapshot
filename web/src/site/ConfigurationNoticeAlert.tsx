import WarningIcon from 'mdi-react/WarningIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'

/**
 * A global alert telling the site admin that the site config has problems.
 */
export const ConfigurationNoticeAlert: React.SFC<{
    className?: string
}> = ({ className = '' }) => (
    <div className={`alert alert-warning ${className}`}>
        <WarningIcon className="icon-inline mr-2" />
        <Link to="/site-admin/configuration">
            <strong>Update site configuration</strong>
        </Link>{' '}
        to resolve problems.
    </div>
)
