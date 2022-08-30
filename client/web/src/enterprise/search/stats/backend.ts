import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { queryGraphQL } from '@sourcegraph/shared/src/backend/graphql'
import * as GQL from '@sourcegraph/shared/src/schema'

export const querySearchResultsStats = (query: string): Observable<GQL.ISearchResultsStats & { limitHit: boolean }> =>
    queryGraphQL(
        gql`
            query SearchResultsStats($query: String!) {
                search(query: $query) {
                    results {
                        limitHit
                    }
                    stats {
                        languages {
                            name
                            totalLines
                        }
                    }
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
