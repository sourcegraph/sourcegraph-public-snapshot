import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { queryGraphQL } from '../../backend/graphql'
import type { SiteAdminPreciseIndexResult } from '../../graphql-operations'

/**
 * Fetch a single precise index by id.
 */
export function fetchPreciseIndex({
    id,
}: {
    id: string
}): Observable<Extract<SiteAdminPreciseIndexResult['node'], { __typename: 'PreciseIndex' }> | null> {
    return queryGraphQL<SiteAdminPreciseIndexResult>(
        gql`
            query SiteAdminPreciseIndex($id: ID!) {
                node(id: $id) {
                    __typename
                    ... on PreciseIndex {
                        projectRoot {
                            commit {
                                repository {
                                    name
                                    url
                                }
                            }
                        }
                    }
                }
            }
        `,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                return null
            }
            if (node.__typename !== 'PreciseIndex') {
                throw new Error(`The given ID is a ${node.__typename}, not a PreciseIndex`)
            }

            return node
        })
    )
}
