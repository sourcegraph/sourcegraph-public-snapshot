import * as React from 'react'

import classNames from 'classnames'
import LockIcon from 'mdi-react/LockIcon'

import { Alert, AlertProps, Icon, Typography } from '@sourcegraph/wildcard'

import styles from './SiteAdminAlert.module.scss'

interface SiteAdminAlertProps {
    className?: string
    variant?: AlertProps['variant']
}

/**
 * An alert message with a site admin lock icon.
 */
export const SiteAdminAlert: React.FunctionComponent<React.PropsWithChildren<SiteAdminAlertProps>> = ({
    children,
    className = '',
    variant = 'warning',
}) => (
    <Alert className={classNames(styles.siteAdminAlert, className)} variant={variant}>
        <Typography.H5 as={Typography.H2}>
            <Icon role="img" as={LockIcon} aria-hidden={true} /> Site admin
        </Typography.H5>
        <div>{children}</div>
    </Alert>
)
