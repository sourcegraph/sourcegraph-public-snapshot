import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../backend/graphql'
import type { RepositoriesByNamesResult, RepositoriesByNamesVariables } from '../../graphql-operations'

export function fetchRepositoriesByNames(
    names: string[]
): Observable<RepositoriesByNamesResult['repositories']['nodes']> {
    const first = names.length
    return requestGraphQL<RepositoriesByNamesResult, RepositoriesByNamesVariables>(
        gql`
            query RepositoriesByNames($names: [String!]!, $first: Int!) {
                repositories(names: $names, first: $first) {
                    nodes {
                        id
                        name
                    }
                }
            }
        `,
        {
            names,
            first,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.repositories.nodes)
    )
}
