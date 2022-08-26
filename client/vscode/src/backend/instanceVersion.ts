import { gql } from '@sourcegraph/http-client'
import { EventSource } from '@sourcegraph/shared/src/graphql-operations'

import { displayWarning } from '../settings/displayWarnings'
import { INSTANCE_VERSION_NUMBER_KEY, LocalStorageService } from '../settings/LocalStorageService'

import { requestGraphQLFromVSCode } from './requestGraphQl'

/**
 * Regular instance version format: ex 3.38.2
 * Insider version format: ex 134683_2022-03-02_5188fes0101
 */
export async function getInstanceVersionNumber(): Promise<string | undefined> {
    try {
        const siteVersionResult = await requestGraphQLFromVSCode<SiteVersionResult>(siteVersionQuery, {})
        if (siteVersionResult.data) {
            // assume instance version longer than 8 is using insider version
            const flattenVersion =
                siteVersionResult.data.site.productVersion.length > 8
                    ? '999999'
                    : siteVersionResult.data.site.productVersion.split('.').join('')
            return flattenVersion
        }
    } catch (error) {
        console.error('Failed to get instance version from host:', error)
    }
    return
}

/**
 * This function will return the EventSource Type based
 * on the instance version
 */
export function initializeInstanceVersionNumber(
    localStorageService: LocalStorageService,
    instanceURL: string,
    accessToken: string | undefined
): EventSource {
    // Check only if a user is trying to connect to a private instance with a valid access token provided
    if (instanceURL !== 'https://sourcegraph.com' && accessToken) {
        getInstanceVersionNumber()
            .then(async flattenVersion => {
                if (flattenVersion) {
                    if (flattenVersion < '3320') {
                        displayWarning(
                            'Your Sourcegraph instance version is not fully compatible with the Sourcegraph extension. Please ask your site admin to upgrade to version 3.32.0 or above. Read more about version support in our [troubleshooting docs](https://docs.sourcegraph.com/admin/how-to/troubleshoot-sg-extension#unsupported-features-by-sourcegraph-version).'
                        ).catch(() => {})
                    }
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
    return EventSource.IDEEXTENSION
}

const siteVersionQuery = gql`
    query SiteProductVersion {
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
