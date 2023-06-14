import React from 'react'

import { mdiCheckCircle, mdiCloseCircle, mdiShieldRemove } from '@mdi/js'
import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { Icon } from '@sourcegraph/wildcard'

import { ProductLicenseFields } from '../../../graphql-operations'
import { isProductLicenseExpired } from '../../../productSubscription/helpers'

const getIcon = (isExpired: boolean, isRevoked: boolean): string => {
    if (isExpired) {
        return mdiCloseCircle
    }
    if (isRevoked) {
        return mdiShieldRemove
    }
    return mdiCheckCircle
}

const ValidityIcon: React.FC<{ isExpired: boolean; isRevoked: boolean }> = ({ isExpired, isRevoked }) => (
    <Icon
        svgPath={getIcon(isExpired, isRevoked)}
        aria-hidden={true}
        className={classNames('mr-1', {
            ['text-success']: !isExpired && !isRevoked,
            ['text-danger']: isExpired && !isRevoked,
            ['text-warning']: !isExpired && isRevoked,
        })}
    />
)

const getText = (isExpired: boolean, isRevoked: boolean): string => {
    if (isExpired) {
        return 'Expired'
    }
    if (isRevoked) {
        return 'Revoked'
    }
    return 'Valid'
}

/**
 * Displays an alert indicating the validity of a product license.
 */
export const ProductLicenseValidity: React.FunctionComponent<
    React.PropsWithChildren<{
        license: ProductLicenseFields
        variant?: 'icon-only' | 'no-icon'
        className?: string
    }>
> = ({ license: { info, revokedAt, revokeReason }, variant, className = '' }) => {
    const expiresAt = info?.expiresAt ?? 0
    const isExpired = isProductLicenseExpired(expiresAt)
    const isRevoked = !!revokedAt
    const timestamp = revokedAt ?? expiresAt
    const timestampSuffix = isExpired || isRevoked ? 'ago' : 'remaining'

    if (variant === 'icon-only') {
        return (
            <div className={className}>
                <ValidityIcon isExpired={isExpired} isRevoked={isRevoked} />
            </div>
        )
    }
    return (
        <div className={className}>
            {variant !== 'no-icon' && <ValidityIcon isExpired={isExpired} isRevoked={isRevoked} />}
            {getText(isExpired, isRevoked)} <Timestamp date={timestamp} noAbout={true} noAgo={true} /> {timestampSuffix}
            {!isExpired && isRevoked && revokeReason && <span> because of "{revokeReason}" reason</span>}
        </div>
    )
}
