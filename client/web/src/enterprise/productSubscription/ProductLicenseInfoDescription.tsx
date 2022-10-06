import React from 'react'

import { ProductLicenseFields } from '../../graphql-operations'
import { formatUserCount } from '../../productSubscription/helpers'

export const ProductLicenseInfoDescription: React.FunctionComponent<
    React.PropsWithChildren<{
        licenseInfo: NonNullable<ProductLicenseFields['info']>
        className?: string
    }>
> = ({ licenseInfo, className = '' }) => (
    <span
        className={className}
        title={licenseInfo.tags.length > 0 ? `Tags: ${licenseInfo.tags.join(', ')}` : 'No tags'}
    >
        {licenseInfo.productNameWithBrand} ({formatUserCount(licenseInfo.userCount)})
    </span>
)
