import format from 'date-fns/format'
import React from 'react'
import { formatRelativeExpirationDate, isProductLicenseExpired } from './helpers'

/** Displays an expiration date (for product subscriptions or licenses). */
export const ExpirationDate: React.SFC<{
    /** The expiration date of a product subscription or license. */
    date: string | number

    /** Show the time of day of expiration. */
    showTime?: boolean

    /** Show "(T remaining)" or "(T ago)" relative times. */
    showRelative?: boolean

    lowercase?: boolean
}> = ({ date, showTime, showRelative, lowercase }) => {
    const isExpired = isProductLicenseExpired(date)
    const text = isExpired ? 'Expired on' : 'Valid until'
    return (
        <span>
            {lowercase ? text.toLowerCase() : text}{' '}
            {showTime ? format(date, 'PPpp') : <span title={format(date, 'PPpp')}>{format(date, 'yyyy-MM-dd')}</span>}
            {showRelative && ` (${formatRelativeExpirationDate(date)})`}
        </span>
    )
}
