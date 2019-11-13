import * as GQL from '../../../../shared/src/graphql/schema'
import { map } from 'rxjs/operators'
import { Observable } from 'rxjs'
import { queryGraphQL } from '../../backend/graphql'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'

/**
 * Fetch counts of LSIF jobs by state.
 */
export function fetchLsifJobStatistics(): Observable<GQL.ILSIFJobStats> {
    return queryGraphQL(
        gql`
            query LsifJobStats {
                lsifJobStats {
                    erroredCount
                    completedCount
                    processingCount
                    queuedCount
                    scheduledCount
                }
            }
        `
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.lsifJobStats)
    )
}

/**
 * Fetch LSIF jobs with the given state.
 */
export function fetchLsifJobs({
    state,
    first,
    after,
    query,
}: GQL.ILsifJobsOnQueryArguments): Observable<GQL.ILSIFJobConnection> {
    return queryGraphQL(
        gql`
            query LsifJobs($state: LSIFJobState!, $first: Int, $after: String, $query: String) {
                lsifJobs(state: $state, first: $first, after: $after, query: $query) {
                    nodes {
                        id
                        type
                        arguments
                        state
                        failure {
                            summary
                        }
                        queuedAt
                        startedAt
                        completedOrErroredAt
                    }

                    totalCount
                    pageInfo {
                        endCursor
                        hasNextPage
                    }
                }
            }
        `,
        { state: state.toUpperCase(), first, after, query }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.lsifJobs)
    )
}

/**
 * Fetch a single LSIF job by id.
 */
export function fetchLsifJob({ id }: GQL.ILsifJobOnQueryArguments): Observable<GQL.ILSIFJob | null> {
    return queryGraphQL(
        gql`
            query LsifJob($id: ID!) {
                node(id: $id) {
                    __typename
                    ... on LSIFJob {
                        id
                        type
                        arguments
                        state
                        failure {
                            summary
                        }
                        queuedAt
                        startedAt
                        completedOrErroredAt
                    }
                }
            }
        `,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                return null
            }
            if (node.__typename !== 'LSIFJob') {
                throw new Error(`The given ID is a ${node.__typename}, not an LSIFJob`)
            }

            return node
        })
    )
}
