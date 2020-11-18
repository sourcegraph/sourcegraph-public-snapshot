import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { requestGraphQL } from '../../../backend/graphql'
import {
    LsifIndexesForRepoResult,
    LsifIndexesForRepoVariables,
    LsifIndexesResult,
    LsifIndexesVariables,
    LsifIndexFields,
    LsifUploadFields,
    LsifUploadsForRepoResult,
    LsifUploadsForRepoVariables,
    LsifUploadsResult,
    LsifUploadsVariables,
} from '../../../graphql-operations'
import { lsifIndexFieldsFragment, lsifUploadFieldsFragment } from '../shared/backend'

interface UploadConnection {
    nodes: LsifUploadFields[]
    totalCount: number | null
    pageInfo: { endCursor: string | null; hasNextPage: boolean }
}

/**
 * Return LSIF uploads. If a repository is given, only uploads for that repository will be returned. Otherwise,
 * uploads across all repositories are returned.
 */
export function fetchLsifUploads({
    repository,
    query,
    state,
    isLatestForRepo,
    first,
    after,
}: { repository?: string } & GQL.ILsifUploadsOnRepositoryArguments): Observable<UploadConnection> {
    const vars = {
        query: query ?? null,
        state: state ?? null,
        isLatestForRepo: isLatestForRepo ?? null,
        first: first ?? null,
        after: after ?? null,
    }

    if (repository) {
        const gqlQuery = gql`
            query LsifUploadsForRepo(
                $repository: ID!
                $state: LSIFUploadState
                $isLatestForRepo: Boolean
                $first: Int
                $after: String
                $query: String
            ) {
                node(id: $repository) {
                    __typename
                    ... on Repository {
                        lsifUploads(
                            query: $query
                            state: $state
                            isLatestForRepo: $isLatestForRepo
                            first: $first
                            after: $after
                        ) {
                            nodes {
                                ...LsifUploadFields
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

            ${lsifUploadFieldsFragment}
        `

        return requestGraphQL<LsifUploadsForRepoResult, LsifUploadsForRepoVariables>(gqlQuery, {
            ...vars,
            repository,
        }).pipe(
            map(dataOrThrowErrors),
            map(({ node }) => {
                if (!node) {
                    throw new Error('Invalid repository')
                }
                if (node.__typename !== 'Repository') {
                    throw new Error(`The given ID is ${node.__typename}, not Repository`)
                }

                return node.lsifUploads
            })
        )
    }

    const gqlQuery = gql`
        query LsifUploads(
            $state: LSIFUploadState
            $isLatestForRepo: Boolean
            $first: Int
            $after: String
            $query: String
        ) {
            lsifUploads(query: $query, state: $state, isLatestForRepo: $isLatestForRepo, first: $first, after: $after) {
                nodes {
                    ...LsifUploadFields
                }
                totalCount
                pageInfo {
                    endCursor
                    hasNextPage
                }
            }
        }

        ${lsifUploadFieldsFragment}
    `

    return requestGraphQL<LsifUploadsResult, LsifUploadsVariables>(gqlQuery, vars).pipe(
        map(dataOrThrowErrors),
        map(({ lsifUploads }) => lsifUploads)
    )
}

interface IndexConnection {
    nodes: LsifIndexFields[]
    totalCount: number | null
    pageInfo: { endCursor: string | null; hasNextPage: boolean }
}

/**
 * Return LSIF indexes. If a repository is given, only indexes for that repository will be returned. Otherwise,
 * indexes across all repositories are returned.
 */
export function fetchLsifIndexes({
    repository,
    query,
    state,
    first,
    after,
}: { repository?: string } & GQL.ILsifIndexesOnRepositoryArguments): Observable<IndexConnection> {
    const vars = {
        repository,
        query: query ?? null,
        state: state ?? null,
        first: first ?? null,
        after: after ?? null,
    }

    if (repository) {
        const gqlQuery = gql`
            query LsifIndexesForRepo(
                $repository: ID!
                $state: LSIFIndexState
                $first: Int
                $after: String
                $query: String
            ) {
                node(id: $repository) {
                    __typename
                    ... on Repository {
                        lsifIndexes(query: $query, state: $state, first: $first, after: $after) {
                            nodes {
                                ...LsifIndexFields
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

            ${lsifIndexFieldsFragment}
        `

        return requestGraphQL<LsifIndexesForRepoResult, LsifIndexesForRepoVariables>(gqlQuery, {
            ...vars,
            repository,
        }).pipe(
            map(dataOrThrowErrors),
            map(({ node }) => {
                if (!node) {
                    throw new Error('Invalid repository')
                }
                if (node.__typename !== 'Repository') {
                    throw new Error(`The given ID is ${node.__typename}, not Repository`)
                }

                return node.lsifIndexes
            })
        )
    }

    const gqlQuery = gql`
        query LsifIndexes($state: LSIFIndexState, $first: Int, $after: String, $query: String) {
            lsifIndexes(query: $query, state: $state, first: $first, after: $after) {
                nodes {
                    ...LsifIndexFields
                }
                totalCount
                pageInfo {
                    endCursor
                    hasNextPage
                }
            }
        }

        ${lsifIndexFieldsFragment}
    `

    return requestGraphQL<LsifIndexesResult, LsifIndexesVariables>(gqlQuery, vars).pipe(
        map(dataOrThrowErrors),
        map(({ lsifIndexes }) => lsifIndexes)
    )
}
