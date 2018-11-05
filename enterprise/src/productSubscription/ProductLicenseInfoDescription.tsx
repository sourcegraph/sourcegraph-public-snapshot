import React from 'react'
import * as GQL from '../../../packages/webapp/src/backend/graphqlschema'
import { formatUserCount } from './helpers'

export const ProductLicenseInfoDescription: React.SFC<{
    licenseInfo: GQL.IProductLicenseInfo
    className?: string
}> = ({ licenseInfo, className = '' }) => (
    <span
        className={className}
        title={licenseInfo.tags.length > 0 ? `Tags: ${licenseInfo.tags.join(', ')}` : 'No tags'}
    >
        {licenseInfo.productNameWithBrand} ({formatUserCount(licenseInfo.userCount)})
    </span>
)
