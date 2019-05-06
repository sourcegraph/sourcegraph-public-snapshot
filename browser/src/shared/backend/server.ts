import { Observable } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { isOptions } from '../../context'
import { sourcegraphUrl } from '../util/context'
import { getContext } from './context'
import { queryGraphQL } from './graphql'

/**
 * @return Observable that emits the client configuration details.
 *         Errors
 */
export const resolveClientConfiguration = (): Observable<GQL.IClientConfigurationDetails> =>
    queryGraphQL({
        ctx: getContext({ repoKey: '' }),
        request: `query ClientConfiguration() {
            clientConfiguration {
                contentScriptUrls
                parentSourcegraph {
                    url
                }
            }
        }`,
        retry: false,
        requestMightContainPrivateInfo: false,
    }).pipe(
        map(result => {
            if (!result || !result.data) {
                throw new Error('No results')
            }
            return result.data.clientConfiguration
        }, catchError((err, caught) => caught))
    )

export const fetchCurrentUser = (useAccessToken = true): Observable<GQL.IUser | undefined> =>
    queryGraphQL({
        ctx: getContext({ repoKey: '' }),
        request: `query CurrentUser() {
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
        useAccessToken,
        retry: false,
        requestMightContainPrivateInfo: false,
    }).pipe(
        map(result => {
            if (!result || !result.data || !result.data.currentUser) {
                return undefined
            }
            return result.data.currentUser
        }, catchError((err, caught) => caught))
    )

export const fetchSite = (url = sourcegraphUrl): Observable<GQL.ISite> =>
    queryGraphQL({
        ctx: getContext({ repoKey: '' }),
        request: `query SiteProductVersion() {
            site {
                productVersion
                buildVersion
                hasCodeIntelligence
            }
        }`,
        retry: false,
        requestMightContainPrivateInfo: false,
        url,
        useAccessToken: !isOptions,
    }).pipe(
        map(result => {
            if (!result || !result.data) {
                throw new Error('unable to fetch site information.')
            }
            return result.data.site
        }, catchError((err, caught) => caught))
    )
