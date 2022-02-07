import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../backend/graphql'
import { RepositoriesByNamesResult, RepositoriesByNamesVariables } from '../../graphql-operations'

export function fetchRepositoriesByNames(
    names: string[]
): Observable<RepositoriesByNamesResult['repositories']['nodes']> {
    return requestGraphQL<RepositoriesByNamesResult, RepositoriesByNamesVariables>(
        gql`
            query RepositoriesByNames($names: [String!]!) {
                repositories(names: $names) {
                    nodes {
                        id
                        name
                    }
                }
            }
        `,
        {
            names,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.repositories.nodes)
    )
}
