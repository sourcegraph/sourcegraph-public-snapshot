import { useCallback } from 'react'

import { useQuery } from '@sourcegraph/http-client'

import type { GetLicenseAndUsageInfoResult, GetLicenseAndUsageInfoVariables } from '../../graphql-operations'

import { GET_LICENSE_AND_USAGE_INFO } from './list/backend'

export interface UseBatchChangesLicenseResult {
    /**
     * The result of the license and usage query.
     */
    licenseAndUsageInfo: GetLicenseAndUsageInfoResult | undefined
    /**
     * Determines if the account is licensed or not.
     */
    isUnlicensed: boolean
    /**
     * The max number of changesets allowed for an unlicensed account.
     */
    maxUnlicensedChangesets: number
    /**
     * Method used to determine if the provided count exceeds the usage of the license.
     *
     * @param count The count to compare against the max allowed changesets.
     */
    exceedsLicense: (count: number) => boolean
}

/**
 * Custom hook to retrieve information about batch changes feature of the account's license.
 */
export const useBatchChangesLicense = (): UseBatchChangesLicenseResult => {
    const { data: licenseAndUsageInfo } = useQuery<GetLicenseAndUsageInfoResult, GetLicenseAndUsageInfoVariables>(
        GET_LICENSE_AND_USAGE_INFO,
        {}
    )

    const isUnlicensed = licenseAndUsageInfo
        ? !licenseAndUsageInfo.batchChanges && !licenseAndUsageInfo.campaigns
        : false
    const maxUnlicensedChangesets = licenseAndUsageInfo ? licenseAndUsageInfo.maxUnlicensedChangesets : 0

    const exceedsLicense = useCallback(
        (count: number) => isUnlicensed && count > maxUnlicensedChangesets,
        [isUnlicensed, maxUnlicensedChangesets]
    )

    return { licenseAndUsageInfo, isUnlicensed, maxUnlicensedChangesets, exceedsLicense }
}
