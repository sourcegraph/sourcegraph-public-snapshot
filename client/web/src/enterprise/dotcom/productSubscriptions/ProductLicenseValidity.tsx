import classNames from 'classnames'
import { parseISO } from 'date-fns'
import format from 'date-fns/format'
import React from 'react'

import * as GQL from '@sourcegraph/shared/src/schema'

import { formatRelativeExpirationDate, isProductLicenseExpired } from '../../../productSubscription/helpers'

/**
 * Displays an alert indicating the validity of a product license.
 */
export const ProductLicenseValidity: React.FunctionComponent<{
    licenseInfo: GQL.IProductLicenseInfo
    primary: boolean
    className?: string
}> = ({ licenseInfo: { expiresAt }, primary, className = '' }) => {
    const isExpired = isProductLicenseExpired(expiresAt)
    const validityClass = isExpired ? 'danger' : 'success'
    return (
        <div
            className={classNames(className, primary && `alert alert-${validityClass} py-1 px-2`)}
            data-tooltip={format(parseISO(expiresAt), 'PPpp')}
        >
            <strong className={classNames(!primary && `text-${validityClass}`)}>
                {isExpired ? 'Expired' : 'Valid'}
            </strong>{' '}
            ({formatRelativeExpirationDate(expiresAt)})
        </div>
    )
}
