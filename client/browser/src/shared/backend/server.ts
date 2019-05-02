import { Observable } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContext } from '../../../../../shared/src/platform/context'

/**
 * @return Observable that emits the client configuration details.
 *         Errors
 */
export const resolveClientConfiguration = (
    queryGraphQL: PlatformContext['requestGraphQL']
): Observable<GQL.IClientConfigurationDetails> =>
    queryGraphQL<GQL.IQuery>(gql`query ClientConfiguration() {
            clientConfiguration {
                contentScriptUrls
                parentSourcegraph {
                    url
                }
            }
        }`).pipe(
        map(dataOrThrowErrors),
        map(({ clientConfiguration }) => clientConfiguration, catchError((err, caught) => caught))
    )

export const fetchCurrentUser = (queryGraphQL: PlatformContext['requestGraphQL']): Observable<GQL.IUser | undefined> =>
    queryGraphQL<GQL.IQuery>(gql`query CurrentUser() {
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
        }`).pipe(
        map(dataOrThrowErrors),
        map(({ currentUser }) => currentUser || undefined, catchError((err, caught) => caught))
    )

export const fetchSite = (queryGraphQL: PlatformContext['requestGraphQL']): Observable<GQL.ISite> =>
    queryGraphQL<GQL.IQuery>(gql`query SiteProductVersion() {
            site {
                productVersion
                buildVersion
                hasCodeIntelligence
            }
        }`).pipe(
        map(dataOrThrowErrors),
        map(({ site }) => site, catchError((err, caught) => caught))
    )
