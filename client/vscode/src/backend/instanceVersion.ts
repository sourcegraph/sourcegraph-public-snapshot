import { noop } from 'lodash'
import { from, Observable } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { satisfies } from 'semver'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { EventSource } from '@sourcegraph/shared/src/graphql-operations'

import { displayWarning } from '../settings/displayWarnings'
import { INSTANCE_VERSION_NUMBER_KEY, LocalStorageService } from '../settings/LocalStorageService'

import { requestGraphQLFromVSCode } from './requestGraphQl'

/**
 * Gets the Sourcegraph instance version number via the GrapQL API.
 *
 * @returns An Observable that emits flattened Sourcegraph instance version number or undefined in case of an error:
 * - regular instance version format: 3.38.2
 * - insiders version format: 134683_2022-03-02_5188fes0101
 */
export const observeInstanceVersionNumber = (): Observable<string | undefined> =>
    from(requestGraphQLFromVSCode<SiteVersionResult>(siteVersionQuery, {})).pipe(
        map(dataOrThrowErrors),
        map(data => data.site.productVersion),
        catchError(error => {
            console.error('Failed to get instance version from host:', error)
            return [undefined]
        })
    )

export const isInsidersVersion = (version: string): boolean => version.split('.').length < 3

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
        observeInstanceVersionNumber()
            .toPromise()
            .then(async version => {
                if (version) {
                    if (!isInsidersVersion(version) && satisfies(version, '<3.32.0')) {
                        displayWarning(
                            'Your Sourcegraph instance version is not fully compatible with the Sourcegraph extension. Please ask your site admin to upgrade to version 3.32.0 or above. Read more about version support in our [troubleshooting docs](https://docs.sourcegraph.com/admin/how-to/troubleshoot-sg-extension#unsupported-features-by-sourcegraph-version).'
                        ).catch(() => {})
                    }
                    await localStorageService.setValue(INSTANCE_VERSION_NUMBER_KEY, version)
                }
            })
            .catch(noop) // We handle potential errors in instanceVersionNumber observable

        const version = localStorageService.getValue(INSTANCE_VERSION_NUMBER_KEY)
        // instances below 3.38.0 does not support EventSource.IDEEXTENSION and should fallback to BACKEND source
        return isInsidersVersion(version) || satisfies(version, '>=3.38.0')
            ? EventSource.IDEEXTENSION
            : EventSource.BACKEND
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
