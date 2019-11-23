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
}: { repository: string } & GQL.ILsifDumpsOnRepositoryArguments): Observable<GQL.ILSIFDumpConnection> {
    return queryGraphQL(
        gql`
            query LsifDumps($repository: ID!, $first: Int, $after: String, $query: String, $isLatestForRepo: Boolean) {
                node(id: $repository) {
                    __typename
                    ... on Repository {
                        lsifDumps(first: $first, after: $after, query: $query, isLatestForRepo: $isLatestForRepo) {
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
                }
            }
        `,
        { repository, first, after, query, isLatestForRepo }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error('Invalid repository')
            }
            if (node.__typename !== 'Repository') {
                throw new Error(`The given ID is a ${node.__typename}, not a Repository`)
            }

            return node.lsifDumps
        })
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
