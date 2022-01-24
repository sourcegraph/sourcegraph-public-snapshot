import classNames from 'classnames'
import LockIcon from 'mdi-react/LockIcon'
import * as React from 'react'

import styles from './SiteAdminAlert.module.scss'

/**
 * An alert message with a site admin lock icon.
 */
export const SiteAdminAlert: React.FunctionComponent<{ children: React.ReactFragment; className?: string }> = ({
    children,
    className = '',
}) => (
    <div className={classNames('alert alert-warning', styles.siteAdminAlert, className)}>
        <h5>
            <LockIcon className="icon-inline" /> Site admin
        </h5>
        <div>{children}</div>
    </div>
)
