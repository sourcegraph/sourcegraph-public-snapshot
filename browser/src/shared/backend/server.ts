import { Observable } from 'rxjs'
import { catchError, map } from 'rxjs/operators'
import * as GQL from '../../../../shared/src/graphql/schema'
import { getContext } from './context'
import { queryGraphQL } from './graphql'

/**
 * @return Observable that emits the client configuration details.
 *         Errors
 */
export const resolveClientConfiguration = (): Observable<GQL.IClientConfigurationDetails> =>
    queryGraphQL({
        ctx: getContext(),
        request: `query ClientConfiguration() {
            clientConfiguration {
                contentScriptUrls
                parentSourcegraph {
                    url
                }
            }
        }`,
        requestMightContainPrivateInfo: false,
    }).pipe(map(({ clientConfiguration }) => clientConfiguration, catchError((err, caught) => caught)))

export const fetchCurrentUser = (): Observable<GQL.IUser | undefined> =>
    queryGraphQL({
        ctx: getContext(),
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
        requestMightContainPrivateInfo: false,
    }).pipe(map(({ currentUser }) => currentUser || undefined, catchError((err, caught) => caught)))

export const fetchSite = (url = sourcegraphUrl): Observable<GQL.ISite> =>
    queryGraphQL({
        ctx: getContext(),
        request: `query SiteProductVersion() {
            site {
                productVersion
                buildVersion
                hasCodeIntelligence
            }
        }`,
        requestMightContainPrivateInfo: false,
        url,
    }).pipe(map(({ site }) => site, catchError((err, caught) => caught)))
