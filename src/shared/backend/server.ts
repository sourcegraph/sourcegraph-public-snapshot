import { IClientConfigurationDetails } from '@sourcegraph/extensions-client-common/lib/schema/graphqlschema'
import { Observable } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import { GQL } from '../../types/gqlschema'
import { getContext } from './context'
import { queryGraphQLNoRetry } from './graphql'

/**
 * @return Observable that emits the client configuration details.
 *         Errors
 */
export const resolveClientConfiguration = (): Observable<IClientConfigurationDetails> =>
    queryGraphQLNoRetry(
        getContext({ repoKey: '' }),
        `query ClientConfiguration() {
            clientConfiguration {
                contentScriptUrls
                parentSourcegraph {
                    url
                }
            }
        }`
    ).pipe(
        map(result => {
            if (!result || !result.data) {
                throw new Error('No results')
            }
            return result.data.clientConfiguration
        }, catchError((err, caught) => caught))
    )

export const fetchCurrentUser = (useToken = true): Observable<GQL.IUser | undefined> =>
    queryGraphQLNoRetry(
        getContext({ repoKey: '' }),
        `query CurrentUser() {
            currentUser {
                id
                displayName
                username
                avatarURL
                url
                emails {
                    email
                }
                siteAdmin
            }
        }`,
        undefined,
        undefined,
        useToken
    ).pipe(
        map(result => {
            if (!result || !result.data || !result.data.currentUser) {
                return undefined
            }
            return result.data.currentUser
        }, catchError((err, caught) => caught))
    )

export const fetchSite = (): Observable<GQL.ISite> =>
    queryGraphQLNoRetry(
        getContext({ repoKey: '' }),
        `query SiteProductVersion() {
            site {
                productVersion
                buildVersion
                hasCodeIntelligence
            }
        }`
    ).pipe(
        map(result => {
            if (!result || !result.data) {
                throw new Error('unable to fetch site information.')
            }
            return result.data.site
        }, catchError((err, caught) => caught))
    )
