import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, dataOrThrowErrors } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { queryGraphQL } from '../../../backend/graphql'

/**
 * Fetch LSIF dumps for a repository.
 */
export function fetchLsifDumps({
    repository,
    first,
    after,
    query,
    isLatestForRepo,
}: GQL.ILsifDumpsOnQueryArguments): Observable<GQL.ILSIFDumpConnection> {
    return queryGraphQL(
        gql`
            query LsifDumps($repository: ID!, $first: Int, $after: String, $query: String, $isLatestForRepo: Boolean) {
                lsifDumps(
                    repository: $repository
                    first: $first
                    after: $after
                    query: $query
                    isLatestForRepo: $isLatestForRepo
                ) {
                    nodes {
                        id
                        projectRoot {
                            commit {
                                abbreviatedOID
                            }
                            path
                            url
                        }
                        processedAt
                    }

                    totalCount
                    pageInfo {
                        endCursor
                        hasNextPage
                    }
                }
            }
        `,
        { repository, first, after, query, isLatestForRepo }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.lsifDumps)
    )
}

/**
 * Fetch LSIF jobs with the given state.
 */
export function fetchLsifJobs({
    state,
    first,
    query,
}: GQL.ILsifJobsOnQueryArguments): Observable<GQL.ILSIFJobConnection> {
    return queryGraphQL(
        gql`
            query LsifJobs($state: LSIFJobState!, $first: Int, $query: String) {
                lsifJobs(state: $state, first: $first, query: $query) {
                    nodes {
                        id
                        arguments
                        state
                        queuedAt
                        startedAt
                        completedOrErroredAt
                    }
                    pageInfo {
                        hasNextPage
                    }
                }
            }
        `,
        { state: state.toUpperCase(), first, query }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.lsifJobs)
    )
}
