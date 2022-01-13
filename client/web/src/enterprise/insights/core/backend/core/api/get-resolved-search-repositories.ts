import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../../../../../backend/graphql'
import { SearchRepositoriesResult } from '../../../../../../graphql-operations'

/**
 * Get list of resolved repositories from the search API.
 *
 * @param query - search query
 */
export const getResolvedSearchRepositories = (query: string): Promise<string[]> =>
    requestGraphQL<SearchRepositoriesResult>(
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
        { query }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(result => result.search?.results?.repositories ?? []),
            // Get only the first 10 repositories to avoid DDoS from the live preview in
            // insight creation UI or at insights page.
            map(result => result.map(repo => repo.name).slice(0, 10))
        )
        .toPromise()
