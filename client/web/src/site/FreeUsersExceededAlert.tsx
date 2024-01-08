import React from 'react'

import { Alert, Link } from '@sourcegraph/wildcard'

/**
 * A global alert that appears telling all users that they have exceeded the limit of free users allowed.
 */
export const FreeUsersExceededAlert: React.FunctionComponent<
    React.PropsWithChildren<{
        noLicenseWarningUserCount: number | null
        className?: string
    }>
> = ({ noLicenseWarningUserCount, className }) => (
    <Alert className={className} variant="danger">
        This Sourcegraph instance has reached{' '}
        {noLicenseWarningUserCount === null ? 'the limit for' : noLicenseWarningUserCount} free users, and an admin must{' '}
        <Link className="site-alert__link" to="https://sourcegraph.com/contact/sales">
            <span className="underline">contact Sourcegraph to start a free trial or purchase a license</span>
        </Link>{' '}
        to add more
    </Alert>
)
