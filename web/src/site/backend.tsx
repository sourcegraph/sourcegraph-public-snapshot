import { Observable } from 'rxjs/Observable'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { tap } from 'rxjs/operators/tap'
import { ReplaySubject } from 'rxjs/ReplaySubject'
import { SiteFlags } from '.'
import { gql, queryGraphQL } from '../backend/graphql'

/**
 * The latest state of the site flags.
 */
export const siteFlags = new ReplaySubject<SiteFlags>(1)

/**
 * refreshSiteFlags refreshes the site flags. The result is available from
 * the siteFlags const.
 */
export function refreshSiteFlags(): Observable<never> {
    return queryGraphQL(gql`
        query SiteFlags {
            site {
                repositoriesCloning: repositories(cloning: true) {
                    nodes {
                        uri
                    }
                    totalCount
                }
                needsRepositoryConfiguration
                hasCodeIntelligence
            }
        }
    `).pipe(
        tap(
            ({
                data,
                errors,
            }: {
                data?: GQL.IQuery & { site: { repositoriesCloning: GQL.IRepositoryConnection } }
                errors?: GQL.IGraphQLResponseError[]
            }) => {
                if (!data || !data.site || (errors && errors.length > 0)) {
                    throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
                }
                siteFlags.next(data.site)
            }
        ),
        mergeMap(() => [])
    )
}

refreshSiteFlags()
    .toPromise()
    .then(() => void 0, err => console.error(err))
