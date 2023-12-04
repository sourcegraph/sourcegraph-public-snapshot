import React from 'react'

import { H3 } from '@sourcegraph/wildcard'

import type { ProductLicenseInfoFields } from '../../graphql-operations'
import { formatUserCount } from '../../productSubscription/helpers'

export const ProductLicenseInfoDescription: React.FunctionComponent<
    React.PropsWithChildren<{
        licenseInfo: ProductLicenseInfoFields
        className?: string
    }>
> = ({ licenseInfo, className = '' }) => (
    <H3 className={className}>
        {licenseInfo.productNameWithBrand} ({formatUserCount(licenseInfo.userCount)})
    </H3>
)
