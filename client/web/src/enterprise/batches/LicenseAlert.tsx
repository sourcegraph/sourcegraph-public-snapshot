import React from 'react'

import { useQuery } from '@sourcegraph/http-client'
import { Alert, AlertProps } from '@sourcegraph/wildcard'

import { GetLicenseAndUsageInfoResult, GetLicenseAndUsageInfoVariables } from '../../graphql-operations'

import { GET_LICENSE_AND_USAGE_INFO } from './list/backend'

export interface LicenseAlertProps {
    variant?: AlertProps['variant']
    additionalCondition?: boolean
    onLicenseRetrieved?: (data: GetLicenseAndUsageInfoResult) => void
}

export const LicenseAlert: React.FunctionComponent<React.PropsWithChildren<LicenseAlertProps>> = ({
    variant = 'info',
    additionalCondition = true,
    onLicenseRetrieved,
}) => {
    const { data: licenseAndUsageInfo } = useQuery<GetLicenseAndUsageInfoResult, GetLicenseAndUsageInfoVariables>(
        GET_LICENSE_AND_USAGE_INFO,
        { onCompleted: onLicenseRetrieved }
    )

    if (!licenseAndUsageInfo?.batchChanges && !licenseAndUsageInfo?.campaigns && additionalCondition) {
        return (
            <Alert variant={variant}>
                <div className="mb-2">
                    <strong>Your license only allows for 5 changesets per Batch Change</strong>
                </div>
                <div>
                    You are running a free version of batch changes. It is fully functional, however it will only
                    generate 5 changesets per batch change. If you would like to learn more about our pricing, contact
                    us.
                </div>
            </Alert>
        )
    }
    return <></>
}
