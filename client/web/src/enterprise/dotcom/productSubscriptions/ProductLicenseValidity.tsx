import React from 'react'

import { mdiCheckCircle, mdiCloseCircle, mdiShieldRemove } from '@mdi/js'
import classNames from 'classnames'

import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { Icon, Label } from '@sourcegraph/wildcard'

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
        className?: string
    }>
> = ({ license: { info, revokedAt, revokeReason }, className = '' }) => {
    const expiresAt = info?.expiresAt ?? 0
    const isExpired = isProductLicenseExpired(expiresAt)
    const isRevoked = !!revokedAt
    const timestamp = revokedAt ?? expiresAt
    const timestampSuffix = isExpired || isRevoked ? 'ago' : 'remaining'

    return (
        <div className={className}>
            <Icon
                svgPath={getIcon(isExpired, isRevoked)}
                aria-hidden={true}
                className={classNames('mr-1', {
                    ['text-success']: !isExpired && !isRevoked,
                    ['text-danger']: isExpired || isRevoked,
                })}
            />
            <strong>{getText(isExpired, isRevoked)}</strong> (
            <Timestamp date={timestamp} noAbout={true} noAgo={true} /> {timestampSuffix})
            {!isExpired && isRevoked && revokeReason && (
                <div className="mt-1 d-flex">
                    <Label>Reason to revoke</Label>
                    <span className="ml-3">{revokeReason}</span>
                </div>
            )}
        </div>
    )
}
