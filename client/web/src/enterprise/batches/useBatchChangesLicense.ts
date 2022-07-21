import { useCallback } from 'react'

import { useQuery } from '@sourcegraph/http-client'

import { GetLicenseAndUsageInfoResult, GetLicenseAndUsageInfoVariables } from '../../graphql-operations'

import { GET_LICENSE_AND_USAGE_INFO } from './list/backend'

export interface UseBatchChangesLicenseResult {
    licenseAndUsageInfo: GetLicenseAndUsageInfoResult | undefined
    isUnlicensed: boolean
    maxUnlicensedChangesets: number
    exceedsLicense: (count: number) => boolean
}

export const useBatchChangesLicense = (): UseBatchChangesLicenseResult => {
    const { data: licenseAndUsageInfo } = useQuery<GetLicenseAndUsageInfoResult, GetLicenseAndUsageInfoVariables>(
        GET_LICENSE_AND_USAGE_INFO,
        {}
    )

    const isUnlicensed = licenseAndUsageInfo
        ? !licenseAndUsageInfo.batchChanges && !licenseAndUsageInfo.campaigns
        : false
    const maxUnlicensedChangesets = licenseAndUsageInfo ? licenseAndUsageInfo.maxUnlicensedChangesets : 0

    const exceedsLicense = useCallback((count: number) => isUnlicensed && count > maxUnlicensedChangesets, [
        isUnlicensed,
        maxUnlicensedChangesets,
    ])

    return { licenseAndUsageInfo, isUnlicensed, maxUnlicensedChangesets, exceedsLicense }
}
