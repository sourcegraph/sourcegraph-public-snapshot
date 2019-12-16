import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { queryGraphQL } from '../../../backend/graphql'

export const querySearchResultsStats = (query: string): Observable<GQL.ISearchResultsStats> =>
    queryGraphQL(
        gql`
            query SearchResultsStats($query: String!) {
                search(query: $query) {
                    results {
                        elapsedMilliseconds
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
            return data.search.stats
        })
    )
