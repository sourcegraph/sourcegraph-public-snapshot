import { Observable, ReplaySubject } from 'rxjs'
import { filter, mergeMap, take, tap } from 'rxjs/operators'
import { SiteFlags } from '.'
import { authRequired } from '../auth'
import { gql, queryGraphQL } from '../backend/graphql'
import { createAggregateError } from '../util/errors'

/**
 * The latest state of the site flags.
 */
export const siteFlags = new ReplaySubject<SiteFlags>(1)

/**
 * refreshSiteFlags refreshes the site flags. The result is available from
 * the siteFlags const.
 */
export function refreshSiteFlags(): Observable<never> {
    return authRequired.pipe(
        take(1),
        filter(authRequired => !authRequired),
        mergeMap(() =>
            queryGraphQL(gql`
                query SiteFlags {
                    site {
                        needsRepositoryConfiguration
                        noRepositoriesEnabled
                        alerts {
                            type
                            message
                            isDismissibleWithKey
                        }
                        hasCodeIntelligence
                        authProviders {
                            nodes {
                                serviceType
                                serviceID
                                clientID
                                displayName
                                isBuiltin
                                authenticationURL
                            }
                        }
                        disableBuiltInSearches
                        sendsEmailVerificationEmails
                        updateCheck {
                            pending
                            checkedAt
                            errorMessage
                            updateVersionAvailable
                        }
                    }
                }
            `)
        ),
        tap(({ data, errors }) => {
            if (!data || !data.site) {
                throw createAggregateError(errors)
            }
            siteFlags.next(data.site)
        }),
        mergeMap(() => [])
    )
}

refreshSiteFlags()
    .toPromise()
    .then(() => void 0, err => console.error(err))
