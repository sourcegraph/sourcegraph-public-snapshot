import LockIcon from 'mdi-react/LockIcon'
import * as React from 'react'

/**
 * An alert message with a site admin lock icon.
 */
export const SiteAdminAlert: React.SFC<{ children: React.ReactFragment; className?: string }> = ({
    children,
    className = '',
}) => (
    <div className={`alert alert-warning site-admin-alert ${className}`}>
        <h5>
            <LockIcon className="icon-inline" /> Site admin
        </h5>
        <div>{children}</div>
    </div>
)
