import { Observable } from 'rxjs'
import { catchError, map, mapTo } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContext } from '../../../../shared/src/platform/context'
import { isAjaxError } from '../../../../shared/src/backend/fetch'
import { isErrorLike } from '../../../../shared/src/util/errors'

export const fetchCurrentUser = (
    requestGraphQL: PlatformContext['requestGraphQL']
): Observable<GQL.IUser | undefined> =>
    requestGraphQL<GQL.IQuery>({
        request: gql`query CurrentUser() {
            currentUser {
                id
                displayName
                username
                avatarURL
                url
                settingsURL
                emails {
                    email
                }
                siteAdmin
            }
        }`,
        variables: {},
        mightContainPrivateInfo: false,
    }).pipe(
        map(dataOrThrowErrors),
        map(({ currentUser }) => currentUser || undefined, catchError((err, caught) => caught))
    )

export const fetchSite = (requestGraphQL: PlatformContext['requestGraphQL']): Observable<GQL.ISite> =>
    requestGraphQL<GQL.IQuery>({
        request: gql`
            query SiteProductVersion {
                site {
                    productVersion
                    buildVersion
                    hasCodeIntelligence
                }
            }
        `,
        variables: {},
        mightContainPrivateInfo: false,
    }).pipe(
        map(dataOrThrowErrors),
        map(({ site }) => site, catchError((err, caught) => caught))
    )

/**
 * Checks whether the user is logged in to the Sourcegraph instance, by calling a {@link fetchSite},
 * and catching Ajax errors with a 401 status code.
 */
export const checkUserLoggedIn = (requestGraphQL: PlatformContext['requestGraphQL']): Observable<boolean> =>
    fetchSite(requestGraphQL).pipe(
        mapTo(true),
        catchError(err => {
            if (isAjaxError(err) && err.response.status === 401) {
                return [false]
            }
            // Workaround for https://github.com/mozilla/webextension-polyfill/issues/210 in the browser extension
            if (isErrorLike(err) && err.message.includes('failed with 401')) {
                return [false]
            }
            throw err
        })
    )
