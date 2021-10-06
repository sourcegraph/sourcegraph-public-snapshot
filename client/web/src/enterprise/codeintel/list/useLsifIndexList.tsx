import { MutationFunctionOptions, FetchResult, ApolloClient, useMutation } from '@apollo/client'
import { parseISO } from 'date-fns'
import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { gql, getDocumentNode } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'

import {
    LsifIndexesForRepoResult,
    LsifIndexesForRepoVariables,
    LsifIndexesResult,
    LsifIndexesVariables,
    LsifIndexFields,
    CodeIntelligenceCommitGraphMetadataResult,
    CodeIntelligenceCommitGraphMetadataVariables,
    QueueAutoIndexJobsForRepoResult,
    QueueAutoIndexJobsForRepoVariables,
    Exact,
    Maybe,
} from '../../../graphql-operations'
import { lsifIndexFieldsFragment } from '../shared/backend'

interface IndexConnection {
    nodes: LsifIndexFields[]
    totalCount: number | null
    pageInfo: { endCursor: string | null; hasNextPage: boolean }
}

const LSIF_INDEX_FOR_REPOSITORY = gql`
    query LsifIndexesForRepo($repository: ID!, $state: LSIFIndexState, $first: Int, $after: String, $query: String) {
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

export const queryLsifIndexListByRepository = (
    { query, state, first, after }: GQL.ILsifIndexesOnRepositoryArguments,
    repository: string,
    client: ApolloClient<object>
): Observable<IndexConnection> => {
    const vars = {
        repository,
        query: query ?? null,
        state: state ?? null,
        first: first ?? null,
        after: after ?? null,
    }

    return from(
        client.query<LsifIndexesForRepoResult, LsifIndexesForRepoVariables>({
            query: getDocumentNode(LSIF_INDEX_FOR_REPOSITORY),
            variables: { ...vars },
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

            return node.lsifIndexes
        })
    )
}

const LSIF_INDEXES = gql`
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

export const queryLsifIndexList = (
    { query, state, first, after }: GQL.ILsifIndexesOnRepositoryArguments,
    client: ApolloClient<object>
): Observable<IndexConnection> => {
    const vars = {
        query: query ?? null,
        state: state ?? null,
        first: first ?? null,
        after: after ?? null,
    }

    return from(
        client.query<LsifIndexesResult, LsifIndexesVariables>({
            query: getDocumentNode(LSIF_INDEXES),
            variables: { ...vars },
        })
    ).pipe(
        map(({ data }) => data),
        map(({ lsifIndexes }) => lsifIndexes)
    )
}

const GRAPH_METADATA = gql`
    query CodeIntelligenceCommitGraphMetadata($repository: ID!) {
        node(id: $repository) {
            __typename
            ... on Repository {
                codeIntelligenceCommitGraph {
                    stale
                    updatedAt
                }
            }
        }
    }
`

export const queryCommitGraphMetadata = (
    repository: string,
    client: ApolloClient<object>
): Observable<{ stale: boolean; updatedAt: Date | null }> =>
    from(
        client.query<CodeIntelligenceCommitGraphMetadataResult, CodeIntelligenceCommitGraphMetadataVariables>({
            query: getDocumentNode(GRAPH_METADATA),
            variables: { repository },
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
            if (!node.codeIntelligenceCommitGraph) {
                throw new Error('Missing code intelligence commit graph value')
            }

            return {
                stale: node.codeIntelligenceCommitGraph.stale,
                updatedAt: node.codeIntelligenceCommitGraph.updatedAt
                    ? parseISO(node.codeIntelligenceCommitGraph.updatedAt)
                    : null,
            }
        })
    )

const QUEUE_AUTO_INDEX_JOBS = gql`
    mutation QueueAutoIndexJobsForRepo($id: ID!, $rev: String) {
        queueAutoIndexJobsForRepo(repository: $id, rev: $rev) {
            ...LsifIndexFields
        }
    }

    ${lsifIndexFieldsFragment}
`

type EnqueueIndexJobResults = Promise<
    FetchResult<QueueAutoIndexJobsForRepoResult, Record<string, any>, Record<string, any>>
>
interface UseEnqueueIndexJobResult {
    handleEnqueueIndexJob: (
        options?:
            | MutationFunctionOptions<QueueAutoIndexJobsForRepoResult, Exact<{ id: string; rev: Maybe<string> }>>
            | undefined
    ) => EnqueueIndexJobResults
}

export const useEnqueueIndexJob = (): UseEnqueueIndexJobResult => {
    const [handleEnqueueIndexJob] = useMutation<QueueAutoIndexJobsForRepoResult, QueueAutoIndexJobsForRepoVariables>(
        getDocumentNode(QUEUE_AUTO_INDEX_JOBS)
    )

    return {
        handleEnqueueIndexJob,
    }
}
