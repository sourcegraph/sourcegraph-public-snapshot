import { gql } from '@sourcegraph/http-client'
import { EventSource } from '@sourcegraph/shared/src/graphql-operations'

import { INSTANCE_VERSION_NUMBER_KEY, LocalStorageService } from '../settings/LocalStorageService'

import { requestGraphQLFromVSCode } from './requestGraphQl'

/**
 * Regular instance version format: ex 3.38.2
 * Insider version format: ex 134683_2022-03-02_5188fes0101
 * This function will return the EventSource Type based
 * on the instance version
 */
export function initializeInstantVersionNumber(localStorageService: LocalStorageService): EventSource {
    requestGraphQLFromVSCode<SiteVersionResult>(siteVersionQuery, {})
        .then(async siteVersionResult => {
            if (siteVersionResult.data) {
                // assume instance version longer than 8 is using insider version
                const flattenVersion =
                    siteVersionResult.data.site.productVersion.length > 8
                        ? '999999'
                        : siteVersionResult.data.site.productVersion.split('.').join('')
                await localStorageService.setValue(INSTANCE_VERSION_NUMBER_KEY, flattenVersion)
            }
        })
        .catch(error => {
            console.error('Failed to get instance version from host:', error)
        })
    const versionNumber = localStorageService.getValue(INSTANCE_VERSION_NUMBER_KEY)
    // instances below 3.38.0 does not support EventSource.IDEEXTENSION and should fallback to BACKEND source
    return versionNumber >= '3380' ? EventSource.IDEEXTENSION : EventSource.BACKEND
}

const siteVersionQuery = gql`
    query {
        site {
            productVersion
        }
    }
`
interface SiteVersionResult {
    site: {
        productVersion: string
    }
}
