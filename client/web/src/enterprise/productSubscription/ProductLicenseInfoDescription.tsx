import React from 'react'

import * as GQL from '@sourcegraph/shared/src/schema'

import { formatUserCount } from '../../productSubscription/helpers'

export const ProductLicenseInfoDescription: React.FunctionComponent<
    React.PropsWithChildren<{
        licenseInfo: GQL.IProductLicenseInfo
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
