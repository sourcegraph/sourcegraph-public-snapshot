import * as React from 'react'

import classNames from 'classnames'
import LockIcon from 'mdi-react/LockIcon'

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
        {/**
         * ally-ignore
         * Rule: "heading-order" (Heading levels should only increase by one)
         * Since `PageHeader` (which is on upper scope), renders `h1`, We could consider using `h2` tag instead, to meet accessibility criteria
         */}
        <h5 className="a11y-ignore">
            <Icon as={LockIcon} /> Site admin
        </h5>
        <div>{children}</div>
    </Alert>
)
