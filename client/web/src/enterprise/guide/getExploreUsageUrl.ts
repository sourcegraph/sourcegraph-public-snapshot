import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../backend/graphql'
import { ExploreUsageURLVariables, ExploreUsageURLResult } from '../../graphql-operations'

export const getExploreUsageURL = (vars: ExploreUsageURLVariables): Observable<string | null> =>
    requestGraphQL<ExploreUsageURLResult, ExploreUsageURLVariables>(
        gql`
            query ExploreUsageURL($repo: String!, $commitID: String!, $path: String!, $line: Int!, $character: Int!) {
                repository(name: $repo) {
                    commit(rev: $commitID) {
                        blob(path: $path) {
                            lsif {
                                exploreUsageURL(line: $line, character: $character)
                            }
                        }
                    }
                }
            }
        `,
        vars
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.repository?.commit?.blob?.lsif?.exploreUsageURL || null)
    )
