import * as GQL from '../../../../shared/src/graphql/schema'
import { map } from 'rxjs/operators'
import { Observable } from 'rxjs'
import { queryGraphQL } from '../../backend/graphql'
import { dataOrThrowErrors, gql } from '../../../../shared/src/graphql/graphql'

/**
 * Fetch counts of LSIF uploads by state.
 */
export function fetchLsifUploadStatistics(): Observable<GQL.ILSIFUploadStats> {
    return queryGraphQL(
        gql`
            query LsifUploadStats {
                lsifUploadStats {
                    erroredCount
                    completedCount
                    processingCount
                    queuedCount
                }
            }
        `
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.lsifUploadStats)
    )
}

/**
 * Fetch LSIF uploads with the given state.
 */
export function fetchLsifUploads({
    state,
    first,
    after,
    query,
}: GQL.ILsifUploadsOnQueryArguments): Observable<GQL.ILSIFUploadConnection> {
    return queryGraphQL(
        gql`
            query LsifUploads($state: LSIFUploadState!, $first: Int, $after: String, $query: String) {
                lsifUploads(state: $state, first: $first, after: $after, query: $query) {
                    nodes {
                        id
                        projectRoot {
                            commit {
                                abbreviatedOID
                                repository {
                                    name
                                }
                            }
                            path
                            url
                        }
                        state
                        failure {
                            summary
                        }
                        inputRepoName
                        inputCommit
                        inputRoot
                        uploadedAt
                        startedAt
                        finishedAt
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
        map(data => data.lsifUploads)
    )
}

/**
 * Fetch a single LSIF upload by id.
 */
export function fetchLsifUpload({ id }: { id: string }): Observable<GQL.ILSIFUpload | null> {
    return queryGraphQL(
        gql`
            query LsifUpload($id: ID!) {
                node(id: $id) {
                    __typename
                    ... on LSIFUpload {
                        id
                        projectRoot {
                            commit {
                                oid
                                abbreviatedOID
                                url
                                repository {
                                    name
                                    url
                                }
                            }
                            path
                            url
                        }
                        inputRepoName
                        inputCommit
                        inputRoot
                        state
                        failure {
                            summary
                        }
                        uploadedAt
                        startedAt
                        finishedAt
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
            if (node.__typename !== 'LSIFUpload') {
                throw new Error(`The given ID is a ${node.__typename}, not an LSIFUpload`)
            }

            return node
        })
    )
}
