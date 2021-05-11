import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import { ISearch } from '@sourcegraph/shared/src/graphql/schema'

import { requestGraphQL } from '../../../../backend/graphql'

/**
 * Fetch insight result.
 *
 * A bulk search is a package of search requests for each data series and each
 * data point of a particular data series.
 * */
export function fetchRawSearchInsightResults(searchQueries: string[]): Observable<Record<string, ISearch>> {
    return requestGraphQL<Record<string, ISearch>>(
        gql`
            query BulkSearch(${searchQueries.map((searchQuery, index) => `$query${index}: String!`).join(', ')}) {
                ${searchQueries
                    .map(
                        (searchQuery, index) => gql`
                        search${index}: search(version: V2, query: $query${index}) {
                            results {
                                matchCount
                            }
                        }`
                    )
                    .join('\n')}
            }`,
        Object.fromEntries(searchQueries.map((query, index) => [`query${index}`, query]))
    ).pipe(map(dataOrThrowErrors))
}

/**
 * Fetch closest commits according to step between points on insight chart.
 * */
export function fetchSearchInsightCommits(commitQueries: string[]): Observable<Record<string, ISearch>> {
    return requestGraphQL<Record<string, ISearch>>(
        gql`
            query BulkSearchCommits(${commitQueries.map((query, index) => `$query${index}: String!`).join(', ')}) {
                ${commitQueries
                    .map(
                        (query, index) => gql`
                    search${index}: search(version: V2, patternType: literal, query: $query${index}) {
                        results {
                            results {
                                ... on CommitSearchResult {
                                    commit {
                                        oid
                                        committer {
                                            date
                                        }
                                    }
                                }
                            }
                        }
                    }
                `
                    )
                    .join('\n')}
            }
        `,
        Object.fromEntries(commitQueries.map((query, index) => [`query${index}`, query]))
    ).pipe(map(dataOrThrowErrors))
}
