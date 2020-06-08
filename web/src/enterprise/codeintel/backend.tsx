import * as GQL from '../../../../shared/src/graphql/schema'
import {
    dataOrThrowErrors,
    gql,
    createInvalidGraphQLMutationResponseError,
} from '../../../../shared/src/graphql/graphql'
import { map } from 'rxjs/operators'
import { Observable } from 'rxjs'
import { queryGraphQL, mutateGraphQL } from '../../backend/graphql'

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
}: { repository?: string } & GQL.ILsifUploadsOnRepositoryArguments): Observable<GQL.ILSIFUploadConnection> {
    if (!repository) {
        return queryGraphQL(
            gql`
                query LsifUploads(
                    $state: LSIFUploadState
                    $isLatestForRepo: Boolean
                    $first: Int
                    $after: String
                    $query: String
                ) {
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
            `,
            { query, state, isLatestForRepo, first, after }
        ).pipe(
            map(dataOrThrowErrors),
            map(({ lsifUploads }) => lsifUploads)
        )
    }

    return queryGraphQL(
        gql`
            query LsifUploads(
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
                        inputCommit
                        inputRoot
                        inputIndexer
                        state
                        failure {
                            summary
                        }
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

export function deleteLsifUpload({ id }: { id: string }): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation DeleteLsifUpload($id: ID!) {
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
 * Return LSIF indexes. If a repository is given, only indexes for that repository will be returned. Otherwise,
 * indexes across all repositories are returned.
 */
export function fetchLsifIndexes({
    repository,
    query,
    state,
    first,
    after,
}: { repository?: string } & GQL.ILsifIndexesOnRepositoryArguments): Observable<GQL.ILSIFIndexConnection> {
    if (!repository) {
        return queryGraphQL(
            gql`
                query LsifIndexes($state: LSIFIndexState, $first: Int, $after: String, $query: String) {
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
            `,
            { query, state, first, after }
        ).pipe(
            map(dataOrThrowErrors),
            map(({ lsifIndexes }) => lsifIndexes)
        )
    }

    return queryGraphQL(
        gql`
            query LsifIndexes($repository: ID!, $state: LSIFIndexState, $first: Int, $after: String, $query: String) {
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

export function fetchLsifIndex({ id }: { id: string }): Observable<GQL.ILSIFIndex | null> {
    return queryGraphQL(
        gql`
            query LsifIndex($id: ID!) {
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
                        failure {
                            summary
                        }
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

export function deleteLsifIndex({ id }: { id: string }): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation DeleteLsifIndex($id: ID!) {
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
