import * as GQL from '../../../../../shared/src/graphql/schema'
import {
    dataOrThrowErrors,
    gql,
    createInvalidGraphQLMutationResponseError,
} from '../../../../../shared/src/graphql/graphql'
import { map } from 'rxjs/operators'
import { Observable } from 'rxjs'
import { queryGraphQL, mutateGraphQL } from '../../../backend/graphql'
import { LsifIndexesForRepoResult, DeleteLsifUploadForRepoResult, LsifIndexesForRepoVariables, LsifIndexForRepoResult, LsifIndexForRepoVariables, LsifUploadsForRepoVariables, LsifUploadsForRepoResult } from '../../../graphql-operations'

/**
 * Fetch LSIF uploads for a repository.
 */
export function fetchLsifUploads({
    repository,
    query,
    state,
    isLatestForRepo,
    first,
    after,
}: LsifUploadsForRepoVariables): Observable<(LsifUploadsForRepoResult['node'] & { __typename: 'Repository' })['lsifUploads']> {
    return queryGraphQL<LsifUploadsForRepoResult>(
        gql`
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
                                id
                                state
                                projectRoot {
                                    commit {
                                        abbreviatedOID
                                        url
                                    }
                                    path
                                    url
                                }
                                inputCommit
                                inputRoot
                                inputIndexer
                                uploadedAt
                                startedAt
                                finishedAt
                                placeInQueue
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
        { repository, query, state, isLatestForRepo, first, after }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error('Invalid repository')
            }
            if (node.__typename !== 'Repository') {
                throw new Error(`The given ID is a ${node.__typename}, not a Repository`)
            }

            return node.lsifUploads
        })
    )
}

/**
 * Fetch a single LSIF upload by id.
 */
export function fetchLsifUpload({ id }: { id: string }): Observable<GQL.LSIFUpload | null> {
    return queryGraphQL<LsifUploadForRepoResult>(
        gql`
            query LsifUploadForRepo($id: ID!) {
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
                        inputCommit
                        inputRoot
                        inputIndexer
                        state
                        failure
                        uploadedAt
                        startedAt
                        finishedAt
                        isLatestForRepo
                        placeInQueue
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

/**
 * Delete an LSIF upload by id.
 */
export function deleteLsifUpload({ id }: { id: string }): Observable<void> {
    return mutateGraphQL<DeleteLsifUploadForRepoResult>(
        gql`
            mutation DeleteLsifUploadForRepo($id: ID!) {
                deleteLSIFUpload(id: $id) {
                    alwaysNil
                }
            }
        `,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.deleteLSIFUpload) {
                throw createInvalidGraphQLMutationResponseError('DeleteLsifUpload')
            }
        })
    )
}

/**
 * Fetch LSIF indexes for a repository.
 */
export function fetchLsifIndexes({
    repository,
    query,
    state,
    first,
    after,
}: LsifIndexesForRepoVariables): Observable<GQL.LSIFIndexConnection> {
    return queryGraphQL<LsifIndexesForRepoResult>(
        gql`
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
                                id
                                state
                                projectRoot {
                                    commit {
                                        abbreviatedOID
                                        url
                                    }
                                    path
                                    url
                                }
                                inputCommit
                                queuedAt
                                startedAt
                                finishedAt
                                placeInQueue
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
        { repository, query, state, first, after }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error('Invalid repository')
            }
            if (node.__typename !== 'Repository') {
                throw new Error(`The given ID is a ${node.__typename}, not a Repository`)
            }

            return node.lsifIndexes
        })
    )
}

/**
 * Fetch a single LSIF index by id.
 */
export function fetchLsifIndex({ id }: LsifIndexForRepoVariables): Observable<GQL.LSIFIndex | null> {
    return queryGraphQL<LsifIndexForRepoResult>(
        gql`
            query LsifIndexForRepo($id: ID!) {
                node(id: $id) {
                    __typename
                    ... on LSIFIndex {
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
                        inputCommit
                        state
                        failure
                        queuedAt
                        startedAt
                        finishedAt
                        placeInQueue
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
            if (node.__typename !== 'LSIFIndex') {
                throw new Error(`The given ID is a ${node.__typename}, not an LSIFIndex`)
            }

            return node
        })
    )
}

/**
 * Delete an LSIF index by id.
 */
export function deleteLsifIndex({ id }: { id: string }): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation DeleteLsifIndexForRepo($id: ID!) {
                deleteLSIFIndex(id: $id) {
                    alwaysNil
                }
            }
        `,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.deleteLSIFIndex) {
                throw createInvalidGraphQLMutationResponseError('DeleteLsifIndex')
            }
        })
    )
}
