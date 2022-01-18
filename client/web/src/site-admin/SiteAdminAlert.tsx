import classNames from 'classnames'
import LockIcon from 'mdi-react/LockIcon'
import * as React from 'react'

import { Alert } from '@sourcegraph/wildcard'

import styles from './SiteAdminAlert.module.scss'

/**
 * An alert message with a site admin lock icon.
 */
export const SiteAdminAlert: React.FunctionComponent<{ children: React.ReactFragment; className?: string }> = ({
    children,
    className = '',
}) => (
    <Alert className={classNames(styles.siteAdminAlert, className)} variant="warning">
        <h5>
            <LockIcon className="icon-inline" /> Site admin
        </h5>
        <div>{children}</div>
    </Alert>
)
