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
                                inputRepoName
                                inputCommit
                                inputRoot
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
 * Fetch LSIF upload with the given state.
 */
export function fetchLsifUploads({
    state,
    first,
    query,
}: GQL.ILsifUploadsOnQueryArguments): Observable<GQL.ILSIFUploadConnection> {
    return queryGraphQL(
        gql`
            query LsifUploads($state: LSIFUploadState!, $first: Int, $query: String) {
                lsifUploads(state: $state, first: $first, query: $query) {
                    nodes {
                        id
                        projectRoot {
                            commit {
                                abbreviatedOID
                            }
                            path
                            url
                        }
                        inputRepoName
                        inputCommit
                        inputRoot
                        state
                        uploadedAt
                        startedAt
                        finishedAt
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
        map(data => data.lsifUploads)
    )
}
