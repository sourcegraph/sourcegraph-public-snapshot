import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../../../../../../../backend/graphql'
import { BulkSearchCommits, BulkSearchFields } from '../../../../../../../../graphql-operations'

const bulkSearchFieldsFragment = gql`
    fragment BulkSearchFields on Search {
        results {
            matchCount
        }
    }
`

/**
 * Fetch insight result.
 *
 * A bulk search is a package of search requests for each data series and each
 * data point of a particular data series.
 * */
export function fetchRawSearchInsightResults(searchQueries: string[]): Observable<Record<string, BulkSearchFields>> {
    return requestGraphQL<Record<string, BulkSearchFields>>(
        `
            query BulkSearch(${searchQueries.map((searchQuery, index) => `$query${index}: String!`).join(', ')}) {
                ${searchQueries
                    .map(
                        (searchQuery, index) => `
                            search${index}: search(version: V2, query: $query${index}) {
                                ...BulkSearchFields
                            }
                        `
                    )
                    .join('\n')}
            }
            ${bulkSearchFieldsFragment}
        `,
        Object.fromEntries(searchQueries.map((query, index) => [`query${index}`, query]))
    ).pipe(map(dataOrThrowErrors))
}

const bulkSearchCommitsFragment = gql`
    fragment BulkSearchCommits on Search {
        results {
            results {
                ... on CommitSearchResult {
                    __typename
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

/**
 * Fetch closest commits according to step between points on insight chart.
 * */
export function fetchSearchInsightCommits(commitQueries: string[]): Observable<Record<string, BulkSearchCommits>> {
    return requestGraphQL<Record<string, BulkSearchCommits>>(
        `
            query BulkSearchCommits(${commitQueries.map((query, index) => `$query${index}: String!`).join(', ')}) {
                ${commitQueries
                    .map(
                        (query, index) => `
                            search${index}: search(version: V2, patternType: literal, query: $query${index}) {
                               ...BulkSearchCommits
                            }
                        `
                    )
                    .join('\n')}
            }
            ${bulkSearchCommitsFragment}
        `,
        Object.fromEntries(commitQueries.map((query, index) => [`query${index}`, query]))
    ).pipe(map(dataOrThrowErrors))
}
