import { ApolloError, MutationFunctionOptions, FetchResult, ApolloClient, useMutation } from '@apollo/client'
import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { gql, getDocumentNode } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'

import {
    LsifUploadFields,
    LsifUploadResult,
    LsifUploadVariables,
    LSIFUploadState,
    LsifUploadsForRepoResult,
    LsifUploadsForRepoVariables,
    LsifUploadConnectionFields,
    LsifUploadsVariables,
    LsifUploadsResult,
    DeleteLsifUploadResult,
    DeleteLsifUploadVariables,
    Exact,
} from '../../../graphql-operations'

const lsifUploadFieldsFragment = gql`
    fragment LsifUploadFields on LSIFUpload {
        __typename
        id
        inputCommit
        inputRoot
        inputIndexer
        projectRoot {
            url
            path
            repository {
                url
                name
            }
            commit {
                url
                oid
                abbreviatedOID
            }
        }
        state
        failure
        isLatestForRepo
        uploadedAt
        startedAt
        finishedAt
        placeInQueue
        associatedIndex {
            id
            state
            queuedAt
            startedAt
            finishedAt
            placeInQueue
        }
    }
`

const LSIF_UPLOAD_FIELDS = gql`
    query LsifUpload($id: ID!) {
        node(id: $id) {
            ...LsifUploadFields
        }
    }

    ${lsifUploadFieldsFragment}
`

export const queryLisfUploadFields = (id: string, client: ApolloClient<object>): Observable<LsifUploadFields | null> =>
    from(
        client.query<LsifUploadResult, LsifUploadVariables>({
            query: getDocumentNode(LSIF_UPLOAD_FIELDS),
            variables: { id },
        })
    ).pipe(
        map(({ data }) => data),
        map(({ node }) => {
            if (!node || node.__typename !== 'LSIFUpload') {
                throw new Error('No such LSIFUpload')
            }
            return node
        })
    )

// List
const LSIF_UPLOAD_LIST_BY_REPO_ID = gql`
    query LsifUploadsForRepo(
        $repository: ID!
        $state: LSIFUploadState
        $isLatestForRepo: Boolean
        $dependencyOf: ID
        $dependentOf: ID
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
                    dependencyOf: $dependencyOf
                    dependentOf: $dependentOf
                    first: $first
                    after: $after
                ) {
                    ...LsifUploadConnectionFields
                }
            }
        }
    }

    fragment LsifUploadConnectionFields on LSIFUploadConnection {
        nodes {
            ...LsifUploadFields
        }
        totalCount
        pageInfo {
            endCursor
            hasNextPage
        }
    }

    ${lsifUploadFieldsFragment}
`

const LSIF_UPLOAD_LIST = gql`
    query LsifUploads(
        $state: LSIFUploadState
        $isLatestForRepo: Boolean
        $dependencyOf: ID
        $dependentOf: ID
        $first: Int
        $after: String
        $query: String
    ) {
        lsifUploads(
            query: $query
            state: $state
            isLatestForRepo: $isLatestForRepo
            dependencyOf: $dependencyOf
            dependentOf: $dependentOf
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

    ${lsifUploadFieldsFragment}
`

export interface UploadListVariables {
    state?: LSIFUploadState
    isLatestForRepo?: boolean
    dependencyOf?: string | null
    dependentOf?: string | null
    first?: number | null
    after?: string | null
    query?: string | null
}

export const queryLsifUploadsList = (
    { query, state, isLatestForRepo, dependencyOf, dependentOf, first, after }: GQL.ILsifUploadsOnRepositoryArguments,
    client: ApolloClient<object>
): Observable<LsifUploadConnectionFields> => {
    const vars: LsifUploadsVariables = {
        query: query ?? null,
        state: state ?? null,
        isLatestForRepo: isLatestForRepo ?? null,
        dependencyOf: dependencyOf ?? null,
        dependentOf: dependentOf ?? null,
        first: first ?? null,
        after: after ?? null,
    }

    return from(
        client.query<LsifUploadsResult, LsifUploadsVariables>({
            query: getDocumentNode(LSIF_UPLOAD_LIST),
            variables: { ...vars },
        })
    ).pipe(
        map(({ data }) => data),
        map(({ lsifUploads }) => lsifUploads)
    )
}

export const queryLsifUploadsByRepository = (
    { query, state, isLatestForRepo, dependencyOf, dependentOf, first, after }: GQL.ILsifUploadsOnRepositoryArguments,
    repository: string,
    client: ApolloClient<object>
): Observable<LsifUploadConnectionFields> => {
    const vars: LsifUploadsVariables = {
        query: query ?? null,
        state: state ?? null,
        isLatestForRepo: isLatestForRepo ?? null,
        dependencyOf: dependencyOf ?? null,
        dependentOf: dependentOf ?? null,
        first: first ?? null,
        after: after ?? null,
    }

    return from(
        client.query<LsifUploadsForRepoResult, LsifUploadsForRepoVariables>({
            query: getDocumentNode(LSIF_UPLOAD_LIST_BY_REPO_ID),
            variables: { ...vars, repository },
        })
    ).pipe(
        map(({ data }) => data),
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

// Mutation
type DeleteLsifUploadResults = Promise<FetchResult<DeleteLsifUploadResult, Record<string, any>, Record<string, any>>>

interface UseDeleteLsifUploadResult {
    handleDeleteLsifUpload: (
        options?:
            | MutationFunctionOptions<
                  DeleteLsifUploadResult,
                  Exact<{
                      id: string
                  }>
              >
            | undefined
    ) => DeleteLsifUploadResults
    deleteError: ApolloError | undefined
}

const DELETE_LSIF_UPLOAD = gql`
    mutation DeleteLsifUpload($id: ID!) {
        deleteLSIFUpload(id: $id) {
            alwaysNil
        }
    }
`

export const useDeleteLsifUpload = (): UseDeleteLsifUploadResult => {
    const [handleDeleteLsifUpload, { error }] = useMutation<DeleteLsifUploadResult, DeleteLsifUploadVariables>(
        getDocumentNode(DELETE_LSIF_UPLOAD)
    )

    return {
        handleDeleteLsifUpload,
        deleteError: error,
    }
}
