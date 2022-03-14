import classNames from 'classnames'
import LockIcon from 'mdi-react/LockIcon'
import * as React from 'react'

import { Alert, AlertProps, Icon } from '@sourcegraph/wildcard'

import styles from './SiteAdminAlert.module.scss'

interface SiteAdminAlertProps {
    className?: string
    variant?: AlertProps['variant']
}

/**
 * An alert message with a site admin lock icon.
 */
export const SiteAdminAlert: React.FunctionComponent<SiteAdminAlertProps> = ({
    children,
    className = '',
    variant = 'warning',
}) => (
    <Alert className={classNames(styles.siteAdminAlert, className)} variant={variant}>
        <h5>
            <Icon as={LockIcon} /> Site admin
        </h5>
        <div>{children}</div>
    </Alert>
)
