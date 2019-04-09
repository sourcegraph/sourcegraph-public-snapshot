import { Observable } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { sourcegraphUrl } from '../util/context'
import { queryGraphQL } from './graphql'

/**
 * @return Observable that emits the client configuration details.
 *         Errors
 */
export const resolveClientConfiguration = (): Observable<GQL.IClientConfigurationDetails> =>
    queryGraphQL(gql`query ClientConfiguration() {
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

export const fetchCurrentUser = (): Observable<GQL.IUser | undefined> =>
    queryGraphQL(gql`query CurrentUser() {
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

export const fetchSite = (url = sourcegraphUrl): Observable<GQL.ISite> =>
    queryGraphQL(gql`query SiteProductVersion() {
            site {
                productVersion
                buildVersion
                hasCodeIntelligence
            }
        }`).pipe(
        map(dataOrThrowErrors),
        map(({ site }) => site, catchError((err, caught) => caught))
    )
