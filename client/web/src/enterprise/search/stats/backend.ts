import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../../backend/graphql'
import { SearchResultsStatsResult, SearchResultsStatsFields } from '../../../graphql-operations'

export const querySearchResultsStats = (query: string): Observable<SearchResultsStatsFields & { limitHit: boolean }> =>
    requestGraphQL<SearchResultsStatsResult>(
        gql`
            query SearchResultsStats($query: String!) {
                search(query: $query) {
                    results {
                        limitHit
                    }
                    stats {
                        ...SearchResultsStatsFields
                    }
                }
            }

            fragment SearchResultsStatsFields on SearchResultsStats {
                languages {
                    name
                    totalLines
                }
            }
        `,
        { query }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.search) {
                throw new Error('no search results')
            }
            return { ...data.search.stats, limitHit: data.search.results.limitHit }
        })
    )
