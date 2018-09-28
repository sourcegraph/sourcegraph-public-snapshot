import * as GQL from '@sourcegraph/webapp/dist/backend/graphqlschema'
import { pluralize } from '@sourcegraph/webapp/dist/util/strings'
import React from 'react'

/**
 * Displays the product subscription's plan and other related information.
 */
export const ProductLicenseInfoPlanDescription: React.SFC<{
    license: GQL.IProductLicenseInfo | null
    className?: string
}> = ({ license, className = '' }) => (
    <span className={className}>
        {license ? (
            <>
                {license.plan} &mdash;{' '}
                {license.userCount === null
                    ? 'Unlimited users'
                    : `${license.userCount} ${pluralize('user', license.userCount)}`}
            </>
        ) : (
            'Invalid license'
        )}
    </span>
)
