import React, { ReactNode } from 'react'

import { useQuery } from '@sourcegraph/http-client'
import { Alert, AlertProps } from '@sourcegraph/wildcard'

import { GetLicenseAndUsageInfoResult, GetLicenseAndUsageInfoVariables } from '../../graphql-operations'

import { GET_LICENSE_AND_USAGE_INFO } from './list/backend'

export interface LicenseAlertProps {
    variant?: AlertProps['variant']
    // By default, the license is enough to determine to display alert. There may be cases where the total number of
    // changesets need to be checked against the max allowed.
    totalChangesetCount?: number
    // Allows the ability to apply additional logic to the parent component (such as disabling a button).
    onLicenseRetrieved?: (data: GetLicenseAndUsageInfoResult) => void
    children?: ReactNode
}

export const LicenseAlert: React.FunctionComponent<React.PropsWithChildren<LicenseAlertProps>> = ({
    variant = 'info',
    totalChangesetCount,
    onLicenseRetrieved,
    children,
}) => {
    const { data: licenseAndUsageInfo } = useQuery<GetLicenseAndUsageInfoResult, GetLicenseAndUsageInfoVariables>(
        GET_LICENSE_AND_USAGE_INFO,
        { onCompleted: onLicenseRetrieved }
    )

    if (!licenseAndUsageInfo) {
        return <></>
    }

    // If totalChangesetCount is not provided then display the alert simply based on if the feature is enabled in the
    // license.
    const exceedsLimit = totalChangesetCount ? totalChangesetCount > licenseAndUsageInfo.maxUnlicensedChangesets : true
    if (!licenseAndUsageInfo.batchChanges && !licenseAndUsageInfo.campaigns && exceedsLimit) {
        return <Alert variant={variant}>{children}</Alert>
    }
    return <></>
}
