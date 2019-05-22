import { isBefore } from 'date-fns'
import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Timestamp } from '../../../components/time/Timestamp'

/**
 * Displays a text label with the date a license expires.
 */
export const LicenseExpiresAtLabel: React.FunctionComponent<{
    productSubscription: Pick<GQL.IProductSubscription, 'activeLicense'>
    className?: string
}> = ({ productSubscription, className = '' }) => (
    <span className={className}>
        {productSubscription.activeLicense && productSubscription.activeLicense.info ? (
            <>
                <Timestamp
                    date={productSubscription.activeLicense.info.expiresAt}
                    className={
                        isBefore(new Date(productSubscription.activeLicense.info.expiresAt), Date.now())
                            ? 'text-danger'
                            : ''
                    }
                />
            </>
        ) : (
            <span className="text-muted font-italic">No license</span>
        )}
    </span>
)
