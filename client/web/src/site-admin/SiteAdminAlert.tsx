import * as React from 'react'

import { mdiLock } from '@mdi/js'
import classNames from 'classnames'

import { Alert, type AlertProps, Icon, H2, H5 } from '@sourcegraph/wildcard'

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
        <H5 as={H2}>
            <Icon aria-hidden={true} svgPath={mdiLock} /> Site admin
        </H5>
        <div>{children}</div>
    </Alert>
)
