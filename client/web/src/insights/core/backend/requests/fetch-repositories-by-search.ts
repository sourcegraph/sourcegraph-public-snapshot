import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../../backend/graphql'
import { SearchRepositoriesResult } from '../../../../graphql-operations'

/**
 * Resolve repositories from the search query.
 * Used for generate repositories field value in 1 click insight creation scenario.
 */
export function fetchRepositoriesBySearch(searchQuery: string): Observable<string[]> {
    return requestGraphQL<SearchRepositoriesResult>(
        gql`
            query SearchRepositories($query: String) {
                search(query: $query) {
                    results {
                        repositories {
                            name
                        }
                    }
                }
            }
        `,
        { query: searchQuery }
    ).pipe(
        map(dataOrThrowErrors),
        map(result => result.search?.results?.repositories ?? []),
        map(result => result.map(repo => repo.name))
    )
}
