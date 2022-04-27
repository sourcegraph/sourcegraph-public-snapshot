import * as React from 'react'

import classNames from 'classnames'
import LockIcon from 'mdi-react/LockIcon'

import { Alert, AlertProps, Icon, H5, H2 } from '@sourcegraph/wildcard'

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
        <H5 as={H2}>
            <Icon as={LockIcon} /> Site admin
        </H5>
        <div>{children}</div>
    </Alert>
)
