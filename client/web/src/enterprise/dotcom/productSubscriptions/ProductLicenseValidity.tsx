import React from 'react'

import { mdiCheckCircle, mdiCloseCircle } from '@mdi/js'
import { parseISO } from 'date-fns'
import format from 'date-fns/format'

import { Icon, Tooltip, Text } from '@sourcegraph/wildcard'

import { ProductLicenseInfoFields } from '../../../graphql-operations'
import { formatRelativeExpirationDate, isProductLicenseExpired } from '../../../productSubscription/helpers'

/**
 * Displays an alert indicating the validity of a product license.
 */
export const ProductLicenseValidity: React.FunctionComponent<
    React.PropsWithChildren<{
        licenseInfo: ProductLicenseInfoFields
        className?: string
    }>
> = ({ licenseInfo: { expiresAt }, className = '' }) => {
    const isExpired = isProductLicenseExpired(expiresAt)
    const tooltip = format(parseISO(expiresAt), 'PPpp')

    return (
        <Tooltip content={tooltip}>
            <Text className={className}>
                {!isExpired && <Icon svgPath={mdiCheckCircle} aria-hidden={true} className="mr-1 text-success" />}
                {isExpired && <Icon svgPath={mdiCloseCircle} aria-hidden={true} className="mr-1 text-danger" />}
                <strong>{isExpired ? 'Expired' : 'Valid'}</strong> ({formatRelativeExpirationDate(expiresAt)})
            </Text>
        </Tooltip>
    )
}
